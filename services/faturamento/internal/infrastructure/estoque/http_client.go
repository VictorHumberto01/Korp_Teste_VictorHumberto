package estoque

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/sony/gobreaker/v2"

	domainerrors "faturamento-service/internal/domain/errors"
)

// HTTPClient implementa port.EstoqueClient chamando o estoque-service via
// HTTP, protegido por um circuit breaker: falhas de infraestrutura (timeout,
// conexão recusada, 5xx) abrem o circuito e passam a falhar rápido sem
// sobrecarregar um serviço já instável. Respostas de negócio (produto não
// encontrado, saldo insuficiente) não contam como falha do circuito — são
// respostas válidas de um serviço saudável.
type HTTPClient struct {
	baseURL    string
	httpClient *http.Client
	breaker    *gobreaker.CircuitBreaker[any]
}

func NewHTTPClient(baseURL string) *HTTPClient {
	breaker := gobreaker.NewCircuitBreaker[any](gobreaker.Settings{
		Name:        "estoque-service",
		MaxRequests: 1,
		Interval:    30 * time.Second,
		Timeout:     10 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 3
		},
		IsSuccessful: func(err error) bool {
			if err == nil {
				return true
			}
			return errors.Is(err, domainerrors.ErrProdutoNaoEncontrado) ||
				errors.Is(err, domainerrors.ErrSaldoInsuficiente)
		},
	})

	return &HTTPClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
		breaker:    breaker,
	}
}

type saldoBody struct {
	Quantidade int `json:"quantidade"`
}

type errorEnvelope struct {
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (c *HTTPClient) DebitarSaldo(ctx context.Context, produtoID string, quantidade int) error {
	return c.chamarSaldo(ctx, produtoID, quantidade, "debitar")
}

func (c *HTTPClient) CreditarSaldo(ctx context.Context, produtoID string, quantidade int) error {
	return c.chamarSaldo(ctx, produtoID, quantidade, "creditar")
}

func (c *HTTPClient) chamarSaldo(ctx context.Context, produtoID string, quantidade int, operacao string) error {
	_, err := c.breaker.Execute(func() (any, error) {
		return nil, c.doRequest(ctx, produtoID, quantidade, operacao)
	})
	if err == nil {
		return nil
	}
	if errors.Is(err, gobreaker.ErrOpenState) || errors.Is(err, gobreaker.ErrTooManyRequests) {
		return domainerrors.ErrEstoqueIndisponivel
	}
	return err
}

func (c *HTTPClient) doRequest(ctx context.Context, produtoID string, quantidade int, operacao string) error {
	body, err := json.Marshal(saldoBody{Quantidade: quantidade})
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/produtos/%s/saldo/%s", c.baseURL, produtoID, operacao)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return domainerrors.ErrEstoqueIndisponivel
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	var envelope errorEnvelope
	_ = json.NewDecoder(resp.Body).Decode(&envelope)
	message := envelope.Error.Message

	switch {
	case resp.StatusCode == http.StatusNotFound:
		return domainerrors.ErrProdutoNaoEncontrado
	case resp.StatusCode == http.StatusConflict && strings.Contains(message, "insuficiente"):
		return domainerrors.ErrSaldoInsuficiente
	case resp.StatusCode >= 500:
		return domainerrors.ErrEstoqueIndisponivel
	default:
		if message == "" {
			message = fmt.Sprintf("erro inesperado do estoque (status %d)", resp.StatusCode)
		}
		return fmt.Errorf("estoque: %s", message)
	}
}

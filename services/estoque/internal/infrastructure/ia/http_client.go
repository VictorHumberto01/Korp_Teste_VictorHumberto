package ia

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/sony/gobreaker/v2"

	domainerrors "estoque-service/internal/domain/errors"
)

// HTTPClient implementa query.AIClient chamando o ia-service via HTTP,
// protegido por um circuit breaker: falhas de infraestrutura (timeout,
// conexão recusada, 5xx) abrem o circuito e passam a falhar rápido sem
// sobrecarregar um serviço já instável.
type HTTPClient struct {
	baseURL    string
	httpClient *http.Client
	breaker    *gobreaker.CircuitBreaker[string]
}

func NewHTTPClient(baseURL string) *HTTPClient {
	breaker := gobreaker.NewCircuitBreaker[string](gobreaker.Settings{
		Name:        "ia-service",
		MaxRequests: 1,
		Interval:    30 * time.Second,
		Timeout:     10 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 3
		},
	})

	return &HTTPClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 15 * time.Second},
		breaker:    breaker,
	}
}

type descricaoRequestBody struct {
	Nome string `json:"nome"`
}

type descricaoResponseBody struct {
	Descricao string `json:"descricao"`
}

type errorEnvelope struct {
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (c *HTTPClient) GenerateDescription(ctx context.Context, nome string) (string, error) {
	descricao, err := c.breaker.Execute(func() (string, error) {
		return c.doRequest(ctx, nome)
	})
	if err == nil {
		return descricao, nil
	}
	if errors.Is(err, gobreaker.ErrOpenState) || errors.Is(err, gobreaker.ErrTooManyRequests) {
		log.Printf("[ia-client] circuito aberto, requisição rejeitada rapidamente (nome=%q): %v", nome, err)
		return "", domainerrors.ErrIAIndisponivel
	}
	return "", err
}

func (c *HTTPClient) doRequest(ctx context.Context, nome string) (string, error) {
	body, err := json.Marshal(descricaoRequestBody{Nome: nome})
	if err != nil {
		log.Printf("[ia-client] falha ao serializar requisição (nome=%q): %v", nome, err)
		return "", err
	}

	url := c.baseURL + "/ia/produtos/descricao"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		log.Printf("[ia-client] falha ao montar requisição: %v", err)
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	start := time.Now()
	resp, err := c.httpClient.Do(req)
	elapsed := time.Since(start)
	if err != nil {
		log.Printf("[ia-client] falha de rede após %s ao chamar %s: %v", elapsed, url, err)
		return "", domainerrors.ErrIAIndisponivel
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		var respBody descricaoResponseBody
		if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
			log.Printf("[ia-client] falha ao decodificar resposta (status=%d, elapsed=%s): %v", resp.StatusCode, elapsed, err)
			return "", fmt.Errorf("falha ao decodificar resposta do ia-service: %w", err)
		}
		return respBody.Descricao, nil
	}

	var envelope errorEnvelope
	_ = json.NewDecoder(resp.Body).Decode(&envelope)
	message := envelope.Error.Message
	if message == "" {
		message = fmt.Sprintf("erro inesperado do ia-service (status %d)", resp.StatusCode)
	}

	log.Printf("[ia-client] resposta de erro do ia-service (status=%d, elapsed=%s, nome=%q): %s", resp.StatusCode, elapsed, nome, message)

	if resp.StatusCode >= 500 {
		return "", domainerrors.ErrIAIndisponivel
	}
	return "", fmt.Errorf("ia: %s", message)
}

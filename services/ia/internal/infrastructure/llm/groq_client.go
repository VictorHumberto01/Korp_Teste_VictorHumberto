package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type GroqClient struct {
	apiKey string
	model  string
	client *http.Client
}

func NewGroqClient(apiKey, model string) *GroqClient {
	if model == "" {
		model = "openai/gpt-oss-20b"
	}
	return &GroqClient{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

type groqRequest struct {
	Model    string    `json:"model"`
	Messages []message `json:"messages"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type groqResponse struct {
	Choices []struct {
		Message message `json:"message"`
	} `json:"choices"`
}

type groqErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

func (c *GroqClient) GenerateDescription(ctx context.Context, nome string) (string, error) {
	if c.apiKey == "" {
		log.Printf("[groq] chamada abortada: GROQ_API_KEY não configurada (nome=%q)", nome)
		return "", errors.New("GROQ_API_KEY não configurada")
	}

	prompt := fmt.Sprintf("Você é um especialista em marketing e e-commerce. O usuário informou o nome básico de um produto: '%s'. Escreva uma descrição comercial curta, atraente e profissional para este produto, em um único parágrafo com NO MÁXIMO 180 caracteres (contando espaços e pontuação). Não use saudações, markdown ou aspas, apenas a descrição direta. É essencial respeitar o limite de 180 caracteres.", nome)

	reqBody := groqRequest{
		Model: c.model,
		Messages: []message{
			{Role: "user", Content: prompt},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		log.Printf("[groq] falha ao serializar requisição: %v", err)
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("[groq] falha ao montar requisição: %v", err)
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	start := time.Now()
	resp, err := c.client.Do(req)
	elapsed := time.Since(start)
	if err != nil {
		log.Printf("[groq] falha de rede após %s (model=%s): %v", elapsed, c.model, err)
		return "", fmt.Errorf("falha ao conectar com a API do Groq: %w", err)
	}
	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		log.Printf("[groq] falha ao ler corpo da resposta (status=%d, elapsed=%s): %v", resp.StatusCode, elapsed, readErr)
		return "", fmt.Errorf("falha ao ler resposta da API do Groq: %w", readErr)
	}

	if resp.StatusCode != http.StatusOK {
		var groqErr groqErrorResponse
		detail := string(body)
		if err := json.Unmarshal(body, &groqErr); err == nil && groqErr.Error.Message != "" {
			detail = fmt.Sprintf("%s (type=%s, code=%s)", groqErr.Error.Message, groqErr.Error.Type, groqErr.Error.Code)
		}
		log.Printf("[groq] erro na API (status=%d, model=%s, elapsed=%s, nome=%q): %s", resp.StatusCode, c.model, elapsed, nome, detail)
		return "", fmt.Errorf("erro na API do Groq (status %d): %s", resp.StatusCode, detail)
	}

	var groqResp groqResponse
	if err := json.Unmarshal(body, &groqResp); err != nil {
		log.Printf("[groq] falha ao decodificar resposta (status=%d, body=%s): %v", resp.StatusCode, string(body), err)
		return "", fmt.Errorf("falha ao decodificar resposta da API do Groq: %w", err)
	}

	if len(groqResp.Choices) == 0 {
		log.Printf("[groq] resposta sem choices (status=%d, model=%s, body=%s)", resp.StatusCode, c.model, string(body))
		return "", errors.New("resposta vazia da IA")
	}

	log.Printf("[groq] descrição gerada com sucesso (model=%s, elapsed=%s, nome=%q)", c.model, elapsed, nome)
	return groqResp.Choices[0].Message.Content, nil
}

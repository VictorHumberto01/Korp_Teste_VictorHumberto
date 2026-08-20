package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"ia-service/internal/application/query"
	httplayer "ia-service/internal/infrastructure/http"
	"ia-service/internal/infrastructure/http/handler"
	"ia-service/internal/infrastructure/llm"
)

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// loadEnvFile tenta cada candidato em ordem e carrega o primeiro que existir.
// godotenv.Load(a, b, c) NÃO faz fallback entre arquivos: ele para no primeiro
// que falhar ao abrir, então passar vários caminhos de uma vez nunca chegava
// a tentar o .env da raiz do projeto quando o serviço era iniciado via `cd
// services/<nome> && go run ./cmd/main.go`.
func loadEnvFile() {
	candidates := []string{".env", "../../.env", "../../../.env", "../../../../.env"}
	for _, path := range candidates {
		if err := godotenv.Load(path); err == nil {
			log.Printf("[ia] variáveis de ambiente carregadas de %s", path)
			return
		}
	}
	log.Printf("[ia] nenhum arquivo .env encontrado em %v — usando apenas variáveis de ambiente já exportadas", candidates)
}

func main() {
	loadEnvFile()

	serverPort := getEnv("IA_SERVER_PORT", "8083")
	groqApiKey := getEnv("GROQ_API_KEY", "")
	groqModel := getEnv("GROQ_MODEL", "openai/gpt-oss-20b")

	if groqApiKey == "" {
		log.Println("[ia] AVISO: GROQ_API_KEY não configurada — chamadas de geração de conteúdo vão falhar")
	}

	groqClient := llm.NewGroqClient(groqApiKey, groqModel)
	gerarDescricaoProdutoQuery := query.NewGerarDescricaoProdutoHandler(groqClient)
	iaHandler := handler.NewIAHandler(gerarDescricaoProdutoQuery)

	router := httplayer.NewRouter(iaHandler)

	srv := &http.Server{
		Addr:    ":" + serverPort,
		Handler: router,
	}

	go func() {
		log.Printf("Starting server on port %s", serverPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen and serve error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}

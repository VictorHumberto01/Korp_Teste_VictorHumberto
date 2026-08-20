package main

import (
	"github.com/joho/godotenv"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"usuario-service/internal/application/command"
	"usuario-service/internal/application/query"
	"usuario-service/internal/infrastructure/database"
	httplayer "usuario-service/internal/infrastructure/http"
	"usuario-service/internal/infrastructure/http/handler"
	"usuario-service/internal/infrastructure/http/middleware"
	"usuario-service/internal/infrastructure/persistence"
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
			log.Printf("[usuario] variáveis de ambiente carregadas de %s", path)
			return
		}
	}
	log.Printf("[usuario] nenhum arquivo .env encontrado em %v — usando apenas variáveis de ambiente já exportadas", candidates)
}

func main() {
	loadEnvFile()
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "korp")
	dbPassword := getEnv("DB_PASSWORD", "korp123")
	dbName := getEnv("DB_NAME", "usuario_db")
	serverPort := getEnv("USUARIO_SERVER_PORT", "8080")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		dbHost, dbUser, dbPassword, dbName, dbPort)

	db, err := database.NewPostgresConnection(dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	err = db.AutoMigrate(&persistence.UsuarioModel{}, &middleware.IdempotencyRecord{})
	if err != nil {
		log.Fatalf("Failed to auto-migrate: %v", err)
	}

	repo := persistence.NewGormUsuarioRepository(db)

	createCmd := command.NewCreateUsuarioHandler(repo)
	updateCmd := command.NewUpdateUsuarioHandler(repo)
	deleteCmd := command.NewDeleteUsuarioHandler(repo)

	getQuery := query.NewGetUsuarioHandler(repo)
	listQuery := query.NewListUsuariosHandler(repo)

	userHandler := handler.NewUsuarioHandler(
		createCmd,
		updateCmd,
		deleteCmd,
		getQuery,
		listQuery,
	)

	router := httplayer.NewRouter(userHandler, db)

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

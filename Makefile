.PHONY: docker-up docker-down run-usuario run-estoque run-frontend test test-concurrency ollama-pull

# ========================
# Docker
# ========================
docker-up:
	docker compose up -d

docker-down:
	docker compose down

# ========================
# Backend — Serviço de Usuários
# ========================
run-usuario:
	cd services/usuario && go run ./cmd/main.go

test:
	cd services/usuario && go test ./... -v -race -count=1

test-concurrency:
	cd services/usuario && go test ./tests/ -run TestConcurrent -v -race -count=1

# ========================
# Backend — Serviço de Estoque
# ========================
run-estoque:
	cd services/estoque && go run ./cmd/main.go

# ========================
# Frontend
# ========================
run-frontend:
	cd frontend && ng serve

install-frontend:
	cd frontend && npm install

# ========================
# Ollama
# ========================
ollama-pull:
	docker exec korp-ollama ollama pull qwen2.5:0.5b

# ========================
# Run All
# ========================
run-all: docker-up
	@echo "Aguardando PostgreSQL..."
	@sleep 3
	@echo "Iniciando serviços..."
	$(MAKE) run-usuario &
	$(MAKE) run-estoque &
	$(MAKE) run-frontend &
	@echo "Todos os serviços iniciados!"

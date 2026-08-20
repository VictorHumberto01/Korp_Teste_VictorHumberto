.PHONY: docker-up docker-down run-usuario run-estoque run-faturamento run-ia run-frontend test test-concurrency \
	stop-usuario stop-estoque stop-faturamento stop-ia stop-frontend stop-all run-all install-frontend

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

stop-usuario:
	-lsof -nP -tiTCP:8080 -sTCP:LISTEN | xargs kill -9 2>/dev/null || true

test:
	cd services/usuario && go test ./... -v -race -count=1

test-concurrency:
	cd services/usuario && go test ./tests/ -run TestConcurrent -v -race -count=1

# ========================
# Backend — Serviço de Estoque
# ========================
run-estoque:
	cd services/estoque && go run ./cmd/main.go

stop-estoque:
	-lsof -nP -tiTCP:8081 -sTCP:LISTEN | xargs kill -9 2>/dev/null || true

# ========================
# Backend — Serviço de Faturamento
# ========================
run-faturamento:
	cd services/faturamento && go run ./cmd/main.go

stop-faturamento:
	-lsof -nP -tiTCP:8082 -sTCP:LISTEN | xargs kill -9 2>/dev/null || true

# ========================
# Backend — Microsserviço de IA
# ========================
run-ia:
	cd services/ia && go run ./cmd/main.go

stop-ia:
	-lsof -nP -tiTCP:8083 -sTCP:LISTEN | xargs kill -9 2>/dev/null || true

# ========================
# Frontend
# ========================
run-frontend:
	cd frontend && NG_CLI_ANALYTICS=false npx ng serve

stop-frontend:
	-lsof -nP -tiTCP:4200 -sTCP:LISTEN | xargs kill -9 2>/dev/null || true

install-frontend:
	cd frontend && npm install

# ========================
# Run All
# ========================
run-all: docker-up
	@echo "Aguardando PostgreSQL..."
	@sleep 3
	@echo "Iniciando serviços..."
	$(MAKE) run-usuario &
	$(MAKE) run-estoque &
	$(MAKE) run-faturamento &
	$(MAKE) run-ia &
	$(MAKE) run-frontend &
	@echo "Todos os serviços iniciados!"

stop-all: stop-usuario stop-estoque stop-faturamento stop-ia stop-frontend
	@echo "Microsserviços e frontend parados."
	@echo "Parando banco de dados (Docker)..."
	docker compose down
	@echo "Tudo parado com sucesso!"

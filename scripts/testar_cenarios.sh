#!/usr/bin/env bash
#
# Testa, contra o sistema rodando de verdade (não mocks), os cenários pedidos
# no enunciado do teste:
#
#   2. Tratamento de falhas: derruba o serviço de estoque, mostra que o
#      faturamento devolve um erro claro em vez de travar, e prova que o
#      sistema se recupera sozinho quando o estoque volta.
#   3. Conexão real com banco de dados: prova que os cadastros feitos pela
#      API aparecem de verdade nas tabelas do Postgres.
#   a. Concorrência: produto com saldo 1, duas notas fiscais tentando debitar
#      ao mesmo tempo — só uma pode vencer.
#   b. IA: gera descrição de produto e valida o resultado.
#   c. Idempotência: repete a mesma escrita com a mesma chave e prova que não
#      duplica o efeito.
#
# Uso: ./scripts/testar_cenarios.sh
# Pré-requisito: make docker-up && make run-all (usuario, estoque,
# faturamento e ia respondendo em :8080/:8081/:8082/:8083).
#
# O script controla o processo do serviço de estoque diretamente (mata e
# reinicia via `go run`) para simular a queda dele no cenário 2. Se você
# tiver o estoque rodando num terminal em foreground, esse terminal vai
# perder o processo — o script sobe uma nova instância em background no
# lugar e a deixa rodando ao final.

set -uo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."

HOST_USUARIO="http://localhost:8080"
HOST_ESTOQUE="http://localhost:8081"
HOST_FATURAMENTO="http://localhost:8082"
HOST_IA="http://localhost:8083"

USUARIO_URL="$HOST_USUARIO/api"
ESTOQUE_URL="$HOST_ESTOQUE/api"
FATURAMENTO_URL="$HOST_FATURAMENTO/api"
IA_URL="$HOST_IA/api"
PG_CONTAINER="korp-postgres"
PG_USER="korp"
PG_PASSWORD="korp123"

PASS=0
FAIL=0
WARN=0

# ---------------------------------------------------------------------------
# Utilidades
# ---------------------------------------------------------------------------

bold() { printf '\033[1m%s\033[0m\n' "$1"; }
header() { echo; bold "══ $1 ══"; }
step() { echo "  → $1"; }
pass() { echo "  ✅ $1"; PASS=$((PASS + 1)); }
fail() { echo "  ❌ $1"; FAIL=$((FAIL + 1)); }
warn() { echo "  ⚠️  $1"; WARN=$((WARN + 1)); }

# req METHOD URL [BODY_JSON] [HEADER_EXTRA]
# Preenche HTTP_STATUS e HTTP_BODY.
req() {
  local method="$1" url="$2" data="${3:-}" extra_header="${4:-}"
  local tmp args
  tmp=$(mktemp)
  args=(-s -o "$tmp" -w "%{http_code}" -X "$method" "$url" -H "Content-Type: application/json")
  [ -n "$extra_header" ] && args+=(-H "$extra_header")
  [ -n "$data" ] && args+=(-d "$data")
  HTTP_STATUS=$(curl "${args[@]}" 2>/dev/null)
  HTTP_BODY=$(cat "$tmp")
  rm -f "$tmp"
}

db_query() {
  local dbname="$1" sql="$2"
  docker exec -e PGPASSWORD="$PG_PASSWORD" -i "$PG_CONTAINER" \
    psql -U "$PG_USER" -d "$dbname" -t -A -c "$sql" 2>/dev/null
}

gen_cpf() {
  python3 - <<'PY'
import random

def dv(digits):
    s = sum(d * (len(digits) + 1 - i) for i, d in enumerate(digits))
    r = s % 11
    return 0 if r < 2 else 11 - r

base = [random.randint(0, 9) for _ in range(9)]
while len(set(base)) == 1:
    base = [random.randint(0, 9) for _ in range(9)]
d1 = dv(base)
d2 = dv(base + [d1])
print(''.join(map(str, base + [d1, d2])))
PY
}

wait_for_health() {
  local host_url="$1" nome="$2" timeout="${3:-20}"
  local elapsed=0
  while [ "$elapsed" -lt "$timeout" ]; do
    code=$(curl -s -o /dev/null -w "%{http_code}" "$host_url/health" 2>/dev/null)
    [ "$code" = "200" ] && return 0
    sleep 1
    elapsed=$((elapsed + 1))
  done
  echo "timeout esperando $nome ficar saudável" >&2
  return 1
}

kill_port() {
  # -sTCP:LISTEN é importante: sem isso, `lsof -ti :porta` também pega
  # processos com uma conexão *aberta* para essa porta (ex: o navegador com
  # o dashboard aberto), não só quem está escutando nela.
  lsof -nP -tiTCP:"$1" -sTCP:LISTEN | xargs kill -9 2>/dev/null || true
}

restart_estoque() {
  step "Reiniciando o serviço de estoque..."
  (cd services/estoque && nohup go run ./cmd/main.go >/tmp/estoque_cenarios.log 2>&1 &)
  if wait_for_health "$HOST_ESTOQUE" "estoque" 25; then
    pass "estoque voltou ao ar e respondendo"
  else
    fail "estoque não voltou ao ar a tempo — veja /tmp/estoque_cenarios.log"
  fi
}

# Garante que o estoque volta no ar mesmo se o script for interrompido no
# meio do cenário de falha.
trap 'code=$(curl -s -o /dev/null -w "%{http_code}" "$HOST_ESTOQUE/health" 2>/dev/null); [ "$code" = "200" ] || restart_estoque' EXIT

# ---------------------------------------------------------------------------
# Pré-requisitos
# ---------------------------------------------------------------------------

header "Pré-requisitos"

check_health() {
  local nome="$1" host_url="$2" obrigatorio="$3"
  code=$(curl -s -o /dev/null -w "%{http_code}" "$host_url/health" 2>/dev/null)
  if [ "$code" = "200" ]; then
    pass "$nome está no ar ($host_url)"
  elif [ "$obrigatorio" = "true" ]; then
    fail "$nome não respondeu em $host_url — rode: make docker-up && make run-all"
    echo
    echo "Abortando: os cenários obrigatórios precisam do usuario, estoque e faturamento no ar."
    exit 1
  else
    warn "$nome não respondeu em $host_url — o cenário de IA vai ser pulado"
  fi
}

check_health "usuario" "$HOST_USUARIO" true
check_health "estoque" "$HOST_ESTOQUE" true
check_health "faturamento" "$HOST_FATURAMENTO" true
check_health "ia" "$HOST_IA" false

docker exec "$PG_CONTAINER" true >/dev/null 2>&1 || {
  fail "container $PG_CONTAINER não está rodando — rode: make docker-up"
  exit 1
}

# ===========================================================================
# Requisito 3: Conexão real com banco de dados
# ===========================================================================

header "Requisito 3 — Persistência real em banco de dados"

CPF_PERSIST=$(gen_cpf)
EMAIL_PERSIST="persistencia.$(date +%s).$RANDOM@teste.com"

step "Criando usuário via API (email: $EMAIL_PERSIST)..."
req POST "$USUARIO_URL/usuarios" "{\"nome\":\"Teste Persistencia\",\"email\":\"$EMAIL_PERSIST\",\"cpf\":\"$CPF_PERSIST\"}"
if [ "$HTTP_STATUS" = "201" ]; then
  pass "usuário criado via API (HTTP 201)"
else
  fail "criação via API falhou (HTTP $HTTP_STATUS): $HTTP_BODY"
fi
USUARIO_ID=$(echo "$HTTP_BODY" | jq -r '.id')

step "Consultando a tabela 'usuarios' direto no Postgres (sem passar pela API)..."
DB_ROW=$(db_query usuario_db "SELECT email FROM usuarios WHERE id = '$USUARIO_ID';")
if [ "$DB_ROW" = "$EMAIL_PERSIST" ]; then
  pass "linha encontrada de verdade na tabela do Postgres — não é um fake em memória"
else
  fail "não encontrei o registro na tabela 'usuarios' (obtido: '$DB_ROW')"
fi

# ===========================================================================
# Requisito 2: Tratamento de falhas e recuperação
# ===========================================================================

header "Requisito 2 — Falha de um microsserviço e recuperação"

CODIGO_FALHA="TESTE-FALHA-$(date +%s)-$RANDOM"
step "Criando produto com saldo 5 (código: $CODIGO_FALHA)..."
req POST "$ESTOQUE_URL/produtos" "{\"codigo\":\"$CODIGO_FALHA\",\"descricao\":\"Produto cenario de falha\",\"saldo\":5}"
PRODUTO_FALHA_ID=$(echo "$HTTP_BODY" | jq -r '.id')
[ "$HTTP_STATUS" = "201" ] && pass "produto criado" || fail "não consegui criar produto (HTTP $HTTP_STATUS)"

step "Criando nota fiscal aberta para esse produto..."
req POST "$FATURAMENTO_URL/notas-fiscais" "{\"itens\":[{\"produto_id\":\"$PRODUTO_FALHA_ID\",\"quantidade\":1}]}"
NOTA_FALHA_ID=$(echo "$HTTP_BODY" | jq -r '.id')
[ "$HTTP_STATUS" = "201" ] && pass "nota fiscal criada (status ABERTA)" || fail "não consegui criar nota (HTTP $HTTP_STATUS)"

step "Derrubando o serviço de estoque (simulando uma falha real de infraestrutura)..."
kill_port 8081
sleep 1
code=$(curl -s -o /dev/null -w "%{http_code}" "$HOST_ESTOQUE/health" 2>/dev/null)
if [ "$code" = "000" ]; then
  pass "estoque confirmadamente fora do ar"
else
  warn "estoque ainda respondeu (HTTP $code) — pode não ter derrubado a tempo"
fi

step "Tentando imprimir a nota com o estoque fora do ar..."
req POST "$FATURAMENTO_URL/notas-fiscais/$NOTA_FALHA_ID/imprimir" ""
if [ "$HTTP_STATUS" = "503" ]; then
  MSG=$(echo "$HTTP_BODY" | jq -r '.error.message')
  pass "faturamento devolveu erro claro ao usuário (HTTP 503: \"$MSG\"), sem travar nem derrubar o processo"
else
  fail "esperava HTTP 503 com mensagem de indisponibilidade, veio HTTP $HTTP_STATUS: $HTTP_BODY"
fi

step "Conferindo que a nota continua ABERTA (nada ficou em estado inconsistente)..."
req GET "$FATURAMENTO_URL/notas-fiscais/$NOTA_FALHA_ID"
STATUS_NOTA=$(echo "$HTTP_BODY" | jq -r '.status')
[ "$STATUS_NOTA" = "ABERTA" ] && pass "nota permanece ABERTA após a falha" || fail "nota deveria continuar ABERTA, está '$STATUS_NOTA'"

restart_estoque

step "Tentando imprimir de novo, agora com o estoque recuperado..."
req POST "$FATURAMENTO_URL/notas-fiscais/$NOTA_FALHA_ID/imprimir" ""
if [ "$HTTP_STATUS" = "200" ]; then
  STATUS_NOTA=$(echo "$HTTP_BODY" | jq -r '.status')
  [ "$STATUS_NOTA" = "FECHADA" ] && pass "sistema se recuperou sozinho: nota fechada com sucesso, sem intervenção manual" \
    || fail "impressão retornou 200 mas status é '$STATUS_NOTA'"
else
  fail "esperava HTTP 200 após recuperação, veio HTTP $HTTP_STATUS: $HTTP_BODY"
fi

# ===========================================================================
# Opcional a: Concorrência — saldo 1, duas notas simultâneas
# ===========================================================================

header "Opcional (a) — Concorrência: saldo 1 disputado por duas notas"

CODIGO_CONC="TESTE-CONCORRENCIA-$(date +%s)-$RANDOM"
step "Criando produto com saldo 1 (código: $CODIGO_CONC)..."
req POST "$ESTOQUE_URL/produtos" "{\"codigo\":\"$CODIGO_CONC\",\"descricao\":\"Produto cenario de concorrencia\",\"saldo\":1}"
PRODUTO_CONC_ID=$(echo "$HTTP_BODY" | jq -r '.id')
[ "$HTTP_STATUS" = "201" ] && pass "produto criado com saldo 1" || fail "não consegui criar produto (HTTP $HTTP_STATUS)"

step "Criando duas notas fiscais abertas, cada uma pedindo 1 unidade do mesmo produto..."
req POST "$FATURAMENTO_URL/notas-fiscais" "{\"itens\":[{\"produto_id\":\"$PRODUTO_CONC_ID\",\"quantidade\":1}]}"
NOTA_A_ID=$(echo "$HTTP_BODY" | jq -r '.id')
req POST "$FATURAMENTO_URL/notas-fiscais" "{\"itens\":[{\"produto_id\":\"$PRODUTO_CONC_ID\",\"quantidade\":1}]}"
NOTA_B_ID=$(echo "$HTTP_BODY" | jq -r '.id')
pass "notas A ($NOTA_A_ID) e B ($NOTA_B_ID) criadas"

step "Imprimindo as duas notas ao mesmo tempo (em paralelo de verdade)..."
TMP_A=$(mktemp)
TMP_B=$(mktemp)
(curl -s -o /dev/null -w "%{http_code}" -X POST "$FATURAMENTO_URL/notas-fiscais/$NOTA_A_ID/imprimir" >"$TMP_A") &
PID_A=$!
(curl -s -o /dev/null -w "%{http_code}" -X POST "$FATURAMENTO_URL/notas-fiscais/$NOTA_B_ID/imprimir" >"$TMP_B") &
PID_B=$!
wait "$PID_A" "$PID_B"
CODE_A=$(cat "$TMP_A")
CODE_B=$(cat "$TMP_B")
rm -f "$TMP_A" "$TMP_B"

step "Nota A -> HTTP $CODE_A | Nota B -> HTTP $CODE_B"
SUCESSOS=0
[ "$CODE_A" = "200" ] && SUCESSOS=$((SUCESSOS + 1))
[ "$CODE_B" = "200" ] && SUCESSOS=$((SUCESSOS + 1))
FALHAS_409=0
[ "$CODE_A" = "409" ] && FALHAS_409=$((FALHAS_409 + 1))
[ "$CODE_B" = "409" ] && FALHAS_409=$((FALHAS_409 + 1))

if [ "$SUCESSOS" -eq 1 ] && [ "$FALHAS_409" -eq 1 ]; then
  pass "exatamente uma nota venceu a disputa pelo saldo, a outra recebeu 409 (saldo insuficiente)"
else
  fail "esperava exatamente 1 sucesso e 1 conflito 409, obtive $SUCESSOS sucesso(s) e $FALHAS_409 conflito(s) (códigos: $CODE_A, $CODE_B)"
fi

step "Conferindo o saldo final do produto..."
req GET "$ESTOQUE_URL/produtos/$PRODUTO_CONC_ID"
SALDO_FINAL=$(echo "$HTTP_BODY" | jq -r '.saldo')
if [ "$SALDO_FINAL" = "0" ]; then
  pass "saldo final é 0 — nunca ficou negativo, mesmo sob concorrência real"
else
  fail "saldo final deveria ser 0, veio '$SALDO_FINAL'"
fi

# ===========================================================================
# Opcional b: Uso de Inteligência Artificial
# ===========================================================================

header "Opcional (b) — Geração de descrição de produto por IA"

code=$(curl -s -o /dev/null -w "%{http_code}" "$HOST_IA/health" 2>/dev/null)
if [ "$code" != "200" ]; then
  warn "serviço de IA fora do ar — pulando este cenário (make run-ia para habilitar)"
else
  step "Pedindo pro estoque gerar uma descrição via IA para 'Cadeira Gamer'..."
  req POST "$ESTOQUE_URL/produtos/suggest-description" '{"nome":"Cadeira Gamer"}'
  if [ "$HTTP_STATUS" = "200" ]; then
    DESCRICAO=$(echo "$HTTP_BODY" | jq -r '.descricao')
    TAMANHO=$(python3 -c "print(len('''$DESCRICAO'''))" 2>/dev/null || echo -1)
    if [ -n "$DESCRICAO" ] && [ "$TAMANHO" -ge 3 ] && [ "$TAMANHO" -le 200 ]; then
      pass "descrição gerada com $TAMANHO caracteres (dentro do limite do domínio): \"$DESCRICAO\""
    else
      fail "descrição gerada fora dos limites esperados (3 a 200 caracteres): \"$DESCRICAO\""
    fi
  else
    warn "chamada à IA falhou (HTTP $HTTP_STATUS) — provavelmente GROQ_API_KEY não configurada ou provedor indisponível: $HTTP_BODY"
  fi
fi

# ===========================================================================
# Opcional c: Idempotência
# ===========================================================================

header "Opcional (c) — Idempotência em escrita repetida"

CPF_IDEMP=$(gen_cpf)
EMAIL_IDEMP="idempotencia.$(date +%s).$RANDOM@teste.com"
CHAVE_IDEMP=$(uuidgen 2>/dev/null || python3 -c "import uuid; print(uuid.uuid4())")

step "Enviando a mesma criação de usuário duas vezes com a mesma Idempotency-Key..."
req POST "$USUARIO_URL/usuarios" "{\"nome\":\"Teste Idempotencia\",\"email\":\"$EMAIL_IDEMP\",\"cpf\":\"$CPF_IDEMP\"}" "Idempotency-Key: $CHAVE_IDEMP"
STATUS_1="$HTTP_STATUS"
ID_1=$(echo "$HTTP_BODY" | jq -r '.id')

req POST "$USUARIO_URL/usuarios" "{\"nome\":\"Teste Idempotencia\",\"email\":\"$EMAIL_IDEMP\",\"cpf\":\"$CPF_IDEMP\"}" "Idempotency-Key: $CHAVE_IDEMP"
STATUS_2="$HTTP_STATUS"
ID_2=$(echo "$HTTP_BODY" | jq -r '.id')

if [ "$STATUS_1" = "201" ] && [ "$STATUS_2" = "201" ] && [ "$ID_1" = "$ID_2" ]; then
  pass "segunda chamada devolveu a resposta original (mesmo id: $ID_1), sem criar um novo registro"
else
  fail "esperava a mesma resposta nas duas chamadas (status $STATUS_1/$STATUS_2, ids $ID_1/$ID_2)"
fi

step "Conferindo no banco que só existe UMA linha com esse email (o efeito colateral não duplicou)..."
QTD=$(db_query usuario_db "SELECT COUNT(*) FROM usuarios WHERE email = '$EMAIL_IDEMP';")
if [ "$QTD" = "1" ]; then
  pass "só 1 linha no banco para esse email, como esperado"
else
  fail "esperava 1 linha no banco, encontrei $QTD"
fi

# ---------------------------------------------------------------------------
# Resumo
# ---------------------------------------------------------------------------

header "Resumo"
echo "  ✅ $PASS passaram"
echo "  ⚠️  $WARN avisos (dependências opcionais indisponíveis)"
echo "  ❌ $FAIL falharam"
echo

if [ "$FAIL" -gt 0 ]; then
  bold "Resultado: FALHOU"
  exit 1
else
  bold "Resultado: OK"
  exit 0
fi

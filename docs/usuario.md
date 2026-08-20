# usuario-service

Cadastro de usuários. É o serviço mais simples do sistema — serve de base para os padrões (DDD, idempotência, concorrência otimista) reutilizados nos outros.

- Código: `services/usuario`
- Porta padrão: `8080` (`USUARIO_SERVER_PORT`)
- Banco: `usuario_db` (testes usam `usuario_test_db`)

## Modelo de domínio

`Usuario` (aggregate root, `internal/domain/entity/usuario.go`):

| Campo | Tipo | Regra |
|---|---|---|
| `id` | string (UUID) | Gerado no `create` |
| `nome` | `valueobject.Nome` | 3 a 100 caracteres |
| `email` | `valueobject.Email` | Formato validado por regex; único no sistema |
| `cpf` | `valueobject.CPF` | 11 dígitos, com verificação real dos dígitos verificadores (módulo 11); único no sistema |
| `bio` | string | Livre, sem validação de tamanho |
| `ativo` | bool | `true` na criação; `false` após "exclusão" |
| `version` | int | Controle de concorrência otimista |

O CPF não é só "tem 11 dígitos": `valueobject/cpf.go` recusa sequências repetidas (`11111111111`) e recalcula os dois dígitos verificadores pelo algoritmo oficial antes de aceitar. Um CPF com formato válido mas dígito verificador errado é rejeitado.

**"Deletar" usuário é soft delete.** `DELETE /api/usuarios/:id` chama `Usuario.Desativar()`, que só muda `ativo` para `false` — o registro nunca é removido do banco. Tentar desativar um usuário já inativo retorna erro (`ErrUsuarioInativo`).

## Endpoints

Prefixo: `/api/usuarios`. Escritas (`POST`/`PUT`) aceitam header `Idempotency-Key` (ver [arquitetura.md](./arquitetura.md#idempotência)).

| Método | Rota | Descrição |
|---|---|---|
| `GET` | `/` | Lista paginada (`?page=1&page_size=10`) |
| `GET` | `/:id` | Busca por ID |
| `POST` | `/` | Cria usuário |
| `PUT` | `/:id` | Atualiza nome/email/bio (parcial) |
| `DELETE` | `/:id` | Desativa (soft delete) |
| `GET` | `/health` | Health check (fora do prefixo `/api`) |

### `POST /api/usuarios`

```json
{
  "nome": "Ana Souza",
  "email": "ana@example.com",
  "cpf": "529.982.247-25"
}
```

Retorna `201` com o usuário criado, ou `409` se email/CPF já existirem.

### `PUT /api/usuarios/:id`

Todos os campos são opcionais (só atualiza o que vier); `version` é obrigatório e precisa bater com a versão atual do registro:

```json
{
  "nome": "Ana Souza Lima",
  "bio": "Gerente de produto",
  "version": 1
}
```

Se `version` não bater com o valor salvo no banco, retorna `409 ErrConcurrencyConflict` — outra requisição alterou o usuário entre a leitura e essa escrita.

## Erros de domínio → HTTP

| Erro | HTTP |
|---|---|
| `ErrUsuarioNotFound` | 404 |
| `ErrEmailAlreadyExists`, `ErrCPFAlreadyExists`, `ErrConcurrencyConflict` | 409 |
| `ErrCPFInvalido`, `ErrEmailInvalido`, `ErrNomeInvalido`, `ErrUsuarioInativo`, `ErrUsuarioAtivo` | 422 |
| Validação de request (`binding`) | 400 |
| Outro | 500 (logado no servidor) |

## Testes

```bash
make test              # todos os testes (precisa do Postgres rodando: make docker-up)
make test-concurrency   # só os testes de concorrência
```

`services/usuario/tests/concurrency_test.go` sobe goroutines concorrentes contra o banco real para provar dois cenários:

- **`TestConcurrentUpdateSameUsuario`**: 10 goroutines tentam atualizar o mesmo usuário ao mesmo tempo lendo a mesma versão — só uma pode vencer, as outras devem receber conflito de concorrência.
- **`TestConcurrentCreateSameEmail`** / **`TestConcurrentCreateSameCPF`**: múltiplas goroutines tentam criar usuários com o mesmo email/CPF simultaneamente — só uma pode ter sucesso.

`integration_test.go` cobre o fluxo HTTP ponta a ponta (criar, buscar, listar, atualizar, desativar).

## Rodando isoladamente

```bash
make run-usuario
# ou
cd services/usuario && go run ./cmd/main.go
```

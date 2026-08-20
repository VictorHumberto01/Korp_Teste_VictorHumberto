# ia-service

Único ponto do sistema que fala com um provedor de IA externo (Groq). Existe como serviço separado, e não como um pacote dentro do `estoque`, por três motivos:

1. **Isolamento de falha.** Se a API do Groq cair, mudar contrato ou descontinuar um modelo (já aconteceu — veja a nota no fim deste documento), o problema fica contido aqui. O `estoque` degrada de forma controlada (só a sugestão de descrição para de funcionar) em vez de propagar o erro para o resto do sistema.
2. **Segredo isolado.** A `GROQ_API_KEY` só precisa existir no ambiente deste serviço, não em todo lugar que eventualmente queira gerar texto.
3. **Reuso futuro.** Qualquer outro serviço (ex: `usuario`, que já teve uma feature de bio gerada por IA via Ollama, removida neste projeto) pode consumir este serviço sem duplicar client de LLM.

- Código: `services/ia`
- Porta padrão: `8083` (`IA_SERVER_PORT`)
- Sem banco de dados próprio — é um serviço stateless, só orquestra a chamada ao provedor.

## Endpoint

| Método | Rota | Descrição |
|---|---|---|
| `POST` | `/api/ia/produtos/descricao` | Gera uma descrição comercial curta a partir do nome de um produto |
| `GET` | `/health` | Health check |

```json
// request
{ "nome": "Cadeira Gamer" }
```
```json
// response
{ "descricao": "Conforto premium e estilo imponente: a Cadeira Gamer ergonômica com apoio lombar..." }
```

Hoje só o `estoque` chama esse endpoint (ver [estoque.md](./estoque.md#post-apiprodutossuggest-description)), mas o contrato foi desenhado para não amarrar no vocabulário do domínio de produtos além do necessário — receber um "insumo" (`nome`) e devolver um texto gerado.

## Fluxo interno

`internal/application/query/gerar_descricao_produto.go` → `internal/infrastructure/llm/groq_client.go`:

1. Monta um prompt pedindo um parágrafo curto, com no máximo 180 caracteres, sem saudações nem markdown.
2. Chama a API de chat completions do Groq (`https://api.groq.com/openai/v1/chat/completions`).
3. **Trunca o resultado em 200 caracteres reais como rede de segurança** — o modelo não obedece limite de tamanho de forma confiável, e o campo `descricao` do produto no `estoque` tem limite rígido de 200 caracteres. O corte respeita fronteira de palavra (não corta no meio) e remove aspas/pontuação solta na borda. Coberto por teste unitário (`gerar_descricao_produto_test.go`).

```go
// services/ia/internal/application/query/gerar_descricao_produto.go
func truncarDescricao(descricao string, max int) string {
    descricao = strings.Trim(strings.TrimSpace(descricao), `"'`)
    runes := []rune(descricao)
    if len(runes) <= max {
        return descricao
    }
    cortado := string(runes[:max])
    if idx := strings.LastIndexAny(cortado, " \n\t"); idx > 0 {
        cortado = cortado[:idx]
    }
    return strings.TrimRight(cortado, " ,.;:-")
}
```

## Configuração

| Variável | Padrão | Descrição |
|---|---|---|
| `IA_SERVER_PORT` | `8083` | Porta HTTP do serviço |
| `GROQ_API_KEY` | _(vazio)_ | Chave da API do Groq. Sem ela, toda chamada falha imediatamente com erro claro no log, sem tentar a rede |
| `GROQ_MODEL` | `openai/gpt-oss-20b` | Modelo usado nas chamadas |

## Logs

Toda chamada ao Groq é logada com o resultado: sucesso (modelo, tempo de resposta), ou falha detalhada (status HTTP, corpo da resposta de erro decodificado, tempo até a falha). Isso foi uma decisão deliberada: erros de IA costumam ser silenciosos por natureza (uma resposta vazia ou genérica não chama atenção), então o log do serviço `ia` é a principal ferramenta de diagnóstico quando a geração para de funcionar.

## Erros → HTTP

Diferente dos outros serviços, este não tem erros de domínio ricos — a maioria das falhas vem do provedor externo:

| Situação | HTTP |
|---|---|
| Request inválido (`nome` ausente) | 400 |
| Qualquer falha ao gerar conteúdo (chave ausente, erro do Groq, rede, resposta vazia) | 502 |

O cliente HTTP do `estoque` (`internal/infrastructure/ia/http_client.go`) reclassifica esse `502` genérico para `503 ErrIAIndisponivel` quando é claramente um problema de infraestrutura (timeout, conexão recusada, 5xx), ou repassa a mensagem com prefixo `ia:` quando é outra coisa — ver [estoque.md](./estoque.md#erros-de-domínio--http).

## Nota histórica: por que o modelo mudou

Este serviço nasceu de um bug real do projeto: o `estoque` chamava o Groq diretamente com o modelo `llama3-8b-8192`, que foi descontinuado pela Groq (`model_decommissioned`) — toda chamada falhava mesmo com a API key correta. A correção envolveu trocar para `openai/gpt-oss-20b` e, na sequência, extrair a chamada de IA para este serviço dedicado, para que uma mudança de provedor/modelo no futuro não exija tocar no código do `estoque`.

## Rodando isoladamente

```bash
make run-ia
# ou
cd services/ia && go run ./cmd/main.go
```

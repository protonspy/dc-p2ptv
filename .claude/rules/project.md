# This project

Rede federada de transmissão compartilhada. Dois projetos, um repositório: `server/` em
Go — plano de controle e nó de encaminhamento — e `client/` em TypeScript, que roda no
navegador. Eles não compartilham código; compartilham o protocolo, que é escrito nos
specs.

## Commands

```bash
# Build
cd server && go build ./...
cd client && npm run build

# Test — the whole suite
cd server && go test ./...
cd client && npm test

# Test — one package or one file (used after every task; scope, not suite)
cd server && go test ./internal/<package>/...
cd client && npx vitest run src/<file>.test.ts

# Lint — the best-practices layer that finds what tests do not
cd server && "$(go env GOPATH)/bin/golangci-lint" run ./...
cd client && ./node_modules/.bin/eslint src

# Format / format check
cd server && go fmt ./...
cd client && npm run format          # npm run format:check to verify only
```

`golangci-lint` não vem instalado com o Go: `go install
github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest`.

**Chame os binários do cliente por `./node_modules/.bin/`, não por `npx`.** Neste
ambiente o `npx` falha com `could not determine executable to run` mesmo com o pacote
instalado, e a falha se parece com um erro de lint.

## Conventions

- **Branch names:** `feat/<slug>`, `fix/<slug>`, um por unidade de trabalho, em worktree
  própria.
- **Commits:** Conventional Commits, escopo pelo projeto — `feat(server): …`,
  `fix(client): …`.
- **Idioma:** identificadores, comentários de código, `README.md`, mensagens de commit e
  corpo de PR em inglês; specs, plans, `docs/` e conversa em português. O glossário liga
  o termo canônico ao identificador.
- **Anything a new contributor gets wrong on their first try:** o nó de encaminhamento
  nunca decide nada — quem entra, em que sala, com que chave, é sempre o plano de
  controle. Código que faça um nó decidir algo sobre uma sala está no lugar errado.

## Boundaries

- `.codegraph/` — índice do CodeGraph. Nunca editar, nunca commitar.
- `docs/wiki/` — pesquisa já destilada, anterior a este plano. Corrigir sim, reescrever
  em massa não.
- `plans/sala-tempo-real.md` — selado por `scc plan approve`. Alterar por
  `scc patch`, nunca por editor.

# AGENTS.md

## Cursor Cloud specific instructions

`openbench` (`ob`) é uma CLI em Go (Go 1.22+, `github.com/spf13/cobra` + `gopkg.in/yaml.v3`) que orquestra
ambiente de desenvolvimento (Docker Compose), gera conventional commits a partir de `git diff` via IA
(OpenAI / OpenRouter / Gemini) e integra com o GitHub CLI (`gh`) para push + PR.

Há também um **app desktop** em migração (**Wails v3** + React), entry na raiz (`main.go`, `frontend/`).
A CLI permanece até a fase de corte do release desktop (ver `docs/architecture/app-desktop-wails.md`).

### Build / lint / test / run

Comandos padrão (rodar da raiz do repo):

- Build (tudo Go): `go build ./...`
- Lint: `go vet ./...`
- Test: `go test ./...`
- Instalar a CLI: `go install ./cmd/ob` (instala como `openbench` em `$(go env GOPATH)/bin`)
- Rodar CLI: `ob --help` ou `openbench --help` (alias `ob` opcional via `install.sh`)

#### App desktop (Wails)

Pré-requisitos: `go install github.com/wailsapp/wails/v3/cmd/wails3@latest` e Node/npm.

- Dev: `wails3 dev`
- Build: `wails3 build` → `bin/openbench`
- Rodar binário: `./bin/openbench`
- Auto-update: ver `docs/release-desktop.md` (chave pública em `build/updater/updater.key.pub`)

Arquitetura: `docs/architecture/app-desktop-wails.md`. Discovery: `docs/discovery/resumo-app-desktop-nativo.md`.

### Versionamento

A versão é **automática**, derivada do número de commits no repositório (sem tags git).
O primeiro commit equivale a `v0.1.0`; cada commit adicional incrementa o patch.
Ex.: 13 commits → `v0.1.12`. O `go install` injeta versão e commit via `-ldflags`.

### Caveats não óbvios

- **Toda** ação de `commit`/`push`/`pr` (inclusive com `--dry-run`) carrega a config e
  faz uma chamada HTTP real ao provider de IA. Sem `api_key` válida o comando falha.
- A chave pode vir do arquivo de config OU da env var `OB_API_KEY`.
- Config: `~/.config/openbench/config.yaml` (ou `.openbench.yaml` local, ou `OB_CONFIG`).
- `ob config` preserva valores existentes — Enter mantém cada campo.
- `ob update` funciona de qualquer diretório (usa clone salvo ou GitHub).
- `ob pr` requer `gh` autenticado (`gh auth login`).
- `ob docker *` requer Docker CLI + daemon; usa `docker compose` via exec.
- Instalação completa: `./install.sh` (Go + binário + PATH + `ob config`).
- O app desktop usa `//go:embed` de `frontend/dist` — o layout Wails fica na **raiz** do módulo
  (não em `cmd/openbench`) por limitação do embed (`..` não é permitido).

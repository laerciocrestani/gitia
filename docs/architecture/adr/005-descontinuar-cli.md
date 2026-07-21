# ADR-005 — Descontinuação total da CLI na v1 do app

## Status

Aceito (decisão de produto)

## Contexto

Descoberta: migrar tudo para o app; CLI **sem** período de deprecação. Hoje existem `cmd/ob`, TUI Bubble Tea e usuários/scripts potenciais.

## Decisão

- **Release v1 desktop:** artefato distribuído = app Wails apenas; README/install passam a ser app-only.
- No repositório, durante desenvolvimento: manter testes do **domain core**; remover entry Cobra/`ob ui` do produto final (pode permanecer branch/histórico).
- Não manter binário `ob` paralelo “só para scripts”.
- Comunicar ruptura nas release notes da v1 (único aviso).

## Consequências

- **+** Um produto, uma UX, menos superfície de manutenção de UI terminal.
- **−** Quebra automação/CI que chama `ob`.
- **−** Power users perdem scripting direto (mitigação fraca: aceitar ruptura, ou no futuro API estável — fora da v1).

## Nota de implementação

Cortar a CLI no **momento do release** (fase 8), não no primeiro commit de scaffold — permite desenvolver o app com o core ainda testável via `go test` sem depender de Cobra.
EOF
# ADR-003 — Preferências locais e multi-projeto

## Status

Aceito

## Contexto

Config de IA/provider já vive em `~/.config/openbench/config.yaml`. O app precisa persistir estado de UI: projetos pinned/recentes, checks “validar commit/PR”, último projeto focado — sem conta cloud.

## Decisão

- Manter **config de domínio** no arquivo atual (`config.yaml` / overrides existentes).
- Novo arquivo de prefs de UI, ex.: `~/.config/openbench/desktop.yaml` (ou `ui.yaml`), separado para não poluir schema da config de IA.
- Modelo: lista de pinned com `path`, `alias` opcional, ordem; recentes; flags `validate_commit`, `validate_pr`.
- Status paralelo via `StatusHub` com limite de pinned ativos (8 na v1).

## Consequências

- **+** Migração suave: quem já tem API key continua funcionando.
- **+** Prefs de UI versionáveis sem quebrar `config.Load`.
- **−** Dois arquivos de config para documentar.

## Alternativas rejeitadas

- Guardar pinned dentro de `config.yaml` misturado com provider (acopla schemas).
- Sync cloud (fora do escopo v1).
EOF
# ADR-004 — Auto-update via Wails Updater + GitHub Releases

## Status

Aceito

## Contexto

Requisito de descoberta: auto-update em todas as plataformas da v1. Wails v3 inclui `updater` com provider GitHub, manifest assinado e `CheckInterval`.

## Decisão

- Usar **Wails v3 Updater** com provider **GitHub** no repositório do openbench.
- Publicar artefatos com `wails3 package` + `wails3 updater manifest` (assinatura).
- Checagem automática periódica (ex.: 6h) + opção “Checar atualização” em Settings.
- Versão do app alinhada à estratégia atual (ldflags / contagem de commits ou semver de release).

## Consequências

- **+** Atende requisito sem serviço próprio.
- **−** Exige disciplina de release (formatos zip/app, tags, chaves).
- **−** Notarização/assinatura de código Apple/Windows continua necessária para boa UX de instalação (paralelo ao updater).

## Alternativas rejeitadas

- Só update manual por download (não atende requisito).
- Sparkle só macOS (quebra paridade Win/Linux).
EOF
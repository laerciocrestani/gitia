# ADR-006 — Docker presets e kits importáveis

## Status

Aceito

## Contexto

CLI/TUI já permitem `docker compose exec`/`sh`, mas sem catálogo reutilizável. O desktop precisa de shell no container, execução de comandos frequentes (Laravel artisan) e o mesmo catálogo na CLI/TUI.

## Decisão

1. **Arquivo por projeto:** `.openbench/docker-presets.yaml` (versionável no repo; separado da config de IA).
2. **Kits embutidos** no binário (`internal/dockerpresets/kits/*.yaml`) via `//go:embed`; fluxo **Importar kit** mescla por `id` sem sobrescrever presets customizados do mesmo id sem intenção (merge: novos ids entram; ids do kit já presentes são atualizados só se `kit` bater e comando for o do kit — v1: skip se id já existe).
3. **Dois caminhos de execução:**
   - **Shell:** PTY `docker compose exec -it <svc> sh|bash` no TerminalPanel.
   - **Preset one-shot:** `docker compose exec -T <svc> <cmd>` com captura de stdout/stderr; sessão encerra; UI mostra modal com resumo (textarea).
   - Presets `interactive: true` (ex. tinker) abrem shell com o comando como processo (`exec -it … php artisan tinker`).
4. **Serviço** é escolhido na execução (não obrigatório no YAML).
5. **Paridade:** APIs em `internal/dockerpresets` + `app`/`desktop`; CLI `ob docker preset|kit`; TUI Environment com atalho de presets.

## Consequências

- **+** Presets versionáveis com o projeto; kits reutilizáveis sem acoplar a detecção de framework.
- **+** Separação clara shell vs one-shot (evita misturar PTY e captura).
- **−** Mais um arquivo de config de projeto para documentar.
- **−** Terminal Windows (ConPTY) continua limitado; shell em container segue a mesma restrição.

## Alternativas rejeitadas

- Presets só em `desktop.yaml` global (não viaja com o repo).
- Misturar presets em `.openbench.yaml` de IA (acopla schemas).
- Sempre detectar Laravel automaticamente (usuário prefere import explícito do kit).

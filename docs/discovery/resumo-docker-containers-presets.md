# Resumo do Entendimento — Docker: containers, shell e presets

## Problema e objetivo

No desktop, o card Docker só cobre lifecycle (Up/Down/Start/Stop/Recreate). O usuário precisa inspecionar os containers do compose, abrir shell interativo e executar comandos frequentes (ex. Laravel artisan) de forma moderna e amigável — no desktop e com o mesmo catálogo na CLI/TUI.

Hoje a CLI/TUI já tem `exec`/`sh` e placeholder `php artisan migrate`, mas sem catálogo de presets, sem kits importáveis e sem fluxo de “resultado em resumo” no desktop.

## Sistema existente

- **Docker core:** `internal/docker` (compose detect, list containers, up/down/exec/shell/logs).
- **CLI/TUI:** `ob docker *` + Environment view (shell `E`, exec custom `x`).
- **Desktop:** `internal/desktop/docker.go` + card em `frontend/src/App.tsx` (lifecycle only).
- **Terminal desktop:** PTY no root do projeto (`TerminalPanel`), não no container.
- **Config:** `~/.config/openbench/config.yaml` / `.openbench.yaml` (AI/git/UI) — sem presets Docker.
- **Padrão de catálogo a reutilizar:** sync modes / branch templates.

## Restrições organizacionais

- Evolução brownfield; reaproveitar `internal/docker` e bindings Wails.
- Preservar CLI até o corte do release desktop (paridade desejada nesta entrega).

## Atores

- **Dev local:** usa desktop (principal) e CLI/TUI; escolhe container, abre shell ou roda preset.
- **Projeto (repo):** guarda presets importados/editados (por projeto).

## Requisitos funcionais

1. Com Docker disponível, clicar no card Docker abre painel/sheet com **todos os containers** do compose do projeto.
2. Por container: ações de lifecycle relevantes + **Shell** (PTY interativo via `docker compose exec`).
3. **Presets de comando** por projeto; usuário **escolhe o container** antes de executar o preset.
4. Execução de preset: `docker compose exec` (não-interativo / one-shot) no serviço escolhido; ao terminar, **fecha a sessão** e mostra **modal de resumo** (textarea com output + botão fechar só no header).
5. Shell manual: usuário entra no container e digita comandos livremente no `TerminalPanel` (ou equivalente PTY).
6. **Kits empacotados** no app (ex. kit Laravel com artisan essenciais); fluxo **Importar kit** mescla/copia presets para o projeto.
7. Mesmo catálogo/presets na **CLI e TUI** (ex. listar/rodar preset; Environment com presets).
8. Usuário pode editar/adicionar/remover presets do projeto após import.

## Requisitos não funcionais

- UX moderna e amigável no desktop (sheet/dialog, não só botões soltos).
- Output do preset deve ser legível no modal (textarea).
- Não expor shell arbitrário fora do domínio Docker/projeto (alinhar com restrição de bindings do desktop).
- Presets versionáveis com o repo (arquivo no projeto).

## Regras de negócio

- Presets **vivem por projeto** (não só globais).
- Kits built-in são **base para import**, não substituem automaticamente presets do projeto sem ação do usuário.
- Kit Laravel inclui comandos artisan essenciais (migrate, seed, tinker, cache:clear, etc. — lista fina na arquitetura).
- Serviço alvo do preset é escolhido **na hora da execução** (não obrigatoriamente fixo no YAML do preset; preset pode sugerir default opcional — PREMISSA: default opcional ok, escolha sempre confirmável).
- Card Docker só aparece quando Docker/compose estiver disponível/visível (comportamento atual).

## Fluxos principais

1. **Abrir ambiente:** card Docker → lista containers (nome, state, serviço).
2. **Shell:** seleciona container → abre PTY `compose exec -it <svc> sh|bash` no TerminalPanel.
3. **Preset:** seleciona container → escolhe preset → app executa `compose exec <svc> <cmd>` → captura stdout/stderr/exit → fecha sessão → modal resumo (textarea + fechar no header).
4. **Importar kit:** UI/CLI “Importar kit Laravel” → mescla presets no arquivo do projeto → lista atualizada.
5. **CLI/TUI:** listar kits/presets do projeto; executar preset pedindo/passando service; TUI Environment com atalho para presets.

## Integrações externas

- **Docker CLI + Compose:** list/ps, exec, shell (já usados).
- Arquivos de kit embutidos no binário (`//go:embed` ou similar) + arquivo de presets no projeto.

## Restrições e premissas

- PREMISSA: presets do projeto em `.openbench.yaml` (seção nova) **ou** arquivo dedicado (ex. `.openbench/docker-presets.yaml`) — decidir na arquitetura; preferência por algo versionável e claro.
- PREMISSA: modal de resumo no desktop; na CLI o “resumo” é stdout do comando (exit code); na TUI, viewport/dialog equivalente quando fizer sentido.
- PREMISSA: kit Laravel sempre disponível para import, independente de detectar Laravel no disco (usuário decide importar).
- Escopo: **desktop + CLI + TUI** na mesma entrega de catálogo.

## Riscos identificados

- **PTY vs one-shot:** shell interativo e exec de preset são caminhos distintos; misturar no mesmo PTY pode confundir — mitigar com APIs separadas (`DockerShell` vs `DockerExecPreset`).
- **Comandos longos/interativos em preset** (ex. `tinker`): one-shot + modal pode ser inadequado — mitigar: presets interativos marcados como “shell” ou abrir shell em vez de one-shot.
- **Merge de kits:** import duplicado — mitigar: merge por id/nome, não sobrescrever custom sem confirmação.
- **Windows:** terminal ConPTY ainda stub — shell em container pode ficar macOS/Linux first; documentar limitação.

## Decisões fechadas (pós-confirmação)

- Arquivo: `.openbench/docker-presets.yaml`
- Schema: `id`, `label`, `description`, `command`, `kit`, `interactive`
- Kit Laravel com artisan essenciais; `tinker` = `interactive: true` (shell)
- Desktop: Sheet de containers + modal de resumo (textarea, fechar no header)
- ADR: `docs/architecture/adr/006-docker-presets-kits.md`

---

**Confirmado — implementação em andamento / entregue.**

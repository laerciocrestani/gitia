# ADR-007 — Transporte CI/CD via GitHub CLI (`gh`)

## Status

Aceito

## Contexto

A descoberta exige orquestrar GitHub Actions (listar/acompanhar runs, logs, re-run, `workflow_dispatch`, minutos) em github.com e GitHub Enterprise, reaproveitando auth já usada no openbench. Hoje `internal/pr` já executa `gh` via `os/exec` no diretório do projeto.

## Decisão

1. **Transporte primário:** subprocesso `gh` (`gh run`, `gh workflow`, `gh api`) no workdir do projeto, no mesmo padrão de `internal/pr`.
2. **Novo pacote de domínio:** `internal/gha` (GitHub Actions) — API tipada que retorna structs; **não** imprime na UI; **não** importa Wails.
3. **Enterprise:** usar o host já resolvido pelo `gh` (`[HOST/]OWNER/REPO`, `GH_HOST` / auth por hostname). Nenhum OAuth próprio na v1 desta feature.
4. **Billing/minutos:** quando `gh run view`/`api` não bastar, chamar `gh api` nos endpoints de Actions/billing; se 403/404, expor `UsageStatus=unavailable` com motivo — nunca omitir o bloco de custo na UI.
5. **SDK Go oficial (`go-github`) fica fora da v1** desta feature, para não duplicar auth/Enterprise e manter uma única superfície operacional (`gh doctor` / onboarding).

## Consequências

- **+** Auth, SSO Enterprise e permissões iguais às que o usuário já configurou no `gh`.
- **+** Paridade com o código existente de PR/checks.
- **−** Parsing de JSON/`watch` acoplado à CLI; precisa testes com fixtures e tolerância a exit codes ≠ 0 (como em `ChecksBestEffort`).
- **−** Dependência de `gh` instalado no PATH (já é pré-requisito de PR).

## Alternativas rejeitadas

| Alternativa | Por que não |
|-------------|-------------|
| `go-github` + token em config openbench | Segunda auth; pior UX Enterprise/SSO; diverge do onboarding atual |
| Abrir só browser (`gh run view --web`) | Viola o objetivo “tudo no openbench” |
| Actions local / act / runner embutido | Fora de escopo da descoberta |

# ADR-008 — Confirmação humana, sanitização de logs e evidência de minutos

## Status

Aceito

## Contexto

Requisitos da descoberta: (1) openbench mastiga o máximo, usuário só confirma efeitos colaterais; (2) nunca expor secrets; (3) minutos de Actions sempre em evidência; (4) IA corrige a partir de logs.

## Decisão

### 1. Preview / Confirm para toda mutação CI

Mesmo contrato mental de commit/push/PR:

| Ação | Preview | Confirm |
|------|---------|---------|
| Push (já existe) | + flag `DefaultBranchWarning` se destino = default | usuário confirma |
| Re-run run/job | impacto (run id, jobs, estimativa de minutos se conhecida) | `ConfirmRerun` |
| `workflow_dispatch` | inputs + workflow id | `ConfirmDispatch` |
| Correção IA → commit/push | diff proposto + mensagem + alerta main | confirma em cadeia |

Nenhuma mutação via tray/chat sem o mesmo Confirm.

### 2. Pipeline de log: Fetch → Redact → (opcional) AI window

```
gh run view/download log
    → redact.Secrets(log)          // nunca persiste versão raw em prefs/events
    → UI mostra versão redigida
    → AI recebe só "failure window" (job/step falho + N linhas de contexto), já redigida
```

Regras mínimas de redação (v1):

- Padrões: `ghp_`, `gho_`, `github_pat_`, `AKIA`, `xox[baprs]-`, `Bearer `, `Authorization:`, PEM private keys, `-----BEGIN .*PRIVATE KEY-----`
- Pares `KEY=VALUE` / `key: value` quando key casa com `(?i)(secret|token|password|api[_-]?key|credential)`
- Substituição estável: `***REDACTED***`
- Se após redação o trecho útil for vazio → IA não é chamada; UI explica “log só com material sensível / vazio”

Raw **não** vai para: eventos Wails de debug, arquivo de prefs, telemetria (inexistente), prompts.

### 3. Evidência de minutos (sempre visível)

Componente/bloco obrigatório `ActionsUsage` em qualquer tela/fluxo de CI, com um dos estados:

| Estado | Conteúdo |
|--------|----------|
| `run` | minutos billable do run atual (quando API devolver) + duração wall-clock |
| `repo_window` | soma dos runs recentes da sessão/branch (best-effort) |
| `org` | restante/usado do billing org (se permissão) |
| `unavailable` | motivo (`forbidden`, `enterprise_unsupported`, `no_permission`, `api_error`) — bloco permanece |

Antes de Confirm de re-run/dispatch: mostrar **aviso de custo** (“vai consumir minutos adicionais”).

### 4. Environments / approvals

Openbench **observa** status `waiting` / conclusion que indique approval; mensagem: “aguardando approval no GitHub”. Não implementa approve/reject de environment na v1.

## Consequências

- **+** Alinha com o modelo mental já usado no app (Preview/Confirm).
- **+** Reduz vazamento acidental para IA/UI.
- **+** Custo não fica escondido mesmo quando a API falha.
- **−** Redação por regex é incompleta (secrets customizados podem passar) — documentar limite; evolução com denylist configurável depois.
- **−** Billing org frequentemente 403 para devs comuns — UX de `unavailable` precisa ser boa.

## Alternativas rejeitadas

- Auto-push após fix da IA sem confirm.
- Enviar log completo bruto ao provider de IA.
- Esconder bloco de minutos quando billing falhar.
- Aprovar deployments de environment pelo app na v1.

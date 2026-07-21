# Resumo do Entendimento — App desktop nativo (openbench)

## Problema e objetivo

O openbench hoje é uma CLI (`ob`) em Go que orquestra ambiente de desenvolvimento (Docker Compose), gera conventional commits a partir de `git diff` via IA e integra com GitHub (`gh`) para push + PR. O uso por linha de comando gera atrito: lembrar comandos, pouca atratividade visual e sensação de produto “antigo”.

O objetivo é **substituir a CLI por um aplicativo desktop nativo**, moderno, com cliques, ícone na taskbar/menu bar para acesso rápido e uma janela completa para trabalho no projeto. O público é geral (produto público); uso em equipe no sentido de várias pessoas no próprio PC é válido. Features colaborativas de time (settings/templates compartilhados) são desejáveis no futuro, não na v1.

## Sistema existente

- CLI Go (`cobra` + YAML), binário único, sem servidor/DB.
- Capacidades: config local, commit via IA (OpenAI / OpenRouter / Gemini), push/PR via `gh`, Docker Compose via CLI Docker.
- Config: `~/.config/openbench/config.yaml` (ou `.openbench.yaml` / `OB_CONFIG`); API key também via `OB_API_KEY`.
- **O que muda de forma ruptura:** a CLI será **totalmente descontinuada na v1 do app**, sem período de deprecação/aviso.
- **O que deve ser preservado:** paridade funcional com os comandos atuais de commit, PR e Docker; aproveitar config local existente para o usuário não perder setup.
- **Stack alvo:** app desktop com **Wails** (backend/core em Go + frontend web na UI), reaproveitando a lógica Go existente sempre que possível.

## Restrições organizacionais

- Sem deadline formal.
- Sem restrição explícita de tamanho de time ou orçamento.
- Prioridade de plataforma: **macOS primeiro**, com **Windows e Linux na v1** (funcionalidade alinhada; polish visual pode ser “bom o suficiente”).
- Stack fixada: **Wails** (Go + UI web).

## Atores

- **Usuário individual (dev):** abre projetos, vê status, executa Commit / PR / Docker via UI e atalhos da tray.
- **Usuário em equipe (v1):** mesmo produto, cada um no seu ambiente local — sem colaboração em nuvem.
- **Operador futuro (fora da v1):** possível dono de políticas/templates compartilhados do time (escopo B adiado).

## Requisitos funcionais

1. App desktop com janela principal + ícone na taskbar/menu bar.
2. **Open project** como entrada; em seguida **dashboard do projeto** com ações Commit, PR e Docker.
3. **Multi-projeto:** vários projetos *pinned*/abas, com status em paralelo.
4. Tray/menu bar dispara também os **comandos mais usados** (além de reabrir a janela).
5. Paridade com a CLI atual nos fluxos de **commit**, **PR** e **Docker** (opções equivalentes ao happy path e flags relevantes de hoje).
6. **Docker opcional:** se Docker não estiver disponível, as opções de Docker não aparecem; o restante do app funciona.
7. Dashboard exibe, sem ação extra: branch atual, dirty working tree, containers up/down (se Docker), PR aberta, status da última geração de commit.
8. **Onboarding** quando faltar pré-requisito (ex.: API key, `gh` auth, remote) — orientar o usuário em vez de só falhar/esconder.
9. Usuário **sempre revisa** a mensagem de commit gerada pela IA antes de confirmar.
10. Preferências/checks de **validar commit** e **validar PR** (confirmação/review adicional antes de executar).
11. Config e API key **somente locais**; app deve localizar/reutilizar config existente.
12. Auto-update do aplicativo em todas as plataformas alvo da v1.
13. Sem conta na nuvem e sem telemetria/analytics na v1.

## Requisitos não funcionais

- **Plataformas:** macOS (prioridade), Windows, Linux na v1.
- **Rede:** offline parcial — status git/docker local permanece útil; Commit/PR falham com erro amigável.
- **Segurança:** segredos (API key) apenas em armazenamento local; sem sync cloud.
- **UX:** mesma simplicidade da CLI, com interação por cliques e visual moderno.
- **Atualização:** auto-update obrigatório na v1.
- **Privacidade:** sem telemetria na v1.
- **Distribuição pública:** instalável por qualquer pessoa (não ferramenta interna apenas).
- **Tecnologia:** Wails; core em Go; UI web embutida.

## Regras de negócio

- Commit via IA nunca aplica sem revisão humana da mensagem.
- Se “validar commit” / “validar PR” estiver ativo, exige passo extra de confirmação/validação antes da execução.
- Docker só é oferecido quando detectado no ambiente.
- Ausência de `gh`, API key ou remote dispara onboarding, não silêncio.
- Um ou mais projetos podem estar pinned; status de cada um atualiza em paralelo.
- Não há login, tenant ou sync de configurações entre máquinas na v1.

## Fluxos principais

1. Instalar app → (auto) carregar config local existente → onboarding se necessário → Open project / pinned.
2. Selecionar ou fixar projeto → dashboard com status (branch, dirty, docker, PR, último commit IA).
3. Commit: gerar mensagem via IA → usuário revisa/edita → (se check) validar → confirmar commit (paridade com CLI).
4. PR: fluxo equivalente ao CLI, com review se “validar PR” ativo → onboarding se `gh`/remote ausente.
5. Docker (se disponível): ações equivalentes ao `ob docker *` a partir do dashboard/tray.
6. Atalhos na tray: comandos mais usados sem necessariamente abrir a UI completa.
7. Offline: consultar status local; ações que dependem de rede/IA/GitHub falham com mensagem clara.
8. Update: app se atualiza automaticamente.

## Integrações externas

- **Providers de IA** (OpenAI / OpenRouter / Gemini): geração de mensagem de commit; requer API key local.
- **Git** (local): status, diff, commit.
- **GitHub CLI (`gh`)**: push/PR; autenticação no SO; onboarding se ausente.
- **Docker CLI / Compose** (opcional): up/down/status; oculto se não instalado/daemon indisponível.
- **Sistema de auto-update** (a definir na arquitetura): distribuição e atualização do binário/app.

## Restrições e premissas

- PREMISSA: “validar commit/PR” = preferência que adiciona confirmação/review extra (detalhe exato de UX na arquitetura).
- PREMISSA: paridade com CLI cobre os fluxos e opções usadas no dia a dia de commit/PR/docker; edge cases raros podem ser priorizados depois se necessário.
- PREMISSA: Windows/Linux na v1 com feature parity; polish visual secundário ao macOS.
- PREMISSA: sem analytics/telemetria na v1 (aceito explicitamente).
- DECISÃO: CLI removida de imediato na v1 do app — sem deprecação gradual.
- DECISÃO: sem backend/conta; tudo local.
- DECISÃO: stack = **Wails** (Go + frontend web).
- FORA DA V1: colaboração de time (templates/políticas compartilhadas).
- ATUALIZAÇÃO (2026-07-20): CLI **permanece** por enquanto; corte adiado (fase 8).

## Riscos identificados

- **Corte abrupto da CLI:** scripts, CI e power users quebram sem migração — mitigação pendente (docs de migração? manter binário interno não documentado? aceitar ruptura).
- **Dependência de `gh` e Docker no host:** UX depende de onboarding sólido; risco de percepção de “app incompleto”.
- **Multi-projeto com status paralelo:** custo de polling/watchers (CPU, bateria, rate limit `gh`) — precisa desenho cuidadoso.
- **Auto-update multi-OS:** complexidade de assinatura de código (Apple/Windows) e canais de release (no Wails, detalhar na arquitetura).
- **Paridade total com CLI:** risco de escopo inchado na v1 se todas as flags forem obrigatórias no dia 1.
- **Produto público sem telemetria:** mais difícil diagnosticar falhas em campo (aceito na v1).
- **Wails / WebView:** diferenças de UI entre macOS, Windows e Linux; tray e empacotamento precisam validação cedo em cada OS.

## Lacunas / decisões pendentes

- Detalhe UX fino dos checks “validar commit” e “validar PR” (fluxo base definido na arquitetura).
- Cópia/ordem exata dos itens do menu tray (lista base na arquitetura; pode ajustar na implementação).
- Frequência exata de refresh PR/`gh` (limites iniciais no StatusHub).

## Confirmação

- Resumo de descoberta confirmado (“pode continuar”).
- Arquitetura: [`docs/architecture/app-desktop-wails.md`](../architecture/app-desktop-wails.md).

# Resumo do Entendimento — Orquestração CI/CD (GitHub Actions) no openbench

## Problema e objetivo

Hoje o ciclo pós-código (push/PR → CI → inspecionar falha → corrigir → repetir) força o time a sair do openbench e ir ao GitHub no browser. Isso quebra o fluxo de trabalho e atrasa a adoção diária do produto.

O objetivo é **orquestrar, observar e reagir à CI que já existe no GitHub** (Actions) **dentro do openbench** (app desktop e CLI), de ponta a ponta: commit → push/PR → acompanhamento dos workflows → logs → correção assistida por IA → confirmação do usuário → novo ciclo ou re-run quando fizer sentido — sem depender do browser do GitHub no dia a dia.

## Sistema existente

- openbench: CLI Go (`ob`) + app desktop Wails (migração em curso); core em Go; UI React no desktop.
- Já integra GitHub via **GitHub CLI (`gh`)** para push/PR; autenticação no SO; onboarding se ausente.
- Já gera commits via IA a partir de `git diff` e orquestra ambiente Docker.
- **Não existe** hoje superfície de GitHub Actions / runs / logs / re-run no app ou na CLI.
- Escopo de sessão: **apenas o repo do projeto aberto** (cada repo traz seus próprios workflows em `.github/workflows`).
- **O que NÃO muda nesta iniciativa:** o CI continua sendo GitHub Actions no GitHub; openbench não vira runner nem substitui o provider de CI.

## Restrições organizacionais

- Prioridade de produto: o time deve **usar o openbench no dia a dia**; esta capacidade é parte do caminho para isso.
- Entrega pedida: **nascer completo** (orquestração + observação + reação com logs e IA); lapidação depois — não um MVP deliberadamente cortado.
- PREMISSA: sem deadline formal nem tamanho de time/orçamento informados nesta descoberta.
- Superfície: **desktop e CLI** na mesma capacidade (paridade de intenção; UX pode diferir).

## Atores

- **Desenvolvedor (usuário do time):** executa o fluxo diário no openbench; confirma ações; interpreta status/custos; aplica/aceita correções.
- **openbench (agente orquestrador):** mastiga o máximo possível (disparar fluxo, pollear/acompanhar runs, obter logs, propor e preparar correção); **nunca** executa efeitos colaterais sem confirmação do usuário.
- **GitHub Actions (sistema externo):** executa workflows; fonte de verdade de status, logs e minutos consumidos.
- PREMISSA: sem papéis distintos (dev vs lead) na v1 — todos do time usam as mesmas capacidades; permissões efetivas vêm da conta/`gh` autenticado no SO (e políticas da org).

## Requisitos funcionais

1. Orquestrar o ciclo: preparar/enviar mudanças (commit → push; PR quando aplicável) e **acompanhar a CI resultante**.
2. Acompanhar **todos** os workflows/runs relevantes do evento; UI/CLI com **filtros** (ex.: só falhos).
3. Exibir status de runs/jobs/steps e, sob demanda, **logs completos** do job/step (logs grandes: download sob demanda, não stream obrigatório de tudo).
4. Em falha: permitir **re-run** de workflow/job quando for falha flaky/infra (sem mudança de código).
5. Em falha de código: fluxo de **correção** — IA analisa o log, prepara a correção no workspace; usuário **confirma**; em seguida commit → push (nova execução), não apenas re-run do run antigo.
6. Após sugestão da IA, o openbench deve **fazer o máximo** (aplicar mudanças no workspace, preparar commit/push); ao usuário resta **sempre confirmar**.
7. Push em branch basta para acompanhar CI; **alertar explicitamente** se o destino for `main`/`master` (ou branch default protegida equivalente) de que o push **vai disparar CI automaticamente**.
8. Suportar acompanhamento dos gatilhos relevantes: push, pull_request e `workflow_dispatch` (disparo manual quando o workflow permitir).
9. Exibir **sempre em evidência** o consumo/custo em **minutos de GitHub Actions** (e impacto de re-runs), na medida em que a API/`gh` e permissões permitirem.
10. **Nunca** expor chaves/secrets/tokens (nem ecoar em UI, CLI, chat, telemetria ou prompts de forma insegura); redigir conteúdo sensível vindo de logs quando detectável.
11. Paridade de capacidade entre **app desktop** e **CLI** (`ob`).
12. Funcionar com **github.com** e **GitHub Enterprise Server** (auth/host via modelo compatível com `gh`).

## Requisitos não funcionais

- Disponibilidade da feature depende de rede + `gh` autenticado (e host Enterprise configurado quando for o caso).
- PREMISSA offline: sem rede, ações de CI falham com mensagem clara; opcionalmente mostrar último status em cache se existir.
- Logs grandes: **sob demanda** (não pré-carregar tudo).
- PREMISSA de latência de status: atualização próxima de real-time via watch/poll do `gh`/API; atraso na ordem de segundos (alvo prático ≤ ~30s para refletir mudança de estado) é aceitável.
- Segurança: zero tolerância a vazamento de secrets; logs tratados como potencialmente sensíveis.
- Transparência de custo: minutos de Actions sempre visíveis no fluxo de CI (incluindo antes/depois de re-run quando possível).
- Escopo de dados: apenas o repositório do projeto atual na sessão.

## Regras de negócio

- Confirmação do usuário é obrigatória para efeitos colaterais: push, abertura/atualização de PR, re-run, commit após correção IA, disparo manual de workflow.
- Dois caminhos pós-falha são válidos e devem coexistir:
  - **Código:** corrigir → commit → push → **novo** run.
  - **Flaky/infra:** re-run do workflow/job sem novo commit.
- Alerta obrigatório ao operar sobre a branch default (`main`/`master` ou equivalente) porque o push dispara CI.
- openbench **não** cria/edita YAML de workflow, **não** gerencia secrets/environments/approvals de deploy, **não** roda Actions localmente como substituto do GitHub — foco é orquestrar + observar + reagir ao CI existente.
- Custo em minutos de Actions é informação de primeira classe (sempre em evidência), não detalhe escondido.

## Fluxos principais

1. **Happy path:** usuário confirma commit → push (alerta se main) → [opcional PR] → openbench acompanha todos os workflows → sucesso → resumo + minutos consumidos.
2. **Falha de código:** CI falha → usuário inspeciona logs (sob demanda) → IA propõe/aplica correção no workspace → usuário confirma → commit → push → novo acompanhamento de CI.
3. **Falha flaky/infra:** CI falha → usuário confirma re-run → openbench re-executa workflow/job → acompanha resultado → minutos atualizados em evidência.
4. **Disparo manual:** usuário escolhe workflow com `workflow_dispatch` → confirma → openbench dispara → acompanha como nos demais.
5. **Filtro:** lista de runs/workflows com filtro (ex.: somente falhos) sem perder a visão do conjunto.
6. **Erro de pré-requisito:** `gh` ausente/não autenticado, sem remote, sem permissão, Enterprise mal configurado → onboarding/mensagem clara, sem falha opaca.
7. **Conteúdo sensível em log:** redacao/bloqueio de exibição de segredos; não enviar material sensível bruto ao provider de IA sem sanitização.

## Integrações externas

- **GitHub CLI (`gh`) / GitHub API:** listar/acompanhar runs, jobs, logs, re-run, workflow_dispatch; auth existente no SO; hosts github.com e Enterprise.
- **GitHub Actions:** execução real dos workflows; billing/minutos (API de billing/usage conforme permissões da conta/org).
- **Provider de IA já usado pelo openbench:** análise de logs e preparação de correção (com sanitização).
- **Git local:** commit/push no repo do projeto aberto.

## Restrições e premissas

- PREMISSA: auth e operações via modelo centrado em `gh` (e API que o `gh` já usa); não há obrigação de OAuth próprio na primeira entrega se `gh` cobrir os casos.
- PREMISSA: escopo = repo do projeto aberto apenas.
- PREMISSA: “nascer completo” significa o conjunto de RFs acima na primeira entrega utilizável pelo time; polish de UX pode evoluir depois.
- PREMISSA: métrica de minutos pode depender de permissões de org/billing; se indisponível, mostrar estado “indisponível + motivo” sem esconder o bloco de custo.
- Fora de escopo: self-hosted runner do openbench; editor de workflows; gestão de secrets/environments; substituir GitHub Actions.

## Riscos identificados

- **Escopo “completo de primeira”:** risco de atrasar a adoção diária; mitigação pendente na arquitetura (fases internas de entrega sem cortar RF da meta, ou fatias verticalmente testáveis).
- **Vazamento de secrets em logs/prompts:** mitigação obrigatória (redacao, allowlist de trechos enviados à IA, nunca logar tokens do `gh`).
- **Visibilidade de minutos/billing:** APIs e permissões variam (org vs user, Enterprise); risco de não conseguir número confiável — precisa fallback explícito.
- **GitHub Enterprise:** diferenças de host, SSO e versão de API; risco de paridade incompleta com github.com.
- **Push em main:** risco de disparar CI/produção acidentalmente — alerta obrigatório não elimina necessidade de proteções no GitHub (branch protection).
- **Dependência de `gh`/rede:** ponto único de falha para o fluxo diário; onboarding e erros claros são obrigatórios.
- **Custo de re-runs:** evidência de minutos deve desencorajar re-run cego; UX deve contrastar “re-run” vs “corrigir e novo push”.

## Lacunas / decisões pendentes

- Formato exato da evidência de minutos (por run, por dia, por repo, projeção pós re-run) — definir na arquitetura com o que a API permitir.
- Política detalhada de sanitização de logs antes de enviar à IA (regras e o que fazer quando o log for só secret-like).
- Comportamento preciso em branch protection / checks obrigatórios / environments com approval (só observar vs mensagem de “aguardando approval no GitHub”).
- Grau de paridade UX CLI vs desktop (mesmos comandos/verbos vs subset ergonômico na CLI).
- Deadline e capacidade do time para “nascer completo” — não definidos; tratados como premissa de intenção de produto.

---
**Status:** entendimento confirmado (2026-07-24). Arquitetura em [`docs/architecture/ci-github-actions.md`](../architecture/ci-github-actions.md).

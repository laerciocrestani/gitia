# Resumo do Entendimento — Índice de contexto do diff (desktop)

## Problema e objetivo

No app desktop, o usuário não vê de forma clara quando o diff que a IA usará no próximo passo está grande demais. Diffs grandes degradam a qualidade da mensagem de commit/PR (truncamento em `max_diff_bytes`), misturam áreas semânticas no mesmo commit e dificultam histórico/rollback.

O objetivo é exibir, no projeto aberto, um **índice de contexto com níveis** (ok / atenção / crítico) e um CTA **“Recomenda-se commit”**, incentivando commits menores e melhor aproveitamento da IA — sem bloquear o fluxo.

## Sistema existente

- Desktop Wails + React (`frontend/src/App.tsx`); dashboard já lista arquivos com `+/-` por arquivo (`ChangedFileView`).
- Commit IA: `DiffForCommit` / `DiffStatForCommit` (staged se houver, senão unstaged); patch truncado em `MaxDiffBytes` (default 120_000).
- PR IA: `DiffBranch(base)` (`base...HEAD`) + `LogOnBranch` — contexto acumulado da branch, **não** o working tree.
- Heurísticas já na CLI e pouco expostas no desktop: `BuildChangeSummary`, `ShouldSuggestSplit` / `ChangeAreasFromStat`, `DescribePreparedInput`; desktop usa `NopProgress` no commit.
- Untracked costuma ter churn `0` no numstat — risco de índice otimista.

**O que NÃO muda:** fluxo de commit/PR existente; limiar de truncamento da IA; não bloquear ações.

## Restrições organizacionais

- Feature incremental no app desktop já em migração.
- Sem deadline formal; priorizar reuso da lógica Go existente.

## Atores

- **Dev (usuário do projeto aberto):** vê saúde do contexto e decide commitar antes que a IA “engasgue”.

## Requisitos funcionais

1. Calcular índice do **próximo commit** = exatamente o que a IA veria em `SuggestCommit` (staged se houver, senão unstaged).
2. Exibir gauge/slider visual no projeto aberto (card de arquivos alterados / status do projeto).
3. Níveis discretos: **ok** / **atenção** / **crítico**.
4. Métrica **composta**: linhas (+/−), quantidade de arquivos, áreas distintas (split) e proximidade de `MaxDiffBytes`.
5. Em atenção/crítico: CTA **“Recomenda-se commit”** que abre o fluxo de commit (não bloqueia).
6. **PR:** fora desta entrega (iteração futura).

## Requisitos não funcionais

- Cálculo barato o bastante para acompanhar refresh do dashboard (sem chamar IA).
- Texto claro em pt-BR; tom de recomendação, não de erro fatal.
- Consistente com `max_diff_bytes` da config do usuário.

## Regras de negócio

- Índice de **commit** ≠ índice de **PR** (inputs diferentes).
- Working tree limpo → índice de commit em estado neutro/vazio (sem CTA de commit).
- Branch com commits ahead e working tree limpo → commit ok; PR pode estar em atenção/crítico.
- Nunca impedir Commit / Push / PR por causa do índice.
- Untracked deve entrar no cálculo de forma honesta (não contar como zero churn).

## Fluxos principais

1. Usuário abre/atualiza projeto → backend calcula `ContextIndex` do próximo commit → UI mostra nível + barra.
2. Nível atenção/crítico → aparece CTA “Recomenda-se commit” → abre preview/fluxo de commit.
3. (Futuro) Índice de PR vs base — fora do escopo atual.

## Integrações externas

- Nenhuma nova: Git local + config `max_diff_bytes` já usada pelos providers de IA.

## Restrições e premissas

- **CONFIRMADO:** v1 entrega **somente** índice de **commit** no dashboard + CTA “Recomenda-se commit”.
- **CONFIRMADO:** índice de **PR** fica para iteração futura.
- **PREMISSA:** limiares iniciais compostos (ajustáveis depois), alinhados a `MaxDiffBytes` e à heurística de split existente.
- CTA só quando há mudanças commitáveis (índice relevante).

## Riscos identificados

- **Untracked sem numstat:** mitigação — contar linhas do arquivo ou marcar peso mínimo por arquivo novo.
- **Falso positivo em refactor mecânico:** mitigação — níveis + não bloquear; áreas distintas pesam mais que só linhas.
- **Custo de `git diff` completo a cada refresh:** mitigação — preferir `--stat`/`--numstat`/shortstat + tamanho estimado; bytes exatos do patch só se necessário ou sob demanda.

## Lacunas / decisões pendentes

- Nenhuma crítica; copy dos níveis definida na implementação.

---

**Confirmado pelo usuário (2026-07-20): só commit agora; PR depois.**
`)
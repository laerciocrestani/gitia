# ADR-002 — Frontend React + Vite + TypeScript

## Status

Aceito (default; revisável se o time preferir Svelte)

## Contexto

Wails embute UI web. Precisamos de dashboard, modais de commit/PR, onboarding e multi-abas com boa DX e tipagem junto aos bindings gerados.

## Decisão

**React 18+ + Vite + TypeScript** no diretório `frontend/`.

Estilo: CSS moderno próprio (variáveis), sem obrigar design system pesado na v1.

## Consequências

- **+** Ecossistema amplo; templates Wails comuns; hiring/familiaridade.
- **+** TS alinha com bindings tipados.
- **−** Bundle maior que Svelte/Solid (aceitável em WebView desktop).

## Alternativas

- **Svelte:** mais leve; ótimo se preferência explícita.
- **Vue:** equivalente ao React em encaixe; sem vantagem clara aqui.
EOF
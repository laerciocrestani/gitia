# ADR-001 — Wails v3 como shell desktop

## Status

Aceito

## Contexto

Precisamos de app multi-OS com **system tray**, janela WebView, bindings Go↔UI e **auto-update**. O core já é Go. Wails v2 é estável; v3 está em **alpha**, mas oferece SystemTray e Updater de primeira classe alinhados aos requisitos.

## Decisão

Usar **Wails v3** (`github.com/wailsapp/wails/v3`).

## Consequências

- **+** Tray e updater nativos; API moderna de app/janelas.
- **+** Reuso direto do domínio Go.
- **−** Risco de breaking changes alpha; exige pin de versão e CI multi-OS cedo.
- **−** Menos exemplos maduros que v2.

## Alternativas rejeitadas

- **Wails v2:** estável, mas tray/updater menos alinhados ao desenho v3; migrar depois custaria.
- **Tauri:** exigiria ponte Rust↔Go ou reescrita.
- **Electron:** peso e dual runtime desnecessários com core Go.
EOF
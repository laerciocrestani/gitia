# Runbook — entregar mudanças no app desktop

Guia curto: **o que fazer depois de alterar código** para testar localmente e/ou publicar update pelo próprio app.

Repo: `laerciocrestani/openbench`  
Updater: GitHub Releases + `manifest.json` assinado  
Versão no binário: `-ldflags` via `APP_VERSION` (obrigatório em release)

---

## Quando usar cada caminho

| Situação | O que fazer |
|----------|-------------|
| Iterar UI / Go rápido | `wails3 dev` — **não** precisa release |
| Validar `.app` como o usuário final | `wails3 package APP_VERSION=…` e abrir `bin/openbench.app` |
| Outros Macs / update automático | Commit → package → **GitHub Release** (versão **maior**) |

Regra: **só o GitHub Release atualiza o app já instalado**. PR/merge sozinho não basta.

---

## 1. Desenvolvimento local (sem release)

Na raiz do repo:

```bash
wails3 dev
```

- Hot reload do frontend; rebuild Go ao salvar.
- Versão pode aparecer como `0.1.N` (git) ou o que estiver no ambiente — ok para dev.
- Feche o `.app` antigo se estiver rodando (evita misturar processos).

---

## 2. Versão — escolha o próximo número

1. Veja o último release: https://github.com/laerciocrestani/openbench/releases  
2. Suba o patch (ex.: `0.2.1` → `0.2.2`).  
3. Use o **mesmo** número em três lugares:

| Onde | Exemplo |
|------|---------|
| Build | `APP_VERSION=0.2.2` |
| Tag do release | `v0.2.2` |
| Manifest | `-version 0.2.2` |

Sem `APP_VERSION` no package, o binário tende a cair em **`0.1.0`** fora do repo → updater fica confuso.

Opcional (Finder / About): alinhar `build/config.yml` e `build/darwin/Info.plist` (`CFBundleShortVersionString` / `CFBundleVersion`).

---

## 3. Commit do código

```bash
git checkout -b feature/minha-mudanca
# …alterações…
git add -A
git status   # NUNCA stagear build/updater/updater.key
git commit -m "feat(desktop): descrição curta"
git push -u origin HEAD
gh pr create   # se quiser review
```

Chave privada (`build/updater/updater.key`) é gitignored — não commitá-la.

---

## 4. Build de produção (macOS arm64)

```bash
# Defina a versão do release
export APP_VERSION=0.2.2

wails3 package APP_VERSION=$APP_VERSION
```

Saída:

- `bin/openbench` — binário
- `bin/openbench.app` — bundle para abrir / zipar

Conferir versão embutida:

```bash
strings bin/openbench.app/Contents/MacOS/openbench | grep -E '^[0-9]+\.[0-9]+\.[0-9]+$' | head
plutil -p bin/openbench.app/Contents/Info.plist | grep -i version
```

Deve aparecer `0.2.2` (não `0.1.0`).

Teste local antes do release:

```bash
open bin/openbench.app
```

---

## 5. Manifest assinado + zip

Arquivo do zip **precisa** ter `darwin` e `arm64` no nome (matcher do updater GitHub).

```bash
export APP_VERSION=0.2.2
export TAG=v$APP_VERSION

rm -rf bin/updates && mkdir -p bin/updates
ditto -c -k --keepParent bin/openbench.app "bin/updates/openbench-darwin-arm64.zip"

cat > /tmp/ob-notes.md <<EOF
## openbench $APP_VERSION

- (escreva 1–3 bullets do que mudou)

**macOS:** baixe \`openbench-darwin-arm64.zip\`, extraia e abra \`openbench.app\`.
EOF

wails3 updater manifest \
  -version "$APP_VERSION" \
  -channel stable \
  -key build/updater/updater.key \
  -name "openbench $APP_VERSION" \
  -notes-file /tmp/ob-notes.md \
  -url-prefix "https://github.com/laerciocrestani/openbench/releases/download/$TAG" \
  -output bin/updates/manifest.json \
  bin/updates/openbench-darwin-arm64.zip

wails3 updater verify \
  -manifest bin/updates/manifest.json \
  -publickey build/updater/updater.key.pub
```

Precisa existir `build/updater/updater.key` (local). A `.pub` já vai no binário.

---

## 6. Publicar GitHub Release

```bash
export APP_VERSION=0.2.2
export TAG=v$APP_VERSION

gh release create "$TAG" \
  --title "openbench $APP_VERSION" \
  --notes-file /tmp/ob-notes.md \
  bin/updates/openbench-darwin-arm64.zip \
  bin/updates/manifest.json
```

Assets obrigatórios:

1. `openbench-darwin-arm64.zip`
2. `manifest.json` (URL `…/releases/latest/download/manifest.json`)

---

## 7. Atualizar o app já instalado

No Mac que roda o openbench:

1. **Settings → Atualizações → Verificar atualizações**  
   (ou esperar a checagem a cada 6h)
2. **Atualizar agora** → reiniciar quando pedir
3. Confirmar que a UI mostra a nova versão (ex.: `v0.2.2`)

Se ainda mostrar `0.1.0` depois do update: o release foi buildado **sem** `APP_VERSION` — refaça o package e publique um patch maior.

---

## Checklist rápido (copiar)

```text
[ ] Código commitado / PR
[ ] APP_VERSION = próximo semver (maior que o último release)
[ ] wails3 package APP_VERSION=…
[ ] strings/plutil mostram a versão certa
[ ] open bin/openbench.app — smoke test
[ ] zip openbench-darwin-arm64.zip + manifest assinado + verify OK
[ ] gh release create vX.Y.Z com zip + manifest.json
[ ] No app: Verificar atualizações → instalou → versão nova na UI
```

---

## Script one-shot (após o package)

Com `bin/openbench.app` já gerado e `APP_VERSION` definido:

```bash
export APP_VERSION=0.2.2
export TAG=v$APP_VERSION

rm -rf bin/updates && mkdir -p bin/updates
ditto -c -k --keepParent bin/openbench.app bin/updates/openbench-darwin-arm64.zip

cat > /tmp/ob-notes.md <<EOF
## openbench $APP_VERSION

- Descreva a mudança

**macOS:** baixe \`openbench-darwin-arm64.zip\`, extraia e abra \`openbench.app\`.
EOF

wails3 updater manifest \
  -version "$APP_VERSION" \
  -channel stable \
  -key build/updater/updater.key \
  -name "openbench $APP_VERSION" \
  -notes-file /tmp/ob-notes.md \
  -url-prefix "https://github.com/laerciocrestani/openbench/releases/download/$TAG" \
  -output bin/updates/manifest.json \
  bin/updates/openbench-darwin-arm64.zip

wails3 updater verify \
  -manifest bin/updates/manifest.json \
  -publickey build/updater/updater.key.pub

gh release create "$TAG" \
  --title "openbench $APP_VERSION" \
  --notes-file /tmp/ob-notes.md \
  bin/updates/openbench-darwin-arm64.zip \
  bin/updates/manifest.json
```

---

## Chaves do updater

| Arquivo | Git? | Uso |
|---------|------|-----|
| `build/updater/updater.key.pub` | sim | Embutida no app (`//go:embed`) |
| `build/updater/updater.key` | **não** | Assina o `manifest.json` |

Gerar de novo (só se necessário — invalida releases antigos assinados com a chave anterior):

```bash
wails3 updater genkey -out build/updater/updater.key
```

---

## Problemas comuns

| Sintoma | Causa | Ação |
|---------|--------|------|
| Update instala mas UI fica `0.1.0` | Package sem `APP_VERSION` | Rebuild com `APP_VERSION` + novo release |
| “Já está na versão mais recente” mas falta feature | Release com versão ≤ app atual | Subir semver |
| Check não acha artefato | Zip sem `darwin`/`arm64` no nome | Renomear para `openbench-darwin-arm64.zip` |
| Signature failed | Manifest com outra chave / `.pub` antiga no binário | Assinar com a `updater.key` atual; app precisa ter a `.pub` correspondente |
| Mudança não aparece sem release | Esperado | Use `wails3 dev` ou abra o `bin/openbench.app` local |

---

## Referências

- Arquitetura: [app-desktop-wails.md](architecture/app-desktop-wails.md)
- ADR updater: [adr/004-auto-update.md](architecture/adr/004-auto-update.md)
- Build geral: `AGENTS.md` (raiz)

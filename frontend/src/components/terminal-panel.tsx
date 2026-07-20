import { useEffect, useRef } from "react"
import { Terminal } from "@xterm/xterm"
import { FitAddon } from "@xterm/addon-fit"
import { WebLinksAddon } from "@xterm/addon-web-links"
import "@xterm/xterm/css/xterm.css"

import { AppService } from "../../bindings/github.com/laerciocrestani/openbench"
import { Events } from "@wailsio/runtime"
import { Button } from "@/components/ui/button"
import { RotateCcw, TerminalSquare } from "lucide-react"

function decodeBase64(b64: string): string {
  const bin = atob(b64)
  const bytes = new Uint8Array(bin.length)
  for (let i = 0; i < bin.length; i++) bytes[i] = bin.charCodeAt(i)
  return new TextDecoder("utf-8", { fatal: false }).decode(bytes)
}

export function TerminalPanel({
  projectPath,
  visible,
}: {
  projectPath: string | null
  visible: boolean
}) {
  const hostRef = useRef<HTMLDivElement>(null)
  const termRef = useRef<Terminal | null>(null)
  const fitRef = useRef<FitAddon | null>(null)
  const startedFor = useRef<string | null>(null)

  // Keep the host mounted so xterm can attach; overlay when no project.
  useEffect(() => {
    const host = hostRef.current
    if (!host || termRef.current) return

    const term = new Terminal({
      cursorBlink: true,
      cursorStyle: "block",
      disableStdin: false,
      fontSize: 12,
      fontFamily:
        'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace',
      theme: {
        background: "#0c0c0c",
        foreground: "#e5e5e5",
        cursor: "#e5e5e5",
        selectionBackground: "#264f78",
      },
      allowProposedApi: true,
    })
    const fit = new FitAddon()
    term.loadAddon(fit)
    term.loadAddon(new WebLinksAddon())
    term.open(host)
    fit.fit()

    term.onData((data) => {
      void AppService.TerminalWrite(data).catch(() => {
        /* PTY still starting / already closed */
      })
    })

    termRef.current = term
    fitRef.current = fit

    const onResize = () => {
      if (!fitRef.current || !termRef.current) return
      fitRef.current.fit()
      const dims = fitRef.current.proposeDimensions()
      if (dims) {
        void AppService.TerminalResize(dims.cols, dims.rows).catch(() => {})
      }
    }
    const ro = new ResizeObserver(() => onResize())
    ro.observe(host)

    // Click / focus so keystrokes reach xterm's textarea.
    const focusTerm = () => term.focus()
    host.addEventListener("mousedown", focusTerm)

    return () => {
      host.removeEventListener("mousedown", focusTerm)
      ro.disconnect()
      term.dispose()
      termRef.current = null
      fitRef.current = null
      startedFor.current = null
    }
  }, [])

  useEffect(() => {
    const offData = Events.On("terminal:data", (ev) => {
      const raw = String((ev?.data as unknown) ?? "")
      if (!raw || !termRef.current) return
      try {
        termRef.current.write(decodeBase64(raw))
      } catch {
        /* ignore bad chunk */
      }
    })
    const offExit = Events.On("terminal:exit", () => {
      startedFor.current = null
      termRef.current?.writeln("\r\n\x1b[90m[processo encerrado]\x1b[0m")
    })
    return () => {
      offData()
      offExit()
    }
  }, [])

  useEffect(() => {
    if (!visible || !projectPath || !termRef.current || !fitRef.current) return
    if (startedFor.current === projectPath) {
      fitRef.current.fit()
      termRef.current.focus()
      return
    }

    const start = async () => {
      fitRef.current?.fit()
      const dims = fitRef.current?.proposeDimensions()
      const cols = dims?.cols ?? 80
      const rows = dims?.rows ?? 24
      termRef.current?.reset()
      try {
        await AppService.TerminalStart(cols, rows)
        startedFor.current = projectPath
        // Defer focus until after layout / WebView paint.
        requestAnimationFrame(() => termRef.current?.focus())
      } catch (e) {
        const msg = e instanceof Error ? e.message : String(e)
        termRef.current?.writeln(`\x1b[31m${msg}\x1b[0m`)
        startedFor.current = null
      }
    }
    void start()
  }, [visible, projectPath])

  const restart = async () => {
    if (!projectPath || !fitRef.current || !termRef.current) return
    fitRef.current.fit()
    const dims = fitRef.current.proposeDimensions()
    termRef.current.reset()
    try {
      await AppService.TerminalRestart(dims?.cols ?? 80, dims?.rows ?? 24)
      startedFor.current = projectPath
      requestAnimationFrame(() => termRef.current?.focus())
    } catch (e) {
      const msg = e instanceof Error ? e.message : String(e)
      termRef.current.writeln(`\x1b[31m${msg}\x1b[0m`)
    }
  }

  return (
    <div className="relative flex h-full min-h-0 flex-col">
      <div className="flex shrink-0 items-center gap-2 border-b px-2 py-1.5">
        <TerminalSquare className="size-3.5 text-muted-foreground" />
        <span
          className="min-w-0 flex-1 truncate font-mono text-[11px] text-muted-foreground"
          title={projectPath ?? undefined}
        >
          {projectPath ?? "sem projeto"}
        </span>
        <Button
          variant="ghost"
          size="icon-xs"
          title="Reiniciar shell"
          disabled={!projectPath}
          onClick={() => void restart()}
        >
          <RotateCcw />
        </Button>
      </div>

      <div
        ref={hostRef}
        className="min-h-0 flex-1 cursor-text overflow-hidden bg-[#0c0c0c] p-1 [&_.xterm]:h-full [&_.xterm-screen]:h-full [&_textarea]:pointer-events-auto"
        tabIndex={0}
        onFocus={() => termRef.current?.focus()}
      />

      {!projectPath && (
        <div className="absolute inset-0 top-10 flex flex-col items-center justify-center gap-2 bg-[#0c0c0c]/95 p-6 text-center text-sm text-muted-foreground">
          <TerminalSquare className="size-8 opacity-40" />
          <p>Abra um projeto para iniciar o shell no root do repositório.</p>
        </div>
      )}
    </div>
  )
}

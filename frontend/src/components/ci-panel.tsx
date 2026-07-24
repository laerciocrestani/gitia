import { useCallback, useEffect, useState } from "react"

import { AppService } from "../../bindings/github.com/laerciocrestani/openbench"
import type {
  CILogView,
  CIRunDetailView,
  CIRunView,
  CIStatusView,
  CIUsageView,
} from "../../bindings/github.com/laerciocrestani/openbench/internal/desktop"

import { Alert, AlertDescription } from "@/components/ui/alert"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { Label } from "@/components/ui/label"
import { ScrollArea } from "@/components/ui/scroll-area"
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { cn } from "@/lib/utils"
import {
  CircleAlert,
  ExternalLink,
  FileText,
  Loader2,
  RefreshCw,
  Workflow,
} from "lucide-react"

function errText(e: unknown): string {
  if (e instanceof Error) return e.message
  if (typeof e === "string") return e
  try {
    return JSON.stringify(e)
  } catch {
    return String(e)
  }
}

function shortSha(sha: string | undefined): string {
  if (!sha) return "—"
  return sha.length > 7 ? sha.slice(0, 7) : sha
}

function conclusionVariant(
  status: string,
  conclusion: string,
  failed: boolean,
): "default" | "secondary" | "outline" | "destructive" {
  if (failed) return "destructive"
  const c = (conclusion || status || "").toLowerCase()
  if (c === "success" || c === "completed") return "default"
  if (c === "in_progress" || c === "queued" || c === "pending" || c === "waiting") {
    return "outline"
  }
  return "secondary"
}

function labelOf(run: CIRunView): string {
  return run.conclusion || run.status || "—"
}

function ActionsUsageBanner({ usage }: { usage: CIUsageView | null | undefined }) {
  if (!usage) return null
  const mins =
    usage.runMinutes ?? usage.windowMinutes ?? usage.orgUsedMinutes ?? null
  return (
    <Alert
      className={cn(
        "shrink-0",
        usage.state === "unavailable" && "border-destructive/40",
      )}
    >
      <CircleAlert className="size-4" />
      <AlertDescription className="text-xs leading-relaxed">
        <span className="font-medium">Actions · {usage.state}</span>
        {mins != null && (
          <span className="ml-2 font-mono tabular-nums">~{mins} min</span>
        )}
        {usage.orgUsedMinutes != null && usage.orgIncludedMinutes != null && (
          <span className="ml-2 font-mono tabular-nums">
            org {usage.orgUsedMinutes}/{usage.orgIncludedMinutes}
            {usage.orgRemainingMinutes != null &&
              ` · resta ${usage.orgRemainingMinutes}`}
          </span>
        )}
        {usage.message ? (
          <span className="mt-0.5 block text-muted-foreground">{usage.message}</span>
        ) : null}
      </AlertDescription>
    </Alert>
  )
}

export function CIPanel({
  open,
  onOpenChange,
  projectPath,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  projectPath: string
}) {
  const [failedOnly, setFailedOnly] = useState(false)
  const [status, setStatus] = useState<CIStatusView | null>(null)
  const [detail, setDetail] = useState<CIRunDetailView | null>(null)
  const [log, setLog] = useState<CILogView | null>(null)
  const [loading, setLoading] = useState(false)
  const [detailLoading, setDetailLoading] = useState(false)
  const [logLoading, setLogLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const refresh = useCallback(async () => {
    if (!projectPath) return
    setLoading(true)
    setError(null)
    try {
      const next = await AppService.CIStatus(failedOnly, 20)
      setStatus(next ?? null)
      setDetail(null)
      setLog(null)
    } catch (e) {
      setError(errText(e))
      setStatus(null)
    } finally {
      setLoading(false)
    }
  }, [failedOnly, projectPath])

  useEffect(() => {
    if (!open || !projectPath) return
    void refresh()
  }, [open, projectPath, refresh])

  const openRun = async (id: number) => {
    setDetailLoading(true)
    setError(null)
    setLog(null)
    try {
      const next = await AppService.CIRunDetail(id)
      setDetail(next ?? null)
    } catch (e) {
      setError(errText(e))
    } finally {
      setDetailLoading(false)
    }
  }

  const loadLog = async (runId: number, jobId: number, onlyFailed: boolean) => {
    setLogLoading(true)
    setError(null)
    try {
      const next = await AppService.CILog(runId, jobId, onlyFailed)
      setLog(next ?? null)
    } catch (e) {
      setError(errText(e))
    } finally {
      setLogLoading(false)
    }
  }

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent
        side="right"
        className="flex w-full flex-col gap-0 overflow-hidden p-0 sm:max-w-xl"
      >
        <SheetHeader className="shrink-0 border-b px-4 py-3">
          <SheetTitle className="flex items-center gap-2 text-base">
            <Workflow className="size-4" />
            CI / GitHub Actions
          </SheetTitle>
          <p className="text-xs text-muted-foreground">
            {status?.branch
              ? `branch ${status.branch} · HEAD ${shortSha(status.headSha)}`
              : "Observar runs do projeto aberto"}
          </p>
        </SheetHeader>

        <div className="flex shrink-0 flex-wrap items-center gap-3 border-b px-4 py-2">
          <div className="flex items-center gap-2">
            <Checkbox
              id="ci-failed-only"
              checked={failedOnly}
              onCheckedChange={(v) => setFailedOnly(v === true)}
            />
            <Label htmlFor="ci-failed-only" className="text-xs font-normal">
              Só falhos
            </Label>
          </div>
          <Button
            size="sm"
            variant="outline"
            onClick={() => void refresh()}
            disabled={loading}
          >
            {loading ? (
              <Loader2 className="size-3.5 animate-spin" />
            ) : (
              <RefreshCw className="size-3.5" />
            )}
            Atualizar
          </Button>
          {detail && (
            <Button
              size="sm"
              variant="ghost"
              onClick={() => setDetail(null)}
              className="ml-auto"
            >
              ← Lista
            </Button>
          )}
        </div>

        <div className="shrink-0 space-y-2 px-4 py-2">
          <ActionsUsageBanner usage={detail?.usage ?? status?.usage} />
          {error && (
            <Alert variant="destructive">
              <AlertDescription className="text-xs">{error}</AlertDescription>
            </Alert>
          )}
        </div>

        <ScrollArea className="min-h-0 flex-1 px-4 pb-4">
          {detailLoading && (
            <div className="flex items-center gap-2 py-8 text-sm text-muted-foreground">
              <Loader2 className="size-4 animate-spin" />
              Carregando run…
            </div>
          )}

          {!detailLoading && detail && (
            <div className="space-y-4">
              <div className="flex flex-wrap items-center gap-2">
                <Badge
                  variant={conclusionVariant(
                    detail.run.status,
                    detail.run.conclusion,
                    detail.run.failed,
                  )}
                >
                  {labelOf(detail.run)}
                </Badge>
                <span className="text-sm font-medium">{detail.run.name}</span>
                {detail.run.url && (
                  <a
                    href={detail.run.url}
                    target="_blank"
                    rel="noreferrer"
                    className="inline-flex items-center gap-1 text-xs text-muted-foreground hover:underline"
                  >
                    GitHub <ExternalLink className="size-3" />
                  </a>
                )}
              </div>
              <p className="text-xs text-muted-foreground">
                {detail.run.workflowName} · {detail.run.event} ·{" "}
                {shortSha(detail.run.headSha)}
              </p>
              <div className="flex flex-wrap gap-2">
                <Button
                  size="sm"
                  variant="outline"
                  disabled={logLoading}
                  onClick={() => void loadLog(detail.run.id, 0, true)}
                >
                  {logLoading ? (
                    <Loader2 className="size-3.5 animate-spin" />
                  ) : (
                    <FileText className="size-3.5" />
                  )}
                  Logs falhos
                </Button>
                <Button
                  size="sm"
                  variant="ghost"
                  disabled={logLoading}
                  onClick={() => void loadLog(detail.run.id, 0, false)}
                >
                  Log completo
                </Button>
                {log && (
                  <Button
                    size="sm"
                    variant="ghost"
                    onClick={() => setLog(null)}
                  >
                    Fechar log
                  </Button>
                )}
              </div>

              {log && (
                <div className="space-y-2 rounded-lg border p-3">
                  <div className="flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
                    <span className="font-medium text-foreground">
                      Log redigido
                    </span>
                    <span>
                      {log.bytes} bytes
                      {log.truncated ? " · truncado" : ""}
                      {log.jobId ? ` · job ${log.jobId}` : ""}
                      {log.failedOnly ? " · só falhos" : ""}
                    </span>
                  </div>
                  {log.message && (
                    <p className="text-xs text-muted-foreground">{log.message}</p>
                  )}
                  <pre className="max-h-[50vh] overflow-auto whitespace-pre-wrap break-all rounded-md bg-muted/50 p-2 font-mono text-[11px] leading-relaxed">
                    {log.redactedText || "(vazio)"}
                  </pre>
                </div>
              )}

              {(detail.jobs ?? []).map((job) => (
                <div key={job.id} className="rounded-lg border p-3">
                  <div className="mb-2 flex flex-wrap items-center gap-2">
                    <Badge
                      variant={conclusionVariant(
                        job.status,
                        job.conclusion,
                        job.failed,
                      )}
                    >
                      {job.conclusion || job.status}
                    </Badge>
                    <span className="text-sm font-medium">{job.name}</span>
                    <Button
                      size="sm"
                      variant="ghost"
                      className="ml-auto h-7 px-2 text-xs"
                      disabled={logLoading}
                      onClick={() =>
                        void loadLog(detail.run.id, job.id, job.failed)
                      }
                    >
                      <FileText className="size-3" />
                      Log
                    </Button>
                  </div>
                  <ul className="space-y-1 text-xs">
                    {(job.steps ?? []).map((step) => (
                      <li
                        key={`${job.id}-${step.number}`}
                        className={cn(
                          "flex justify-between gap-2 font-mono",
                          step.failed && "text-destructive",
                        )}
                      >
                        <span>
                          {step.number}. {step.name}
                        </span>
                        <span className="text-muted-foreground">
                          {step.conclusion || step.status}
                        </span>
                      </li>
                    ))}
                  </ul>
                </div>
              ))}
              {(detail.jobs ?? []).length === 0 && (
                <p className="text-sm text-muted-foreground">Run sem jobs.</p>
              )}
            </div>
          )}

          {!detailLoading && !detail && (
            <>
              {loading && !status ? (
                <div className="flex items-center gap-2 py-8 text-sm text-muted-foreground">
                  <Loader2 className="size-4 animate-spin" />
                  Listando runs…
                </div>
              ) : (status?.runs ?? []).length === 0 ? (
                <p className="py-8 text-sm text-muted-foreground">
                  Nenhum workflow run encontrado
                  {failedOnly ? " com falha" : ""} nesta branch.
                </p>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Status</TableHead>
                      <TableHead>Run</TableHead>
                      <TableHead>Event</TableHead>
                      <TableHead>SHA</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {(status?.runs ?? []).map((run) => (
                      <TableRow
                        key={run.id}
                        className="cursor-pointer"
                        onClick={() => void openRun(run.id)}
                      >
                        <TableCell>
                          <Badge
                            variant={conclusionVariant(
                              run.status,
                              run.conclusion,
                              run.failed,
                            )}
                          >
                            {labelOf(run)}
                          </Badge>
                        </TableCell>
                        <TableCell className="max-w-[12rem] truncate text-sm">
                          {run.name || run.workflowName}
                        </TableCell>
                        <TableCell className="text-xs text-muted-foreground">
                          {run.event}
                        </TableCell>
                        <TableCell className="font-mono text-xs">
                          {shortSha(run.headSha)}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              )}
            </>
          )}
        </ScrollArea>
      </SheetContent>
    </Sheet>
  )
}

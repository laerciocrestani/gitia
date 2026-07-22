import { Bot, MessageSquare, X } from "lucide-react"

import { ProjectChatPanel } from "@/components/project-chat-panel"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"

export function FloatingChat({
  projectPath,
  open,
  onOpenChange,
}: {
  projectPath: string | null
  open: boolean
  onOpenChange: (open: boolean) => void
}) {
  if (!projectPath) return null

  return (
    <>
      {open ? (
        <div
          className={cn(
            "fixed right-4 bottom-4 z-50 flex w-[min(100vw-2rem,26rem)] flex-col overflow-hidden",
            "h-[min(72vh,40rem)] rounded-xl border border-border bg-popover text-popover-foreground shadow-xl",
          )}
        >
          <div className="flex shrink-0 items-center gap-2 border-b px-3 py-2">
            <Bot className="size-4 text-muted-foreground" />
            <span className="text-sm font-medium">Chat IA</span>
            <span className="text-[11px] text-muted-foreground">flutuante</span>
            <Button
              variant="ghost"
              size="icon-xs"
              className="ml-auto"
              title="Fechar chat"
              onClick={() => onOpenChange(false)}
            >
              <X />
            </Button>
          </div>
          <div className="min-h-0 flex-1">
            <ProjectChatPanel
              projectPath={projectPath}
              visible={open}
              className="bg-popover"
              hideChrome
            />
          </div>
        </div>
      ) : (
        <Button
          size="lg"
          className="fixed right-4 bottom-4 z-50 h-12 gap-2 rounded-full px-4 shadow-lg"
          onClick={() => onOpenChange(true)}
          title="Abrir chat com a IA"
        >
          <MessageSquare className="size-4" />
          Chat IA
        </Button>
      )}
    </>
  )
}

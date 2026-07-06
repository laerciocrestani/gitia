package tui

import "strings"

func helpContent() string {
	lines := []string{
		"Atalhos do dashboard",
		"",
		"  c       Commit com IA (preview → Enter confirma)",
		"  p       Push para origin",
		"  P       Criar Pull Request com IA",
		"  d       Ver diff (working tree ou branch)",
		"  s       Sync com origin (quando behind)",
		"  o       Abrir PR no browser",
		"  u       Relatório de uso/custo de IA",
		"  r       Atualizar dashboard",
		"  ?       Esta ajuda",
		"  q       Sair",
		"",
		"Na tela de diff/report",
		"  ↑↓      Scroll",
		"  esc     Voltar",
		"",
		"No modal de PR",
		"  d       Alternar draft",
		"  Enter   Confirmar",
		"  esc     Cancelar",
		"",
		"Preferências em config.yaml",
		"  interactive_ui   TUI ao rodar gitai (padrão: sim)",
		"  ui_color         Cores na CLI e TUI (padrão: sim)",
		"",
		"Variáveis de ambiente (sobrescrevem config)",
		"  GITAI_NO_UI=1   Força overview CLI em vez da TUI",
		"  NO_COLOR=1      Sem cores (convenção Unix; ver no-color.org)",
		"  CI=1            Sem TUI nem cores",
	}

	var b strings.Builder
	b.WriteString(styleSection.Render("Ajuda"))
	b.WriteString("\n\n")
	for _, line := range lines {
		if line == "" {
			b.WriteString("\n")
			continue
		}
		if !strings.HasPrefix(line, "  ") {
			b.WriteString(styleTitle.Render(line))
		} else {
			b.WriteString(styleHint.Render(line))
		}
		b.WriteString("\n")
	}
	return b.String()
}

func helpHelpLine() string {
	return styleKey.Render("esc") + " ou " + styleKey.Render("?") + " fechar"
}

package components

import (
	"fmt"
	"strings"

	gitpkg "github.com/laerciocrestani/gitai/internal/git"
	"github.com/laerciocrestani/gitai/internal/tui/theme"
)

// SelectMarker renders (●) or ( ) — mesmo estilo do wizard de config.
func SelectMarker(selected bool) string {
	if selected {
		return theme.S.Current.Render("(●)")
	}
	return theme.S.Hint.Render("( )")
}

// BulletMarker renders ● ou ○ para itens da lista.
func BulletMarker(selected bool) string {
	if selected {
		return theme.S.Current.Render("●")
	}
	return theme.S.Hint.Render("○")
}

func rowPrefix(current bool) string {
	if current {
		return theme.S.Current.Render("> ")
	}
	return "  "
}

// RenderAddTodosLine renders the "Todos" row at the top of the add screen.
func RenderAddTodosLine(allSelected, current bool) string {
	var b strings.Builder
	b.WriteString(rowPrefix(current))
	b.WriteString(SelectMarker(allSelected))
	b.WriteString(" ")
	if current {
		b.WriteString(theme.S.Current.Render("Todos"))
	} else {
		b.WriteString(theme.S.Hint.Render("Todos"))
	}
	return b.String()
}

// RenderAddFileLine renders one selectable file row in the add screen.
func RenderAddFileLine(selected, current bool, f gitpkg.FileChange) string {
	tag := statusTag(f.Status)
	path := f.Path
	if current {
		path = theme.S.Current.Render(path)
	} else {
		path = fileRowStyle(f.Status).Render(path)
	}

	var b strings.Builder
	b.WriteString(rowPrefix(current))
	b.WriteString(BulletMarker(selected))
	b.WriteString(theme.S.Hint.Render(fmt.Sprintf(" %-*s ", tagWidth, tag)))
	b.WriteString(path)
	return b.String()
}

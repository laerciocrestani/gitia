package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

const (
	maxCIFixFiles     = 8
	maxCIFixFileRunes = 120_000
	maxCIFixLogRunes  = 40_000
)

// CIFixFile is one file the model wants to write.
type CIFixFile struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// CIFixSuggestion is the structured AI response for fixing a CI failure.
type CIFixSuggestion struct {
	Summary       string      `json:"summary"`
	CommitMessage string      `json:"commit_message"`
	Files         []CIFixFile `json:"files"`
	Notes         []string    `json:"notes"`
}

func buildCIFixPrompt(logWindow, lang, branch string) string {
	var b strings.Builder
	b.WriteString(`Você é um engenheiro corrigindo uma falha de CI (GitHub Actions).

Com base no log redigido abaixo, proponha a menor correção possível no código do repositório.

Responda SOMENTE com JSON válido, sem markdown:
{
  "summary": "o que falhou e o que você corrige (1-3 frases)",
  "commit_message": "mensagem conventional commits completa (title + body opcional)",
  "files": [
    {"path": "caminho/relativo/ao/repo", "content": "conteúdo COMPLETO do arquivo após a correção"}
  ],
  "notes": ["avisos opcionais"]
}

Regras:
- Idioma: `)
	b.WriteString(lang)
	b.WriteString(`
- path sempre relativo ao root do repo; sem .. nem caminhos absolutos
- no máximo `)
	b.WriteString(fmt.Sprintf("%d", maxCIFixFiles))
	b.WriteString(` arquivos
- content deve ser o arquivo inteiro pronto para gravar (não um diff)
- não invente arquivos sem evidência no log
- não inclua secrets; o log já está redigido
- se não for possível corrigir com confiança, retorne files=[] e explique em summary/notes
- commit_message no formato Conventional Commits (ex.: fix: …)

Branch atual: `)
	b.WriteString(strings.TrimSpace(branch))
	b.WriteString(`

Log da CI (failure window, redigido):
`)
	b.WriteString(truncateRunes(logWindow, maxCIFixLogRunes))
	return b.String()
}

func parseCIFixSuggestion(raw string) (*CIFixSuggestion, error) {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	var sug CIFixSuggestion
	if err := json.Unmarshal([]byte(raw), &sug); err != nil {
		return nil, fmt.Errorf("parse JSON da IA: %w\nresposta: %s", err, raw)
	}
	sug.Summary = strings.TrimSpace(sug.Summary)
	sug.CommitMessage = strings.TrimSpace(sug.CommitMessage)
	if sug.Summary == "" {
		return nil, fmt.Errorf("resposta da IA incompleta: summary é obrigatório")
	}
	if len(sug.Files) > maxCIFixFiles {
		sug.Files = sug.Files[:maxCIFixFiles]
	}
	cleaned := make([]CIFixFile, 0, len(sug.Files))
	for _, f := range sug.Files {
		path := strings.TrimSpace(f.Path)
		path = strings.TrimPrefix(path, "./")
		if path == "" || strings.Contains(path, "..") || strings.HasPrefix(path, "/") {
			continue
		}
		if strings.ContainsAny(path, "\x00") {
			continue
		}
		content := f.Content
		if runeCount(content) > maxCIFixFileRunes {
			return nil, fmt.Errorf("arquivo %s excede limite de tamanho da correção", path)
		}
		cleaned = append(cleaned, CIFixFile{Path: path, Content: content})
	}
	sug.Files = cleaned
	return &sug, nil
}

func suggestCIFixWithRetry(ctx context.Context, logWindow, lang, branch string, call apiCall) (*CIFixSuggestion, error) {
	prompt := buildCIFixPrompt(logWindow, lang, branch)
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		raw, err := call(ctx, prompt, "ci-fix")
		if err != nil {
			return nil, err
		}
		sug, err := parseCIFixSuggestion(raw)
		if err == nil {
			return sug, nil
		}
		lastErr = err
		prompt = buildCIFixPrompt(logWindow, lang, branch) + "\n\nA resposta anterior era inválida. Retorne APENAS JSON válido."
	}
	return nil, lastErr
}

func truncateRunes(s string, max int) string {
	if max <= 0 || runeCount(s) <= max {
		return s
	}
	r := []rune(s)
	return string(r[:max]) + "\n… [truncado]"
}

func runeCount(s string) int {
	return len([]rune(s))
}

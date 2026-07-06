package app

import (
	"fmt"
	"os"

	gitpkg "github.com/laerciocrestani/gitai/internal/git"
	"github.com/laerciocrestani/gitai/internal/ui"
	"golang.org/x/term"
)

// PruneBranchAction é a escolha do usuário ao podar uma branch com divergência.
type PruneBranchAction int

const (
	PruneBranchDeleteForce PruneBranchAction = iota
	PruneBranchKeep
)

type pruneBranchChoice struct {
	action PruneBranchAction
	label  string
}

// PruneBranchRecommendation resume a ação sugerida para uma divergência.
type PruneBranchRecommendation struct {
	Action PruneBranchAction
	Reason string
}

// RecommendPruneBranchAction sugere apagar ou manter com base na divergência.
func RecommendPruneBranchAction(issue *gitpkg.BranchPruneIssue) PruneBranchRecommendation {
	if issue == nil {
		return PruneBranchRecommendation{
			Action: PruneBranchDeleteForce,
			Reason: "sem divergência com o upstream",
		}
	}

	switch {
	case issue.LocalAhead > 0 && issue.RemoteAhead == 0:
		return PruneBranchRecommendation{
			Action: PruneBranchDeleteForce,
			Reason: "a branch já está mergeada na base; commits locais extras não foram enviados ao upstream",
		}
	case issue.RemoteAhead > 0 && issue.LocalAhead == 0:
		return PruneBranchRecommendation{
			Action: PruneBranchKeep,
			Reason: "o upstream tem commits que não existem na branch local — revise antes de apagar",
		}
	default:
		return PruneBranchRecommendation{
			Action: PruneBranchKeep,
			Reason: "branch local e upstream divergiram — revise antes de apagar",
		}
	}
}

func promptPruneBranchConflict(sess *ui.Session, issue *gitpkg.BranchPruneIssue) (PruneBranchAction, error) {
	rec := RecommendPruneBranchAction(issue)

	sess.Section("Divergência ao podar " + issue.Name)
	sess.KV("Upstream", issue.Upstream)
	if issue.LocalAhead > 0 {
		sess.KV("Local à frente", fmt.Sprintf("%d commit(s) não enviado(s)", issue.LocalAhead))
		for _, commit := range issue.LocalCommits {
			sess.Bullet(commit)
		}
		if issue.LocalAhead > len(issue.LocalCommits) {
			sess.Bullet(fmt.Sprintf("… e mais %d", issue.LocalAhead-len(issue.LocalCommits)))
		}
	}
	if issue.RemoteAhead > 0 {
		sess.KV("Upstream à frente", fmt.Sprintf("%d commit(s) ausente(s) localmente", issue.RemoteAhead))
		for _, commit := range issue.RemoteCommits {
			sess.Bullet(commit)
		}
		if issue.RemoteAhead > len(issue.RemoteCommits) {
			sess.Bullet(fmt.Sprintf("… e mais %d", issue.RemoteAhead-len(issue.RemoteCommits)))
		}
	}

	sess.Bullet("Recomendado: " + pruneActionLabel(rec.Action) + " — " + rec.Reason)

	choices := []pruneBranchChoice{
		{action: PruneBranchDeleteForce, label: "Apagar com git branch -D (forçado)"},
		{action: PruneBranchKeep, label: "Manter branch local"},
	}

	defaultIdx := 0
	for i, choice := range choices {
		if choice.action == rec.Action {
			defaultIdx = i
			break
		}
	}

	if !prunePromptInteractive() {
		sess.Warn("Terminal não interativo — mantendo " + issue.Name + " (" + pruneActionLabel(rec.Action) + " seria o padrão)")
		return PruneBranchKeep, nil
	}

	idx, err := sess.Choose("O que fazer?", mapPruneChoiceLabels(choices), defaultIdx)
	if err != nil {
		return PruneBranchKeep, err
	}
	return choices[idx].action, nil
}

func mapPruneChoiceLabels(choices []pruneBranchChoice) []string {
	labels := make([]string, len(choices))
	for i, choice := range choices {
		labels[i] = choice.label
	}
	return labels
}

func pruneActionLabel(action PruneBranchAction) string {
	switch action {
	case PruneBranchDeleteForce:
		return "apagar com -D"
	default:
		return "manter"
	}
}

func prunePromptInteractive() bool {
	if os.Getenv("CI") != "" {
		return false
	}
	return term.IsTerminal(int(os.Stdin.Fd()))
}

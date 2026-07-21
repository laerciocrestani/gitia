package desktop

import (
	"context"
	"fmt"
	"strings"

	"github.com/laerciocrestani/openbench/internal/app"
	gitpkg "github.com/laerciocrestani/openbench/internal/git"
)

// CommitPreview is the AI-generated message awaiting human review.
type CommitPreview struct {
	Message string   `json:"message"`
	Notes   []string `json:"notes"`
	Title   string   `json:"title"`
}

// CommitOutcome is returned after a successful confirm.
type CommitOutcome struct {
	Message string `json:"message"`
	Path    string `json:"path"`
	Pushed  bool   `json:"pushed"`
}

// PreviewCommit generates a conventional commit message for projectPath (dry-run).
func PreviewCommit(ctx context.Context, projectPath string) (*CommitPreview, error) {
	if strings.TrimSpace(projectPath) == "" {
		return nil, fmt.Errorf("no project open")
	}
	result, err := app.PreviewCommit(ctx, app.Options{
		WorkDir:  projectPath,
		Progress: app.NopProgress(),
	})
	if err != nil {
		return nil, err
	}
	if result == nil || strings.TrimSpace(result.Message) == "" {
		return nil, fmt.Errorf("IA não retornou mensagem de commit")
	}
	preview := &CommitPreview{Message: result.Message}
	if result.Suggestion != nil {
		preview.Title = result.Suggestion.Title
		preview.Notes = append([]string{}, result.Suggestion.Notes...)
	}
	return preview, nil
}

// ConfirmCommit writes the (possibly edited) message to the repo.
func ConfirmCommit(ctx context.Context, projectPath, message string) (*CommitOutcome, error) {
	if strings.TrimSpace(projectPath) == "" {
		return nil, fmt.Errorf("no project open")
	}
	message = strings.TrimSpace(message)
	if message == "" {
		return nil, fmt.Errorf("mensagem de commit vazia")
	}
	result, err := app.ConfirmCommit(ctx, &app.Result{Message: message}, app.Options{
		WorkDir:  projectPath,
		Progress: app.NopProgress(),
	})
	if err != nil {
		return nil, err
	}
	out := &CommitOutcome{Path: projectPath, Message: message}
	if result != nil && result.Message != "" {
		out.Message = result.Message
	}
	return out, nil
}

// ConfirmCommitAndPush commits with the reviewed message and pushes to origin.
func ConfirmCommitAndPush(ctx context.Context, projectPath, message string) (*CommitOutcome, error) {
	if strings.TrimSpace(projectPath) == "" {
		return nil, fmt.Errorf("no project open")
	}
	message = strings.TrimSpace(message)
	if message == "" {
		return nil, fmt.Errorf("mensagem de commit vazia")
	}
	result, err := app.ConfirmPush(ctx, &app.Result{Message: message}, app.Options{
		WorkDir:  projectPath,
		Progress: app.NopProgress(),
	})
	if err != nil {
		return nil, err
	}
	out := &CommitOutcome{Path: projectPath, Message: message, Pushed: true}
	if result != nil && result.Message != "" {
		out.Message = result.Message
	}
	return out, nil
}

// CreateBranch creates and checks out a new branch from fromName (defaults to main).
func CreateBranch(projectPath, name, fromName string) (*Dashboard, error) {
	if strings.TrimSpace(projectPath) == "" {
		return nil, fmt.Errorf("no project open")
	}
	name = strings.TrimSpace(name)
	fromName = strings.TrimSpace(fromName)
	if name == "" {
		return nil, fmt.Errorf("nome da branch vazio")
	}
	if fromName == "" {
		fromName = "main"
	}
	repo, err := gitpkg.Open(projectPath)
	if err != nil {
		return nil, err
	}
	if err := repo.CreateBranch(name, fromName); err != nil {
		return nil, err
	}
	return LoadDashboard(projectPath)
}

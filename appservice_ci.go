package main

import "github.com/laerciocrestani/openbench/internal/desktop"

// CIStatus lists GitHub Actions runs for the open project (Observe slice).
func (s *AppService) CIStatus(failedOnly bool, limit int) (*desktop.CIStatusView, error) {
	return desktop.LoadCIStatus(s.currentPath(), failedOnly, limit)
}

// CIRunDetail returns one workflow run with jobs/steps.
func (s *AppService) CIRunDetail(runID int64) (*desktop.CIRunDetailView, error) {
	return desktop.LoadCIRunDetail(s.currentPath(), runID)
}

// CILog fetches a redacted Actions log on demand (failed steps and/or job).
func (s *AppService) CILog(runID, jobID int64, failedOnly bool) (*desktop.CILogView, error) {
	return desktop.LoadCILog(s.currentPath(), runID, jobID, failedOnly)
}

// PreviewCIRerun prepares a re-run confirmation with cost warning.
func (s *AppService) PreviewCIRerun(runID, jobID int64, failedOnly bool) (*desktop.CIRerunPreviewView, error) {
	return desktop.PreviewCIRerun(s.currentPath(), runID, jobID, failedOnly)
}

// ConfirmCIRerun executes a re-run after the user confirms in the UI.
func (s *AppService) ConfirmCIRerun(runID, jobID int64, failedOnly bool) error {
	return desktop.ConfirmCIRerun(s.currentPath(), runID, jobID, failedOnly)
}

// ListCIWorkflows lists workflows for optional workflow_dispatch.
func (s *AppService) ListCIWorkflows() ([]desktop.CIWorkflowView, error) {
	return desktop.ListCIWorkflows(s.currentPath())
}

// PreviewCIDispatch prepares workflow_dispatch confirmation (fields as key=value).
func (s *AppService) PreviewCIDispatch(workflow, ref string, fields []string) (*desktop.CIDispatchPreviewView, error) {
	return desktop.PreviewCIDispatch(s.currentPath(), workflow, ref, fields)
}

// ConfirmCIDispatch executes workflow_dispatch after UI confirmation.
func (s *AppService) ConfirmCIDispatch(workflow, ref string, fields []string) error {
	return desktop.ConfirmCIDispatch(s.currentPath(), workflow, ref, fields)
}

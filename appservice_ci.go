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

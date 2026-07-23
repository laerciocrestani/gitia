package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/laerciocrestani/openbench/internal/app"
)

type dashKey int

const (
	dashKeyNone dashKey = iota
	dashKeyCommit
	dashKeyPush
	dashKeyPR
	dashKeyDiff
	dashKeyBranches
	dashKeyAdd
	dashKeySync
	dashKeySyncOptions
	dashKeyOpenPR
	dashKeyCopyHash
	dashKeyLogs
	dashKeyReport
	dashKeyDoctor
	dashKeyDockerUp
	dashKeyDockerDown
	dashKeyDockerLogs
	dashKeyDockerShell
	dashKeyEnvironment
	dashKeyHelp
)

func parseGlobalKey(msg tea.KeyMsg) (keyMsg, bool) {
	switch msg.String() {
	case "q", "ctrl+c":
		return keyQuit, true
	case "r":
		return keyRefresh, true
	}
	return 0, false
}

func parseDashboardKey(msg tea.KeyMsg, snap *app.WorkspaceSnapshot) (dashKey, bool) {
	switch msg.String() {
	case "?":
		return dashKeyHelp, true
	case "u":
		return dashKeyReport, true
	case "h":
		return dashKeyDoctor, true
	case "c":
		if snap != nil && snap.Overview != nil && snap.Overview.IsDirty() && snap.ConfigErr == nil {
			return dashKeyCommit, true
		}
	case "p":
		if app.CanPush(snap) {
			return dashKeyPush, true
		}
	case "P", "shift+p":
		if app.CanPR(snap) {
			return dashKeyPR, true
		}
	case "d":
		return dashKeyDiff, true
	case "b":
		if snap != nil && snap.Overview != nil && len(snap.Overview.Branches) > 0 {
			return dashKeyBranches, true
		}
	case "a":
		if app.CanAdd(snap) {
			return dashKeyAdd, true
		}
	case "s":
		if snap != nil && snap.Overview != nil && app.CanSync(snap) &&
			(snap.Overview.Behind > 0 || snap.Overview.BaseBehind > 0) {
			return dashKeySync, true
		}
	case "S", "shift+s":
		if app.CanHygiene(snap) {
			return dashKeySyncOptions, true
		}
	case "o":
		if snap != nil && snap.OpenPR != nil {
			return dashKeyOpenPR, true
		}
	case "y":
		if snap != nil && snap.Overview != nil && snap.Overview.HeadHash != "" {
			return dashKeyCopyHash, true
		}
	case "l":
		if snap != nil && snap.Overview != nil && len(snap.Overview.RecentCommits) > 0 {
			return dashKeyLogs, true
		}
	case "L", "shift+l":
		if app.CanDockerLogs(snap) {
			return dashKeyDockerLogs, true
		}
	case "U", "shift+u":
		if app.CanDockerUp(snap) {
			return dashKeyDockerUp, true
		}
	case "D", "shift+d":
		if app.CanDockerDown(snap) {
			return dashKeyDockerDown, true
		}
	case "E", "shift+e":
		if app.CanDockerShell(snap) {
			return dashKeyDockerShell, true
		}
	case "i":
		if app.CanDockerEnvironment(snap) {
			return dashKeyEnvironment, true
		}
	}
	return dashKeyNone, false
}

func dashboardHelpLine() string {
	return ""
}

type keyMsg int

const (
	keyRefresh keyMsg = iota
	keyQuit
)

package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/laerciocrestani/gitai/internal/app"
)

type dashKey int

const (
	dashKeyNone dashKey = iota
	dashKeyCommit
	dashKeyPush
	dashKeyPR
	dashKeyDiff
	dashKeySync
	dashKeyOpenPR
	dashKeyReport
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
	case "c":
		if snap != nil && snap.Overview != nil && snap.Overview.IsDirty() && snap.ConfigErr == nil {
			return dashKeyCommit, true
		}
	case "p":
		if CanPush(snap) {
			return dashKeyPush, true
		}
	case "P", "shift+p":
		if CanPR(snap) {
			return dashKeyPR, true
		}
	case "d":
		return dashKeyDiff, true
	case "s":
		if snap != nil && snap.Overview != nil && snap.Overview.Behind > 0 {
			return dashKeySync, true
		}
	case "o":
		if snap != nil && snap.OpenPR != nil {
			return dashKeyOpenPR, true
		}
	}
	return dashKeyNone, false
}

// CanPush indica se push está disponível no snapshot atual.
func CanPush(snap *app.WorkspaceSnapshot) bool {
	if snap == nil || snap.Overview == nil || snap.ConfigErr != nil {
		return false
	}
	o := snap.Overview
	return o.IsDirty() || o.Ahead > 0
}

// CanPR indica se criar PR está disponível.
func CanPR(snap *app.WorkspaceSnapshot) bool {
	if snap == nil || snap.Overview == nil || snap.ConfigErr != nil || !snap.HasGH {
		return false
	}
	o := snap.Overview
	return o.CommitsAheadOfBase > 0 || o.IsDirty()
}

func dashboardHelpLine() string {
	return styleKey.Render("c") + " commit  " +
		styleKey.Render("p") + " push  " +
		styleKey.Render("P") + " pr  " +
		styleKey.Render("d") + " diff  " +
		styleKey.Render("u") + " usage  " +
		styleKey.Render("?") + " help  " +
		styleKey.Render("q") + " quit"
}

type keyMsg int

const (
	keyRefresh keyMsg = iota
	keyQuit
)

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
	case "c":
		if snap != nil && snap.Overview != nil && snap.Overview.IsDirty() && snap.ConfigErr == nil {
			return dashKeyCommit, true
		}
	case "p":
		if canPush(snap) {
			return dashKeyPush, true
		}
	case "P":
		if canPR(snap) {
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

func canPush(snap *app.WorkspaceSnapshot) bool {
	if snap == nil || snap.Overview == nil || snap.ConfigErr != nil {
		return false
	}
	o := snap.Overview
	return o.IsDirty() || o.Ahead > 0
}

func canPR(snap *app.WorkspaceSnapshot) bool {
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
		styleKey.Render("s") + " sync  " +
		styleKey.Render("o") + " open pr  " +
		styleKey.Render("r") + " refresh  " +
		styleKey.Render("q") + " quit"
}

type keyMsg int

const (
	keyRefresh keyMsg = iota
	keyQuit
)

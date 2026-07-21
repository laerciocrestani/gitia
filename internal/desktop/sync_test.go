package desktop

import (
	"testing"
)

func TestSyncFlags(t *testing.T) {
	prune, remote, err := syncFlags(SyncModeStandard)
	if err != nil || prune || remote {
		t.Fatalf("standard: prune=%v remote=%v err=%v", prune, remote, err)
	}
	prune, remote, err = syncFlags(SyncModePruneRemote)
	if err != nil || prune || !remote {
		t.Fatalf("prune_remote: prune=%v remote=%v err=%v", prune, remote, err)
	}
	prune, remote, err = syncFlags(SyncModePruneFull)
	if err != nil || !prune || remote {
		t.Fatalf("prune_full: prune=%v remote=%v err=%v", prune, remote, err)
	}
	if _, _, err := syncFlags("nope"); err == nil {
		t.Fatal("expected error for invalid mode")
	}
}

func TestSyncModesCatalog(t *testing.T) {
	modes := SyncModes()
	if len(modes) != 3 {
		t.Fatalf("modes=%d", len(modes))
	}
	if modes[0].ID != SyncModeStandard || modes[2].ID != SyncModePruneFull {
		t.Fatalf("unexpected order: %+v", modes)
	}
}

func TestRunSync_emptyPath(t *testing.T) {
	if _, err := RunSync("", SyncModeStandard, "main"); err == nil {
		t.Fatal("expected error for empty path")
	}
}

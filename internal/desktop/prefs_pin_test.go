package desktop

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestPinAndUnpin(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := filepath.Join(home, "proj")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}

	prefs, err := PinProject(dir, "demo")
	if err != nil {
		t.Fatal(err)
	}
	if len(prefs.Pinned) != 1 || prefs.Pinned[0].Alias != "demo" {
		t.Fatalf("pinned=%+v", prefs.Pinned)
	}

	prefs, err = UnpinProject(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(prefs.Pinned) != 0 {
		t.Fatalf("expected empty pinned, got %+v", prefs.Pinned)
	}
}

func TestPinnedOrderMostRecentFirst(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	a := filepath.Join(home, "a")
	b := filepath.Join(home, "b")
	c := filepath.Join(home, "c")
	for _, dir := range []string{a, b, c} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	if _, err := PinProject(a, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := PinProject(b, ""); err != nil {
		t.Fatal(err)
	}
	prefs, err := PinProject(c, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(prefs.Pinned) != 3 {
		t.Fatalf("pinned=%+v", prefs.Pinned)
	}
	if prefs.Pinned[0].Path != c || prefs.Pinned[1].Path != b || prefs.Pinned[2].Path != a {
		t.Fatalf("want c,b,a got %+v", prefs.Pinned)
	}

	prefs, err = RememberProject(a)
	if err != nil {
		t.Fatal(err)
	}
	if prefs.Pinned[0].Path != a || prefs.Pinned[1].Path != c || prefs.Pinned[2].Path != b {
		t.Fatalf("after touch a: want a,c,b got %+v", prefs.Pinned)
	}
}

func TestPinLimit(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	for i := 0; i < MaxPinned; i++ {
		dir := filepath.Join(home, "p", strconv.Itoa(i))
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if _, err := PinProject(dir, ""); err != nil {
			t.Fatalf("pin %d: %v", i, err)
		}
	}
	extra := filepath.Join(home, "p", "extra")
	if err := os.MkdirAll(extra, 0o755); err != nil {
		t.Fatal(err)
	}
	_, err := PinProject(extra, "")
	if err != ErrTooManyPinned {
		t.Fatalf("want ErrTooManyPinned, got %v", err)
	}
}

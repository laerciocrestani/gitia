package dockerpresets

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadKitLaravel(t *testing.T) {
	kit, err := LoadKit("laravel")
	if err != nil {
		t.Fatal(err)
	}
	if kit.ID != "laravel" {
		t.Fatalf("id=%q", kit.ID)
	}
	if len(kit.Presets) < 5 {
		t.Fatalf("expected several artisan presets, got %d", len(kit.Presets))
	}
	var foundMigrate, foundTinker bool
	for _, p := range kit.Presets {
		if p.ID == "laravel.migrate" && p.Command == "php artisan migrate" {
			foundMigrate = true
		}
		if p.ID == "laravel.tinker" && p.Interactive {
			foundTinker = true
		}
	}
	if !foundMigrate {
		t.Fatal("missing laravel.migrate")
	}
	if !foundTinker {
		t.Fatal("missing interactive laravel.tinker")
	}
}

func TestImportKitSkipsExisting(t *testing.T) {
	dir := t.TempDir()
	f := &File{
		Version: FileVersion,
		Presets: []Preset{{
			ID:      "laravel.migrate",
			Label:   "custom migrate",
			Command: "php artisan migrate --force",
		}},
	}
	if err := SaveProject(dir, f); err != nil {
		t.Fatal(err)
	}
	added, err := ImportKit(dir, "laravel")
	if err != nil {
		t.Fatal(err)
	}
	if added < 1 {
		t.Fatalf("expected new presets, added=%d", added)
	}
	got, err := LoadProject(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range got.Presets {
		if p.ID == "laravel.migrate" && p.Command != "php artisan migrate --force" {
			t.Fatalf("existing preset was overwritten: %+v", p)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, RelativePath)); err != nil {
		t.Fatal(err)
	}
}

func TestParseCommand(t *testing.T) {
	got := ParseCommand("  php artisan migrate:fresh --seed  ")
	want := []string{"php", "artisan", "migrate:fresh", "--seed"}
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v want %v", got, want)
		}
	}
}

func TestListKits(t *testing.T) {
	kits, err := ListKits()
	if err != nil {
		t.Fatal(err)
	}
	if len(kits) == 0 {
		t.Fatal("expected at least laravel kit")
	}
}

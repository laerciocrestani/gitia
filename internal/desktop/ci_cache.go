package desktop

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/laerciocrestani/openbench/internal/config"
	"github.com/laerciocrestani/openbench/internal/gha"
)

type ciCacheFile struct {
	SavedAt time.Time   `json:"savedAt"`
	Summary gha.Summary `json:"summary"`
}

func ciCachePath(projectPath string) (string, error) {
	dir, err := config.DataDir()
	if err != nil {
		return "", err
	}
	cacheDir := filepath.Join(dir, "ci-cache")
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte(strings.TrimSpace(projectPath)))
	return filepath.Join(cacheDir, hex.EncodeToString(sum[:8])+".json"), nil
}

func saveCISummaryCache(projectPath string, sum *gha.Summary) {
	if sum == nil || sum.FromCache {
		return
	}
	path, err := ciCachePath(projectPath)
	if err != nil {
		return
	}
	payload := ciCacheFile{SavedAt: time.Now().UTC(), Summary: *sum}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(path, data, 0o600)
}

func loadCISummaryCache(projectPath string) *gha.Summary {
	path, err := ciCachePath(projectPath)
	if err != nil {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var payload ciCacheFile
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil
	}
	// ignore very stale cache (>7 days)
	if !payload.SavedAt.IsZero() && time.Since(payload.SavedAt) > 7*24*time.Hour {
		return nil
	}
	s := payload.Summary
	s.FromCache = true
	return &s
}

// LoadCISummaryForProject fetches CI badge data, falling back to disk cache offline.
func LoadCISummaryForProject(projectPath, branch string) *gha.Summary {
	client, err := gha.Open(projectPath)
	if err != nil {
		return gha.SummaryFromError(projectPath, err, loadCISummaryCache(projectPath))
	}
	sum, err := client.LoadSummary(branch)
	if err != nil {
		return gha.SummaryFromError(projectPath, err, loadCISummaryCache(projectPath))
	}
	saveCISummaryCache(projectPath, sum)
	return sum
}

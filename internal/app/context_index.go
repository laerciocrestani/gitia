package app

import (
	"strings"

	"github.com/laerciocrestani/openbench/internal/ai"
	"github.com/laerciocrestani/openbench/internal/config"
	gitpkg "github.com/laerciocrestani/openbench/internal/git"
)

// Níveis do índice de contexto do próximo commit.
const (
	ContextLevelOK        = "ok"
	ContextLevelAttention = "attention"
	ContextLevelCritical  = "critical"
)

const (
	contextBytesPerLine       = 48
	contextUntrackedLineFloor = 40
	contextFileSoftCap        = 25
)

// CommitContextIndex estima o peso do diff que a IA verá no próximo commit
// (fluxo padrão com git add .), sem chamar o modelo.
type CommitContextIndex struct {
	Score           int    `json:"score"`
	Level           string `json:"level"`
	Label           string `json:"label"`
	RecommendCommit bool   `json:"recommendCommit"`
	FileCount       int    `json:"fileCount"`
	Insertions      int    `json:"insertions"`
	Deletions       int    `json:"deletions"`
	AreaCount       int    `json:"areaCount"`
	EstimatedBytes  int    `json:"estimatedBytes"`
	MaxDiffBytes    int    `json:"maxDiffBytes"`
	NearTruncate    bool   `json:"nearTruncate"`
}

// BuildCommitContextIndex calcula o índice a partir do overview e da config.
// Retorna nil quando não há alterações commitáveis.
func BuildCommitContextIndex(o *gitpkg.Overview, cfg *config.Config) *CommitContextIndex {
	if o == nil || len(o.FileChanges) == 0 || !o.IsDirty() {
		return nil
	}

	maxBytes := config.Default().MaxDiffBytes
	if cfg != nil && cfg.MaxDiffBytes > 0 {
		maxBytes = cfg.MaxDiffBytes
	}

	paths := make([]string, 0, len(o.FileChanges))
	insertions, deletions := 0, 0
	estimated := 0

	for _, f := range o.FileChanges {
		ins, del := estimateCommitChurn(f)
		insertions += ins
		deletions += del
		estimated += (ins + del) * contextBytesPerLine
		estimated += len(f.Path) * 8
		paths = append(paths, f.Path)
	}

	areas := ai.ChangeAreasFromPaths(paths)
	areaCount := len(areas)

	byteScore := percentOf(estimated, maxBytes)
	fileScore := percentOf(len(o.FileChanges), contextFileSoftCap)
	areaScore := areaPressureScore(areaCount)

	score := (byteScore*55 + fileScore*25 + areaScore*20) / 100
	if score > 100 {
		score = 100
	}

	nearTruncate := estimated >= maxBytes
	level := ContextLevelOK
	switch {
	case nearTruncate || score >= 70:
		level = ContextLevelCritical
		if score < 70 {
			score = 70
		}
	case score >= 40 || areaCount >= 2:
		level = ContextLevelAttention
		if score < 40 && areaCount >= 2 {
			score = 40
		}
	}

	idx := &CommitContextIndex{
		Score:           score,
		Level:           level,
		Label:           contextLevelLabel(level),
		RecommendCommit: level == ContextLevelAttention || level == ContextLevelCritical,
		FileCount:       len(o.FileChanges),
		Insertions:      insertions,
		Deletions:       deletions,
		AreaCount:       areaCount,
		EstimatedBytes:  estimated,
		MaxDiffBytes:    maxBytes,
		NearTruncate:    nearTruncate,
	}
	return idx
}

func estimateCommitChurn(f gitpkg.FileChange) (insertions, deletions int) {
	insertions, deletions = f.Insertions, f.Deletions
	if strings.EqualFold(f.Status, "untracked") && insertions+deletions == 0 {
		return contextUntrackedLineFloor, 0
	}
	return insertions, deletions
}

func percentOf(value, cap int) int {
	if cap <= 0 || value <= 0 {
		return 0
	}
	pct := value * 100 / cap
	if pct > 100 {
		return 100
	}
	return pct
}

func areaPressureScore(areaCount int) int {
	switch {
	case areaCount >= 4:
		return 100
	case areaCount >= 3:
		return 75
	case areaCount >= 2:
		return 50
	default:
		return 0
	}
}

func contextLevelLabel(level string) string {
	switch level {
	case ContextLevelCritical:
		return "Contexto alto — recomenda-se commit"
	case ContextLevelAttention:
		return "Diff crescendo — considere commit"
	default:
		return "Contexto saudável"
	}
}

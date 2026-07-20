package desktop

import (
	"sync"
	"time"
)

// StatusEmitter sends project status updates to the UI.
type StatusEmitter func(ProjectStatus)

// StatusHub polls pinned projects in parallel with capped concurrency.
type StatusHub struct {
	mu       sync.Mutex
	paths    []string
	aliases  map[string]string
	active   string
	cache    map[string]ProjectStatus
	emit     StatusEmitter
	stopCh   chan struct{}
	running  bool
	tick     int
	interval time.Duration
}

// NewStatusHub creates a hub. Call Start after setting emitter.
func NewStatusHub(emit StatusEmitter) *StatusHub {
	return &StatusHub{
		aliases:  map[string]string{},
		cache:    map[string]ProjectStatus{},
		emit:     emit,
		interval: 5 * time.Second,
	}
}

// Start begins background polling.
func (h *StatusHub) Start() {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.running {
		return
	}
	h.stopCh = make(chan struct{})
	h.running = true
	go h.loop(h.stopCh)
}

// Stop ends background polling.
func (h *StatusHub) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.running {
		return
	}
	close(h.stopCh)
	h.running = false
}

// SetPinned replaces watched paths (max MaxPinned) and optional aliases.
func (h *StatusHub) SetPinned(pinned []PinnedProject, active string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.paths = h.paths[:0]
	h.aliases = map[string]string{}
	for i, p := range pinned {
		if i >= MaxPinned {
			break
		}
		h.paths = append(h.paths, p.Path)
		if p.Alias != "" {
			h.aliases[p.Path] = p.Alias
		}
	}
	h.active = active
}

// SetActive marks the focused project (gets PR checks more often).
func (h *StatusHub) SetActive(path string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.active = path
}

// Snapshot returns cached statuses in pinned order.
func (h *StatusHub) Snapshot() []ProjectStatus {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]ProjectStatus, 0, len(h.paths))
	for _, path := range h.paths {
		st, ok := h.cache[path]
		if !ok {
			st = ProjectStatus{Path: path, RepoName: path}
		}
		st.Active = samePath(path, h.active)
		if alias := h.aliases[path]; alias != "" {
			st.Alias = alias
		}
		out = append(out, st)
	}
	return out
}

// RefreshNow runs one poll cycle synchronously (light, no PR except active).
func (h *StatusHub) RefreshNow() []ProjectStatus {
	h.poll(false)
	return h.Snapshot()
}

func (h *StatusHub) loop(stop <-chan struct{}) {
	// immediate first pass
	h.poll(true)
	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			h.poll(false)
		}
	}
}

func (h *StatusHub) poll(forcePR bool) {
	h.mu.Lock()
	paths := append([]string{}, h.paths...)
	active := h.active
	aliases := map[string]string{}
	for k, v := range h.aliases {
		aliases[k] = v
	}
	h.tick++
	tick := h.tick
	h.mu.Unlock()

	if len(paths) == 0 {
		return
	}

	includePRActive := forcePR || tick%12 == 0 // ~60s with 5s interval

	var wg sync.WaitGroup
	sem := make(chan struct{}, 3) // cap parallelism
	results := make([]ProjectStatus, len(paths))

	for i, path := range paths {
		wg.Add(1)
		go func(i int, path string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			withPR := includePRActive && samePath(path, active)
			st := LoadProjectStatus(path, withPR)
			if alias := aliases[path]; alias != "" {
				st.Alias = alias
			}
			st.Active = samePath(path, active)
			results[i] = st
		}(i, path)
	}
	wg.Wait()

	h.mu.Lock()
	for _, st := range results {
		// Preserve previous PR info if this tick skipped gh.
		if !includePRActive || !samePath(st.Path, active) {
			if prev, ok := h.cache[st.Path]; ok && prev.HasOpenPR && !st.HasOpenPR && st.Error == "" {
				st.HasOpenPR = prev.HasOpenPR
				st.PRTitle = prev.PRTitle
			}
		}
		h.cache[st.Path] = st
	}
	emit := h.emit
	h.mu.Unlock()

	if emit != nil {
		for _, st := range results {
			emit(st)
		}
	}
}

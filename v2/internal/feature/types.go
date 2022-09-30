package feature

import "sync"

type Flags struct {
	mu       sync.Mutex
	features map[Flag]bool
}

func (f *Flags) Enable(flag Flag) {
	f.mu.Lock()
	f.features[flag] = true
	f.mu.Unlock()
}

func (f *Flags) Enabled(flag Flag) bool {
	return f.features[flag]
}

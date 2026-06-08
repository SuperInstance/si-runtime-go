package siruntime

import (
	"fmt"
	"strings"
	"sync"
)

// Capability describes a skill or function an agent can perform.
type Capability struct {
	Name     string
	Version  string
	Provides []string
	Requires []string
}

// Registry stores and queries capabilities.
type Registry struct {
	caps map[string]Capability
	mu   sync.RWMutex
}

// NewRegistry creates an empty capability registry.
func NewRegistry() *Registry {
	return &Registry{caps: make(map[string]Capability)}
}

// Register adds a capability to the registry.
func (r *Registry) Register(c Capability) error {
	if strings.TrimSpace(c.Name) == "" {
		return fmt.Errorf("capability name cannot be empty")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.caps[c.Name] = c
	return nil
}

// Get retrieves a capability by name.
func (r *Registry) Get(name string) (Capability, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.caps[name]
	return c, ok
}

// List returns all registered capabilities.
func (r *Registry) List() []Capability {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Capability, 0, len(r.caps))
	for _, c := range r.caps {
		out = append(out, c)
	}
	return out
}

// Remove deletes a capability from the registry.
func (r *Registry) Remove(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.caps, name)
}

// MatchResult scores how well an agent satisfies task requirements.
type MatchResult struct {
	AgentID string
	Score   float64
	Matched []string
	Missing []string
}

// Match scores compatibility between agent capabilities and required functionalities.
func Match(agentID string, agentCaps []string, required []string) MatchResult {
	result := MatchResult{AgentID: agentID}
	if len(required) == 0 {
		result.Score = 1.0
		return result
	}
	capSet := make(map[string]struct{})
	for _, c := range agentCaps {
		capSet[c] = struct{}{}
	}
	matched := 0
	for _, req := range required {
		if _, ok := capSet[req]; ok {
			result.Matched = append(result.Matched, req)
			matched++
		} else {
			result.Missing = append(result.Missing, req)
		}
	}
	result.Score = float64(matched) / float64(len(required))
	return result
}

// Resolve selects the best matching agent from candidates.
func Resolve(candidates []MatchResult) (MatchResult, bool) {
	if len(candidates) == 0 {
		return MatchResult{}, false
	}
	best := candidates[0]
	for _, c := range candidates[1:] {
		if c.Score > best.Score {
			best = c
		}
	}
	return best, true
}

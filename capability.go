package siruntime

import (
	"fmt"
	"strings"
	"sync"
)

// Capability represents a skill or function an agent can perform.
type Capability struct {
	Name     string
	Version  string
	Score    float64 // 0.0 to 1.0, higher = more proficient
	Metadata map[string]string
}

// CapabilityRegistry stores and queries capabilities.
type CapabilityRegistry struct {
	caps map[string]Capability
	mu   sync.RWMutex
}

// NewCapabilityRegistry creates an empty registry.
func NewCapabilityRegistry() *CapabilityRegistry {
	return &CapabilityRegistry{
		caps: make(map[string]Capability),
	}
}

// Register adds a capability to the registry.
// Overwrites any existing capability with the same name.
func (r *CapabilityRegistry) Register(c Capability) error {
	if strings.TrimSpace(c.Name) == "" {
		return fmt.Errorf("capability name cannot be empty")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.caps[c.Name] = c
	return nil
}

// Get retrieves a capability by name.
func (r *CapabilityRegistry) Get(name string) (Capability, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.caps[name]
	return c, ok
}

// List returns all registered capabilities.
func (r *CapabilityRegistry) List() []Capability {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Capability, 0, len(r.caps))
	for _, c := range r.caps {
		out = append(out, c)
	}
	return out
}

// Remove deletes a capability from the registry.
func (r *CapabilityRegistry) Remove(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.caps, name)
}

// MatchResult scores how well an agent matches a task requirement.
type MatchResult struct {
	AgentID    string
	Score      float64 // 0.0 to 1.0
	Matched    []string
	Missing    []string
	Partial    []string // matched but score < threshold
}

// Match computes compatibility between an agent's capabilities and a
// set of required capabilities. Each requirement is a capability name;
// the agent's score for that capability contributes to the overall match.
func Match(agentID string, agentCaps []Capability, required []string, threshold float64) MatchResult {
	result := MatchResult{AgentID: agentID}
	if len(required) == 0 {
		result.Score = 1.0
		return result
	}

	capMap := make(map[string]Capability)
	for _, c := range agentCaps {
		capMap[c.Name] = c
	}

	var totalScore float64
	for _, req := range required {
		if cap, ok := capMap[req]; ok {
			if cap.Score >= threshold {
				result.Matched = append(result.Matched, req)
			} else {
				result.Partial = append(result.Partial, req)
			}
			totalScore += cap.Score
		} else {
			result.Missing = append(result.Missing, req)
		}
	}

	if len(result.Missing) == 0 && len(result.Partial) == 0 {
		result.Score = 1.0
	} else {
		result.Score = totalScore / float64(len(required))
	}
	return result
}

// BestMatch selects the highest-scoring agent from a list of candidates.
func BestMatch(candidates []MatchResult) (MatchResult, bool) {
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

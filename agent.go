package siruntime

import (
	"fmt"
	"math"
	"sync"
)

// Agent represents a computational agent in the SuperInstance fleet.
type Agent struct {
	ID           string
	State        map[string]float64
	Capabilities []string
	Homeostasis  map[string]float64
	mu           sync.RWMutex
}

// NewAgent creates an agent with the given ID.
func NewAgent(id string) *Agent {
	return &Agent{
		ID:           id,
		State:        make(map[string]float64),
		Capabilities: make([]string, 0),
		Homeostasis:  make(map[string]float64),
	}
}

// SetState sets a scalar state variable.
func (a *Agent) SetState(key string, value float64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.State[key] = value
}

// GetState retrieves a scalar state variable.
func (a *Agent) GetState(key string) (float64, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	v, ok := a.State[key]
	return v, ok
}

// SetHomeostasis sets the target value for a state variable.
func (a *Agent) SetHomeostasis(key string, target float64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Homeostasis[key] = target
}

// AddCapability adds a capability name to the agent.
func (a *Agent) AddCapability(name string) error {
	if name == "" {
		return fmt.Errorf("capability name cannot be empty")
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, existing := range a.Capabilities {
		if existing == name {
			return fmt.Errorf("agent %s already has capability %s", a.ID, name)
		}
	}
	a.Capabilities = append(a.Capabilities, name)
	return nil
}

// RemoveCapability removes a capability by name.
func (a *Agent) RemoveCapability(name string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	for i, c := range a.Capabilities {
		if c == name {
			a.Capabilities = append(a.Capabilities[:i], a.Capabilities[i+1:]...)
			return true
		}
	}
	return false
}

// ListCapabilities returns a copy of the agent's capability names.
func (a *Agent) ListCapabilities() []string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := make([]string, len(a.Capabilities))
	copy(out, a.Capabilities)
	return out
}

// UpdateHomeostasis drives each state variable toward its target.
func (a *Agent) UpdateHomeostasis(rate float64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for key, target := range a.Homeostasis {
		current := a.State[key]
		a.State[key] = current + rate*(target-current)
	}
}

// HomeostasisError computes RMS deviation from targets.
func (a *Agent) HomeostasisError() float64 {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if len(a.Homeostasis) == 0 {
		return 0
	}
	var sumSq float64
	for key, target := range a.Homeostasis {
		d := a.State[key] - target
		sumSq += d * d
	}
	return math.Sqrt(sumSq / float64(len(a.Homeostasis)))
}

// String returns a concise representation.
func (a *Agent) String() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return fmt.Sprintf("Agent[%s] caps=%d states=%d", a.ID, len(a.Capabilities), len(a.State))
}

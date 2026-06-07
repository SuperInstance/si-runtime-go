package siruntime

import (
	"fmt"
	"math"
	"sync"
)

// Agent represents a computational agent in the SuperInstance fleet.
type Agent struct {
	ID            string
	State         map[string]float64 // arbitrary scalar state variables
	Capabilities  []Capability
	Homeostasis   map[string]float64 // target values for state variables
	Budget        *AgentBudget
	mu            sync.RWMutex
}

// NewAgent creates an agent with the given ID.
func NewAgent(id string) *Agent {
	return &Agent{
		ID:           id,
		State:        make(map[string]float64),
		Capabilities: make([]Capability, 0),
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

// AddCapability adds a capability to the agent.
func (a *Agent) AddCapability(c Capability) error {
	if c.Name == "" {
		return fmt.Errorf("capability name cannot be empty")
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, existing := range a.Capabilities {
		if existing.Name == c.Name {
			return fmt.Errorf("agent %s already has capability %s", a.ID, c.Name)
		}
	}
	a.Capabilities = append(a.Capabilities, c)
	return nil
}

// RemoveCapability removes a capability by name.
func (a *Agent) RemoveCapability(name string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	for i, c := range a.Capabilities {
		if c.Name == name {
			a.Capabilities = append(a.Capabilities[:i], a.Capabilities[i+1:]...)
			return true
		}
	}
	return false
}

// ListCapabilities returns a copy of the agent's capabilities.
func (a *Agent) ListCapabilities() []Capability {
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := make([]Capability, len(a.Capabilities))
	copy(out, a.Capabilities)
	return out
}

// UpdateHomeostasis drives each state variable toward its target:
//   state[k] += rate * (homeostasis[k] - state[k])
func (a *Agent) UpdateHomeostasis(rate float64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for key, target := range a.Homeostasis {
		current := a.State[key]
		a.State[key] = current + rate*(target-current)
	}
}

// HomeostasisError computes the root-mean-square deviation from targets.
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

// AttachBudget links a budget to this agent.
func (a *Agent) AttachBudget(total float64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Budget = &AgentBudget{
		AgentID: a.ID,
		Budget:  NewBudget(total),
	}
}

// TotalCapabilityScore returns the sum of all capability scores.
func (a *Agent) TotalCapabilityScore() float64 {
	a.mu.RLock()
	defer a.mu.RUnlock()
	var sum float64
	for _, c := range a.Capabilities {
		sum += c.Score
	}
	return sum
}

// String returns a concise representation of the agent.
func (a *Agent) String() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	budgetStr := "no budget"
	if a.Budget != nil {
		budgetStr = fmt.Sprintf("budget(total=%.0f γ=%.0f η=%.0f)",
			a.Budget.Budget.Total, a.Budget.Budget.Gamma, a.Budget.Budget.Eta)
	}
	return fmt.Sprintf("Agent[%s] caps=%d %s", a.ID, len(a.Capabilities), budgetStr)
}

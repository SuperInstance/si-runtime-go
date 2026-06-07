package siruntime

import (
	"fmt"
	"math"
	"sort"
	"sync"
)

// Fleet is a collection of agents with shared conservation and spectral
// infrastructure.
type Fleet struct {
	ID       string
	Agents   map[string]*Agent
	Adjacency *AdjacencyMatrix
	mu       sync.RWMutex
}

// NewFleet creates a fleet with the given identifier.
func NewFleet(id string) *Fleet {
	return &Fleet{
		ID:      id,
		Agents:  make(map[string]*Agent),
	}
}

// AddAgent registers an agent with the fleet.
func (f *Fleet) AddAgent(a *Agent) error {
	if a == nil {
		return fmt.Errorf("cannot add nil agent")
	}
	if a.ID == "" {
		return fmt.Errorf("agent ID cannot be empty")
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, exists := f.Agents[a.ID]; exists {
		return fmt.Errorf("agent %s already in fleet %s", a.ID, f.ID)
	}
	f.Agents[a.ID] = a
	f.Adjacency = nil // invalidate cached matrix
	return nil
}

// RemoveAgent unregisters an agent.
func (f *Fleet) RemoveAgent(id string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, exists := f.Agents[id]; !exists {
		return false
	}
	delete(f.Agents, id)
	f.Adjacency = nil
	return true
}

// GetAgent retrieves an agent by ID.
func (f *Fleet) GetAgent(id string) (*Agent, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	a, ok := f.Agents[id]
	return a, ok
}

// ListAgents returns all agents in the fleet.
func (f *Fleet) ListAgents() []*Agent {
	f.mu.RLock()
	defer f.mu.RUnlock()
	out := make([]*Agent, 0, len(f.Agents))
	for _, a := range f.Agents {
		out = append(out, a)
	}
	return out
}

// AgentCount returns the number of agents.
func (f *Fleet) AgentCount() int {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return len(f.Agents)
}

// BuildAdjacencyMatrix constructs an affinity matrix from agent states.
// The affinity between agents i and j is computed by the provided function.
func (f *Fleet) BuildAdjacencyMatrix(affinity func(a, b *Agent) float64) (*AdjacencyMatrix, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	agents := f.agentSlice()
	n := len(agents)
	if n == 0 {
		return nil, fmt.Errorf("fleet %s has no agents", f.ID)
	}

	m := NewAdjacencyMatrix(n)
	for i := 0; i < n; i++ {
		for j := i; j < n; j++ {
			aff := affinity(agents[i], agents[j])
			m.Data[i][j] = aff
			m.Data[j][i] = aff
		}
	}
	return m, nil
}

// SpectralRank ranks agents by eigenvector centrality using their
// current adjacency matrix. If no matrix exists, one is built from
// workload state similarity.
func (f *Fleet) SpectralRank() ([]RankedAgent, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.Adjacency == nil {
		agents := f.agentSlice()
		m := NewAdjacencyMatrix(len(agents))
		for i := 0; i < len(agents); i++ {
			for j := i; j < len(agents); j++ {
				var aff float64
				if i == j {
					aff = 1.0
				} else {
					wi, _ := agents[i].GetState("workload")
					wj, _ := agents[j].GetState("workload")
					d := wi - wj
					aff = math.Exp(-d * d)
				}
				m.Data[i][j] = aff
				m.Data[j][i] = aff
			}
		}
		f.Adjacency = m
	}

	return SpectralRank(f.Adjacency)
}

// ConservationAudit verifies that every agent's budget satisfies
// gamma + eta == total, and computes fleet-wide aggregates.
func (f *Fleet) ConservationAudit() AuditResult {
	f.mu.RLock()
	defer f.mu.RUnlock()

	var budgets []*AgentBudget
	for _, a := range f.Agents {
		if a.Budget != nil {
			budgets = append(budgets, a.Budget)
		}
	}
	return Audit(budgets)
}

// TotalFleetBudget returns the sum of all agent budget totals.
func (f *Fleet) TotalFleetBudget() float64 {
	f.mu.RLock()
	defer f.mu.RUnlock()
	var sum float64
	for _, a := range f.Agents {
		if a.Budget != nil {
			sum += a.Budget.Budget.Total
		}
	}
	return sum
}

// FleetHomeostasisError returns the average homeostasis error across all agents.
func (f *Fleet) FleetHomeostasisError() float64 {
	f.mu.RLock()
	defer f.mu.RUnlock()
	if len(f.Agents) == 0 {
		return 0
	}
	var sum float64
	for _, a := range f.Agents {
		sum += a.HomeostasisError()
	}
	return sum / float64(len(f.Agents))
}

// BestAgentForTask finds the agent most capable of handling the given requirements.
func (f *Fleet) BestAgentForTask(required []string, threshold float64) (MatchResult, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if len(f.Agents) == 0 {
		return MatchResult{}, fmt.Errorf("fleet %s has no agents", f.ID)
	}

	var candidates []MatchResult
	for _, a := range f.Agents {
		caps := a.ListCapabilities()
		result := Match(a.ID, caps, required, threshold)
		candidates = append(candidates, result)
	}

	best, ok := BestMatch(candidates)
	if !ok {
		return MatchResult{}, fmt.Errorf("no matching agent found")
	}
	return best, nil
}

// SortAgentsByCapability sorts agents by total capability score descending.
func (f *Fleet) SortAgentsByCapability() []*Agent {
	agents := f.ListAgents()
	sort.Slice(agents, func(i, j int) bool {
		return agents[i].TotalCapabilityScore() > agents[j].TotalCapabilityScore()
	})
	return agents
}

// RebalanceBudgets redistributes eta proportionally so that every agent
// has the same eta fraction of its total budget.
func (f *Fleet) RebalanceBudgets() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	var totalEta, totalBudget float64
	var withBudget []*Agent
	for _, a := range f.Agents {
		if a.Budget != nil {
			withBudget = append(withBudget, a)
			totalEta += a.Budget.Budget.Eta
			totalBudget += a.Budget.Budget.Total
		}
	}

	if len(withBudget) == 0 || totalBudget == 0 {
		return nil
	}

	// Target eta fraction for each agent = fleet average
	targetFraction := totalEta / totalBudget
	for _, a := range withBudget {
		targetEta := a.Budget.Budget.Total * targetFraction
		delta := targetEta - a.Budget.Budget.Eta
		a.Budget.Budget.Eta += delta
		a.Budget.Budget.Gamma -= delta
		// Clamp to non-negative
		if a.Budget.Budget.Gamma < 0 {
			a.Budget.Budget.Eta += a.Budget.Budget.Gamma
			a.Budget.Budget.Gamma = 0
		}
		if a.Budget.Budget.Eta < 0 {
			a.Budget.Budget.Gamma += a.Budget.Budget.Eta
			a.Budget.Budget.Eta = 0
		}
	}
	return nil
}

// helper: stable slice of agents for indexing
func (f *Fleet) agentSlice() []*Agent {
	out := make([]*Agent, 0, len(f.Agents))
	for _, a := range f.Agents {
		out = append(out, a)
	}
	// Sort by ID for determinism
	sort.Slice(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})
	return out
}

package siruntime

import (
	"fmt"
	"math"
	"sort"
	"sync"
)

// Fleet is a collection of agents with shared spectral and conservation infrastructure.
type Fleet struct {
	ID        string
	Agents    map[string]*Agent
	Budgets   map[string]*AgentBudget
	Adjacency *AdjacencyMatrix
	mu        sync.RWMutex
}

// NewFleet creates a fleet with the given identifier.
func NewFleet(id string) *Fleet {
	return &Fleet{
		ID:      id,
		Agents:  make(map[string]*Agent),
		Budgets: make(map[string]*AgentBudget),
	}
}

// AddAgent registers an agent and optionally its budget.
func (f *Fleet) AddAgent(a *Agent, budget *Budget) error {
	if a == nil || a.ID == "" {
		return fmt.Errorf("agent cannot be nil and must have an ID")
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, exists := f.Agents[a.ID]; exists {
		return fmt.Errorf("agent %s already in fleet %s", a.ID, f.ID)
	}
	f.Agents[a.ID] = a
	if budget != nil {
		f.Budgets[a.ID] = &AgentBudget{AgentID: a.ID, Budget: budget}
	}
	f.Adjacency = nil
	return nil
}

// RemoveAgent unregisters an agent and its budget.
func (f *Fleet) RemoveAgent(id string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.Agents[id]; !ok {
		return false
	}
	delete(f.Agents, id)
	delete(f.Budgets, id)
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

// ListAgents returns all agents sorted by ID.
func (f *Fleet) ListAgents() []*Agent {
	f.mu.RLock()
	defer f.mu.RUnlock()
	out := make([]*Agent, 0, len(f.Agents))
	for _, a := range f.Agents {
		out = append(out, a)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// AgentCount returns the number of agents.
func (f *Fleet) AgentCount() int {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return len(f.Agents)
}

// BuildAdjacencyMatrix constructs an affinity matrix from agent states.
func (f *Fleet) BuildAdjacencyMatrix(affinity func(a, b *Agent) float64) (*AdjacencyMatrix, error) {
	agents := f.ListAgents()
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

// SpectralRank ranks agents by eigenvector centrality.
func (f *Fleet) SpectralRank() ([]RankedAgent, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.Adjacency == nil {
		agents := f.sortedAgents()
		n := len(agents)
		m := NewAdjacencyMatrix(n)
		for i := 0; i < n; i++ {
			for j := i; j < n; j++ {
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

// ConservationAudit verifies gamma+eta==total for all budgeted agents.
func (f *Fleet) ConservationAudit() AuditResult {
	f.mu.RLock()
	defer f.mu.RUnlock()
	budgets := make([]*AgentBudget, 0, len(f.Budgets))
	for _, ab := range f.Budgets {
		budgets = append(budgets, ab)
	}
	return Audit(budgets)
}

// HealthReport holds fleet-wide homeostasis diagnostics.
type HealthReport struct {
	AgentErrors map[string]float64
	FleetAvg    float64
	WorstAgent  string
}

// HealthReport computes homeostasis error for every agent.
func (f *Fleet) HealthReport() HealthReport {
	f.mu.RLock()
	defer f.mu.RUnlock()
	rpt := HealthReport{AgentErrors: make(map[string]float64)}
	if len(f.Agents) == 0 {
		return rpt
	}
	var sum float64
	worst := 0.0
	for id, a := range f.Agents {
		err := a.HomeostasisError()
		rpt.AgentErrors[id] = err
		sum += err
		if err > worst {
			worst = err
			rpt.WorstAgent = id
		}
	}
	rpt.FleetAvg = sum / float64(len(f.Agents))
	return rpt
}

// BestAgentForTask finds the agent with the most matching capabilities.
func (f *Fleet) BestAgentForTask(required []string) (MatchResult, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	if len(f.Agents) == 0 {
		return MatchResult{}, fmt.Errorf("fleet %s has no agents", f.ID)
	}
	var candidates []MatchResult
	for _, a := range f.Agents {
		candidates = append(candidates, Match(a.ID, a.ListCapabilities(), required))
	}
	best, ok := Resolve(candidates)
	if !ok {
		return MatchResult{}, fmt.Errorf("no agent matches requirements")
	}
	return best, nil
}

// RebalanceBudgets equalizes the eta fraction across all budgeted agents.
func (f *Fleet) RebalanceBudgets() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	var totalEta, totalBudget float64
	var withBudget []*AgentBudget
	for _, ab := range f.Budgets {
		withBudget = append(withBudget, ab)
		totalEta += ab.Budget.Eta
		totalBudget += ab.Budget.Total
	}
	if len(withBudget) == 0 || totalBudget == 0 {
		return nil
	}
	targetFraction := totalEta / totalBudget
	for _, ab := range withBudget {
		targetEta := ab.Budget.Total * targetFraction
		delta := targetEta - ab.Budget.Eta
		ab.Budget.Eta += delta
		ab.Budget.Gamma -= delta
		if ab.Budget.Gamma < 0 {
			ab.Budget.Eta += ab.Budget.Gamma
			ab.Budget.Gamma = 0
		}
		if ab.Budget.Eta < 0 {
			ab.Budget.Gamma += ab.Budget.Eta
			ab.Budget.Eta = 0
		}
	}
	return nil
}

func (f *Fleet) sortedAgents() []*Agent {
	out := make([]*Agent, 0, len(f.Agents))
	for _, a := range f.Agents {
		out = append(out, a)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

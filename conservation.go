package siruntime

import (
	"fmt"
	"sync"
)

// Budget represents a conserved resource pool.
// Invariant: Gamma + Eta == Total.
type Budget struct {
	Total float64
	Gamma float64
	Eta   float64
	mu    sync.RWMutex
}

// NewBudget creates a budget with the given total.
func NewBudget(total float64) *Budget {
	return &Budget{Total: total, Gamma: 0, Eta: total}
}

// Allocate sets gamma and eta, enforcing gamma+eta==total.
func (b *Budget) Allocate(gamma, eta float64) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if gamma < 0 || eta < 0 {
		return fmt.Errorf("gamma and eta must be non-negative")
	}
	if gamma+eta != b.Total {
		return fmt.Errorf("invariant violated: gamma(%.2f)+eta(%.2f)=%.2f != total(%.2f)", gamma, eta, gamma+eta, b.Total)
	}
	b.Gamma = gamma
	b.Eta = eta
	return nil
}

// Transfer moves amount from Eta to Gamma.
func (b *Budget) Transfer(amount float64) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if amount < 0 {
		return fmt.Errorf("transfer amount must be non-negative")
	}
	if amount > b.Eta {
		return fmt.Errorf("overspend: cannot transfer %.2f from eta %.2f", amount, b.Eta)
	}
	b.Gamma += amount
	b.Eta -= amount
	return nil
}

// Remaining returns Gamma + Eta.
func (b *Budget) Remaining() float64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.Gamma + b.Eta
}

// AgentBudget couples an agent identifier with its budget.
type AgentBudget struct {
	AgentID string
	Budget  *Budget
}

// Allocate is a convenience wrapper.
func (ab *AgentBudget) Allocate(gamma, eta float64) error {
	return ab.Budget.Allocate(gamma, eta)
}

// Transfer moves productive budget from this agent to another.
func (ab *AgentBudget) Transfer(to *AgentBudget, amount float64) error {
	ab.Budget.mu.Lock()
	defer ab.Budget.mu.Unlock()
	to.Budget.mu.Lock()
	defer to.Budget.mu.Unlock()
	if amount < 0 {
		return fmt.Errorf("transfer amount must be non-negative")
	}
	if ab.Budget.Gamma < amount {
		return fmt.Errorf("insufficient gamma in %s", ab.AgentID)
	}
	ab.Budget.Gamma -= amount
	to.Budget.Gamma += amount
	return nil
}

// AuditResult holds a fleet-wide budget audit outcome.
type AuditResult struct {
	Valid      bool
	FleetTotal float64
	FleetGamma float64
	FleetEta   float64
	Violations []string
}

// Audit verifies gamma+eta==total for every agent budget.
func Audit(budgets []*AgentBudget) AuditResult {
	result := AuditResult{Valid: true}
	for _, ab := range budgets {
		ab.Budget.mu.RLock()
		g, e, tot := ab.Budget.Gamma, ab.Budget.Eta, ab.Budget.Total
		ab.Budget.mu.RUnlock()
		result.FleetTotal += tot
		result.FleetGamma += g
		result.FleetEta += e
		if g+e != tot {
			result.Valid = false
			result.Violations = append(result.Violations, fmt.Sprintf("agent %s: gamma(%.2f)+eta(%.2f)!=total(%.2f)", ab.AgentID, g, e, tot))
		}
		if g < 0 || e < 0 {
			result.Valid = false
			result.Violations = append(result.Violations, fmt.Sprintf("agent %s: negative gamma(%.2f) or eta(%.2f)", ab.AgentID, g, e))
		}
	}
	return result
}

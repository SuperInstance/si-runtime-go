// Package siruntime provides the SuperInstance unified runtime.
// The conservation package models budget invariants using a physics-inspired
// analogy: total budget C is split into productive spend gamma and overhead eta.
package siruntime

import (
	"fmt"
	"sync"
)

// Budget models a conserved resource pool.
// The invariant gamma + eta == total must hold at all times.
type Budget struct {
	Total float64 // C: total budget ceiling
	Gamma float64 // γ: productive spend (doing useful work)
	Eta   float64 // η: overhead / waste / idle capacity
	mu    sync.RWMutex
}

// NewBudget creates a budget with the given total ceiling.
// Initially all capacity is overhead (eta = total, gamma = 0).
func NewBudget(total float64) *Budget {
	return &Budget{
		Total: total,
		Gamma: 0,
		Eta:   total,
	}
}

// Allocate sets gamma and eta directly, enforcing the invariant.
// Returns an error if gamma + eta != total.
func (b *Budget) Allocate(gamma, eta float64) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if gamma < 0 || eta < 0 {
		return fmt.Errorf("gamma and eta must be non-negative: gamma=%.2f eta=%.2f", gamma, eta)
	}
	if gamma+eta != b.Total {
		return fmt.Errorf("invariant violated: gamma(%.2f) + eta(%.2f) = %.2f != total(%.2f)", gamma, eta, gamma+eta, b.Total)
	}
	b.Gamma = gamma
	b.Eta = eta
	return nil
}

// Transfer moves amount from this budget's eta to its gamma.
// Models converting idle capacity into productive work.
func (b *Budget) Transfer(amount float64) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if amount < 0 {
		return fmt.Errorf("transfer amount must be non-negative: %.2f", amount)
	}
	if amount > b.Eta {
		return fmt.Errorf("overspend: cannot transfer %.2f from eta %.2f", amount, b.Eta)
	}
	b.Gamma += amount
	b.Eta -= amount
	return nil
}

// Overspend attempts to spend amount from gamma. If gamma is insufficient,
// the shortfall is taken from eta (breaking the ideal split but preserving
// the total ceiling). Returns the actual shortfall.
func (b *Budget) Overspend(amount float64) (shortfall float64, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if amount < 0 {
		return 0, fmt.Errorf("spend amount must be non-negative: %.2f", amount)
	}
	if amount <= b.Gamma {
		b.Gamma -= amount
		return 0, nil
	}
	shortfall = amount - b.Gamma
	b.Gamma = 0
	if shortfall > b.Eta {
		return 0, fmt.Errorf("catastrophic overspend: need %.2f, have %.2f total", amount, b.Total)
	}
	b.Eta -= shortfall
	return shortfall, nil
}

// Remaining returns the sum of gamma and eta (should equal total).
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

// Transfer moves budget between two agents. The total fleet budget
// is conserved: source loses amount, destination gains amount.
func Transfer(from, to *AgentBudget, amount float64) error {
	if from.Budget.Gamma < amount {
		return fmt.Errorf("insufficient gamma in %s: need %.2f, have %.2f", from.AgentID, amount, from.Budget.Gamma)
	}
	from.Budget.Gamma -= amount
	to.Budget.Gamma += amount
	return nil
}

// Audit verifies that for every agent budget, gamma + eta == total.
// It also computes fleet-wide aggregates.
type AuditResult struct {
	Valid        bool
	FleetTotal   float64
	FleetGamma   float64
	FleetEta     float64
	Violations   []string
}

// Audit checks the conservation invariant across all agent budgets.
func Audit(budgets []*AgentBudget) AuditResult {
	result := AuditResult{Valid: true}
	for _, ab := range budgets {
		ab.Budget.mu.RLock()
		gamma := ab.Budget.Gamma
		eta := ab.Budget.Eta
		total := ab.Budget.Total
		ab.Budget.mu.RUnlock()

		result.FleetTotal += total
		result.FleetGamma += gamma
		result.FleetEta += eta

		if gamma+eta != total {
			result.Valid = false
			result.Violations = append(result.Violations,
				fmt.Sprintf("agent %s: gamma(%.2f) + eta(%.2f) != total(%.2f)", ab.AgentID, gamma, eta, total))
		}
		if gamma < 0 || eta < 0 {
			result.Valid = false
			result.Violations = append(result.Violations,
				fmt.Sprintf("agent %s: negative gamma(%.2f) or eta(%.2f)", ab.AgentID, gamma, eta))
		}
	}
	return result
}

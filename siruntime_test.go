package siruntime

import (
	"fmt"
	"math"
	"testing"
)

// --- Budget tests ---

func TestNewBudget(t *testing.T) {
	b := NewBudget(1000)
	if b.Total != 1000 {
		t.Errorf("expected total 1000, got %f", b.Total)
	}
	if b.Gamma != 0 {
		t.Errorf("expected gamma 0, got %f", b.Gamma)
	}
	if b.Eta != 1000 {
		t.Errorf("expected eta 1000, got %f", b.Eta)
	}
}

func TestBudgetAllocateInvariant(t *testing.T) {
	b := NewBudget(1000)
	if err := b.Allocate(600, 400); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.Gamma != 600 || b.Eta != 400 {
		t.Errorf("allocate failed: gamma=%f eta=%f", b.Gamma, b.Eta)
	}
	if err := b.Allocate(600, 401); err == nil {
		t.Error("expected invariant violation error")
	}
}

func TestBudgetTransfer(t *testing.T) {
	b := NewBudget(1000)
	b.Allocate(200, 800)
	if err := b.Transfer(300); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.Gamma != 500 || b.Eta != 500 {
		t.Errorf("transfer failed: gamma=%f eta=%f", b.Gamma, b.Eta)
	}
	if err := b.Transfer(600); err == nil {
		t.Error("expected overspend error")
	}
}

func TestBudgetOverspend(t *testing.T) {
	b := NewBudget(1000)
	b.Allocate(300, 700)
	shortfall, err := b.Overspend(500)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if shortfall != 200 {
		t.Errorf("expected shortfall 200, got %f", shortfall)
	}
	if b.Gamma != 0 {
		t.Errorf("expected gamma 0, got %f", b.Gamma)
	}
	if b.Eta != 500 {
		t.Errorf("expected eta 500, got %f", b.Eta)
	}
	_, err = b.Overspend(1000)
	if err == nil {
		t.Error("expected catastrophic overspend error")
	}
}

func TestTransferBetweenAgents(t *testing.T) {
	a := &AgentBudget{AgentID: "A", Budget: NewBudget(1000)}
	b := &AgentBudget{AgentID: "B", Budget: NewBudget(1000)}
	a.Budget.Allocate(500, 500)
	b.Budget.Allocate(200, 800)

	if err := Transfer(a, b, 100); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Budget.Gamma != 400 || b.Budget.Gamma != 300 {
		t.Errorf("transfer failed: A.gamma=%f B.gamma=%f", a.Budget.Gamma, b.Budget.Gamma)
	}
}

func TestAudit(t *testing.T) {
	a := &AgentBudget{AgentID: "A", Budget: NewBudget(100)}
	b := &AgentBudget{AgentID: "B", Budget: NewBudget(100)}
	a.Budget.Allocate(60, 40)
	b.Budget.Allocate(30, 70)

	result := Audit([]*AgentBudget{a, b})
	if !result.Valid {
		t.Errorf("expected valid audit, got violations: %v", result.Violations)
	}
	if result.FleetTotal != 200 {
		t.Errorf("expected fleet total 200, got %f", result.FleetTotal)
	}
	if result.FleetGamma != 90 {
		t.Errorf("expected fleet gamma 90, got %f", result.FleetGamma)
	}
}

// --- Spectral tests ---

func TestAdjacencyMatrix(t *testing.T) {
	m := NewAdjacencyMatrix(3)
	if err := m.Set(0, 1, 0.5); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v, _ := m.Get(1, 0)
	if v != 0.5 {
		t.Errorf("expected symmetric value 0.5, got %f", v)
	}
	_, err := m.Get(3, 0)
	if err == nil {
		t.Error("expected out of bounds error")
	}
}

func TestPowerIteration(t *testing.T) {
	m := NewAdjacencyMatrix(2)
	m.Set(0, 0, 2)
	m.Set(1, 1, 1)
	m.Set(0, 1, 0)

	pair, err := PowerIteration(m, 100, 1e-10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if math.Abs(pair.Value-2.0) > 1e-6 {
		t.Errorf("expected dominant eigenvalue ~2.0, got %f", pair.Value)
	}
}

func TestSpectralRank(t *testing.T) {
	m := NewAdjacencyMatrix(3)
	m.Set(0, 0, 1)
	m.Set(0, 1, 1)
	m.Set(0, 2, 1)
	m.Set(1, 1, 1)
	m.Set(1, 2, 0)
	m.Set(2, 2, 1)

	ranked, err := SpectralRank(m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ranked) != 3 {
		t.Errorf("expected 3 ranked agents, got %d", len(ranked))
	}
	if ranked[0].Index != 0 {
		t.Errorf("expected agent 0 to be top ranked, got %d", ranked[0].Index)
	}
}

// --- Capability tests ---

func TestCapabilityRegistry(t *testing.T) {
	r := NewCapabilityRegistry()
	c := Capability{Name: "compute", Version: "1.0", Score: 0.9}
	if err := r.Register(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok := r.Get("compute")
	if !ok || got.Score != 0.9 {
		t.Errorf("expected capability compute with score 0.9, got %v", got)
	}
	if err := r.Register(Capability{Name: "", Score: 0.5}); err == nil {
		t.Error("expected error for empty name")
	}
}

func TestMatch(t *testing.T) {
	caps := []Capability{
		{Name: "read", Score: 1.0},
		{Name: "write", Score: 0.5},
	}
	result := Match("agent1", caps, []string{"read", "write"}, 0.6)
	if result.Score != 0.75 {
		t.Errorf("expected score 0.75, got %f", result.Score)
	}
	if len(result.Matched) != 1 || result.Matched[0] != "read" {
		t.Errorf("expected read matched, got %v", result.Matched)
	}
	if len(result.Partial) != 1 || result.Partial[0] != "write" {
		t.Errorf("expected write partial, got %v", result.Partial)
	}
}

func TestBestMatch(t *testing.T) {
	candidates := []MatchResult{
		{AgentID: "A", Score: 0.5},
		{AgentID: "B", Score: 0.9},
		{AgentID: "C", Score: 0.7},
	}
	best, ok := BestMatch(candidates)
	if !ok || best.AgentID != "B" {
		t.Errorf("expected best agent B, got %v", best)
	}
}

// --- Cell tests ---

func TestCellUpdate(t *testing.T) {
	c := NewCell("c1", 10.0, 20.0)
	c.Update(0.0, 0.5)
	if math.Abs(c.State-15.0) > 1e-9 {
		t.Errorf("expected state 15.0, got %f", c.State)
	}
}

func TestGridEquilibrium(t *testing.T) {
	g := NewGrid(3, 3, 0.0, 10.0)
	g.WireNeighbors()
	for i := 0; i < 50; i++ {
		g.UpdateAll(0.2, 0.1)
	}
	if !g.EquilibriumCheck(0.5) {
		t.Errorf("expected grid near equilibrium after 50 steps, variance=%f", g.Variance())
	}
}

// --- Agent tests ---

func TestAgentHomeostasis(t *testing.T) {
	a := NewAgent("test-agent")
	a.SetState("energy", 50.0)
	a.SetHomeostasis("energy", 100.0)
	a.UpdateHomeostasis(0.1)
	v, _ := a.GetState("energy")
	if math.Abs(v-55.0) > 1e-9 {
		t.Errorf("expected energy 55.0, got %f", v)
	}
	err := a.HomeostasisError()
	if math.Abs(err-45.0) > 1e-9 {
		t.Errorf("expected error 45.0, got %f", err)
	}
}

func TestAgentCapability(t *testing.T) {
	a := NewAgent("agent-x")
	if err := a.AddCapability(Capability{Name: "fly", Score: 0.8}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := a.AddCapability(Capability{Name: "fly", Score: 0.9}); err == nil {
		t.Error("expected duplicate capability error")
	}
	if !a.RemoveCapability("fly") {
		t.Error("expected fly to be removed")
	}
	if a.RemoveCapability("swim") {
		t.Error("expected swim removal to fail")
	}
}

// --- Fleet tests ---

func TestFleetAddRemove(t *testing.T) {
	f := NewFleet("test-fleet")
	a := NewAgent("a1")
	if err := f.AddAgent(a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.AgentCount() != 1 {
		t.Errorf("expected 1 agent, got %d", f.AgentCount())
	}
	if !f.RemoveAgent("a1") {
		t.Error("expected a1 to be removed")
	}
	if f.AgentCount() != 0 {
		t.Errorf("expected 0 agents, got %d", f.AgentCount())
	}
}

func TestFleetConservationAudit(t *testing.T) {
	f := NewFleet("fleet-a")
	a1 := NewAgent("x")
	a1.AttachBudget(100)
	a1.Budget.Budget.Allocate(60, 40)
	a2 := NewAgent("y")
	a2.AttachBudget(100)
	a2.Budget.Budget.Allocate(30, 70)
	f.AddAgent(a1)
	f.AddAgent(a2)

	result := f.ConservationAudit()
	if !result.Valid {
		t.Errorf("expected valid audit, got %v", result.Violations)
	}
	if result.FleetTotal != 200 {
		t.Errorf("expected fleet total 200, got %f", result.FleetTotal)
	}
}

func TestFleetSpectralRank(t *testing.T) {
	f := NewFleet("fleet-b")
	for i := 0; i < 3; i++ {
		a := NewAgent(fmt.Sprintf("agent-%d", i))
		a.SetState("workload", float64(i*10))
		f.AddAgent(a)
	}
	ranked, err := f.SpectralRank()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ranked) != 3 {
		t.Errorf("expected 3 ranked agents, got %d", len(ranked))
	}
}

func TestFleetBestAgentForTask(t *testing.T) {
	f := NewFleet("fleet-c")
	a1 := NewAgent("planner")
	a1.AddCapability(Capability{Name: "plan", Score: 0.9})
	a2 := NewAgent("coder")
	a2.AddCapability(Capability{Name: "code", Score: 0.8})
	f.AddAgent(a1)
	f.AddAgent(a2)

	result, err := f.BestAgentForTask([]string{"plan"}, 0.5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AgentID != "planner" {
		t.Errorf("expected planner, got %s", result.AgentID)
	}
}

func TestFleetRebalanceBudgets(t *testing.T) {
	f := NewFleet("fleet-d")
	a1 := NewAgent("rich")
	a1.AttachBudget(100)
	a1.Budget.Budget.Allocate(90, 10)
	a2 := NewAgent("poor")
	a2.AttachBudget(100)
	a2.Budget.Budget.Allocate(10, 90)
	f.AddAgent(a1)
	f.AddAgent(a2)

	f.RebalanceBudgets()
	if a1.Budget.Budget.Eta != 50.0 {
		t.Errorf("expected rich eta 50, got %f", a1.Budget.Budget.Eta)
	}
	if a2.Budget.Budget.Eta != 50.0 {
		t.Errorf("expected poor eta 50, got %f", a2.Budget.Budget.Eta)
	}
}

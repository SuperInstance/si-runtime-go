package siruntime

import (
	"testing"
)

func TestFleetAddRemove(t *testing.T) {
	f := NewFleet("test")
	tests := []struct {
		name  string
		id    string
		add   bool
		want  int
	}{
		{"add a1", "a1", true, 1},
		{"add a2", "a2", true, 2},
		{"duplicate a1", "a1", true, 2},
		{"remove a1", "a1", false, 1},
		{"remove unknown", "x", false, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.add {
				f.AddAgent(NewAgent(tt.id), nil)
			} else {
				f.RemoveAgent(tt.id)
			}
			if f.AgentCount() != tt.want {
				t.Errorf("count=%d want %d", f.AgentCount(), tt.want)
			}
		})
	}
}

func TestFleetConservationAudit(t *testing.T) {
	f := NewFleet("f1")
	a1 := NewAgent("x")
	a2 := NewAgent("y")
	f.AddAgent(a1, NewBudget(100))
	f.AddAgent(a2, NewBudget(100))
	f.Budgets["x"].Allocate(60, 40)
	f.Budgets["y"].Allocate(30, 70)

	result := f.ConservationAudit()
	if !result.Valid {
		t.Errorf("expected valid audit, got %v", result.Violations)
	}
	if result.FleetTotal != 200 {
		t.Errorf("fleetTotal=%.0f want 200", result.FleetTotal)
	}
}

func TestFleetSpectralRank(t *testing.T) {
	f := NewFleet("f2")
	for i := 0; i < 3; i++ {
		a := NewAgent(string(rune('a' + i)))
		a.SetState("workload", float64(i*10))
		f.AddAgent(a, nil)
	}
	ranked, err := f.SpectralRank()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ranked) != 3 {
		t.Errorf("expected 3 agents, got %d", len(ranked))
	}
}

func TestFleetBestAgentForTask(t *testing.T) {
	f := NewFleet("f3")
	a1 := NewAgent("planner")
	a1.AddCapability("plan")
	a2 := NewAgent("coder")
	a2.AddCapability("code")
	f.AddAgent(a1, nil)
	f.AddAgent(a2, nil)

	result, err := f.BestAgentForTask([]string{"plan"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AgentID != "planner" || result.Score != 1.0 {
		t.Errorf("expected planner score=1.0, got %s score=%.2f", result.AgentID, result.Score)
	}
}

func TestFleetRebalanceBudgets(t *testing.T) {
	f := NewFleet("f4")
	a1 := NewAgent("rich")
	a2 := NewAgent("poor")
	f.AddAgent(a1, NewBudget(100))
	f.AddAgent(a2, NewBudget(100))
	f.Budgets["rich"].Allocate(90, 10)
	f.Budgets["poor"].Allocate(10, 90)

	f.RebalanceBudgets()
	if f.Budgets["rich"].Budget.Eta != 50.0 {
		t.Errorf("rich eta=%.0f want 50", f.Budgets["rich"].Budget.Eta)
	}
	if f.Budgets["poor"].Budget.Eta != 50.0 {
		t.Errorf("poor eta=%.0f want 50", f.Budgets["poor"].Budget.Eta)
	}
}

func TestFleetHealthReport(t *testing.T) {
	f := NewFleet("f5")
	a1 := NewAgent("healthy")
	a1.SetState("temp", 36.5)
	a1.SetHomeostasis("temp", 37.0)
	a2 := NewAgent("sick")
	a2.SetState("temp", 40.0)
	a2.SetHomeostasis("temp", 37.0)
	f.AddAgent(a1, nil)
	f.AddAgent(a2, nil)

	rpt := f.HealthReport()
	if rpt.WorstAgent != "sick" {
		t.Errorf("expected worst=sick, got %s", rpt.WorstAgent)
	}
	if rpt.FleetAvg <= 0 {
		t.Error("expected positive fleet average error")
	}
}

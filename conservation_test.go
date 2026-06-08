package siruntime

import "testing"

func TestBudgetAllocateInvariant(t *testing.T) {
	tests := []struct {
		name    string
		gamma   float64
		eta     float64
		wantErr bool
	}{
		{"valid split", 600, 400, false},
		{"all gamma", 1000, 0, false},
		{"all eta", 0, 1000, false},
		{"invalid oversum", 600, 500, true},
		{"negative gamma", -10, 1010, true},
	}
	b := NewBudget(1000)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := b.Allocate(tt.gamma, tt.eta)
			if (err != nil) != tt.wantErr {
				t.Errorf("Allocate(%.0f,%.0f) error=%v wantErr=%v", tt.gamma, tt.eta, err, tt.wantErr)
			}
			if err == nil && b.Remaining() != 1000 {
				t.Errorf("remaining=%.0f, want 1000", b.Remaining())
			}
		})
	}
}

func TestBudgetTransfer(t *testing.T) {
	tests := []struct {
		name    string
		amount  float64
		wantErr bool
		wantG   float64
		wantE   float64
	}{
		{"small transfer", 100, false, 300, 700},
		{"exact eta", 800, false, 1000, 0},
		{"overspend", 801, true, 1000, 0},
		{"negative", -1, true, 1000, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBudget(1000)
			b.Allocate(200, 800)
			err := b.Transfer(tt.amount)
			if (err != nil) != tt.wantErr {
				t.Errorf("Transfer(%.0f) error=%v wantErr=%v", tt.amount, err, tt.wantErr)
			}
			if err == nil {
				if b.Gamma != tt.wantG || b.Eta != tt.wantE {
					t.Errorf("gamma=%.0f eta=%.0f, want gamma=%.0f eta=%.0f", b.Gamma, b.Eta, tt.wantG, tt.wantE)
				}
			}
		})
	}
}

func TestAgentBudgetTransfer(t *testing.T) {
	tests := []struct {
		name    string
		amount  float64
		wantErr bool
		wantA   float64
		wantB   float64
	}{
		{"valid", 50, false, 30, 80},
		{"too much", 100, true, 80, 30},
		{"negative", -1, true, 80, 30},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alice := &AgentBudget{AgentID: "alice", Budget: NewBudget(100)}
			bob := &AgentBudget{AgentID: "bob", Budget: NewBudget(100)}
			alice.Allocate(80, 20)
			bob.Allocate(30, 70)
			err := alice.Transfer(bob, tt.amount)
			if (err != nil) != tt.wantErr {
				t.Errorf("Transfer error=%v wantErr=%v", err, tt.wantErr)
			}
			if err == nil {
				if alice.Budget.Gamma != tt.wantA || bob.Budget.Gamma != tt.wantB {
					t.Errorf("alice=%.0f bob=%.0f, want %.0f/%.0f", alice.Budget.Gamma, bob.Budget.Gamma, tt.wantA, tt.wantB)
				}
			}
		})
	}
}

func TestAudit(t *testing.T) {
	tests := []struct {
		name      string
		budgets   []*AgentBudget
		wantValid bool
		wantTotal float64
		wantGamma float64
		wantEta   float64
	}{
		{
			name: "valid fleet",
			budgets: []*AgentBudget{
				{AgentID: "a", Budget: NewBudget(100)},
				{AgentID: "b", Budget: NewBudget(200)},
			},
			wantValid: true,
			wantTotal: 300,
			wantGamma: 0,
			wantEta:   300,
		},
		{
			name: "allocated fleet",
			budgets: func() []*AgentBudget {
				a := &AgentBudget{AgentID: "a", Budget: NewBudget(100)}
				b := &AgentBudget{AgentID: "b", Budget: NewBudget(100)}
				a.Allocate(60, 40)
				b.Allocate(30, 70)
				return []*AgentBudget{a, b}
			}(),
			wantValid: true,
			wantTotal: 200,
			wantGamma: 90,
			wantEta:   110,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Audit(tt.budgets)
			if result.Valid != tt.wantValid {
				t.Errorf("valid=%v want %v", result.Valid, tt.wantValid)
			}
			if result.FleetTotal != tt.wantTotal {
				t.Errorf("fleetTotal=%.0f want %.0f", result.FleetTotal, tt.wantTotal)
			}
			if result.FleetGamma != tt.wantGamma {
				t.Errorf("fleetGamma=%.0f want %.0f", result.FleetGamma, tt.wantGamma)
			}
			if result.FleetEta != tt.wantEta {
				t.Errorf("fleetEta=%.0f want %.0f", result.FleetEta, tt.wantEta)
			}
		})
	}
}

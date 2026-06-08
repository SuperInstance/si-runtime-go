package siruntime

import "testing"

func TestRegistry(t *testing.T) {
	r := NewRegistry()
	tests := []struct {
		name    string
		cap     Capability
		wantErr bool
	}{
		{"valid", Capability{Name: "compute", Version: "1.0"}, false},
		{"empty name", Capability{Name: "   ", Version: "1.0"}, true},
		{"another valid", Capability{Name: "storage", Version: "2.0"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := r.Register(tt.cap)
			if (err != nil) != tt.wantErr {
				t.Errorf("Register error=%v wantErr=%v", err, tt.wantErr)
			}
		})
	}
	if len(r.List()) != 2 {
		t.Errorf("expected 2 capabilities, got %d", len(r.List()))
	}
	r.Remove("compute")
	if _, ok := r.Get("compute"); ok {
		t.Error("expected compute to be removed")
	}
}

func TestMatch(t *testing.T) {
	tests := []struct {
		name     string
		agentCaps []string
		required []string
		wantScore float64
		wantMatch int
		wantMiss  int
	}{
		{"perfect", []string{"read", "write", "delete"}, []string{"read", "write"}, 1.0, 2, 0},
		{"partial", []string{"read"}, []string{"read", "write"}, 0.5, 1, 1},
		{"none", []string{"read"}, []string{"write"}, 0.0, 0, 1},
		{"empty required", []string{"read"}, []string{}, 1.0, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Match("a1", tt.agentCaps, tt.required)
			if result.Score != tt.wantScore {
				t.Errorf("score=%.2f want %.2f", result.Score, tt.wantScore)
			}
			if len(result.Matched) != tt.wantMatch {
				t.Errorf("matched=%d want %d", len(result.Matched), tt.wantMatch)
			}
			if len(result.Missing) != tt.wantMiss {
				t.Errorf("missing=%d want %d", len(result.Missing), tt.wantMiss)
			}
		})
	}
}

func TestResolve(t *testing.T) {
	candidates := []MatchResult{
		{AgentID: "a", Score: 0.3},
		{AgentID: "b", Score: 0.9},
		{AgentID: "c", Score: 0.7},
	}
	best, ok := Resolve(candidates)
	if !ok || best.AgentID != "b" {
		t.Errorf("expected best=b, got %v", best)
	}
	_, ok = Resolve([]MatchResult{})
	if ok {
		t.Error("expected no best from empty candidates")
	}
}

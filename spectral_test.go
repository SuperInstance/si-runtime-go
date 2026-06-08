package siruntime

import (
	"math"
	"testing"
)

func TestAdjacencyMatrix(t *testing.T) {
	m := NewAdjacencyMatrix(3)
	tests := []struct {
		name  string
		i, j  int
		val   float64
		want  float64
		wantE bool
	}{
		{"set 0,1", 0, 1, 0.5, 0.5, false},
		{"symmetric 1,0", 1, 0, 0, 0.5, false},
		{"set 1,1", 1, 1, 2.0, 2.0, false},
		{"oob", 3, 0, 0, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.val != 0 {
				if err := m.Set(tt.i, tt.j, tt.val); err != nil {
					t.Fatalf("Set error: %v", err)
				}
			}
			v, err := m.Get(tt.i, tt.j)
			if (err != nil) != tt.wantE {
				t.Errorf("Get(%d,%d) error=%v wantErr=%v", tt.i, tt.j, err, tt.wantE)
			}
			if err == nil && v != tt.want {
				t.Errorf("Get(%d,%d)=%.2f want %.2f", tt.i, tt.j, v, tt.want)
			}
		})
	}
}

func TestPowerIteration(t *testing.T) {
	m := NewAdjacencyMatrix(2)
	m.Set(0, 0, 2)
	m.Set(1, 1, 1)
	pair, err := PowerIteration(m, 100, 1e-10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if math.Abs(pair.Value-2.0) > 1e-6 {
		t.Errorf("eigenvalue=%.6f want ~2.0", pair.Value)
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
		t.Fatalf("expected 3 agents, got %d", len(ranked))
	}
	if ranked[0].Index != 0 {
		t.Errorf("expected agent 0 top ranked, got %d", ranked[0].Index)
	}
	if ranked[0].Centrality <= ranked[1].Centrality {
		t.Error("expected strictly decreasing centrality")
	}
}

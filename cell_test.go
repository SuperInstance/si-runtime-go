package siruntime

import (
	"math"
	"testing"
)

func TestCellUpdate(t *testing.T) {
	tests := []struct {
		name   string
		state  float64
		target float64
		alpha  float64
		beta   float64
		want   float64
	}{
		{"pure homeostasis", 10, 20, 0, 0.5, 15},
		{"no change", 10, 10, 0, 0, 10},
		{"full reach", 10, 20, 0, 1.0, 20},
	}
	cells := []*Cell{NewCell("c0", 0, 0)}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCell("c", tt.state, tt.target)
			c.Update(cells, tt.alpha, tt.beta)
			if math.Abs(c.State-tt.want) > 1e-9 {
				t.Errorf("state=%.2f want %.2f", c.State, tt.want)
			}
		})
	}
}

func TestGridEquilibrium(t *testing.T) {
	g := NewGrid(3, 3, 0.0, 10.0)
	g.WireNeighbors()
	g.Cells[0].State = 100.0
	for i := 0; i < 50; i++ {
		g.UpdateAll(0.3, 0.1)
	}
	if !g.EquilibriumCheck(2.0) {
		t.Errorf("expected near equilibrium, variance=%.2f", g.Variance())
	}
}

func TestCellAddNeighbor(t *testing.T) {
	c := NewCell("c", 0, 0)
	if err := c.AddNeighbor(1); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(c.Neighbors) != 1 {
		t.Errorf("expected 1 neighbor, got %d", len(c.Neighbors))
	}
	if err := c.AddNeighbor(-1); err == nil {
		t.Error("expected error for negative index")
	}
}

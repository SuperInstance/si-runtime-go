package siruntime

import (
	"fmt"
	"math"
)

// Cell represents a discrete unit in a cellular automaton or agent mesh.
// Each cell has an internal state and connections to neighbors.
type Cell struct {
	ID        string
	State     float64
	Target    float64 // homeostatic target value
	Neighbors []*Cell
}

// NewCell creates a cell with the given ID, initial state, and target.
func NewCell(id string, state, target float64) *Cell {
	return &Cell{
		ID:     id,
		State:  state,
		Target: target,
	}
}

// AddNeighbor links two cells bidirectionally.
func (c *Cell) AddNeighbor(other *Cell) error {
	if other == nil {
		return fmt.Errorf("cannot add nil neighbor to cell %s", c.ID)
	}
	if c == other {
		return fmt.Errorf("cell %s cannot be its own neighbor", c.ID)
	}
	c.Neighbors = append(c.Neighbors, other)
	other.Neighbors = append(other.Neighbors, c)
	return nil
}

// AverageNeighborState returns the mean state of all neighbors.
func (c *Cell) AverageNeighborState() float64 {
	if len(c.Neighbors) == 0 {
		return c.State
	}
	sum := 0.0
	for _, n := range c.Neighbors {
		sum += n.State
	}
	return sum / float64(len(c.Neighbors))
}

// Update applies a single step of homeostatic smoothing:
//   state_new = state + alpha * (avg_neighbors - state) + beta * (target - state)
//
// Alpha controls diffusion from neighbors; beta controls attraction to target.
func (c *Cell) Update(alpha, beta float64) {
	if len(c.Neighbors) == 0 {
		c.State += beta * (c.Target - c.State)
		return
	}
	avg := c.AverageNeighborState()
	diffusion := alpha * (avg - c.State)
	homeostasis := beta * (c.Target - c.State)
	c.State += diffusion + homeostasis
}

// Grid is a rectangular lattice of cells.
type Grid struct {
	Width  int
	Height int
	Cells  []*Cell
}

// NewGrid creates a width×height grid of cells, each initialized to
// the given state and target.
func NewGrid(width, height int, state, target float64) *Grid {
	g := &Grid{
		Width:  width,
		Height: height,
		Cells:  make([]*Cell, 0, width*height),
	}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			id := fmt.Sprintf("cell-%d-%d", x, y)
			g.Cells = append(g.Cells, NewCell(id, state, target))
		}
	}
	return g
}

// WireNeighbors connects each cell to its von Neumann neighbors
// (up, down, left, right).
func (g *Grid) WireNeighbors() {
	for y := 0; y < g.Height; y++ {
		for x := 0; x < g.Width; x++ {
			idx := y*g.Width + x
			cell := g.Cells[idx]
			// Right
			if x+1 < g.Width {
				cell.AddNeighbor(g.Cells[idx+1])
			}
			// Down
			if y+1 < g.Height {
				cell.AddNeighbor(g.Cells[idx+g.Width])
			}
		}
	}
}

// GridState returns a flat slice of all cell states.
func (g *Grid) GridState() []float64 {
	states := make([]float64, len(g.Cells))
	for i, c := range g.Cells {
		states[i] = c.State
	}
	return states
}

// UpdateAll runs one synchronous update on every cell.
func (g *Grid) UpdateAll(alpha, beta float64) {
	// Snapshot current states to avoid order-dependent bias
	snapshots := make([]float64, len(g.Cells))
	for i, c := range g.Cells {
		snapshots[i] = c.State
	}
	for i, c := range g.Cells {
		avg := 0.0
		if len(c.Neighbors) > 0 {
			sum := 0.0
			for _, n := range c.Neighbors {
				sum += snapshots[g.indexOf(n)]
			}
			avg = sum / float64(len(c.Neighbors))
		} else {
			avg = snapshots[i]
		}
		diffusion := alpha * (avg - snapshots[i])
		homeostasis := beta * (c.Target - snapshots[i])
		c.State = snapshots[i] + diffusion + homeostasis
	}
}

func (g *Grid) indexOf(cell *Cell) int {
	for i, c := range g.Cells {
		if c == cell {
			return i
		}
	}
	return -1
}

// Variance computes the population variance of cell states.
func (g *Grid) Variance() float64 {
	if len(g.Cells) == 0 {
		return 0
	}
	mean := 0.0
	for _, c := range g.Cells {
		mean += c.State
	}
	mean /= float64(len(g.Cells))
	variance := 0.0
	for _, c := range g.Cells {
		d := c.State - mean
		variance += d * d
	}
	return variance / float64(len(g.Cells))
}

// EquilibriumCheck returns true if all cells are within tolerance of their target.
func (g *Grid) EquilibriumCheck(tolerance float64) bool {
	for _, c := range g.Cells {
		if math.Abs(c.State-c.Target) > tolerance {
			return false
		}
	}
	return true
}

package siruntime

import (
	"fmt"
	"math"
)

// Cell represents a discrete unit in a cellular mesh.
type Cell struct {
	ID        string
	State     float64
	Target    float64
	Neighbors []int
}

// NewCell creates a cell with the given ID, initial state, and target.
func NewCell(id string, state, target float64) *Cell {
	return &Cell{ID: id, State: state, Target: target}
}

// AddNeighbor appends a neighbor index.
func (c *Cell) AddNeighbor(idx int) error {
	if idx < 0 {
		return fmt.Errorf("neighbor index cannot be negative: %d", idx)
	}
	c.Neighbors = append(c.Neighbors, idx)
	return nil
}

// AverageNeighborState computes the mean state of neighbors.
func (c *Cell) AverageNeighborState(cells []*Cell) float64 {
	if len(c.Neighbors) == 0 {
		return c.State
	}
	sum := 0.0
	for _, idx := range c.Neighbors {
		if idx >= 0 && idx < len(cells) {
			sum += cells[idx].State
		}
	}
	return sum / float64(len(c.Neighbors))
}

// Update applies one step of homeostatic smoothing.
func (c *Cell) Update(cells []*Cell, alpha, beta float64) {
	avg := c.AverageNeighborState(cells)
	c.State += alpha*(avg-c.State) + beta*(c.Target-c.State)
}

// Grid is a rectangular lattice of cells.
type Grid struct {
	Width  int
	Height int
	Cells  []*Cell
}

// NewGrid creates a width×height grid with uniform initial state and target.
func NewGrid(width, height int, state, target float64) *Grid {
	g := &Grid{Width: width, Height: height, Cells: make([]*Cell, 0, width*height)}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			g.Cells = append(g.Cells, NewCell(fmt.Sprintf("c-%d-%d", x, y), state, target))
		}
	}
	return g
}

// WireNeighbors connects each cell to its von Neumann neighbors.
func (g *Grid) WireNeighbors() {
	for y := 0; y < g.Height; y++ {
		for x := 0; x < g.Width; x++ {
			idx := y*g.Width + x
			if x+1 < g.Width {
				g.Cells[idx].AddNeighbor(idx + 1)
				g.Cells[idx+1].AddNeighbor(idx)
			}
			if y+1 < g.Height {
				g.Cells[idx].AddNeighbor(idx + g.Width)
				g.Cells[idx+g.Width].AddNeighbor(idx)
			}
		}
	}
}

// UpdateAll performs one synchronous update on every cell.
func (g *Grid) UpdateAll(alpha, beta float64) {
	snapshots := make([]float64, len(g.Cells))
	for i, c := range g.Cells {
		snapshots[i] = c.State
	}
	for i, c := range g.Cells {
		avg := 0.0
		if len(c.Neighbors) > 0 {
			sum := 0.0
			count := 0
			for _, nidx := range c.Neighbors {
				if nidx >= 0 && nidx < len(snapshots) {
					sum += snapshots[nidx]
					count++
				}
			}
			if count > 0 {
				avg = sum / float64(count)
			}
		} else {
			avg = snapshots[i]
		}
		c.State = snapshots[i] + alpha*(avg-snapshots[i]) + beta*(c.Target-snapshots[i])
	}
}

// Variance computes population variance of cell states.
func (g *Grid) Variance() float64 {
	if len(g.Cells) == 0 {
		return 0
	}
	mean := 0.0
	for _, c := range g.Cells {
		mean += c.State
	}
	mean /= float64(len(g.Cells))
	v := 0.0
	for _, c := range g.Cells {
		d := c.State - mean
		v += d * d
	}
	return v / float64(len(g.Cells))
}

// EquilibriumCheck returns true when all cells are within tolerance of target.
func (g *Grid) EquilibriumCheck(tolerance float64) bool {
	for _, c := range g.Cells {
		if math.Abs(c.State-c.Target) > tolerance {
			return false
		}
	}
	return true
}

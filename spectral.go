package siruntime

import (
	"fmt"
	"math"
	"sort"
)

// AdjacencyMatrix is a dense symmetric matrix of agent affinities.
type AdjacencyMatrix struct {
	Data [][]float64
	Size int
}

// NewAdjacencyMatrix creates an n×n zero matrix.
func NewAdjacencyMatrix(n int) *AdjacencyMatrix {
	d := make([][]float64, n)
	for i := range d {
		d[i] = make([]float64, n)
	}
	return &AdjacencyMatrix{Data: d, Size: n}
}

// Set sets both (i,j) and (j,i) to value.
func (m *AdjacencyMatrix) Set(i, j int, value float64) error {
	if i < 0 || i >= m.Size || j < 0 || j >= m.Size {
		return fmt.Errorf("index out of bounds (%d,%d) for size %d", i, j, m.Size)
	}
	m.Data[i][j] = value
	m.Data[j][i] = value
	return nil
}

// Get returns element (i,j).
func (m *AdjacencyMatrix) Get(i, j int) (float64, error) {
	if i < 0 || i >= m.Size || j < 0 || j >= m.Size {
		return 0, fmt.Errorf("index out of bounds (%d,%d) for size %d", i, j, m.Size)
	}
	return m.Data[i][j], nil
}

// Eigenpair holds one eigenvalue and its eigenvector.
type Eigenpair struct {
	Value  float64
	Vector []float64
}

func dot(a, b []float64) float64 {
	sum := 0.0
	for i := range a {
		sum += a[i] * b[i]
	}
	return sum
}

func normalize(v []float64) {
	n := math.Sqrt(dot(v, v))
	if n == 0 {
		return
	}
	for i := range v {
		v[i] /= n
	}
}

// PowerIteration finds the dominant eigenpair of a symmetric matrix.
func PowerIteration(m *AdjacencyMatrix, maxIter int, tol float64) (*Eigenpair, error) {
	n := m.Size
	if n == 0 {
		return nil, fmt.Errorf("empty matrix")
	}
	vec := make([]float64, n)
	for i := range vec {
		vec[i] = 1.0
	}
	normalize(vec)
	var ev float64
	for iter := 0; iter < maxIter; iter++ {
		next := make([]float64, n)
		for i := 0; i < n; i++ {
			for j := 0; j < n; j++ {
				next[i] += m.Data[i][j] * vec[j]
			}
		}
		newEV := dot(next, vec)
		normalize(next)
		if math.Abs(newEV-ev) < tol {
			return &Eigenpair{Value: newEV, Vector: next}, nil
		}
		ev = newEV
		vec = next
	}
	return &Eigenpair{Value: ev, Vector: vec}, nil
}

// RankedAgent pairs an agent index with its eigenvector centrality.
type RankedAgent struct {
	Index      int
	Centrality float64
}

// SpectralRank returns agent indices sorted by eigenvector centrality.
func SpectralRank(m *AdjacencyMatrix) ([]RankedAgent, error) {
	pair, err := PowerIteration(m, 1000, 1e-8)
	if err != nil {
		return nil, err
	}
	ranked := make([]RankedAgent, m.Size)
	for i, v := range pair.Vector {
		ranked[i] = RankedAgent{Index: i, Centrality: math.Abs(v)}
	}
	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].Centrality > ranked[j].Centrality
	})
	return ranked, nil
}

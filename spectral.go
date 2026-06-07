package siruntime

import (
	"fmt"
	"math"
	"sort"
)

// AdjacencyMatrix represents agent-agent affinity as a dense symmetric matrix.
type AdjacencyMatrix struct {
	Data [][]float64
	Size int
}

// NewAdjacencyMatrix creates an n×n zero matrix.
func NewAdjacencyMatrix(n int) *AdjacencyMatrix {
	data := make([][]float64, n)
	for i := range data {
		data[i] = make([]float64, n)
	}
	return &AdjacencyMatrix{Data: data, Size: n}
}

// Set symmetric entry (i,j) and (j,i) to value.
func (m *AdjacencyMatrix) Set(i, j int, value float64) error {
	if i < 0 || i >= m.Size || j < 0 || j >= m.Size {
		return fmt.Errorf("index out of bounds: (%d,%d) for size %d", i, j, m.Size)
	}
	m.Data[i][j] = value
	m.Data[j][i] = value
	return nil
}

// Get returns entry (i,j).
func (m *AdjacencyMatrix) Get(i, j int) (float64, error) {
	if i < 0 || i >= m.Size || j < 0 || j >= m.Size {
		return 0, fmt.Errorf("index out of bounds: (%d,%d) for size %d", i, j, m.Size)
	}
	return m.Data[i][j], nil
}

// FromAgentAffinities builds an adjacency matrix from a function that
// returns the affinity between any two agent indices.
func FromAgentAffinities(n int, affinity func(i, j int) float64) *AdjacencyMatrix {
	m := NewAdjacencyMatrix(n)
	for i := 0; i < n; i++ {
		for j := i; j < n; j++ {
			m.Data[i][j] = affinity(i, j)
			m.Data[j][i] = m.Data[i][j]
		}
	}
	return m
}

// Eigenpair holds one eigenvalue and its corresponding eigenvector.
type Eigenpair struct {
	Value  float64
	Vector []float64
}

// EigenDecomposition holds the top-k eigenpairs of a symmetric matrix.
type EigenDecomposition struct {
	Pairs []Eigenpair
}

// PowerIteration performs power iteration on a matrix to find its
// dominant eigenpair. Returns the eigenvalue and eigenvector.
func PowerIteration(m *AdjacencyMatrix, maxIter int, tol float64) (*Eigenpair, error) {
	n := m.Size
	if n == 0 {
		return nil, fmt.Errorf("empty matrix")
	}

	// Random-ish initial vector (deterministic for reproducibility)
	vec := make([]float64, n)
	for i := range vec {
		vec[i] = 1.0
	}
	normalize(vec)

	var eigenvalue float64
	for iter := 0; iter < maxIter; iter++ {
		// Multiply: newVec = M * vec
		newVec := make([]float64, n)
		for i := 0; i < n; i++ {
			for j := 0; j < n; j++ {
				newVec[i] += m.Data[i][j] * vec[j]
			}
		}

		// Rayleigh quotient for eigenvalue estimate
		newEigenvalue := dot(newVec, vec)
		normalize(newVec)

		if math.Abs(newEigenvalue-eigenvalue) < tol {
			return &Eigenpair{Value: newEigenvalue, Vector: newVec}, nil
		}
		eigenvalue = newEigenvalue
		vec = newVec
	}

	return &Eigenpair{Value: eigenvalue, Vector: vec}, nil
}

// TopKEigenpairs finds the k largest eigenpairs using power iteration
// with Hotelling deflation.
func TopKEigenpairs(m *AdjacencyMatrix, k, maxIter int, tol float64) (*EigenDecomposition, error) {
	if k > m.Size {
		return nil, fmt.Errorf("k=%d exceeds matrix size %d", k, m.Size)
	}

	// Work on a copy because deflation modifies the matrix
	work := NewAdjacencyMatrix(m.Size)
	for i := 0; i < m.Size; i++ {
		copy(work.Data[i], m.Data[i])
	}

	pairs := make([]Eigenpair, 0, k)
	for p := 0; p < k; p++ {
		pair, err := PowerIteration(work, maxIter, tol)
		if err != nil {
			return nil, err
		}
		pairs = append(pairs, *pair)

		// Deflate: A_{i+1} = A_i - λ * v * v^T
		for i := 0; i < work.Size; i++ {
			for j := 0; j < work.Size; j++ {
				work.Data[i][j] -= pair.Value * pair.Vector[i] * pair.Vector[j]
			}
		}
	}

	return &EigenDecomposition{Pairs: pairs}, nil
}

// SpectralRank returns agent indices sorted by their centrality in the
// dominant eigenvector. Higher values mean higher rank.
type RankedAgent struct {
	Index     int
	Centrality float64
}

// SpectralRank computes eigenvector centrality ranking for all agents.
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

// --- helpers ---

func dot(a, b []float64) float64 {
	sum := 0.0
	for i := range a {
		sum += a[i] * b[i]
	}
	return sum
}

func normalize(v []float64) {
	norm := math.Sqrt(dot(v, v))
	if norm == 0 {
		return
	}
	for i := range v {
		v[i] /= norm
	}
}

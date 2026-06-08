# si-runtime-go

**Go runtime for the SuperInstance ecosystem.** Conservation budgets, spectral ranking, capability matching, cellular automata, agent homeostasis, and fleet orchestration — all in pure Go with zero external dependencies.

---

## Quick Start

```bash
git clone https://github.com/SuperInstance/si-runtime-go.git
cd si-runtime-go

# Run all tests
go test -v ./...

# Use in your own project
go get github.com/SuperInstance/si-runtime-go
```

**go.mod:**

```
module github.com/SuperInstance/si-runtime-go
go 1.21
```

Zero external dependencies.

---

## Overview

```go
package main

import (
    "fmt"
    siruntime "github.com/SuperInstance/si-runtime-go"
)

func main() {
    // Create agents
    a1 := siruntime.NewAgent("worker-1")
    a1.SetState("load", 0.8)
    a1.SetHomeostasis("load", 0.5)
    a1.AddCapability("compute")
    a1.AddCapability("network")

    // Create a budget
    budget := siruntime.NewBudget(1000)
    budget.Allocate(600, 400) // gamma=600, eta=400

    // Fleet orchestration
    fleet := siruntime.NewFleet("production")
    fleet.AddAgent(a1, budget)

    // Conservation audit
    result := fleet.ConservationAudit()
    fmt.Printf("Valid: %v, Fleet Total: %.0f\n", result.Valid, result.FleetTotal)

    // Spectral ranking
    ranked, _ := fleet.SpectralRank()
    for _, r := range ranked {
        fmt.Printf("Agent %d: centrality=%.4f\n", r.Index, r.Centrality)
    }
}
```

---

## API Reference

### Budget — `conservation.go`

A conserved resource pool. **Invariant: Gamma + Eta == Total.**

```go
type Budget struct {
    Total float64
    Gamma float64
    Eta   float64
}

func NewBudget(total float64) *Budget
func (b *Budget) Allocate(gamma, eta float64) error
func (b *Budget) Transfer(amount float64) error
func (b *Budget) Remaining() float64
```

**Examples:**

```go
// Create a 1000-unit budget (initially all in eta)
b := siruntime.NewBudget(1000)
// b.Gamma=0, b.Eta=1000

// Allocate: gamma=600, eta=400
err := b.Allocate(600, 400)
if err != nil { panic(err) }
fmt.Printf("Gamma=%.0f Eta=%.0f Total=%.0f\n", b.Gamma, b.Eta, b.Total)
// Gamma=600 Eta=400 Total=1000

// Transfer 100 from Eta to Gamma
err = b.Transfer(100)
// b.Gamma=700, b.Eta=300

// Error: overspend
err = b.Transfer(500)
// err: "overspend: cannot transfer 500 from eta 300"

// Error: invariant violation
err = b.Allocate(600, 500)
// err: "invariant violated: gamma(600.00)+eta(500.00)=1100.00 != total(1000.00)"

// Error: negative values
err = b.Allocate(-10, 1010)
// err: "gamma and eta must be non-negative"
```

### AgentBudget — `conservation.go`

Couples an agent ID with its budget.

```go
type AgentBudget struct {
    AgentID string
    Budget  *Budget
}

func (ab *AgentBudget) Allocate(gamma, eta float64) error
func (ab *AgentBudget) Transfer(to *AgentBudget, amount float64) error
```

**Examples:**

```go
alice := &siruntime.AgentBudget{AgentID: "alice", Budget: siruntime.NewBudget(100)}
bob := &siruntime.AgentBudget{AgentID: "bob", Budget: siruntime.NewBudget(100)}

alice.Allocate(80, 20)  // alice: gamma=80, eta=20
bob.Allocate(30, 70)    // bob: gamma=30, eta=70

// Transfer 50 gamma from alice to bob
err := alice.Transfer(bob, 50)
// alice.Gamma=30, bob.Gamma=80

// Error: insufficient
err = alice.Transfer(bob, 100)
// err: "insufficient gamma in alice"
```

### Audit — `conservation.go`

```go
type AuditResult struct {
    Valid      bool
    FleetTotal float64
    FleetGamma float64
    FleetEta   float64
    Violations []string
}

func Audit(budgets []*AgentBudget) AuditResult
```

**Examples:**

```go
budgets := []*siruntime.AgentBudget{
    {AgentID: "a", Budget: siruntime.NewBudget(100)},
    {AgentID: "b", Budget: siruntime.NewBudget(200)},
}
budgets[0].Allocate(60, 40)
budgets[1].Allocate(120, 80)

result := siruntime.Audit(budgets)
fmt.Printf("Valid: %v\n", result.Valid)           // true
fmt.Printf("Fleet Total: %.0f\n", result.FleetTotal) // 300
fmt.Printf("Fleet Gamma: %.0f\n", result.FleetGamma) // 180
fmt.Printf("Fleet Eta: %.0f\n", result.FleetEta)     // 120
```

---

### Agent — `agent.go`

An autonomous agent with state, capabilities, and homeostatic regulation.

```go
type Agent struct {
    ID           string
    State        map[string]float64
    Capabilities []string
    Homeostasis  map[string]float64
}

func NewAgent(id string) *Agent
func (a *Agent) SetState(key string, value float64)
func (a *Agent) GetState(key string) (float64, bool)
func (a *Agent) SetHomeostasis(key string, target float64)
func (a *Agent) AddCapability(name string) error
func (a *Agent) RemoveCapability(name string) bool
func (a *Agent) ListCapabilities() []string
func (a *Agent) UpdateHomeostasis(rate float64)
func (a *Agent) HomeostasisError() float64
func (a *Agent) String() string
```

**Examples:**

```go
a := siruntime.NewAgent("worker-1")

// Set state variables
a.SetState("cpu", 0.8)
a.SetState("memory", 0.6)
val, ok := a.GetState("cpu")
fmt.Printf("CPU: %.2f (exists: %v)\n", val, ok) // CPU: 0.80 (exists: true)

// Capabilities
a.AddCapability("compute")
a.AddCapability("network")
a.AddCapability("compute") // error: already has "compute"
fmt.Println(a.ListCapabilities()) // [compute network]

a.RemoveCapability("network")
a.RemoveCapability("nonexistent") // false

// Homeostatic regulation
a.SetState("temperature", 72.0)
a.SetHomeostasis("temperature", 65.0)

// Drive toward target at rate 0.1
a.UpdateHomeostasis(0.1)
temp, _ := a.GetState("temperature")
fmt.Printf("Temperature: %.2f\n", temp) // 71.3 (moved 10% toward 65)

// RMS error from targets
err := a.HomeostasisError()
fmt.Printf("Homeostasis error: %.4f\n", err)

// Full convergence
for i := 0; i < 50; i++ {
    a.UpdateHomeostasis(0.1)
}
temp, _ = a.GetState("temperature")
fmt.Printf("Temperature: %.2f\n", temp) // ~65.0

// String representation
a2 := siruntime.NewAgent("x")
a2.SetState("load", 1.0)
a2.AddCapability("compute")
fmt.Println(a2.String()) // Agent[x] caps=1 states=1
```

---

### Capability — `capability.go`

Capability discovery, matching, and resolution.

```go
type Capability struct {
    Name     string
    Version  string
    Provides []string
    Requires []string
}

type Registry struct { ... }

func NewRegistry() *Registry
func (r *Registry) Register(c Capability) error
func (r *Registry) Get(name string) (Capability, bool)
func (r *Registry) List() []Capability
func (r *Registry) Remove(name string)

type MatchResult struct {
    AgentID string
    Score   float64
    Matched []string
    Missing []string
}

func Match(agentID string, agentCaps []string, required []string) MatchResult
func Resolve(candidates []MatchResult) (MatchResult, bool)
```

**Examples:**

```go
reg := siruntime.NewRegistry()
reg.Register(siruntime.Capability{
    Name: "http", Version: "1.0",
    Provides: []string{"network", "http"},
})
reg.Register(siruntime.Capability{
    Name: "grpc", Version: "1.0",
    Provides: []string{"network", "rpc"},
})
reg.Register(siruntime.Capability{
    Name: "cache", Version: "2.0",
    Provides: []string{"storage"},
    Requires: []string{"network"},
})

// List all
caps := reg.List()
fmt.Printf("%d capabilities registered\n", len(caps)) // 3

// Match agents to task requirements
result := siruntime.Match("agent-1",
    []string{"compute", "network", "storage"},
    []string{"compute", "network"},
)
fmt.Printf("Score: %.1f, Matched: %v, Missing: %v\n",
    result.Score, result.Matched, result.Missing)
// Score: 0.5, Matched: [network], Missing: [compute]

// Resolve: find best agent
candidates := []siruntime.MatchResult{
    {AgentID: "a", Score: 0.3},
    {AgentID: "b", Score: 0.9},
    {AgentID: "c", Score: 0.7},
}
best, ok := siruntime.Resolve(candidates)
fmt.Printf("Best: %s (%.1f)\n", best.AgentID, best.Score) // Best: b (0.9)
```

---

### Cell & Grid — `cell.go`

Cellular automata with pluggable update rules.

```go
type Cell struct {
    ID        string
    State     float64
    Target    float64
    Neighbors []int
}

func NewCell(id string, state, target float64) *Cell
func (c *Cell) AddNeighbor(idx int) error
func (c *Cell) AverageNeighborState(cells []*Cell) float64
func (c *Cell) Update(cells []*Cell, alpha, beta float64)

type Grid struct {
    Width  int
    Height int
    Cells  []*Cell
}

func NewGrid(width, height int, state, target float64) *Grid
func (g *Grid) WireNeighbors()
func (g *Grid) UpdateAll(alpha, beta float64)
func (g *Grid) Variance() float64
func (g *Grid) EquilibriumCheck(tolerance float64) bool
```

**Examples:**

```go
// 3x3 grid, all cells start at 0, target 10
grid := siruntime.NewGrid(3, 3, 0.0, 10.0)
grid.WireNeighbors() // Connect von Neumann neighbors

// Set one cell high
grid.Cells[0].State = 100.0

// Run updates until equilibrium
for i := 0; i < 50; i++ {
    grid.UpdateAll(0.3, 0.1) // alpha=diffusion, beta=homeostasis
}

fmt.Printf("Variance: %.4f\n", grid.Variance())
fmt.Printf("At equilibrium: %v\n", grid.EquilibriumCheck(2.0)) // true

// Custom cell
c := siruntime.NewCell("custom", 50.0, 100.0)
c.AddNeighbor(0)
c.AddNeighbor(1)
c.AddNeighbor(-1) // error: negative index
```

---

### Spectral — `spectral.go`

Eigenvector centrality via power iteration.

```go
type AdjacencyMatrix struct {
    Data [][]float64
    Size int
}

func NewAdjacencyMatrix(n int) *AdjacencyMatrix
func (m *AdjacencyMatrix) Set(i, j int, value float64) error
func (m *AdjacencyMatrix) Get(i, j int) (float64, error)

type Eigenpair struct {
    Value  float64
    Vector []float64
}

func PowerIteration(m *AdjacencyMatrix, maxIter int, tol float64) (*Eigenpair, error)

type RankedAgent struct {
    Index      int
    Centrality float64
}

func SpectralRank(m *AdjacencyMatrix) ([]RankedAgent, error)
```

**Examples:**

```go
// 3-node graph
m := siruntime.NewAdjacencyMatrix(3)
m.Set(0, 1, 1.0)
m.Set(1, 2, 1.0)
m.Set(0, 2, 0.5)

// Find dominant eigenpair
pair, err := siruntime.PowerIteration(m, 1000, 1e-8)
fmt.Printf("Eigenvalue: %.6f\n", pair.Value)
fmt.Printf("Eigenvector: %v\n", pair.Vector)

// Rank by centrality
ranked, err := siruntime.SpectralRank(m)
for _, r := range ranked {
    fmt.Printf("Node %d: centrality=%.4f\n", r.Index, r.Centrality)
}
// Node 1 is most central (bridges 0 and 2)
```

---

### Fleet — `fleet.go`

Multi-agent orchestration with spectral ranking and conservation.

```go
type Fleet struct {
    ID        string
    Agents    map[string]*Agent
    Budgets   map[string]*AgentBudget
    Adjacency *AdjacencyMatrix
}

func NewFleet(id string) *Fleet
func (f *Fleet) AddAgent(a *Agent, budget *Budget) error
func (f *Fleet) RemoveAgent(id string) bool
func (f *Fleet) GetAgent(id string) (*Agent, bool)
func (f *Fleet) ListAgents() []*Agent
func (f *Fleet) AgentCount() int
func (f *Fleet) BuildAdjacencyMatrix(affinity func(a, b *Agent) float64) (*AdjacencyMatrix, error)
func (f *Fleet) SpectralRank() ([]RankedAgent, error)
func (f *Fleet) ConservationAudit() AuditResult
func (f *Fleet) HealthReport() HealthReport
func (f *Fleet) BestAgentForTask(required []string) (MatchResult, error)
func (f *Fleet) RebalanceBudgets() error
```

**Examples:**

```go
// Create a fleet
fleet := siruntime.NewFleet("production")

// Add agents with budgets
a1 := siruntime.NewAgent("planner")
a1.AddCapability("plan")
a1.AddCapability("schedule")

a2 := siruntime.NewAgent("executor")
a2.AddCapability("compute")
a2.AddCapability("network")

fleet.AddAgent(a1, siruntime.NewBudget(100))
fleet.AddAgent(a2, siruntime.NewBudget(200))

fleet.Budgets["planner"].Allocate(60, 40)
fleet.Budgets["executor"].Allocate(120, 80)

// Conservation audit
audit := fleet.ConservationAudit()
fmt.Printf("Valid: %v\n", audit.Valid)
fmt.Printf("Fleet total: %.0f\n", audit.FleetTotal) // 300

// Best agent for a task
result, err := fleet.BestAgentForTask([]string{"compute"})
fmt.Printf("Best: %s (score=%.1f)\n", result.AgentID, result.Score)
// Best: executor (score=1.0)

// Spectral ranking
ranked, _ := fleet.SpectralRank()
for _, r := range ranked {
    fmt.Printf("Rank %d: agent index %d\n", r.Index, r.Index)
}

// Health report
health := fleet.HealthReport()
fmt.Printf("Fleet avg error: %.4f\n", health.FleetAvg)
fmt.Printf("Worst agent: %s\n", health.WorstAgent)

// Rebalance budgets (equalize eta fraction)
fleet.RebalanceBudgets()
```

---

## Docker

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod ./
COPY *.go ./
RUN go build -v ./...
RUN go test -v ./...

# Runtime stage
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/si-runtime-go .
CMD ["./si-runtime-go"]
```

```bash
docker build -t si-runtime-go .
docker run si-runtime-go
```

---

## Working Examples

### Full Fleet Simulation

```go
package main

import (
    "fmt"
    siruntime "github.com/SuperInstance/si-runtime-go"
)

func main() {
    fleet := siruntime.NewFleet("demo")

    // Create 5 agents with varying budgets
    for i := 0; i < 5; i++ {
        id := fmt.Sprintf("agent-%d", i)
        a := siruntime.NewAgent(id)
        a.SetState("workload", float64(i*10))
        a.AddCapability(fmt.Sprintf("cap-%d", i))

        budget := siruntime.NewBudget(float64(100 - i*10))
        budget.Allocate(float64(60-i*5), float64(40-i*5))

        fleet.AddAgent(a, budget)
    }

    // Conservation audit
    audit := fleet.ConservationAudit()
    fmt.Printf("Conservation valid: %v\n", audit.Valid)
    fmt.Printf("Fleet total: %.0f\n", audit.FleetTotal)
    if !audit.Valid {
        for _, v := range audit.Violations {
            fmt.Printf("  Violation: %s\n", v)
        }
    }

    // Spectral ranking
    ranked, err := fleet.SpectralRank()
    if err != nil {
        panic(err)
    }
    fmt.Println("Spectral ranking:")
    for i, r := range ranked {
        fmt.Printf("  #%d: agent-%d (centrality=%.4f)\n", i+1, r.Index, r.Centrality)
    }

    // Health
    health := fleet.HealthReport()
    fmt.Printf("Avg error: %.4f, Worst: %s\n", health.FleetAvg, health.WorstAgent)
}
```

### Grid Diffusion

```go
package main

import (
    "fmt"
    siruntime "github.com/SuperInstance/si-runtime-go"
)

func main() {
    grid := siruntime.NewGrid(5, 5, 0.0, 10.0)
    grid.WireNeighbors()

    // Inject heat at center
    center := 2*5 + 2
    grid.Cells[center].State = 100.0

    fmt.Printf("Initial variance: %.2f\n", grid.Variance())

    for i := 0; i < 100; i++ {
        grid.UpdateAll(0.3, 0.1)
    }

    fmt.Printf("Final variance: %.4f\n", grid.Variance())
    fmt.Printf("Equilibrium: %v\n", grid.EquilibriumCheck(2.0))

    // Print grid state
    for y := 0; y < 5; y++ {
        for x := 0; x < 5; x++ {
            fmt.Printf("%6.2f ", grid.Cells[y*5+x].State)
        }
        fmt.Println()
    }
}
```

---

## Tests

```bash
go test -v ./...
```

Test files cover every module:

| File | Tests |
|------|-------|
| `agent_test.go` | Homeostasis, capabilities, String |
| `capability_test.go` | Registry, Match, Resolve |
| `cell_test.go` | Cell update, grid equilibrium |
| `conservation_test.go` | Allocate invariant, transfer, audit |
| `fleet_test.go` | Add/remove, conservation audit, spectral rank, task matching, rebalance, health |
| `spectral_test.go` | Adjacency matrix, power iteration, spectral rank |

---

## Architecture

```
si-runtime-go/
├── agent.go            # Agent with state, capabilities, homeostasis
├── capability.go       # Capability registry, matching, resolution
├── cell.go             # Cellular automata grid
├── conservation.go     # Budget, AgentBudget, Audit
├── fleet.go            # Fleet orchestration
├── spectral.go         # Adjacency matrix, power iteration, spectral rank
├── *_test.go           # Tests for each module
├── Dockerfile          # Multi-stage Docker build
└── go.mod              # Module definition (Go 1.21, zero deps)
```

---

## Related Repos

| Repo | Language | Description |
|------|----------|-------------|
| [`conservation-law`](https://github.com/SuperInstance/conservation-law) | Rust | Core conservation law crate |
| [`si-runtime-python`](https://github.com/SuperInstance/si-runtime-python) | Python | Pure Python runtime (same API) |
| [`si-conservation-python`](https://github.com/SuperInstance/si-conservation-python) | Rust/Python | PyO3 bindings for heavy compute |
| [`si-cli`](https://github.com/SuperInstance/si-cli) | Rust | CLI for fleet management |
| [`si-fleet-api`](https://github.com/SuperInstance/si-fleet-api) | TypeScript | REST API for fleet budgets |

---

## License

MIT

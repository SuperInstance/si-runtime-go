# si-runtime-go

Go runtime for the SuperInstance ecosystem.

This library unifies conservation-law budgeting, spectral fleet ranking,
capability matching, cellular homeostasis, and agent orchestration into a
single coherent Go package.

```bash
go get github.com/SuperInstance/si-runtime-go
```

---

## Table of Contents

1. [Concepts](#concepts)
2. [Conservation Budgets](#conservation-budgets)
3. [Spectral Ranking](#spectral-ranking)
4. [Capability Registry](#capability-registry)
5. [Cellular Homeostasis](#cellular-homeostasis)
6. [Agents](#agents)
7. [Fleet Orchestration](#fleet-orchestration)
8. [Running Tests](#running-tests)
9. [API Overview](#api-overview)

---

## Concepts

The SuperInstance runtime treats an agent fleet as a physical system:

- **Budget** — A conserved quantity split into productive spend (`gamma`)
  and overhead (`eta`). The invariant `gamma + eta == total` is enforced
  at compile time via the API.
- **Spectral ranking** — Agents are nodes in a graph; eigenvector
  centrality identifies the most connected / influential agents.
- **Capabilities** — Each agent advertises skills. The fleet matches tasks
  to the best-qualified agent.
- **Homeostasis** — Agents maintain internal state variables near target
  values, like biological cells regulating temperature or pH.
- **Cells** — A cellular-automaton layer for spatial diffusion and
  emergent pattern formation.

---

## Conservation Budgets

A `Budget` has three fields: `Total`, `Gamma`, and `Eta`. The API
guarantees (and `Audit` verifies) that `Gamma + Eta == Total`.

```go
package main

import (
    "fmt"
    "log"

    siruntime "github.com/SuperInstance/si-runtime-go"
)

func main() {
    // Create a fleet-wide budget of 1000 tokens
    b := siruntime.NewBudget(1000)
    fmt.Printf("initial: total=%.0f gamma=%.0f eta=%.0f\n", b.Total, b.Gamma, b.Eta)
    // Output: initial: total=1000 gamma=0 eta=1000

    // Allocate 600 to productive work, 400 to overhead
    if err := b.Allocate(600, 400); err != nil {
        log.Fatal(err)
    }
    fmt.Printf("allocated: total=%.0f gamma=%.0f eta=%.0f\n", b.Total, b.Gamma, b.Eta)
    // Output: allocated: total=1000 gamma=600 eta=400

    // Move 200 from eta into gamma (converting idle capacity to work)
    if err := b.Transfer(200); err != nil {
        log.Fatal(err)
    }
    fmt.Printf("transferred: gamma=%.0f eta=%.0f\n", b.Gamma, b.Eta)
    // Output: transferred: gamma=800 eta=200

    // Overspend: try to spend 900 from gamma (only 800 available)
    shortfall, err := b.Overspend(900)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("overspend shortfall=%.0f gamma=%.0f eta=%.0f\n", shortfall, b.Gamma, b.Eta)
    // Output: overspend shortfall=100 gamma=0 eta=100
}
```

### Inter-agent transfers

```go
alice := &siruntime.AgentBudget{AgentID: "alice", Budget: siruntime.NewBudget(100)}
alice.Budget.Allocate(80, 20)

bob := &siruntime.AgentBudget{AgentID: "bob", Budget: siruntime.NewBudget(100)}
bob.Budget.Allocate(30, 70)

// Move 50 tokens of productive budget from alice to bob
if err := siruntime.Transfer(alice, bob, 50); err != nil {
    log.Fatal(err)
}
fmt.Printf("alice gamma=%.0f, bob gamma=%.0f\n", alice.Budget.Gamma, bob.Budget.Gamma)
// Output: alice gamma=30, bob gamma=80
```

### Fleet-wide audit

```go
agents := []*siruntime.AgentBudget{alice, bob}
result := siruntime.Audit(agents)
fmt.Printf("valid=%v fleet_total=%.0f fleet_gamma=%.0f fleet_eta=%.0f\n",
    result.Valid, result.FleetTotal, result.FleetGamma, result.FleetEta)
// Output: valid=true fleet_total=200 fleet_gamma=110 fleet_eta=90
```

---

## Spectral Ranking

Given an affinity matrix between agents, power iteration finds the
dominant eigenvector. Each component gives the centrality of the
corresponding agent.

```go
package main

import (
    "fmt"
    "log"

    siruntime "github.com/SuperInstance/si-runtime-go"
)

func main() {
    m := siruntime.NewAdjacencyMatrix(3)
    m.Set(0, 0, 1.0)
    m.Set(0, 1, 0.8)
    m.Set(0, 2, 0.2)
    m.Set(1, 1, 1.0)
    m.Set(1, 2, 0.5)
    m.Set(2, 2, 1.0)

    ranked, err := siruntime.SpectralRank(m)
    if err != nil {
        log.Fatal(err)
    }

    for _, r := range ranked {
        fmt.Printf("agent %d: centrality=%.4f\n", r.Index, r.Centrality)
    }
    // Output:
    // agent 0: centrality=0.7399
    // agent 1: centrality=0.6148
    // agent 2: centrality=0.2712
}
```

### Top-k eigenpairs

```go
pairs, err := siruntime.TopKEigenpairs(m, 2, 1000, 1e-8)
if err != nil {
    log.Fatal(err)
}
for i, pair := range pairs.Pairs {
    fmt.Printf("eigenvalue %d: %.4f\n", i, pair.Value)
}
```

---

## Capability Registry

Agents advertise capabilities. The registry stores them; the `Match`
function scores task-agent compatibility.

```go
package main

import (
    "fmt"

    siruntime "github.com/SuperInstance/si-runtime-go"
)

func main() {
    reg := siruntime.NewCapabilityRegistry()
    reg.Register(siruntime.Capability{Name: "plan", Version: "1.0", Score: 0.9})
    reg.Register(siruntime.Capability{Name: "code", Version: "2.0", Score: 0.8})
    reg.Register(siruntime.Capability{Name: "review", Version: "1.0", Score: 0.7})

    c, ok := reg.Get("code")
    fmt.Printf("found=%v score=%.1f\n", ok, c.Score)
    // Output: found=true score=0.8

    caps := reg.List()
    fmt.Printf("registered capabilities: %d\n", len(caps))
    // Output: registered capabilities: 3
}
```

### Matching agents to tasks

```go
agentCaps := []siruntime.Capability{
    {Name: "read", Score: 1.0},
    {Name: "write", Score: 0.5},
}
required := []string{"read", "write"}

result := siruntime.Match("agent-1", agentCaps, required, 0.6)
fmt.Printf("score=%.2f matched=%v partial=%v missing=%v\n",
    result.Score, result.Matched, result.Partial, result.Missing)
// Output: score=0.75 matched=[read] partial=[write] missing=[]
```

---

## Cellular Homeostasis

A `Grid` of `Cell` values diffuses state toward neighbors and toward a
homeostatic target. This models spatial load balancing or heat diffusion.

```go
package main

import (
    "fmt"

    siruntime "github.com/SuperInstance/si-runtime-go"
)

func main() {
    // 5×5 grid, all cells start at 0.0, target is 10.0
    g := siruntime.NewGrid(5, 5, 0.0, 10.0)
    g.WireNeighbors()

    // Heat one corner cell
    g.Cells[0].State = 100.0

    for step := 0; step < 20; step++ {
        g.UpdateAll(0.3, 0.1)
    }

    fmt.Printf("corner=%.2f center=%.2f variance=%.2f\n",
        g.Cells[0].State, g.Cells[12].State, g.Variance())
    // The heat diffuses from the corner toward the rest of the grid.
}
```

### Single-cell update

```go
c := siruntime.NewCell("c1", 10.0, 20.0)
c.Update(0.0, 0.5) // no diffusion, pure homeostasis
fmt.Printf("state=%.2f\n", c.State)
// Output: state=15.00
```

---

## Agents

An `Agent` combines state, capabilities, homeostasis targets, and a budget.

```go
package main

import (
    "fmt"

    siruntime "github.com/SuperInstance/si-runtime-go"
)

func main() {
    a := siruntime.NewAgent("planner-1")
    a.SetState("energy", 50.0)
    a.SetHomeostasis("energy", 100.0)
    a.AddCapability(siruntime.Capability{Name: "plan", Score: 0.95})
    a.AddCapability(siruntime.Capability{Name: "negotiate", Score: 0.7})
    a.AttachBudget(500)
    a.Budget.Budget.Allocate(300, 200)

    fmt.Println(a.String())
    // Output: Agent[planner-1] caps=2 budget(total=500 γ=300 η=200)

    // Drive energy toward its target
    a.UpdateHomeostasis(0.1)
    energy, _ := a.GetState("energy")
    fmt.Printf("energy=%.2f error=%.2f\n", energy, a.HomeostasisError())
    // Output: energy=55.00 error=45.00
}
```

---

## Fleet Orchestration

A `Fleet` owns agents, builds their adjacency matrix, ranks them
spectrally, audits their budgets, and matches tasks.

```go
package main

import (
    "fmt"
    "log"

    siruntime "github.com/SuperInstance/si-runtime-go"
)

func main() {
    fleet := siruntime.NewFleet("production")

    // Create agents
    a1 := siruntime.NewAgent("api-gateway")
    a1.SetState("workload", 80.0)
    a1.AttachBudget(1000)
    a1.Budget.Budget.Allocate(700, 300)
    a1.AddCapability(siruntime.Capability{Name: "route", Score: 0.95})

    a2 := siruntime.NewAgent("ml-inference")
    a2.SetState("workload", 95.0)
    a2.AttachBudget(1000)
    a2.Budget.Budget.Allocate(600, 400)
    a2.AddCapability(siruntime.Capability{Name: "inference", Score: 0.9})

    a3 := siruntime.NewAgent("cache-layer")
    a3.SetState("workload", 40.0)
    a3.AttachBudget(1000)
    a3.Budget.Budget.Allocate(200, 800)
    a3.AddCapability(siruntime.Capability{Name: "route", Score: 0.6})

    fleet.AddAgent(a1)
    fleet.AddAgent(a2)
    fleet.AddAgent(a3)

    // Conservation audit
    audit := fleet.ConservationAudit()
    fmt.Printf("audit valid=%v fleet_total=%.0f\n", audit.Valid, audit.FleetTotal)
    // Output: audit valid=true fleet_total=3000

    // Spectral rank by workload similarity
    ranked, err := fleet.SpectralRank()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Spectral ranking:")
    for _, r := range ranked {
        fmt.Printf("  agent %d: centrality=%.4f\n", r.Index, r.Centrality)
    }

    // Best agent for routing
    best, err := fleet.BestAgentForTask([]string{"route"}, 0.5)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("best router: %s (score=%.2f)\n", best.AgentID, best.Score)
    // Output: best router: api-gateway (score=0.95)

    // Rebalance budgets so every agent has the same eta fraction
    fleet.RebalanceBudgets()
    fmt.Println("After rebalancing:")
    for _, a := range fleet.ListAgents() {
        fmt.Printf("  %s: gamma=%.0f eta=%.0f\n", a.ID,
            a.Budget.Budget.Gamma, a.Budget.Budget.Eta)
    }
    // Every agent now has eta ≈ 500 (half of 1000)
}
```

---

## Running Tests

```bash
go test -v ./...
```

The test suite covers:

- Budget allocation, transfer, overspend, and inter-agent transfer
- Fleet-wide conservation audit
- Adjacency matrix construction and symmetry
- Power iteration and eigenpair extraction
- Spectral ranking and centrality ordering
- Capability registry registration and retrieval
- Task-agent matching with threshold scoring
- Best-match selection from candidates
- Cell homeostatic update
- Grid diffusion toward equilibrium
- Agent state, homeostasis error, capability add/remove
- Fleet add/remove, conservation audit, spectral rank
- Task-to-agent matching in a fleet
- Budget rebalancing across fleet members

Run benchmarks:

```bash
go test -bench=. ./...
```

---

## API Overview

### Conservation (`conservation.go`)

| Type | Purpose |
|------|---------|
| `Budget` | Conserved resource pool with `Total`, `Gamma`, `Eta` |
| `AgentBudget` | Couples an agent ID to a `Budget` |
| `Transfer(from, to, amount)` | Move gamma between agents |
| `Audit(budgets)` | Verify `gamma + eta == total` fleet-wide |

### Spectral (`spectral.go`)

| Type | Purpose |
|------|---------|
| `AdjacencyMatrix` | Dense symmetric affinity matrix |
| `Eigenpair` | One `(eigenvalue, eigenvector)` pair |
| `PowerIteration(m, maxIter, tol)` | Dominant eigenpair via power method |
| `TopKEigenpairs(m, k, maxIter, tol)` | Top-k eigenpairs with deflation |
| `SpectralRank(m)` | Agents sorted by eigenvector centrality |

### Capabilities (`capability.go`)

| Type | Purpose |
|------|---------|
| `Capability` | Named skill with version and score |
| `CapabilityRegistry` | Thread-safe capability store |
| `Match(agentID, caps, required, threshold)` | Score task compatibility |
| `BestMatch(candidates)` | Select highest-scoring candidate |

### Cells (`cell.go`)

| Type | Purpose |
|------|---------|
| `Cell` | Single unit with state, target, and neighbors |
| `Grid` | Rectangular lattice of cells |
| `Update(alpha, beta)` | Diffusion + homeostasis step |
| `UpdateAll(alpha, beta)` | Synchronous grid update |
| `EquilibriumCheck(tolerance)` | All cells near target |

### Agents (`agent.go`)

| Type | Purpose |
|------|---------|
| `Agent` | ID, state map, capabilities, homeostasis targets |
| `AddCapability(c)` | Register a capability |
| `RemoveCapability(name)` | Unregister by name |
| `UpdateHomeostasis(rate)` | Drive state toward targets |
| `HomeostasisError()` | RMS deviation from targets |
| `AttachBudget(total)` | Link a conserved budget |

### Fleet (`fleet.go`)

| Type | Purpose |
|------|---------|
| `Fleet` | Container for agents with shared infrastructure |
| `AddAgent(a)` | Register an agent |
| `RemoveAgent(id)` | Unregister by ID |
| `BuildAdjacencyMatrix(affinity)` | Construct affinity from agent states |
| `SpectralRank()` | Rank agents by centrality |
| `ConservationAudit()` | Verify all budgets |
| `BestAgentForTask(required, threshold)` | Match task to best agent |
| `RebalanceBudgets()` | Equalize eta fractions fleet-wide |

---

## License

MIT OR Apache-2.0

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
10. [Docker](#docker)

---

## Concepts

The SuperInstance runtime treats an agent fleet as a physical system:

- **Budget** — A conserved quantity split into productive spend (`Gamma`)
  and overhead (`Eta`). The invariant `Gamma + Eta == Total` is enforced
  by the API.
- **Spectral ranking** — Agents are nodes in a graph; eigenvector
  centrality identifies the most connected / influential agents.
- **Capabilities** — Each agent advertises skills. The fleet matches tasks
  to the best-qualified agent.
- **Homeostasis** — Agents maintain internal state variables near target
  values, like biological cells regulating temperature.
- **Cells** — A cellular-automaton layer for spatial diffusion and
  emergent pattern formation.

---

## Conservation Budgets

A `Budget` has three fields: `Total`, `Gamma`, and `Eta`. The API
enforces `Gamma + Eta == Total`.

```go
package main

import (
    "fmt"
    "log"

    siruntime "github.com/SuperInstance/si-runtime-go"
)

func main() {
    b := siruntime.NewBudget(1000)
    fmt.Printf("initial: total=%.0f gamma=%.0f eta=%.0f\n", b.Total, b.Gamma, b.Eta)
    // Output: initial: total=1000 gamma=0 eta=1000

    if err := b.Allocate(600, 400); err != nil {
        log.Fatal(err)
    }
    fmt.Printf("allocated: gamma=%.0f eta=%.0f\n", b.Gamma, b.Eta)

    if err := b.Transfer(200); err != nil {
        log.Fatal(err)
    }
    fmt.Printf("transferred: gamma=%.0f eta=%.0f\n", b.Gamma, b.Eta)
}
```

### Inter-agent transfers

```go
alice := &siruntime.AgentBudget{AgentID: "alice", Budget: siruntime.NewBudget(100)}
alice.Allocate(80, 20)

bob := &siruntime.AgentBudget{AgentID: "bob", Budget: siruntime.NewBudget(100)}
bob.Allocate(30, 70)

if err := alice.Transfer(bob, 50); err != nil {
    log.Fatal(err)
}
fmt.Printf("alice gamma=%.0f, bob gamma=%.0f\n", alice.Budget.Gamma, bob.Budget.Gamma)
```

### Fleet-wide audit

```go
agents := []*siruntime.AgentBudget{alice, bob}
result := siruntime.Audit(agents)
fmt.Printf("valid=%v total=%.0f gamma=%.0f eta=%.0f\n",
    result.Valid, result.FleetTotal, result.FleetGamma, result.FleetEta)
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
}
```

---

## Capability Registry

Agents advertise capabilities. The registry stores them; `Match` scores
task-agent compatibility.

```go
package main

import (
    "fmt"

    siruntime "github.com/SuperInstance/si-runtime-go"
)

func main() {
    reg := siruntime.NewRegistry()
    reg.Register(siruntime.Capability{Name: "plan", Version: "1.0", Provides: []string{"strategy"}})
    reg.Register(siruntime.Capability{Name: "code", Version: "2.0", Provides: []string{"implementation"}})

    c, ok := reg.Get("code")
    fmt.Printf("found=%v version=%s\n", ok, c.Version)
}
```

### Matching agents to tasks

```go
agentCaps := []string{"read", "write"}
required := []string{"read", "write"}
result := siruntime.Match("agent-1", agentCaps, required)
fmt.Printf("score=%.2f matched=%v missing=%v\n", result.Score, result.Matched, result.Missing)
```

---

## Cellular Homeostasis

A `Grid` of `Cell` values diffuses state toward neighbors and toward a
homeostatic target.

```go
package main

import (
    "fmt"

    siruntime "github.com/SuperInstance/si-runtime-go"
)

func main() {
    g := siruntime.NewGrid(5, 5, 0.0, 10.0)
    g.WireNeighbors()
    g.Cells[0].State = 100.0
    for step := 0; step < 20; step++ {
        g.UpdateAll(0.3, 0.1)
    }
    fmt.Printf("corner=%.2f center=%.2f variance=%.2f\n",
        g.Cells[0].State, g.Cells[12].State, g.Variance())
}
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
    a.AddCapability("plan")
    a.AddCapability("negotiate")

    fmt.Println(a.String())

    a.UpdateHomeostasis(0.1)
    energy, _ := a.GetState("energy")
    fmt.Printf("energy=%.2f error=%.2f\n", energy, a.HomeostasisError())
}
```

---

## Fleet Orchestration

A `Fleet` owns agents, ranks them spectrally, audits budgets, and matches tasks.

```go
package main

import (
    "fmt"
    "log"

    siruntime "github.com/SuperInstance/si-runtime-go"
)

func main() {
    fleet := siruntime.NewFleet("production")

    a1 := siruntime.NewAgent("api-gateway")
    a1.SetState("workload", 80.0)
    a1.AddCapability("route")
    fleet.AddAgent(a1, siruntime.NewBudget(1000))
    fleet.Budgets["api-gateway"].Allocate(700, 300)

    a2 := siruntime.NewAgent("ml-inference")
    a2.SetState("workload", 95.0)
    a2.AddCapability("inference")
    fleet.AddAgent(a2, siruntime.NewBudget(1000))
    fleet.Budgets["ml-inference"].Allocate(600, 400)

    a3 := siruntime.NewAgent("cache-layer")
    a3.SetState("workload", 40.0)
    a3.AddCapability("route")
    fleet.AddAgent(a3, siruntime.NewBudget(1000))
    fleet.Budgets["cache-layer"].Allocate(200, 800)

    // Conservation audit
    audit := fleet.ConservationAudit()
    fmt.Printf("audit valid=%v fleet_total=%.0f\n", audit.Valid, audit.FleetTotal)

    // Spectral rank
    ranked, err := fleet.SpectralRank()
    if err != nil {
        log.Fatal(err)
    }
    for _, r := range ranked {
        fmt.Printf("agent %d: centrality=%.4f\n", r.Index, r.Centrality)
    }

    // Best agent for routing
    best, err := fleet.BestAgentForTask([]string{"route"})
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("best router: %s (score=%.2f)\n", best.AgentID, best.Score)

    // Health report
    rpt := fleet.HealthReport()
    fmt.Printf("fleet avg error=%.2f worst=%s\n", rpt.FleetAvg, rpt.WorstAgent)

    // Rebalance budgets
    fleet.RebalanceBudgets()
    for _, a := range fleet.ListAgents() {
        ab := fleet.Budgets[a.ID]
        fmt.Printf("%s: gamma=%.0f eta=%.0f\n", a.ID, ab.Budget.Gamma, ab.Budget.Eta)
    }
}
```

---

## Running Tests

```bash
go test -v ./...
```

The test suite includes table-driven tests covering:

- Budget allocation with invariant enforcement
- Budget transfer and overspend handling
- Inter-agent budget transfers
- Fleet-wide conservation audit
- Adjacency matrix symmetry and bounds
- Power iteration for dominant eigenpair
- Spectral ranking ordering
- Capability registry CRUD operations
- Task-agent matching with thresholds
- Best-match resolution
- Cell homeostatic update
- Grid diffusion toward equilibrium
- Neighbor index validation
- Agent homeostasis convergence
- Capability add/remove deduplication
- Fleet add/remove agent lifecycle
- Fleet conservation audit with budgets
- Fleet spectral ranking from workload
- Fleet task-to-agent matching
- Fleet budget rebalancing
- Fleet health report computation

---

## API Overview

### Conservation (`conservation.go`)

| Type | Purpose |
|------|---------|
| `Budget` | Conserved resource pool with `Total`, `Gamma`, `Eta` |
| `AgentBudget` | Couples an agent ID to a `Budget` |
| `Transfer(to, amount)` | Move gamma between agents |
| `Audit(budgets)` | Verify `gamma + eta == total` fleet-wide |

### Spectral (`spectral.go`)

| Type | Purpose |
|------|---------|
| `AdjacencyMatrix` | Dense symmetric affinity matrix |
| `Eigenpair` | One `(eigenvalue, eigenvector)` pair |
| `PowerIteration(m, maxIter, tol)` | Dominant eigenpair via power method |
| `SpectralRank(m)` | Agents sorted by eigenvector centrality |

### Capabilities (`capability.go`)

| Type | Purpose |
|------|---------|
| `Capability` | Named skill with `Provides` and `Requires` |
| `Registry` | Thread-safe capability store |
| `Match(agentID, caps, required)` | Score task compatibility |
| `Resolve(candidates)` | Select highest-scoring candidate |

### Cells (`cell.go`)

| Type | Purpose |
|------|---------|
| `Cell` | Single unit with state, target, and neighbor indices |
| `Grid` | Rectangular lattice of cells |
| `Update(cells, alpha, beta)` | Diffusion + homeostasis step |
| `UpdateAll(alpha, beta)` | Synchronous grid update |
| `EquilibriumCheck(tolerance)` | All cells near target |

### Agents (`agent.go`)

| Type | Purpose |
|------|---------|
| `Agent` | ID, state map, capability names, homeostasis targets |
| `AddCapability(name)` | Register a capability |
| `RemoveCapability(name)` | Unregister by name |
| `UpdateHomeostasis(rate)` | Drive state toward targets |
| `HomeostasisError()` | RMS deviation from targets |

### Fleet (`fleet.go`)

| Type | Purpose |
|------|---------|
| `Fleet` | Container for agents with shared infrastructure |
| `AddAgent(a, budget)` | Register an agent with optional budget |
| `RemoveAgent(id)` | Unregister by ID |
| `BuildAdjacencyMatrix(affinity)` | Construct affinity from agent states |
| `SpectralRank()` | Rank agents by centrality |
| `ConservationAudit()` | Verify all budgets |
| `HealthReport()` | Per-agent homeostasis diagnostics |
| `BestAgentForTask(required)` | Match task to best agent |
| `RebalanceBudgets()` | Equalize eta fractions fleet-wide |

---

## Docker

Build and test in a container:

```bash
docker build -t si-runtime-go .
```

The Dockerfile uses a multi-stage build: the builder stage compiles and
runs tests; the runtime stage copies the artifact into a minimal Alpine
image.

---

## License

MIT OR Apache-2.0

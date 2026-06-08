# INTEGRATION.md — si-runtime-go

Cross-language integration guide for the **SuperInstance Go runtime** (`si-runtime-go`).
This document shows the same conservation budget operation in all 7 supported languages,
how this library connects to the broader SuperInstance ecosystem, and FFI binding patterns.

---

## Table of Contents

1. [Same Operation in 7 Languages](#1-same-operation-in-7-languages)
2. [Cross-Repo Integration](#2-cross-repo-integration)
3. [FFI Bindings](#3-ffi-bindings)

---

## 1. Same Operation in 7 Languages

The canonical operation: **create a conservation budget of C=1000, allocate gamma=600 and eta=400, verify the invariant γ+η=C, then transfer budget and run a fleet-wide audit.**

### Go (si-runtime-go — this repo)

```go
package main

import (
    "fmt"
    siruntime "github.com/SuperInstance/si-runtime-go"
)

func main() {
    // ── Create budget with total C = 1000 ──
    budget := siruntime.NewBudget(1000)

    // ── Allocate gamma (productive) and eta (waste) ──
    err := budget.Allocate(600, 400)
    if err != nil {
        panic(err)
    }

    // ── Verify invariant: gamma + eta == total ──
    fmt.Printf("gamma=%.1f eta=%.1f total=%.1f remaining=%.1f\n",
        budget.Gamma, budget.Eta, budget.Total, budget.Remaining())
    // Output: gamma=600.0 eta=400.0 total=1000.0 remaining=1000.0

    // ── Transfer 50 from eta to gamma ──
    err = budget.Transfer(50)
    if err != nil {
        panic(err)
    }
    fmt.Printf("After transfer: gamma=%.1f eta=%.1f total=%.1f\n",
        budget.Gamma, budget.Eta, budget.Total)
    // Output: gamma=650.0 eta=350.0 total=1000.0

    // ── Fleet-wide audit ──
    agentBudgets := []*siruntime.AgentBudget{
        {AgentID: "agent-a", Budget: budget},
        {AgentID: "agent-b", Budget: siruntime.NewBudget(500)},
    }
    agentBudgets[1].Allocate(300, 200)

    result := siruntime.Audit(agentBudgets)
    fmt.Printf("Fleet audit: valid=%v fleetTotal=%.0f fleetGamma=%.0f fleetEta=%.0f\n",
        result.Valid, result.FleetTotal, result.FleetGamma, result.FleetEta)

    // ── Agent with homeostasis ──
    agent := siruntime.NewAgent("demo-agent")
    agent.SetState("energy", 50)
    agent.SetHomeostasis("energy", 100)
    agent.UpdateHomeostasis(0.1)
    fmt.Printf("Agent: %s, energy error=%.2f\n", agent, agent.HomeostasisError())
}
```

### Rust (conservation-law-rs — reference implementation)

```rust
use conservation_law::ConservationBudget;

fn main() {
    let mut budget = ConservationBudget::new(1000.0);
    budget.allocate(600.0, 400.0).expect("allocation failed");

    let audit = budget.audit();
    assert!((audit.gamma + audit.eta - audit.total).abs() < 1e-10);
    println!("gamma={} eta={} total={}", audit.gamma, audit.eta, audit.total);

    budget.transfer("gamma", "eta", 50.0).expect("transfer failed");
    let audit = budget.audit();
    println!("After transfer: gamma={} eta={}", audit.gamma, audit.eta);
}
```

### C (si-core-c)

```c
#include "si_core.h"
#include <stdio.h>
#include <assert.h>

int main(void) {
    si_init();
    SiBudget *budget = budget_create(1000.0);
    budget_allocate(budget, 600.0, 400.0);

    BudgetReport rpt = budget_audit(budget);
    assert(rpt.violation == 0);
    printf("gamma=%.1f eta=%.1f total=%.1f\n", rpt.gamma, rpt.eta, rpt.total_budget);

    budget_transfer(budget, 0, 1, 50.0);
    rpt = budget_audit(budget);
    printf("After transfer: gamma=%.1f eta=%.1f\n", rpt.gamma, rpt.eta);

    budget_free(budget);
    si_shutdown();
    return 0;
}
```

### Python (si-runtime-python)

```python
from si_runtime import Budget, AgentBudget, audit, transfer

budget = Budget(total=1000.0, gamma=600.0, eta=400.0)
assert abs(budget.gamma + budget.eta - budget.total) < 1e-9
print(f"gamma={budget.gamma} eta={budget.eta} total={budget.total}")
```

### TypeScript (si-runtime-js)

```typescript
import { ConservationBudget } from 'si-runtime-js';

const budget = new ConservationBudget(1000);
budget.allocate(600, 400);
const report = budget.audit();
console.log(`gamma=${report.gamma} eta=${report.eta} total=${report.C}`);
budget.transfer('gamma', 'eta', 50);
```

### Zig (si-runtime-zig)

```zig
const conservation = @import("conservation.zig");

pub fn main() !void {
    var budget = conservation.ConservationBudget.init(1000.0);
    try budget.allocate(600.0, 400.0);
    const report = try budget.audit();
    std.debug.print("gamma={d:.1} eta={d:.1} total={d:.1}\n",
        .{ report.gamma, report.eta, report.total });
    try budget.transfer(true, 50.0);
}
```

### WASM (si-runtime-wasm — from JavaScript)

```javascript
import init, { Budget } from 'si-runtime-wasm';

async function run() {
    await init();
    const budget = new Budget(1000);
    budget.allocate(300);
    budget.transfer_gamma_to_eta(50);
    console.log(`Audit: ${budget.audit()}, gamma=${budget.gamma()}`);
}
```

---

## 2. Cross-Repo Integration

### conservation-law-rs (Mathematical Foundation)

The Go `Budget` struct enforces the same γ+η=C invariant as `conservation-law-rs`. The
`Allocate()` method rejects allocations where gamma+eta≠total. Go's `AgentBudget` adds
fleet-level tracking on top of the base conservation law.

**Connection points:**
- `NewBudget(total)` ↔ `ConservationBudget::new(C)`
- `budget.Allocate(γ, η)` ↔ `ConservationBudget::allocate(γ, η)` — same rejection logic
- `budget.Transfer(amount)` ↔ `ConservationBudget::transfer("eta", "gamma", amount)`
- `Audit(budgets)` ↔ Rust fleet-wide audit

### spectral-fleet-rs (Fleet Ranking)

The Go `spectral.go` module provides `PowerIteration()`, `SpectralRank()`, and
`AdjacencyMatrix` using the same algorithm as `spectral-fleet-rs`. Go agents contribute
eigenvector centrality scores to fleet-wide ranking.

**Connection points:**
- `PowerIteration(matrix, maxIter, tol)` ↔ Rust `power_iteration()`
- `SpectralRank(matrix)` ↔ Rust `rank()` — returns `[]RankedAgent`
- `AdjacencyMatrix.Set(i, j, value)` ↔ Rust matrix construction
- `Fleet.SpectralRank()` combines adjacency + ranking for fleet agents

### si-cli (CLI Discovery)

`si-cli` discovers Go-based agents by calling Go shared library exports via cgo. The Go
runtime exposes `Agent`, `Budget`, and `Fleet` types that the CLI queries for fleet
membership and status.

**Connection points:**
- `Agent.ListCapabilities()` → CLI capability discovery
- `Budget.Remaining()` → CLI budget display
- `Fleet.HealthReport()` → CLI health dashboard
- `Fleet.BestAgentForTask(required)` → CLI task assignment

### si-fleet-api (REST API Layer)

The fleet API serves Go agent state as JSON. The `Fleet.ConservationAudit()` method
provides `AuditResult` data that the API serializes for `GET /fleet/audit`.

**Connection points:**
- `AuditResult` → `GET /fleet/audit` response body
- `RankedAgent` → `POST /fleet/rank` response
- `Agent.String()` → `GET /agents/:id` display
- `MatchResult` → `POST /fleet/resolve` task assignment

### Supabase Fleet Registry (Data Backend)

Go agents persist state to Supabase via the fleet API. The `Budget`, `AgentBudget`,
and `AuditResult` structs map to Supabase table schemas.

**Connection points:**
- `Budget{Total, Gamma, Eta}` → `agent_budgets` table
- `Agent{ID, State, Capabilities}` → `agents` table
- `AuditResult{Valid, Violations}` → `fleet_audits` table
- `HealthReport` → `fleet_health` materialized view

### sunset-ecosystem (Fleet Coordination)

`sunset-ecosystem` coordinates multi-fleet operations. Go agents participate via the
`Fleet` type, which provides budget rebalancing, spectral ranking, and health reporting.

**Connection points:**
- `Fleet.RebalanceBudgets()` for fleet-wide budget equalization
- `Fleet.SpectralRank()` for agent ordering
- `Fleet.HealthReport()` for fleet health monitoring
- `Fleet.BestAgentForTask()` for task-to-agent assignment
- `AgentBudget.Transfer()` for cross-agent budget movement

---

## 3. FFI Bindings

### Calling si-runtime-go from C (via cgo export)

```go
// export.go
package main

import "C"
import siruntime "github.com/SuperInstance/si-runtime-go"

//export GoBudgetCreate
func GoBudgetCreate(total C.double) unsafe.Pointer {
    budget := siruntime.NewBudget(float64(total))
    return unsafe.Pointer(budget)
}

//export GoBudgetAllocate
func GoBudgetAllocate(bptr unsafe.Pointer, gamma C.double, eta C.double) C.int {
    budget := (*siruntime.Budget)(bptr)
    err := budget.Allocate(float64(gamma), float64(eta))
    if err != nil { return -1 }
    return 0
}

//export GoBudgetFree
func GoBudgetFree(bptr unsafe.Pointer) {
    // Go GC handles it
}

func main() {} // required for cgo
```

```c
// caller.c
extern void* GoBudgetCreate(double total);
extern int   GoBudgetAllocate(void* budget, double gamma, double eta);

int main(void) {
    void* budget = GoBudgetCreate(1000.0);
    GoBudgetAllocate(budget, 600.0, 400.0);
    return 0;
}
```

### Calling si-runtime-go from Rust (via C ABI)

```rust
use std::os::raw::c_double;

extern "C" {
    fn GoBudgetCreate(total: c_double) -> *mut std::ffi::c_void;
    fn GoBudgetAllocate(b: *mut std::ffi::c_void, gamma: c_double, eta: c_double) -> i32;
}

fn main() {
    unsafe {
        let budget = GoBudgetCreate(1000.0);
        let err = GoBudgetAllocate(budget, 600.0, 400.0);
        assert_eq!(err, 0);
    }
}
```

### Calling si-runtime-go from Python (via ctypes)

```python
import ctypes

lib = ctypes.CDLL("./libgoruntime.so")

lib.GoBudgetCreate.restype = ctypes.c_void_p
lib.GoBudgetCreate.argtypes = [ctypes.c_double]

lib.GoBudgetAllocate.argtypes = [ctypes.c_void_p, ctypes.c_double, ctypes.c_double]
lib.GoBudgetAllocate.restype = ctypes.c_int

budget = lib.GoBudgetCreate(1000.0)
err = lib.GoBudgetAllocate(budget, 600.0, 400.0)
assert err == 0
print(f"Go budget allocated: err={err}")
```

### Calling si-runtime-go from TypeScript (via HTTP bridge)

```typescript
// Go runtime runs as a microservice
const response = await fetch('http://localhost:8080/budget/create', {
    method: 'POST',
    body: JSON.stringify({ total: 1000 }),
});
const { budget_id } = await response.json();

await fetch(`http://localhost:8080/budget/${budget_id}/allocate`, {
    method: 'POST',
    body: JSON.stringify({ gamma: 600, eta: 400 }),
});
```

### Calling C from Go (via cgo)

```go
package main

/*
#cgo LDFLAGS: -lsi_core
#include "si_core.h"
*/
import "C"
import "fmt"

func main() {
    C.si_init()
    budget := C.budget_create(C.double(1000.0))
    err := C.budget_allocate(budget, C.double(600.0), C.double(400.0))
    fmt.Printf("C allocate result: %d\n", int(err))
    C.budget_free(budget)
    C.si_shutdown()
}
```

### Calling Rust from Go (via cgo + C ABI)

```go
package main

/*
#cgo LDFLAGS: -lconservation_law
extern void* conservation_budget_new(double total);
extern int   conservation_budget_allocate(void* budget, double gamma, double eta);
extern void  conservation_budget_free(void* budget);
*/
import "C"
import "fmt"

func main() {
    budget := C.conservation_budget_new(1000.0)
    err := C.conservation_budget_allocate(budget, 600.0, 400.0)
    fmt.Printf("Rust allocate result: %d\n", int(err))
    C.conservation_budget_free(budget)
}
```

### Calling si-runtime-go from Zig (via C ABI)

```zig
const c = @cImport(@cInclude("go_runtime.h"));

pub fn callGoBudget() !void {
    const budget = c.GoBudgetCreate(1000.0) orelse return error.CreationFailed;
    const err = c.GoBudgetAllocate(budget, 600.0, 400.0);
    if (err != 0) return error.AllocationFailed;
    c.GoBudgetFree(budget);
}
```

---

## Integration Test Matrix

| From → To | C | Rust | Python | TypeScript | Zig | Go | WASM |
|---|---|---|---|---|---|---|---|
| **Go** | cgo | cgo + C ABI | ctypes | HTTP bridge | cgo | ✅ native | N/A |
| **C** | ✅ native | cdylib | ctypes | ffi-napi | `@cImport` | cgo | emscripten |
| **Rust** | extern "C" | ✅ native | PyO3 | wasm-bindgen | C ABI | cgo | wasm-bindgen |
| **Python** | ctypes | PyO3 | ✅ native | pythonia | ctypes | ctypes | N/A |
| **TypeScript** | ffi-napi | wasm-bindgen | pythonia | ✅ native | ffi-napi | HTTP | wasm API |
| **Zig** | `@cImport` | C ABI | C API | ffi-napi | ✅ native | cgo | N/A |
| **WASM** | emscripten | wasm-bindgen | N/A | JS import | N/A | N/A | ✅ native |

---

## Go Package API Summary

| Type/Function | Description |
|---|---|
| `Budget` | Conserved resource pool: γ+η=Total |
| `NewBudget(total)` | Create budget |
| `budget.Allocate(γ, η)` | Set gamma and eta (must sum to total) |
| `budget.Transfer(amount)` | Move from eta to gamma |
| `budget.Remaining()` | γ+η |
| `AgentBudget` | Agent + Budget pair |
| `AgentBudget.Transfer(to, amount)` | Cross-agent budget transfer |
| `Audit(budgets)` | Fleet-wide audit → `AuditResult` |
| `Agent` | Agent with state, capabilities, homeostasis |
| `NewAgent(id)` | Create agent |
| `agent.UpdateHomeostasis(rate)` | PID regulation |
| `agent.HomeostasisError()` | RMS deviation from targets |
| `Cell` | Discrete unit in cellular mesh |
| `Grid` | Rectangular lattice of cells |
| `AdjacencyMatrix` | Dense symmetric affinity matrix |
| `PowerIteration(m, iter, tol)` | Dominant eigenpair |
| `SpectralRank(m)` | Agent indices by eigenvector centrality |
| `Fleet` | Multi-agent collection |
| `Fleet.SpectralRank()` | Fleet-wide ranking |
| `Fleet.ConservationAudit()` | Fleet-wide budget audit |
| `Fleet.RebalanceBudgets()` | Equalize eta fraction |
| `Fleet.HealthReport()` | Fleet health diagnostics |
| `Fleet.BestAgentForTask(reqs)` | Task-to-agent matching |
| `Capability` | Skill/function registration |
| `Registry` | Capability storage and query |
| `Match(agentID, agentCaps, required)` | Compatibility scoring |
| `Resolve(candidates)` | Best match selection |

---

*Generated for SuperInstance cross-language integration — si-runtime-go v0.1.0*

# MAESTRO_CONTEXT

## Stato documentazione

Completati:

- identity.md
- philosophy.md
- principles.md
- vision.md
- architecture.md
- roadmap.md
- design-decisions.md
- README.md

ADR completate:

- ADR-0001 — Scelta del nome Maestro
- ADR-0002 — Provider Abstraction
- ADR-0003 — Plugin Architecture
- ADR-0004 — Capability-Based Runtime Architecture

---

## Stato architetturale

L'architettura logica del progetto è stata consolidata.

Macro-componenti identificati:

- Runtime
- Provider Layer
- Plugin System
- Context Engine
- Tool System
- Agent System
- Gestor

Il Runtime rappresenta il cuore dell'intero sistema.

Principi architetturali adottati:

- API First
- Composition over Inheritance
- Capability-Based Architecture
- Separation of Identity, Capability, State and Orchestration
- Minimal Public Contracts
- Internal Implementations hidden under internal/

Il contratto pubblico (`pkg/runtime`) è stato definito e costituisce la base stabile per lo sviluppo futuro.

---

## Runtime Public API

Sono stati definiti i contratti pubblici per:

- Runtime
- Component
- Context
- Registry
- Service
- EventBus
- Event
- StateManager
- LifecycleManager
- Config
- Logger

Sono stati inoltre introdotti i tipi fondamentali:

- ComponentID
- Metadata
- Dependency
- Capability
- State
- Transition
- ComponentState

Le operazioni del ciclo di vita non fanno parte di `Component`.

Ogni componente espone esclusivamente:

```go
type Component interface {
    Metadata() Metadata
}
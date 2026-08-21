# Milestone 1 — Report retrospettivo Fase 1

Fase: Core Types & Public Interfaces

Stato: Completata

Data di completamento: 2026-08-04

Natura del documento: ricostruzione retrospettiva

---

# Obiettivo

Definire il contratto pubblico minimale del Runtime Core prima di introdurre
registry, dependency graph e implementazioni concrete.

---

# Risultati consegnati

- Package pubblico `pkg/runtime` per i contratti stabili del core.
- Interfacce per Runtime, Component, Context, Registry, EventBus,
  StateManager, LifecycleManager, Config e Logger.
- Value object per identità, metadata, dipendenze, capability, stati e
  transizioni.
- Lifecycle modellato tramite capability opzionali: Configurer, Initializer,
  Starter, Stopper, Reloader e HealthChecker.
- Errori pubblici minimali e idiomatici.
- ADR-0004 per l'architettura capability-based.

---

# Decisioni principali

- `Component` descrive l'identità; non incorpora un lifecycle obbligatorio.
- Le capacità operative sono interfacce indipendenti e componibili.
- Identità, capability, stato e orchestrazione restano responsabilità separate.
- Le API pubbliche vivono sotto `pkg/`; le implementazioni concrete saranno
  mantenute sotto `internal/`.

---

# Evidenze storiche

La prima superficie pubblica fu introdotta dal commit `1fa7488` e consolidata
nel commit `8c55498`; ADR-0004 fu registrata in `f4af1ad`. Il contesto di
progetto successivo registra la fase come completata e conserva l'elenco dei
contratti e dei tipi consegnati.

---

# Handoff alla Fase 2

La Fase 2 può implementare registry, grafo, resolver, validator e builder senza
ampliare il contratto pubblico stabilito in questa fase.

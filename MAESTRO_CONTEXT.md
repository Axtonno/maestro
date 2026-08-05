# MAESTRO_CONTEXT

## Stato del progetto

**Nome:** Maestro

Maestro è un Runtime per sistemi AI locali, progettato per orchestrare componenti, provider, plugin e servizi attraverso un'architettura modulare, capability-based e fortemente orientata alla composizione.

L'obiettivo del progetto non è fornire un singolo agente AI, ma costituire il Runtime sul quale costruire un intero ecosistema di strumenti intelligenti.

---

# Stato della documentazione

Completati:

* README.md
* identity.md
* philosophy.md
* principles.md
* vision.md
* architecture.md
* roadmap.md
* design-decisions.md
* runtime-internals.md

---

# ADR approvate

## ADR-0001

Scelta del nome **Maestro**.

---

## ADR-0002

Provider Abstraction.

I provider rappresentano un layer indipendente dal Runtime.

---

## ADR-0003

Plugin Architecture.

I plugin sono componenti registrabili e orchestrati dal Runtime.

---

## ADR-0004

Capability-Based Runtime Architecture.

Il contratto pubblico del Runtime è costruito secondo il principio:

* Component descrive l'identità.
* Capability descrivono ciò che il componente sa fare.
* Runtime orchestra il comportamento.

Le capacità operative sono espresse tramite interfacce opzionali (`Starter`, `Initializer`, `Stopper`, ecc.).

---

# Principi architetturali consolidati

L'architettura di Maestro è oggi fondata sui seguenti principi.

## API First

Prima vengono progettati i contratti pubblici.

Successivamente vengono realizzate le implementazioni interne.

---

## Composition over Inheritance

Il comportamento emerge dalla composizione di piccoli componenti.

---

## Capability-Based Architecture

Le capacità operative sono rappresentate da interfacce indipendenti.

---

## Separation of Responsibilities

Sono completamente separate:

* Identità
* Capacità
* Stato
* Orchestrazione

---

## Internal First

Le implementazioni concrete rimangono sotto `internal/`.

Le API pubbliche rimangono minimali e stabili.

---

## Ownership of Invariants

Ogni tipo interno protegge gli invarianti del proprio livello di responsabilità.

Gli invarianti locali appartengono al tipo proprietario.

Gli invarianti che coinvolgono più oggetti appartengono all'aggregato che li coordina.

Questo principio è formalizzato nel documento:

```
docs/runtime-internals.md
```

---

# Stato del Runtime Core

## Fase 1 — Core Types & Public Interfaces

Completata.

Contratti pubblici definiti.

Package:

```
pkg/runtime
```

API pubbliche disponibili:

* Runtime
* Component
* Context
* Registry
* EventBus
* StateManager
* LifecycleManager
* Config
* Logger

Tipi fondamentali:

* ComponentID
* Metadata
* Dependency
* Capability
* State
* Transition
* ComponentState

Capability:

* Configurer
* Initializer
* Starter
* Stopper
* Reloader
* HealthChecker

---

## Fase 2 — Dependency Container & Registry

Completata.

Package:

```
internal/runtime
```

Componenti implementati:

### runtime

Composition root del Runtime.

Coordina:

* Registry
* Builder
* Dependency Graph

---

### registry

Registry thread-safe.

Responsabilità:

* registrazione componenti
* lookup
* prevenzione duplicati

---

### node

Rappresentazione interna di un nodo del grafo.

Responsabilità:

* componente
* dipendenze
* dipendenti

---

### graph

Gestione del DAG.

Responsabilità:

* aggiunta nodi
* collegamento dipendenze
* mantenimento delle relazioni

---

### resolver

Costruzione del grafo leggendo i Metadata.

Responsabilità:

* creazione nodi
* risoluzione dipendenze
* gestione dipendenze opzionali

---

### validator

Validazione del grafo.

Controlli implementati:

* Metadata validi
* Dependency duplicate
* Capability duplicate
* Self dependency
* Dependency cycle

Algoritmo:

Depth First Search.

---

### builder

Coordina:

Resolver

↓

Validator

↓

Graph validato

Builder completamente stateless.

---

# Convenzioni implementative

Per `internal/runtime` sono state adottate le seguenti regole.

* Campi sempre privati.
* Nessuna struttura mutabile viene esposta direttamente.
* Le modifiche avvengono tramite metodi intenzionali.
* I costruttori garantiscono gli invarianti iniziali.
* node mantiene gli invarianti locali.
* graph mantiene gli invarianti delle relazioni.
* registry mantiene gli invarianti della registrazione.
* resolver costruisce ma non valida.
* validator valida ma non modifica.
* builder coordina senza conservare stato.
* runtime orchestra senza duplicare responsabilità.

---

# Stato dei test

Suite di test completata.

Package:

```
internal/runtime
```

Copertura funzionale:

* Registry
* Node
* Graph
* Resolver
* Validator
* Builder
* Runtime

Verifica effettuata con:

```
go test ./internal/runtime
```

Risultato:

Tutti i test superati.

---

# Stato del repository

Documentazione consolidata.

API pubbliche consolidate.

Runtime interno implementato fino alla costruzione e validazione del Dependency Graph.

Il progetto è ora pronto per introdurre il Lifecycle Engine.

---

# Roadmap corrente

## Runtime Core

### ✅ Fase 1

Core Types & Public Interfaces

---

### ✅ Fase 2

Dependency Container & Registry

---

### 🔜 Fase 3

Lifecycle Engine

Obiettivi:

* LifecycleManager
* Component State
* State Machine
* Topological Sort
* Startup ordinato
* Shutdown ordinato
* Gestione delle Capability
* Context Runtime
* Bootstrap completo

---

### Fase 4

Event System

---

### Fase 5

Provider Runtime

---

### Fase 6

Plugin Runtime

---

# Stato architetturale

Le fondamenta del Runtime possono essere considerate stabili.

Le API pubbliche risultano minimali e orientate all'estensibilità.

Le implementazioni interne rispettano il principio di proprietà degli invarianti e la separazione delle responsabilità.

Le prossime evoluzioni interesseranno principalmente il Lifecycle Engine, senza richiedere modifiche significative ai contratti pubblici già definiti.

Il Runtime Core dispone ora di una base sufficientemente solida per sostenere le fasi successive dello sviluppo di Maestro.

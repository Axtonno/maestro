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
* event-system.md
* provider-runtime.md
* ollama-provider.md
* llamacpp-provider.md
* plugin-runtime.md
* laravel-plugin.md

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

## ADR-0005

Synchronous In-Process Event Bus.

L'Event Bus usa dispatch sincrono, topic esatti, fan-out ordinato e snapshot dei
subscriber. Le operazioni sono thread-safe e i callback non vengono eseguiti
mentre il bus mantiene lock interni.

---

## ADR-0006

Capability-Based Provider Runtime.

Il provider descrive la propria identità. Completion, streaming, embedding e
model listing sono capability opzionali. Il Provider Runtime registra, risolve
e orchestra i provider senza dipendere dalle loro implementazioni concrete.

---

## ADR-0007

Component-Based In-Process Plugin Runtime.

I plugin sono componenti registrati attraverso un registry dedicato. Il Plugin
Runtime possiede classificazione e risoluzione; dependency graph, stato e
lifecycle restano responsabilità del Runtime Core. Il modello supporta plugin Go
fidati caricati e registrati in-process.

---

## ADR-0008

Trusted In-Process Plugin Catalog.

Discovery e caricamento usano un catalogo thread-safe di loader in-process. Il
manifest dichiara la versione dell'API Plugin Runtime e i loader vengono
eseguiti senza lock. Plugin e loader sono codice fidato con i privilegi del
processo Maestro.

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

## Fase 3 — Lifecycle Engine

Completata.

Package:

```
internal/runtime
```

Componenti implementati:

### stateManager

State manager thread-safe.

Responsabilità:

* tracciamento dello stato dei componenti
* inizializzazione a `StateCreated` durante la registrazione
* transizioni controllate tramite state machine interna
* marcatura dei componenti falliti con causa

---

### lifecycleManager

Motore lifecycle per singolo componente.

Responsabilità:

* esecuzione ordinata delle capability opzionali
* `Configure`
* `Initialize`
* `Start`
* `Stop`
* `Reload`
* `Health`
* aggiornamento coerente dello stato
* propagazione degli errori e passaggio a `StateFailed`

---

### runtimeContext

Context interno passato ai componenti durante il lifecycle.

Espone:

* Config
* Logger
* EventBus
* Registry

---

### graph lifecycle ordering

Il grafo espone ora:

* ordinamento topologico stabile
* ordinamento topologico inverso

Questi ordinamenti sono usati per garantire:

* startup dependency-first
* shutdown dependent-first

---

### runtime bootstrap

Il Runtime interno implementa ora il contratto pubblico `pkg/runtime.Runtime`.

Responsabilità aggiunte:

* bootstrap automatico del Dependency Graph al primo `Start`
* avvio ordinato dei componenti
* shutdown ordinato dei componenti
* protezione da start duplicati
* protezione da stop prima dello start
* blocco delle registrazioni durante o dopo lo start
* accesso pubblico a Registry, EventBus e StateManager

---

## Fase 4 — Event System

Completata.

Package:

```
pkg/runtime
internal/runtime
```

Componenti implementati:

### eventBus

Event Bus in-process thread-safe.

Responsabilità:

* validazione di eventi, topic e handler
* dispatch sincrono
* più handler per topic
* ordine deterministico di sottoscrizione
* snapshot dei subscriber per pubblicazione
* callback eseguiti senza lock interni
* rimozione idempotente di tutti gli handler di un topic

Semantica e limiti della prima versione sono descritti in:

```
docs/event-system.md
```

---

## Fase 5 — Provider Runtime/Configuration

In corso.

Primo incremento implementato nei package:

```
pkg/provider
internal/provider
pkg/runtime
internal/runtime
```

Componenti disponibili:

* identità `Provider` separata dalle capability operative
* capability `Completer`, `Streamer`, `Embedder` e `ModelLister`
* tipi condivisi per messaggi, completion, stream, embedding e modelli
* Provider Runtime thread-safe
* registrazione e risoluzione provider
* selezione esplicita del provider predefinito
* routing delle capability senza lock durante codice esterno
* streaming pull-based con chiusura esplicita
* configurazione minimale tramite `runtime.NewConfig`
* composition root pubblico `maestro.New`
* Provider Runtime e Config condivisi con il `runtime.Context`

Semantica e limiti sono descritti in:

```
docs/provider-runtime.md
```

Su questi contratti è ora disponibile l'adapter Ollama. Restano da introdurre
ulteriori provider e le eventuali policy di resilienza richieste da casi d'uso
reali.

### Secondo incremento — Adapter Ollama

Completato.

Package:

```
pkg/provider/ollama
internal/provider/ollama
```

Funzionalità:

* facade pubblica e configurazione tipizzata
* implementazione privata basata sulla libreria standard `net/http`
* completion tramite `POST /api/chat`
* streaming NDJSON tramite `POST /api/chat`
* embedding multipli tramite `POST /api/embed`
* model listing tramite `GET /api/tags`
* risoluzione tra modello esplicito e modello predefinito
* validazione semantica delle risposte
* error handling HTTP con body limitato
* propagazione di cancellazione e deadline
* test HTTP in-memory
* test d'integrazione opzionali tramite build tag

Semantica, configurazione e limiti sono descritti in:

```
docs/ollama-provider.md
```

Gate di chiusura pendente:

* smoke test live contro un'istanza Ollama
* listing dei modelli
* completion non-streaming
* streaming fino a chiusura regolare
* embedding con un modello compatibile
* cancellazione dello stream e chiusura delle risorse

L'indisponibilità di un'istanza Ollama nell'ambiente corrente non costituisce
un difetto dell'adapter, ma impedisce di dichiarare completata la verifica live.

---

## Fase 6 — Plugin Runtime

Completata.

Implementazione disponibile nei package:

```
pkg/plugin
pkg/plugin/laravel
internal/plugin
internal/plugin/laravel
internal/runtime
```

Componenti disponibili:

* contratto pubblico `Plugin` basato su `runtime.Component`
* alias dell'identità plugin a `runtime.ComponentID`
* manifest con versione dell'API Plugin Runtime
* Plugin Runtime pubblico con registrazione, risoluzione e listing
* registry e catalogo loader thread-safe
* discovery deterministica tramite `Available` e `Registered`
* loader interface e adapter `LoaderFunc`
* caricamento cancellabile senza lock durante codice esterno
* validazione di ID, risultato del loader e compatibilità
* validazione di plugin nil, typed nil e ID
* coordinamento della registrazione con il Runtime Core
* collisioni coerenti tra plugin e normali componenti
* eventi per loader registrato, plugin registrato e plugin caricato
* integrazione nel composition root `maestro.New`
* riuso del dependency graph e del lifecycle globali
* primo plugin Laravel con detection, versione framework e health
* test unitari, concorrenti e d'integrazione

Semantica e limiti sono descritti in:

```
docs/plugin-runtime.md
docs/laravel-plugin.md
```

Il gate della fase usa plugin Go fidati caricati in-process. Packaging remoto,
firme, sandbox, process isolation, unload e hot replacement sono estensioni
dell'ecosistema e non modificano i contratti completati.

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
* stateManager mantiene gli invarianti degli stati dei componenti.
* lifecycleManager orchestra le capability senza conoscere la logica di dominio.
* eventBus protegge i subscriber e non mantiene lock durante i callback.
* provider runtime protegge registry e default senza mantenere lock durante le chiamate esterne.
* plugin runtime protegge registry e catalogo, esegue loader senza lock e delega stato e lifecycle al Runtime Core.
* runtime orchestra senza duplicare responsabilità.

---

# Stato dei test

Suite di test completata.

Package:

```
internal/runtime
internal/provider
internal/provider/ollama
internal/provider/llamacpp
internal/plugin
internal/plugin/laravel
pkg/runtime
pkg/plugin
pkg/plugin/laravel
pkg/provider/ollama
pkg/provider/llamacpp
```

Copertura funzionale:

* Registry
* Node
* Graph
* Resolver
* Validator
* Builder
* Runtime
* Event Bus
* Provider Runtime
* Plugin Runtime
* Plugin Laravel
* Configurazione
* Adapter Ollama
* Adapter llama.cpp

Verifica effettuata con:

```
GOCACHE=/tmp/maestro-go-build go test ./...
```

Risultato:

Tutti i test superati.

---

# Stato del repository

Documentazione consolidata.

API pubbliche estese con i contratti provider, plugin e il composition root.

Runtime interno implementato fino alla chiusura del Plugin Runtime.

Il progetto dispone di due adapter provider concreti, di un catalogo per plugin
trusted in-process e del primo plugin framework-aware Laravel.

La Milestone 2 — Provider Layer è iniziata con l'adapter llama.cpp. Facade
pubblica, protocollo HTTP interno, completion, streaming SSE, embedding, model
listing, autenticazione Bearer opzionale, test isolati e smoke test opzionale
sono implementati. Resta pendente la verifica live contro `llama-server`.

---

# Roadmap corrente

## Runtime Core

### ✅ Fase 1

Core Types & Public Interfaces

---

### ✅ Fase 2

Dependency Container & Registry

---

### ✅ Fase 3

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

### ✅ Fase 4

Event System

---

### 🚧 Fase 5

Provider Runtime/Configuration

Implementazione completata. Fase ancora in corso esclusivamente per lo smoke
test live dell'adapter Ollama.

---

### ✅ Fase 6

Plugin Runtime

Completata: contratti, manifest, registry, catalogo loader, discovery,
caricamento, eventi, lifecycle condiviso e primo plugin Laravel.

---

# Stato architetturale

Le fondamenta del Runtime possono essere considerate stabili.

Le API pubbliche risultano minimali e orientate all'estensibilità.

Le implementazioni interne rispettano il principio di proprietà degli invarianti e la separazione delle responsabilità.

Gli adapter e le policy provider future appartengono alla Milestone 2 — Provider
Layer e possono evolvere indipendentemente dal Plugin Runtime completato.

Il Runtime Core dispone ora di una base sufficientemente solida per sostenere
plugin trusted in-process senza duplicare grafo, stati o lifecycle.

---

# Provider Layer — Fase 1

## 🚧 Adapter llama.cpp

Implementazione e test isolati completati. Gate di chiusura pendente:

* smoke test live contro `llama-server`;
* model listing;
* completion non-streaming;
* streaming SSE fino a `[DONE]`;
* embedding con un modello compatibile;
* cancellazione dello stream e chiusura delle risorse.

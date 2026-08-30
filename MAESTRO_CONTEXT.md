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
* provider-model-lifecycle.md
* provider-model-acquisition.md
* provider-model-residency.md
* provider-capability-introspection.md
* provider-layer-plan.md
* benchmark-evaluation-plan.md
* benchmark-runtime.md
* smoke-benchmark.md
* plugin-runtime.md
* laravel-plugin.md
* plugin-system-design.md
* plugin-system-development-plan.md
* plugin-api-compatibility-audit.md
* context-engine-api-compatibility-audit.md
* context-engine-indexing.md
* context-engine-analysis.md
* context-engine-retrieval.md
* context-engine-cache.md
* context-engine-design.md
* context-engine-development-plan.md
* agent-system-api-compatibility-audit.md
* tool-runtime.md
* agent-permissions.md
* agent-sessions.md
* agent-runtime.md
* agent-workspace.md
* agent-system-design.md
* agent-system-development-plan.md
* release-readiness-audit.md
* milestone-8-design.md
* milestone-8-development-plan.md
* configuration.md
* cli.md
* operational-experience.md
* installation.md
* packaging-candidate.md
* v0.2.0-development-plan.md
* milestone-10-development-plan.md
* milestone-11-development-plan.md
* milestone-12-development-plan.md
* milestone-13-field-validation-plan.md
* field-validation-task-matrix.md
* milestone-14-interaction-modes-direct-chat-plan.md
* milestone-15-reference-hardware-readonly-baseline-plan.md
* milestone-16-controlled-mutation-recovery-plan.md
* milestone-17-mutation-qualification-plan.md
* milestone-18-productization-release-v0.3.0-plan.md
* milestone-18-productization-v0.4.0-plan.md
* milestone-19-post-release-adoption-lower-bound-validation-plan.md
* mutation-qualification.md
* mutation-benchmark.md
* reports/milestone-11-final.md
* reports/milestone-12-phase-1.md
* reports/milestone-12-phase-2.md
* reports/milestone-12-phase-3.md
* reports/milestone-12-phase-4.md
* reports/milestone-18-phase-1.md
* reports/milestone-18-phase-2.md
* reports/milestone-18-phase-3.md
* reports/milestone-18-phase-4.md
* reports/milestone-18-phase-5.md
* reports/milestone-19-thinkpad-adoption.md

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

## ADR-0009

Capability-Based Model Discovery and Lifecycle.

Discovery arricchita, load e unload sono capability provider opzionali e
indipendenti. Lo stato modello è uno snapshot autorevole del provider e non
viene duplicato nel Runtime.

---

## ADR-0026

Release-Oriented v0.1.0 Product Boundary.

Il gate post-Milestone 7 dà GO alla Milestone 8 come productization. La
piattaforma iniziale supportata è Linux `amd64`; Ollama con `llama3.1:8b` e
`embeddinggemma:latest` è positivo per provider-level e reference agent
read-only, ma non per il reference agent mutante. llama.cpp resta
condizionato alla presenza del report live. CLI, configurazione e package
pubblici sono sperimentali; il modello resta trusted in-process e gli ambiti
rinviati non entrano nella v0.1.0.

---

## ADR-0027

Apache License 2.0 for Maestro.

Il repository e ogni artifact includono il testo Apache-2.0. `NOTICE` e
`THIRD_PARTY_LICENSES.txt` conservano attribution e termini delle dipendenze
distribuite. La scelta è vincolante dalla Fase 4; la Fase 6 ne verifica la
pubblicazione definitiva.

---

## ADR-0031

Contratto Controlled Mutation v0.2.0.

Il candidato limita l'autorità a una singola patch esatta su un file PHP
esistente sotto `app/`, con read verificata, preview, TTY, approval one-shot,
commit atomico Linux e reindex obbligatorio prima del final.

---

## ADR-0032

Rinvio della Controlled Mutation dopo la qualificazione.

La matrice deterministica e il preflight sono positivi, ma Gate A fallisce al
primo tentativo. L'esito è `mutation_deferred`; Gate B/C non vengono eseguiti e
la Milestone 12 riceve un GO limitato alla productization read-only.

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

Completata.

Primo incremento implementato nei package:

```
pkg/provider
internal/provider
pkg/runtime
internal/runtime
```

Componenti disponibili:

* identità `Provider` separata dalle capability operative
* capability `Completer`, `Streamer`, `Embedder`, `ModelLister`,
  `ModelDiscoverer`, `ModelLoader` e `ModelUnloader`
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

Smoke test trasferito al Livello 1 della Milestone 3 — Benchmark & Evaluation
Layer:

* smoke test live contro un'istanza Ollama
* listing dei modelli
* completion non-streaming
* streaming fino a chiusura regolare
* embedding con un modello compatibile
* cancellazione dello stream e chiusura delle risorse

L'indisponibilità di un'istanza Ollama nell'ambiente corrente non costituisce
un difetto dell'adapter e non blocca più la chiusura della fase del Runtime
Core.

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
* Model discovery e lifecycle provider
* Model acquisition, progress e removal provider
* Model residency policies provider
* Provider capability introspection

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
sono implementati. La verifica live contro `llama-server` confluisce nello
Smoke Benchmark della Milestone 3.

La Fase 2 della Provider Layer aggiunge `ModelInfo`, stati neutrali e capability
indipendenti per discovery, load e unload. Il Provider Runtime espone il routing
senza possedere stato modello. Ollama unisce `/api/tags` e `/api/ps` e governa
il lifecycle tramite `keep_alive`; llama.cpp usa il catalogo e gli endpoint del
router mode.

La Fase 3 aggiunge pull con progresso e rimozione attraverso capability
indipendenti. Ollama usa gli endpoint nativi `/api/pull` e `/api/delete`;
llama.cpp avvia il download con `/models`, osserva `/models/sse`, usa
`/models/unload` per la cancellazione remota e rimuove soltanto dalla cache del
router. Il Provider Runtime non possiede trasferimenti né accede al filesystem.

La Fase 4 aggiunge `ModelResidencyPolicy` per coppia provider–model. Autoload è
opt-in; il rilascio può essere immediato, a TTL o allo shutdown. Discovery resta
la fonte osservabile dello stato remoto, mentre il Runtime coordina lease e
timer soltanto per le residenze caricate da Maestro. Completion, stream ed
embedding sono coperti da test deterministici, inclusa la concorrenza.

La Fase 5 aggiunge `CapabilityInspector` e report neutrali per adapter, istanza
e modello. `Support` descrive il contratto implementato da Maestro;
`Availability` descrive lo snapshot operativo come available, unavailable o
unknown. Ollama usa `/api/tags` e `/api/show`; llama.cpp usa `/models`, stato
router e argomenti del processo. I report non sono memorizzati e non effettuano
selezione automatica.

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

### ✅ Fase 5

Provider Runtime/Configuration

Implementazione completata. Gli smoke test live provider confluiscono nello
Smoke Benchmark della Milestone 3.

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

# Provider Layer

## ✅ Fase 1 — Adapter llama.cpp

Implementazione, test isolati e documentazione completati. Lo smoke test live
confluisce nello Smoke Benchmark della Milestone 3.

## ✅ Fase 2 — Model Discovery & Lifecycle

Completati:

* `ModelInfo` e `ModelState` neutrali;
* capability `ModelDiscoverer`, `ModelLoader` e `ModelUnloader`;
* routing tramite `provider.Runtime`;
* discovery Ollama attraverso `/api/tags` e `/api/ps`;
* load/unload Ollama tramite `keep_alive`;
* discovery e lifecycle llama.cpp in router mode;
* test isolati e d'integrazione opzionali;
* ADR-0009 e documentazione dedicata.

## ✅ Fase 3 — Model Acquisition & Removal

Completati:

* capability `ModelPuller` e `ModelRemover`;
* `ModelPullStream` pull-based con stage neutrali;
* routing tramite `provider.Runtime`;
* pull Ollama tramite `/api/pull` e remove tramite `/api/delete`;
* pull llama.cpp tramite `/models` e `/models/sse`;
* cancellazione remota llama.cpp tramite `/models/unload`;
* remove llama.cpp tramite `DELETE /models`;
* validazione del progresso, chiusura delle risorse e test isolati;
* ADR-0010 e documentazione dedicata.

Gli smoke test live che modificano il catalogo confluiscono nello Smoke
Benchmark della Milestone 3.

## ✅ Fase 4 — Model Residency Policies

Completati:

* contratto pubblico `ModelResidencyPolicy` e configurazione per provider e
  model ID esatto;
* autoload opt-in prima di completion, streaming ed embedding;
* rilascio immediato, TTL deterministico e permanenza fino allo shutdown;
* coalescing dei load concorrenti e lease mantenuto fino alla fine degli stream;
* ownership limitata ai modelli caricati dalla policy;
* integrazione dello shutdown del Provider Runtime nel Runtime Core;
* semantica comune sui loader/unloader Ollama e llama.cpp;
* ADR-0011 e documentazione dedicata.

Gli smoke test live delle policy confluiscono nello Smoke Benchmark della
Milestone 3.

## ✅ Fase 5 — Capability Introspection

Completati:

* capability neutrali per completion, streaming, embedding, catalogo,
  lifecycle, acquisition, structured output e tool calling;
* target espliciti adapter, instance e model;
* separazione tra supporto strutturale e disponibilità operativa;
* routing e validazione canonica dei report nel Provider Runtime;
* introspection Ollama tramite catalogo e `/api/show`;
* introspection llama.cpp tramite `/models`, modalità router e argomenti;
* test di assenza I/O, variazione del catalogo e concorrenza;
* ADR-0012 e documentazione dedicata.

Gli smoke test live dell'introspection confluiscono nello Smoke Benchmark della
Milestone 3.

## ✅ Fase 6 — Error Semantics

Completati:

* envelope pubblico `ProviderError` con kind, operazione, provider, modello,
  status, ritentabilità e dettagli remoti strutturati;
* sentinel neutrali e compatibilità con `ErrInvalidRequest`,
  `ErrInvalidResponse` ed `ErrUnsupportedCapability`;
* preservation di cause, `context.Canceled`, `context.DeadlineExceeded` e
  `io.EOF` tramite semantica idiomatica `errors.Is`/`errors.As`;
* mapping HTTP comune e mapping del `type` OpenAI-like di llama.cpp;
* classificazione di operazioni sincrone, trasporto e stream per Ollama e
  llama.cpp;
* normalizzazione e limite dei dettagli remoti;
* ADR-0013 e documentazione dedicata.

La classificazione non effettua retry: idempotenza, backoff e circuit breaker
sono applicati separatamente dalla Fase 7.

## ✅ Fase 7 — Resilience Policies

Completati:

* `ResiliencePolicy` per provider, operazione e modello opzionale;
* retry finiti con backoff esponenziale saturato, jitter e budget temporale;
* matrice esplicita di ripetibilità delle operazioni;
* retry dello streaming soltanto prima del primo chunk consegnato;
* circuit breaker closed/open/half-open con probe concorrenti limitati;
* snapshot tipizzati tramite `CircuitState`;
* composizione con discovery, load e unload delle policy di residenza;
* clock, attesa e jitter sostituibili nei test;
* ADR-0014 e documentazione dedicata.

Le policy sono opt-in e usano esclusivamente `ProviderError.Retryable`; fallback
e routing multi-provider rimangono fuori scope.

## ✅ Fase 8 — Provider Observability

Completati:

* contratti pubblici `ProviderObserver`, `ProviderObserverFunc` e
  `ProviderEvent` senza dipendenze da SDK telemetrici;
* correlazione tramite operation ID di start, tentativi, retry, transizioni del
  circuito e terminale;
* copertura dei confini pubblici per completion, stream, embedding, catalogo,
  lifecycle, acquisition, removal e capability introspection;
* terminale unico per EOF, errore, cancellazione, pull completato e chiusura
  anticipata degli stream;
* redazione strutturale di prompt, risposte, chunk, embedding, credenziali e
  payload remoti;
* isolamento di errori e panic degli observer e invocazione senza lock interni;
* fast path senza observer privo di tracker, eventi e allocazioni;
* ADR-0015 e documentazione dedicata.

Gli adapter verso logging, metriche e tracing restano applicativi; le misure
live e di risorse appartengono alla Milestone 3.

## ✅ Fase 9 — Advanced Generation Baseline

Completati:

* opzioni comuni per limite token, temperatura, `top_p` e stop sequence;
* output strutturati JSON e JSON Schema;
* definizioni tool, choice, chiamate, risultati e storia conversazionale;
* delta tool negli stream e validazione del risultato terminale;
* traduzione e test isolati per Ollama e llama.cpp;
* capability introspection aggiornata con disponibilità e limiti operativi;
* ADR-0016 e documentazione dedicata.

Multimodalità, reasoning e opzioni proprietarie restano fuori scope.

## ✅ Fase 10 — Hardening & Provider Handoff

La Milestone 2 — Provider Layer è conclusa. Il gate comprende suite completa,
race detector, vet, audit delle API e della documentazione e compilazione delle
suite di integrazione senza servizi live. Il manifest
`docs/provider-smoke-benchmark-manifest.yaml` assegna gli scenari live alla
Milestone 3 con modelli fixture, configurazione, protezioni delle mutazioni,
cleanup e redazione espliciti.

Nuovi adapter, fallback multi-provider, selezione hardware-aware, supervisione
dei processi, multimodalità e reasoning non sono requisiti di chiusura della
Milestone 2.

## Gate deterministico della Milestone 2

Stato: Superato.

* audit di compatibilità delle API pubbliche;
* routing e capability coperti da test isolati;
* adapter Ollama e llama.cpp verificati tramite trasporti HTTP in-memory;
* completion, streaming, embedding, lifecycle, pull e remove coperti senza
  servizi esterni;
* introspection, resilienza e osservabilità verificate in modo deterministico;
* output strutturati e tool calling coperti sui due adapter;
* suite completa, race detector, vet e audit della documentazione;
* manifest degli scenari live consegnato alla Milestone 3.

## Milestone 3 — Benchmark & Evaluation Layer

Stato: Completata — Fasi 1–5 e decisione live concluse.

La Fase 1 — Benchmark Contracts & Runner consegna:

* contratti pubblici in `pkg/benchmark`;
* manifest schema `1` con loader YAML strict;
* runner con warmup, run, timeout, cleanup e classificazione degli errori;
* aggregati deterministici e p95 da almeno 20 campioni;
* report JSON schema `1.0.0` con redazione al confine di serializzazione;
* base CLI `maestro bench` e comando `maestro bench validate`;
* documentazione in `docs/benchmark-runtime.md`;
* report finale in `docs/reports/milestone-3-phase-1.md`;
* ADR-0017 per contratti versionati e runner deterministico interno.

La Fase 2 — Smoke Benchmark consegna:

* composition root live per Ollama e llama.cpp;
* tutti i 14 scenari del manifest con capability preflight;
* distinzione tra `unsupported`, `skipped` e `failed`;
* modelli fixture per ruolo e mutation guard acquisition;
* cleanup di stream, pull, lifecycle, resilience e observer;
* comando `maestro bench smoke` e report JSON atomico `0600`;
* report schema `1.1.0` con `configuration.models`;
* documentazione in `docs/smoke-benchmark.md`;
* report finale in `docs/reports/milestone-3-phase-2.md`;
* ADR-0018 per matrice live esplicita e mutation-safe.

La Fase 3 — Runtime Benchmark consegna:

* manifest Runtime con scenari provider e modello separati;
* comandi `maestro bench provider` e `maestro bench model`;
* latenza completion, TTFT e throughput stream quando l'usage è disponibile;
* cancellazione misurata di generazione e pull;
* embedding batch, lifecycle load/unload e confronto cold/warm;
* retry e circuit breaker attraverso fault transitori controllati;
* sampler Linux opzionale per CPU e RAM con scope di processo dichiarato;
* assenza esplicita di metriche non osservabili, inclusa la VRAM;
* documentazione e report finale in `docs/reports/milestone-3-phase-3.md`;
* ADR-0019 per sampling scoped e fault controllati.

La Fase 4 — Developer Benchmark consegna:

* dataset embedded `maestro-laravel-mini@1.0.0` con fixture PHP/Laravel;
* cinque task generativi e un retrieval tramite embedding;
* lifecycle reale del plugin Laravel su workspace temporaneo privato;
* rubriche deterministiche e trasparenti 0–3 senza evaluator LLM;
* separazione tra stato tecnico e `QualityEvaluation` nel report `1.2.0`;
* `rationale_code` validato senza testo libero sensibile;
* comando `maestro bench laravel` con gate tecnici e qualitativi opt-in;
* documentazione in `docs/developer-benchmark.md`;
* report finale in `docs/reports/milestone-3-phase-4.md`;
* ADR-0020 per dataset embedded e rubriche non-LLM.

La Fase 5 — Reporting & Hardware Profiles consegna:

* JSON `1.2.0` confermato come fonte canonica;
* renderer Markdown deterministico e redatto;
* comando `maestro bench render` con decoder strict e protezione input/output;
* flag `--markdown` per smoke, provider, model e laravel;
* scrittura atomica `0600` per entrambi i formati;
* profilo comune con runtime, procfs Linux e build metadata;
* GPU, backend e VRAM come metadata opt-in senza probe esterni;
* documentazione e gate in `docs/benchmark-reporting.md`;
* report finale in `docs/reports/milestone-3-phase-5.md`;
* ADR-0021 per JSON canonico, Markdown derivato e hardware dichiarativo.

Il completamento della Fase 5 richiedeva una decisione live separata. ADR-0030
registra quella decisione e chiude formalmente la Milestone 3.

Validazione live Ollama del 2026-08-09:

* integration test provider superato per listing, discovery, completion,
  streaming, cancellazione ed embedding;
* Smoke Benchmark: 9 passed, 3 skipped, 2 failed;
* failure: `tool_call_missing` e `tool_stream_terminal_missing` con
  `qwen2.5-coder:7b`;
* la ripetizione diretta su `POST /api/chat`, con temperatura 0, non produce
  `message.tool_calls` né in modalità non-stream né in alcuno dei 27 chunk
  stream; la chiamata viene resa come JSON testuale e il terminale usa
  `done_reason: stop`, escludendo una perdita nella traduzione o aggregazione
  dell'adapter Maestro per questa fixture;
* embedding Smoke saltato perché l'introspection richiede l'ID catalogo esatto
  `embeddinggemma:latest`, mentre era configurato l'alias `embeddinggemma`;
* gate deterministico, race detector, vet e tre manifest superati;
* report in `docs/reports/milestone-3-live-ollama-validation.md`.

La run Smoke live senza scenari failed è stata ottenuta con la fixture positiva
`llama3.1:8b`; ADR-0030 conserva questo risultato come baseline positiva della
Milestone 3 completata.

La seconda fixture `llama3.1:8b` supera la prova diretta: non-stream restituisce
una tool call nativa e lo stream la emette nel primo chunk, seguito da un chunk
terminale con `done_reason: stop`. L'adapter Ollama normalizza ora quel terminale
in `tool_calls` soltanto dopo aver tradotto una tool call nello stesso stream;
le altre cause e gli stream senza tool call restano invariati. La stessa regola
allinea le completion non-stream. Test mirati, gate Go, integration suite,
embedding con ID catalogo esatto e lifecycle passano. Lo Smoke completo
post-correzione produce 13 passed, 1 skipped e 0 failed: il gate live Ollama è
verde. `qwen2.5-coder:7b` resta il caso negativo documentato e
`llama3.1:8b` la fixture positiva. La chiusura formale è registrata separatamente
da ADR-0030.

Decisione fixture del gate Ollama:

* `llama3.1:8b` è la fixture positiva validata per chat, streaming, structured
  output JSON e JSON Schema, lifecycle, tool calling non-stream e tool calling
  stream dopo la normalizzazione Maestro;
* `embeddinggemma:latest` è la fixture embedding validata nella stessa
  configurazione live;
* `qwen2.5-coder:7b` è il caso negativo canonico: Ollama dichiara `tools`, il
  modello comprende semanticamente la richiesta, ma serializza la chiamata nel
  contenuto o non la espone come `message.tool_calls`; la capability non è
  validata operativamente con runtime e template Ollama correnti;
* `tool_stream_terminal_missing` con `llama3.1:8b` era un difetto di
  normalizzazione dell'adapter Maestro, ora corretto e coperto da regressioni;
* il gate live Ollama è superato e la relativa documentazione è conclusa;
* non servono altre fixture Ollama per questo gate.

Il preflight conclusivo llama.cpp non trova binario, server, endpoint o profilo
single-model configurato. Dopo due OOM storici in router mode, nessuna nuova
matrice viene avviata sullo stesso host. llama.cpp resta sperimentale/non
supportato; l'assenza della prova non è un PASS.

Checkpoint di chiusura:

| Punto | Decisione |
|---|---|
| Milestone 3 | Completata con decisione ADR-0030 |
| Ollama | Gate live superato con `llama3.1:8b` |
| Qwen | `qwen2.5-coder:7b` è il caso negativo canonico |
| llama.cpp | Sperimentale/non supportato; preflight incompatibile |
| Motivo | Nessun profilo live valido; router mode ha causato OOM sul target |

Regola futura: un support claim llama.cpp richiede una nuova matrice su un
profilo hardware–server–modello dichiarato; non riapre retroattivamente la
Milestone 3.

La milestone misura configurazioni complete hardware–provider–modello–plugin,
non costruisce classifiche assolute tra modelli.

Livelli:

1. Smoke Benchmark — matrice live e assorbimento degli smoke test provider;
2. Runtime Benchmark — latenza, throughput, risorse e cancellazione;
3. Developer Benchmark — task reali PHP/Laravel, refactor, test ed embedding.

Output previsti:

* `maestro bench smoke`;
* `maestro bench provider`;
* `maestro bench model`;
* `maestro bench laravel`;
* report JSON e Markdown;
* profili hardware documentati;
* dataset minimale di task reali.

Il piano dettagliato è in `docs/benchmark-evaluation-plan.md`. Le milestone
successive sono rinumerate: Gestor diventa Milestone 4, Plugin System 5, Context
Engine 6, Agent System 7 ed Ecosistema 8.

## Milestone 4 — Gestor

Stato: completata — Fasi 1–5 e gate finale completati.

Il documento `docs/gestor-design.md` apre formalmente la milestone.
Il piano operativo è in `docs/gestor-development-plan.md`.

Decisioni iniziali:

* Gestor è l'indice e il risolutore centrale delle capability;
* non esegue codice e non possiede lifecycle o stato;
* riusa il Registry e il dependency graph autorevoli del Runtime Core senza
  duplicarli;
* integra component metadata e provider capability introspection attraverso
  sorgenti esplicite;
* distingue capability dichiarata da disponibilità operativa;
* usa snapshot atomici, listing deterministico e preferenze esplicite;
* non applica ranking nascosti e segnala le risoluzioni ambigue;
* non interroga sorgenti esterne né pubblica eventi mantenendo lock interni.

Le cinque fasi previste sono: contratti e ADR, Snapshot Registry, discovery
sources, resolver con dependency graph, composition root e osservabilità.

Stato delle fasi:

1. Contratti, modello di dominio e ADR-0022 — completata;
2. Snapshot Registry — completata;
3. Discovery sources Runtime e Provider — completata;
4. Resolver e dependency graph — completata;
5. Composition root, osservabilità e gate finale — completata.

Ogni fase deve produrre un report in `docs/reports/`; la Fase 5 produce anche
`docs/reports/milestone-4-final.md`. Nessuna fase viene dichiarata completata
senza i test e i deliverable obbligatori descritti nel piano.

### Fase 1 — Contratti, modello di dominio e ADR

Completata nei package:

```text
pkg/gestor
internal/gestor
```

Sono disponibili:

* ID namespaced validati e ordinabili per capability e source;
* target `component` e `provider` con scope esplicito;
* availability `unknown`, `available` e `unavailable` separata dalla
  dichiarazione;
* descriptor, query, snapshot metadata e resolution immutabili tramite copie
  difensive;
* preferenze esatte e ordinate senza ranking implicito;
* sentinel compatibili con `errors.Is`;
* contratti pubblici minimi `Source`, `Registry` e `Resolver`;
* mapping interno esaustivo delle capability Runtime e Provider note;
* ADR-0022 Accepted.

La generazione appartiene allo snapshot e non viene duplicata nei descriptor.

### Fase 2 — Snapshot Registry

Completata in `internal/gestor` con estensione additiva di `pkg/gestor` per le
sorgenti consultate senza descriptor.

Sono disponibili:

* catalogo sorgenti thread-safe con source ID univoci;
* esecuzione sequenziale delle sorgenti in ordine lessicografico e senza lock;
* refresh all-or-nothing con candidato locale;
* snapshot immutabili current/stale e generazioni monotone;
* epoch che impedisce a refresh superati da registrazione o invalidazione di
  pubblicare risultati obsoleti;
* conservazione dell'ultimo snapshot valido su errori e cancellazione;
* collision detection capability–target tra e dentro le sorgenti;
* indici interni capability → descriptor e target → descriptor;
* copie difensive per snapshot, metadata, listing e indici;
* fixture in-memory e test concorrenti con race detector.

Lo snapshot iniziale è stale a generazione zero. Ogni refresh riuscito, anche
con zero sorgenti, incrementa la generazione. Registrazione e `Invalidate`
marcano stale senza cambiare generazione. Le sorgenti Runtime e Provider reali
sono implementate nella Fase 3.

### Fase 3 — Discovery sources

Completata in:

```text
internal/gestor
internal/provider
internal/runtime
```

Sono disponibili:

* `RuntimeComponentSource` sulla vista autorevole `Components()`;
* mapping delle sei capability lifecycle in ID `runtime.*`;
* conservazione delle capability custom già namespaced;
* plugin scoperti una sola volta come componenti del Registry globale;
* listing interno ordinato `Registered()` nel concrete Provider Runtime;
* `ProviderCapabilitySource` per target adapter, instance e model esatti;
* omissione delle capability provider unsupported;
* traduzione conservativa di availability unknown/available/unavailable;
* model target espliciti e copiati, senza uso implicito del default;
* invalidazione additiva dopo registrazioni riuscite di componenti, plugin e
  provider, eseguita fuori lock;
* propagazione di errori e cancellazione senza snapshot parziali;
* fixture Qwen dichiarata ma non operativa e Llama 3.1 operativa;
* test concorrenti e race detector.

Le interfacce pubbliche `runtime.Runtime`, `provider.Runtime` e `plugin.Runtime`
non sono cambiate. Il wiring effettivo delle sorgenti e dell'invalidatore nel
composition root appartiene alla Fase 5.

### Fase 4 — Resolver e dependency graph

Completata in:

```text
internal/gestor
internal/runtime
```

Sono disponibili:

* `Resolver` concreto conforme al contratto pubblico `pkg/gestor.Resolver`;
* filtri esatti per capability, target kind, scope e model ID;
* distinzione tra `ErrNotFound`, `ErrUnavailable`, `ErrAmbiguous` ed
  `ErrStaleSnapshot`;
* esclusione dei target unavailable e requisito opzionale di evidenza
  available;
* preferenze target esatte e ordinate, senza vincitore lessicografico
  implicito;
* vista read-only del dependency graph autorevole del Runtime Core;
* verifica dell'eleggibilità dei componenti e piano transitivo dependency-first
  in ordine topologico;
* generazioni indipendenti per snapshot Gestor, catalogo componenti e grafo,
  ricontrollate prima di produrre un risultato;
* identità dei nodi acquisita alla costruzione del grafo, così `Resolve` non
  invoca `Metadata` o codice del candidato;
* test con dipendenze richieste e opzionali, variazioni concorrenti e race
  detector.

Il Resolver non esegue capability, probe o introspection. Composition root,
refresh iniziale ed eventi redatti sono consegnati dalla Fase 5.

### Fase 5 — Composition root, osservabilità e gate finale

Completata in:

```text
pkg/gestor
internal/gestor
internal/runtime
maestro.go
```

Sono disponibili:

* contratto additivo `gestor.Service`, composto da Registry e Resolver;
* accesso pubblico `Runtime.Gestor()` nel composition root Maestro senza
  modificare `pkg/runtime.Runtime`;
* registrazione automatica delle sorgenti `runtime.components` e
  `provider.capabilities`;
* snapshot iniziale current a generazione 1, anche con cataloghi vuoti;
* refresh successivi espliciti e all-or-nothing;
* invalidazione coordinata dopo registrazioni riuscite di componenti, plugin,
  provider o nuove sorgenti;
* topic pubblici stabili per refresh e resolution;
* payload redatti senza error string, source detail, target/model ID o dati
  operativi;
* observer invocati senza lock, con errori e panic isolati dal risultato;
* test end-to-end dal Runtime pubblico con componenti, provider, plugin e grafo
  reale;
* cinque report di fase e report finale della Milestone 4.

La sorgente provider built-in interroga adapter e instance. I target model non
sono inferiti: dichiarazioni model-specific ulteriori devono arrivare da una
`Source` esplicita. Gestor restituisce descriptor e dependency plan e non
esegue la capability risolta.

La Milestone 4 è chiusa. La Milestone 3 resta sospesa con la matrice live
llama.cpp ancora pendente; la chiusura di Gestor non ne modifica lo stato.

## Milestone 5 — Plugin System

Stato: completata — Fasi 1–5 e gate finale superati.

Il design iniziale è definito in `docs/plugin-system-design.md` e il piano
operativo in `docs/plugin-system-development-plan.md`.

La milestone consolida il modello trusted in-process già introdotto dalla Fase
6 del Runtime Core. Non crea un secondo lifecycle e non anticipa packaging,
sandbox, permission model o plugin di terze parti.

Le cinque fasi previste sono:

1. contratti, audit della baseline e ADR-0023 — completata;
2. catalogo, registry e caricamento — completata;
3. lifecycle, dependency graph e Gestor — completata;
4. Laravel reference plugin — completata;
5. osservabilità, hardening e gate finale — completata.

Ogni fase produce un report in `docs/reports/`; la Fase 5 produce anche
`docs/reports/milestone-5-final.md`. Il codice plugin esistente è considerato
una baseline da verificare e non rende automaticamente completate le fasi.

### Fase 1 — Contratti, audit della baseline e ADR-0023

Completata in:

```text
pkg/plugin
internal/plugin
docs/adr/ADR-0023.md
docs/plugin-api-compatibility-audit.md
```

L'audit conferma senza modifiche breaking `Plugin`, `Manifest`, `Loader`,
`LoaderFunc`, `Runtime`, eventi, sentinel e facade Laravel. Available indica il
catalogo, registered il registry, loaded un'operazione riuscita e running lo
stato posseduto dal Runtime Core. Non viene introdotto uno stato plugin
parallelo.

ADR-0023 stabilisce modello trusted in-process, registrazione pre-start,
ownership globale di graph/stato/lifecycle e discovery Gestor attraverso il
Registry componenti. Descriptor di catalogo e contratto workspace generico non
vengono introdotti senza un consumer concreto.

Sono stati aggiunti commenti pubblici ai sentinel, assertion di compilazione e
regressioni sulla compatibilità esatta del manifest. Suite completa, race
detector e vet sono verdi. Il report è disponibile in
`docs/reports/milestone-5-phase-1.md`.

### Fase 2 — Catalogo, registry e caricamento

Completata in:

```text
pkg/plugin
internal/plugin
```

Le firme pubbliche restano invariate. Il contratto ora esplicita che ogni
`Load` è un tentativo indipendente: factory concorrenti sullo stesso ID possono
essere invocate più volte, mentre il Registry globale consente una sola
registrazione riuscita. Non viene introdotto singleflight implicito.

Fixture bloccanti dimostrano che loader, registrar ed eventi vengono invocati
senza lock interni. Sono coperti snapshot concorrenti, load sullo stesso ID e
su ID differenti, failure atomiche, cancellazione e composizione dei sentinel.
Suite completa, race detector e vet sono verdi. Il report è disponibile in
`docs/reports/milestone-5-phase-2.md`.

### Fase 3 — Lifecycle, dependency graph e Gestor

Completata in:

```text
maestro_test.go
internal/runtime/runtime.go
```

La matrice end-to-end copre plugin passivi e lifecycle completo, failure di
configure/initialize/start/stop, dipendenze plugin-componente in entrambe le
direzioni, plugin-plugin, required/optional e cicli. Startup e shutdown seguono
lo stesso graph globale e gli stati restano nello `StateManager` unico.

Fixture bloccanti verificano il rifiuto di Register e Load durante startup e
shutdown. È stato corretto l'ordine dei controlli del Runtime Core affinché lo
shutdown restituisca `ErrInvalidState` senza essere mascherato dal flag started.

Una capability custom caricata dal catalogo invalida Gestor, viene scoperta una
sola volta dopo refresh e produce un dependency plan senza eseguire codice del
plugin. Suite completa, race detector e vet sono verdi. Il report è disponibile
in `docs/reports/milestone-5-phase-3.md`.

### Fase 4 — Laravel reference plugin

Completata in:

```text
pkg/plugin
pkg/plugin/laravel
internal/plugin/laravel
```

Il contratto additivo `plugin.CapabilityWorkspaceDetection` usa l'ID
`plugin.workspace-detection`. Laravel `0.2.0` lo dichiara nei metadata e Gestor
lo indicizza una sola volta prima del lifecycle.

La facade concreta resta invariata: Root è assoluta e immutabile;
FrameworkVersion viene pubblicata atomicamente dopo Initialize. Health e
inizializzazioni fallite conservano l'ultimo snapshot valido. Root relative o
inesistenti, manifest mancanti/malformati/oltre limite, constraint vuoti,
mutation e concorrenza sono coperti deterministicamente.

Non viene introdotto un contratto workspace framework-neutral senza il consumer
del Context Engine. Suite completa, race detector e vet sono verdi. Il report è
disponibile in `docs/reports/milestone-5-phase-4.md`.

### Fase 5 — Osservabilità, hardening e gate finale

Completata in:

```text
pkg/plugin
internal/plugin
maestro_test.go
```

Ordine e cardinalità degli eventi sono coperti per successi, failure e
cancellazione. La pubblicazione resta sincrona e fuori lock, ma errori e panic
dell'Event Bus sono isolati affinché non alterino operazioni già committate. Un
subscriber lento mantiene la backpressure definita da ADR-0005.

Il payload conserva ID e riferimento plugin trusted in-process senza copiare
configurazione, error string o contenuti del workspace; non è un envelope
telemetrico serializzabile. L'audit API finale non rileva modifiche breaking o
nuove dipendenze.

Suite ripetuta, suite completa, race detector, vet e diff check sono verdi. Il
report di fase è in `docs/reports/milestone-5-phase-5.md`; il report conclusivo
è in `docs/reports/milestone-5-final.md`.

La Milestone 5 è chiusa. La Milestone 3 resta sospesa per la matrice live
llama.cpp già documentata; non è stata modificata da questo gate.

## Milestone 6 — Context Engine

Stato: completata — Fasi 1–6 completate.

Il design iniziale è definito in `docs/context-engine-design.md` e il piano
operativo in `docs/context-engine-development-plan.md`.

La milestone introduce un servizio provider-agnostic separato da
`context.Context` e da `runtime.Context`. Il package pubblico è
`pkg/contextengine`; l'implementazione concreta resta in
`internal/contextengine`. Il composition root espone `Runtime.ContextEngine`
senza modificare il contratto di basso livello `pkg/runtime.Runtime`.

La pipeline prevista è:

```text
workspace -> snapshot -> analisi -> retrieval -> selezione -> context bundle
```

Le sei fasi sono:

1. contratti, ownership e ADR-0024 — completata;
2. workspace indexing e snapshot — completata;
3. analisi strutturata e AST — completata;
4. retrieval, Context Builder e budget — completata;
5. cache e aggiornamento incrementale — completata;
6. integrazione, osservabilità e gate finale — completata.

Decisioni iniziali del design:

* workspace e path pubblici sono framework-neutral; i path sono logici e non
  escono dalla root;
* gli snapshot sono immutabili, generazionali e pubblicati atomicamente;
* refresh falliti o cancellati conservano l'ultimo snapshot valido;
* analyzer language-specific sono registrabili e restano fuori dal Runtime
  Core;
* retrieval lessicale e strutturale funzionano offline;
* embedding e semantic retrieval sono opt-in e riusano il Provider Runtime con
  provider e modello espliciti;
* il Context Builder conserva provenance, metodo, costo stimato e budget;
* una stima token non viene presentata come conteggio esatto;
* la cache iniziale è in-memory, bounded, content-addressed e non autorevole;
* Gestor descrive capability ma non esegue indexing o analyzer;
* Laravel fornisce il primo workspace attraverso un contratto generico senza
  introdurre conoscenza framework-specific nel Context Engine;
* eventi e log non contengono query, testo, embedding o path assoluti.

Ogni fase ha un report in `docs/reports/`; il report conclusivo è
`docs/reports/milestone-6-final.md`.

Restano fuori scope memoria conversazionale, tool execution, permission model,
watcher filesystem, persistenza distribuita, vector database, ranking LLM e
selezione implicita di provider o modello. Questi confini impediscono alla
Milestone 6 di anticipare Agent System ed Ecosistema.

### Fase 1 — Contratti, ownership e ADR-0024

Completata in:

```text
pkg/contextengine
docs/adr/ADR-0024.md
docs/context-engine-api-compatibility-audit.md
```

La baseline pubblica introduce workspace, policy, documenti content-addressed,
analisi strutturate, snapshot generazionali, query di retrieval, budget,
estimator e context bundle immutabili. Source, analyzer ed engine sono contratti
provider- e framework-neutral.

Path documento, digest, intervalli, riferimenti, query e budget vengono validati
prima della costruzione. Slice e mappe sono difensive; gli errori sentinel
restano ispezionabili con `errors.Is`. Nessuna API Runtime, Gestor, Provider o
Plugin è stata modificata.

ADR-0024 formalizza ownership, refresh atomico, analyzer sostituibili,
retrieval semantico opt-in, budget dichiarati e confine di riservatezza. La
suite `pkg/contextengine` è verde. Il report è disponibile in
`docs/reports/milestone-6-phase-1.md`.

### Fase 2 — Workspace indexing e snapshot

Completata in:

```text
internal/contextengine
pkg/contextengine
docs/context-engine-indexing.md
```

L'engine registra la source filesystem built-in e pubblica snapshot in-memory
ordinati con generazioni monotone. La scansione applica path containment,
include/exclude, limiti per file e workspace, esclusione di hidden e dipendenze,
normalizzazione UTF-8 e classificazione media/language deterministica.

Symlink non vengono seguiti; identità, dimensione e modification time vengono
verificati attorno alla lettura. Binari sono esclusi per default o conservati
come `application/octet-stream` su richiesta. Failure, cancellazione e output
source invalidi non sostituiscono l'ultimo snapshot.

Source esterne vengono invocate senza lock globali. Test bloccanti e concorrenti
verificano registrazione indipendente, venti refresh simultanei e snapshot
completi. Il report è disponibile in
`docs/reports/milestone-6-phase-2.md`.

### Fase 3 — Analisi strutturata e AST

Completata in:

```text
internal/contextengine
pkg/contextengine
docs/context-engine-analysis.md
```

Il registry analyzer valida ID e versioni, rifiuta nil/typed nil e duplicati e
invoca `Supports` e `Analyze` fuori lock. Senza configurazione, più analyzer
applicabili producono `ErrAmbiguous`; `WorkspaceOptions.Analyzers` dichiara una
composizione esplicita e difensiva.

`context.go-ast@1` usa `go/parser` e produce package, import, type, field,
constant, variable, function, method, relazioni e chunk. Parse incompleti
conservano l'AST parziale con diagnostica `go_parse_error`. Errori operativi,
panic, cancellazione o output incoerenti non pubblicano snapshot.

Test bloccanti dimostrano callback fuori lock e cancellazione atomica. Il report
è disponibile in `docs/reports/milestone-6-phase-3.md`.

### Fase 4 — Retrieval, Context Builder e budget

Completata in:

```text
internal/contextengine
pkg/contextengine
internal/benchmark/developer
docs/context-engine-retrieval.md
```

Retrieval lessicale usa term coverage deterministica; retrieval strutturale usa
simboli e intervalli; semantic retrieval usa embedding provider con target
esplicito e validazione di cardinalità, dimensione, finitezza e norma. Query
multi-metodo richiedono Reciprocal Rank Fusion esplicita.

`context.utf8-estimator@1` fornisce una stima conservativa offline. Il builder
deduplica intervalli, tronca su confini UTF-8 e produce sezioni con provenance e
costo senza superare evidence budget, riserva e safety margin.

Lo scenario retrieval del Developer Benchmark passa ora dal Context Engine
senza cambiare dataset o rubrica. Il report è disponibile in
`docs/reports/milestone-6-phase-4.md`.

### Fase 5 — Cache e aggiornamento incrementale

Completata in:

```text
internal/contextengine
pkg/contextengine
docs/context-engine-cache.md
```

La cache LRU in-memory applica limiti per entry e byte e pubblica statistiche
aggregate. Analysis, embedding e stime usano chiavi che includono digest e
versioni semantiche; rename, analyzer, provider, modello, dimensione ed
estimator invalidano soltanto gli artefatti coinvolti.

Vettori sono clonati e validati prima del commit. Un cambio dimensione elimina
le entry del target e fallisce la richiesta corrente; il retry riparte pulito.
Failure, panic e cancellazioni non diventano hit.

Cold e warm producono ranking e bundle equivalenti. Due richieste cold
concorrenti restano indipendenti e non introducono singleflight implicito. Il
report è disponibile in `docs/reports/milestone-6-phase-5.md`.

### Fase 6 — Integrazione, osservabilità e gate finale

Completata in:

```text
internal/runtime
internal/plugin/laravel
pkg/contextengine
docs/context-engine-runtime.md
```

Il composition root crea una sola istanza del Context Engine e condivide il
Provider Runtime per gli embedding e l'Event Bus per osservabilità. L'accessor
pubblico è `Runtime.ContextEngine`; il Context Engine non entra nel lifecycle
dei componenti.

Laravel `0.3.0` implementa `WorkspaceProvider` e dichiara
`context.workspace-provider`. Gestor indicizza e risolve la capability senza
invocare workspace, indexing o analyzer. Il percorso end-to-end usa il
workspace generico Laravel per produrre uno snapshot.

Gli eventi `context.index.*`, `context.build.*` e `context.cache.observed`
espongono soltanto ID, generazione, conteggi, statistiche aggregate e codici di
failure. Query, testo, embedding, root, path, provider, modello ed error string
non entrano nei payload. Errori e panic degli observer sono best-effort e non
alterano operazioni già committate.

Suite completa, race detector, vet, test ripetuti, diff check e audit API sono
verdi. I report sono `docs/reports/milestone-6-phase-6.md` e
`docs/reports/milestone-6-final.md`. Memoria, tool permissions, watcher,
persistenza e ranking LLM restano fuori scope.

## Milestone 7 — Agent System

Stato: completata — Fasi 1–7 e gate finale superati.

Il design iniziale è definito in `docs/agent-system-design.md` e il piano
operativo in `docs/agent-system-development-plan.md`.

La milestone introduce due servizi con ownership separate:

```text
pkg/tool     + internal/tool   -> catalogo, prepare, autorizzazione, execute
pkg/agent    + internal/agent  -> sessione, piano, budget, loop modello-tool
```

Il composition root li espone in modo additivo tramite `Runtime.Tools()` e
`Runtime.Agents()` senza modificare `pkg/runtime.Runtime`. L'Agent Runtime
coordina Provider Runtime, Context Engine, Tool Runtime e Gestor; non duplica
registry, snapshot o lifecycle autorevoli.

Le sette fasi pianificate sono:

1. contratti, ownership e ADR-0025 — completata;
2. Tool catalog ed execution boundary — completata;
3. permission model e approval flow — completata;
4. sessioni, piani e budget — completata;
5. loop agentico e tool calling — completata;
6. workspace awareness e reference tool — completata;
7. integrazione, osservabilità e gate finale — completata.

Decisioni iniziali del design:

* gli arguments prodotti dal modello sono input non fidati;
* ogni effetto attraversa `Prepare`, autorizzazione ed `Execute`;
* l'autorizzazione valuta action concrete e normalizzate, non il solo nome del
  tool;
* una policy senza regola applicabile nega l'azione;
* `prompt` senza un Approver configurato non equivale ad allow;
* agenti, modelli e tool non possono concedersi grant;
* invocazione del modello e disclosure del workspace sono effetti
  autorizzabili separatamente;
* provider, modello, workspace, agente, policy e limiti sono input espliciti;
* sessioni e memoria iniziali sono in-memory, bounded e con terminale unico;
* piani prodotti dal modello vengono validati prima di diventare stato;
* ogni loop applica hard ceiling su durata, turni, tool, token e byte;
* tool mutanti non vengono ritentati implicitamente;
* il Context Engine fornisce evidenza e provenance, non memoria agente;
* una mutazione marca il contesto stale e richiede refresh a un checkpoint
  esplicito;
* Gestor descrive agenti e tool senza eseguirli;
* eventi e log non contengono prompt, contenuti workspace, arguments o output;
* il permission model trusted in-process non viene presentato come sandbox.

### Fase 1 — Contratti, ownership e ADR-0025

Completata in:

```text
pkg/tool
pkg/agent
docs/adr/ADR-0025.md
docs/agent-system-api-compatibility-audit.md
docs/reports/milestone-7-phase-1.md
```

`pkg/tool` introduce descriptor versionati, invocation e prepared invocation
immutabili, action tipizzate, permission request atomiche, richiesta modello
distinta, disclosure manifest redatto, decision/approval, limiti, risultati,
sentinel, envelope ed event allowlist. Arguments e schema JSON sono
canonicalizzati e difensivi; il fingerprint lega tool, versione, call, run,
arguments e action.

`pkg/agent` introduce descriptor e capability, run request con target e limiti
espliciti, piani aciclici, step transition, planning request, session snapshot,
contatori, stale bit, terminal precedence, run result, Runtime, sentinel,
envelope ed event allowlist.

ADR-0025 stabilisce che `Decision` e `Approval` non sono permit. Il percorso
pubblico `tool.Runtime.Invoke` incorpora autorizzazione ed esecuzione e non
accetta un valore allow dal chiamante. Il permit operativo futuro è interno,
issuer-bound e vincolato a run ID e permission fingerprint. La Fase 2 ne
implementa controllo e issuer deterministico di test; la Fase 3 implementa
policy, Approver e grant reali.

Le permission modello usano lo stesso modello di action ma subject distinto:
il bundle viene costruito localmente, ne viene derivato un manifest redatto,
quindi `model.invoke` e `model.disclose` vengono autorizzati prima della chiamata
provider. Tool con più action sono autorizzati atomicamente.

La precedenza pre-commit dei terminali è `deadline > canceled > limit >
permission_denied > blocked > provider_failure > tool_failure >
planning_failure > internal_failure > completed`; il primo terminale committato
resta definitivo. L'avvio di `workspace.mutate` marca il contesto stale anche
su errore o esito ambiguo.

Nessuna API Runtime, Provider, Context Engine, Gestor, Plugin o composition root
è stata modificata. I contratti nuovi dipendono soltanto da package pubblici e
non importano implementazioni interne.

### Fase 2 — Tool catalog ed execution boundary

Completata in:

```text
internal/tool
docs/tool-runtime.md
docs/reports/milestone-7-phase-2.md
```

Il catalogo registra tool trusted in-process con ID e nome provider univoci e
listing deterministico. `Prepare`, authorizer ed `Execute` vengono invocati
fuori lock. Il Runtime valida nuovamente identità, versione, fingerprint e
effect dichiarati prima di costruire la permission request atomica.

Il permit operativo è privato, issuer-bound, legato a run, permission e
prepared fingerprint e consumato con compare-and-swap. Replay, issuer diverso
o mismatch vengono rifiutati. Il runtime di produzione resta default-deny; la
suite usa un authorizer deterministico interno per testare l'executor senza
anticipare le policy della Fase 3.

Deadline, item e byte limit sono applicati al boundary. Output eccedente viene
troncato su confini UTF-8; panic di tool e cancellazione sono classificati senza
retry implicito. Test bloccanti dimostrano assenza di callback sotto lock e il
race detector copre registrazioni concorrenti e consumo dei permit.

### Fase 3 — Permission model e approval flow

Completata in:

```text
pkg/tool/rule.go
internal/tool/policy.go
internal/tool/authorization.go
docs/agent-permissions.md
docs/reports/milestone-7-phase-3.md
```

`StaticPolicy` applica esclusivamente matcher exact effect/resource/workspace;
non esistono wildcard, prefix grant o ordine di preferenza. Regole duplicate
sono invalide e una action senza regola produce deny terminale. Su richieste
multi-action deny prevale, prompt richiede approval atomica e allow richiede
tutte le action consentite.

Il Runtime risolve il `PolicyID` esatto senza fallback. Prompt senza Approver
diventa deny terminale; approval allow/deny conserva scope o disposition. Grant
run-scoped sono indicizzati da policy, run e permission fingerprint; allow
one-shot non viene memorizzato e usa soltanto il permit consumabile della Fase
2.

Policy e Approver vengono invocati fuori lock e i loro output vengono validati.
Errori, cancellazioni e panic non producono grant. Lo stesso authorizer governa
le permission modello con subject distinto dalle tool invocation.

### Fase 4 — Sessioni, piani e budget

Completata in:

```text
internal/agent
docs/agent-sessions.md
docs/reports/milestone-7-phase-4.md
```

Il registry bounded assegna un solo coordinatore a ogni run e non riusa ID
terminali. Snapshot, generazioni e contatori sono monotoni; il terminale viene
committato una sola volta. I piani immutabili validano dipendenze e transizioni,
mentre le revisioni sequenziali conservano una storia bounded.

`ProviderPlanner` richiede structured output e il runtime autorizza
`model.invoke` e `model.disclose` prima di inviare istruzione o contesto al
provider. Durata, turni, tool, token, step, revisioni e byte restano hard
ceiling locali. Il confine della fase termina con `blocked`; il loop
modello-tool appartiene alla Fase 5.

### Fase 5 — Loop agentico e tool calling

Completata in:

```text
internal/agent/loop.go
internal/agent/stream.go
docs/agent-runtime.md
docs/reports/milestone-7-phase-5.md
```

Il loop esegue step ready in sequenza usando il Provider Runtime e soltanto i
tool registrati e inclusi nella request. Call multiple mantengono ordine e
correlazione; i result vengono restituiti come JSON tipizzato e attraversano
sempre `Tool Runtime.Invoke`. Completion e streaming condividono validation e
terminali; l'assembler limita testo, cardinalità e delta arguments. Budget,
cancellazione e deadline vengono applicati prima di ogni nuova iterazione.

### Fase 6 — Workspace awareness e reference tool

Completata in:

```text
internal/tool/workspace_registry.go
internal/tool/workspace_tools.go
docs/agent-workspace.md
docs/reports/milestone-7-phase-6.md
```

La request può associare un Workspace immutabile al run senza mostrare la root
al modello. I reference tool list/read/search/write/patch usano path logici,
`os.Root`, controllo Lstat di ogni componente, rifiuto dei symlink e
precondizioni SHA-256. Una mutazione marca la sessione stale quando Execute
inizia; Index e Build devono pubblicare una generazione successiva prima di
renderla fresh. Il percorso usa il contratto WorkspaceProvider ed è verificato
con Laravel senza dipendenze framework nel runtime.

### Fase 7 — Integrazione, osservabilità e gate finale

Completata in:

```text
maestro.Runtime.Tools / maestro.Runtime.Agents
internal/gestor/agent_source.go
internal/gestor/tool_source.go
internal/agent/reference.go
docs/reports/milestone-7-phase-7.md
docs/reports/milestone-7-final.md
```

Il composition root condivide una singola istanza di Provider Runtime, Context
Engine, Tool Runtime, Agent Runtime ed Event Bus. Reference tool e agent sono
registrati senza policy permissive; il chiamante deve fornire policy e target
espliciti. Gestor descrive agenti e tool tramite sorgenti read-only e viene
invalidato dalle nuove registrazioni. Eventi session/plan/step/turn e
permission/invocation usano payload allowlist senza prompt, path, arguments o
output.

Uno scenario deterministico completa read, patch, reindex e risposta finale su
workspace temporaneo. Suite completa, race detector, test ripetuti, benchmark,
vet, diff check e audit API chiudono la Milestone 7.

La baseline workspace prevista comprende listing, read, search e write/patch
con path logici, containment, controllo symlink, output limitati e
precondizione content-addressed per le mutazioni. Shell completa, Git write,
Docker, Composer, Artisan e PHPUnit non sono requisiti del primo reference
agent.

Restano fuori scope memoria persistente, recovery dopo restart, multi-agent,
esecuzione distribuita, sandbox, plugin/tool di terze parti, secret manager,
CLI completa e selezione automatica di provider o modello. Questi confini sono
assegnati alla Milestone 8 o a evoluzioni successive.

## Gate post-Milestone 7 e percorso v0.1.0

Stato: Milestone 8 e Fasi 1–6 completate; artifact e tag `v0.1.0` validati.

Il gate in `docs/release-readiness-audit.md` accetta la Milestone 7 come
baseline ingegneristica e registra il divario di prodotto senza riaprirne
l'architettura. Suite e scenario deterministico sono verdi. Le Fasi 2–4 hanno
chiuso configurazione, CLI minima, controllo operativo, packaging e prova
d'installazione pulita; restano documentazione pubblica di sicurezza e
scenario agentico live su Laravel.

La Milestone 8 è ridefinita come productization per v0.1.0. Il release
contract, il design e il piano sono approvati in:

```text
docs/adr/ADR-0026.md
docs/milestone-8-design.md
docs/milestone-8-development-plan.md
```

La superficie minima prevista è:

```text
maestro doctor
maestro models
maestro agents
maestro run
maestro version
```

Provider, modello, workspace, agente, policy, tool set e limiti restano
espliciti. Il formato di configurazione iniziale è YAML strict con
`version: 1`; i secret sono referenziati tramite ambiente e non inseriti nel
file. Il permission model rimane default-deny e l'approvazione CLI non viene
presentata come sandbox.

La Fase 2 è implementata in:

```text
internal/productconfig
internal/application
internal/buildinfo
cmd/maestro
configs/maestro.example.yaml
docs/configuration.md
docs/cli.md
docs/reports/milestone-8-phase-2.md
```

Il loader rifiuta unknown/duplicate fields, multi-document e alias. La
composition registra un solo provider, il plugin Laravel e una policy che
valida PermissionRequest concrete. Doctor usa probe read-only; run indicizza il
Workspace autorevole ed esegue `agent.reference`.

La Fase 3 aggiunge Approver terminale cancellabile, renderer degli eventi
redatti e output finale stabile. Una decisione `prompt` può essere negata,
approvata one-shot o concessa per la stessa action durante il run; EOF, input
invalido e no-TTY negano in sicurezza. I dettagli sono in
`docs/operational-experience.md` e nel report
`docs/reports/milestone-8-phase-3.md`.

La matrice live llama.cpp resta il debito formale della Milestone 3. Il report
indicato da decisioni successive non è presente nel repository né nella
cronologia Git disponibile: deve essere recuperato e verificato o la matrice
deve essere rieseguita prima della release candidate. In alternativa è
necessaria una decisione esplicita che classifichi llama.cpp sperimentale nella
v0.1.0; uno stato ambiguo non soddisfa il gate.

Le fasi approvate sono:

1. Release contract e audit — completata;
2. Configurazione e CLI minima — completata;
3. Esperienza operativa — completata;
4. Packaging e installazione — completata con packaging candidate, non release candidate;
5. Validazione live e release candidate;
6. Documentazione pubblica e v0.1.0.

SDK stabile, packaging di plugin/tool terzi, sandbox, recovery persistente,
multi-agent, shell, Git, Docker e selezione automatica di provider/modello sono
rinviati oltre v0.1.0.

La Fase 4 produce `maestro-v0.1.0-pc.1-linux-amd64.tar.gz` dal commit
`4578c132682e6b715317a6b4d1de958459cfc086`. Due build indipendenti sono
byte-identiche; archive, checksum, manifest, installazione pulita, config e
fixture Laravel superano il gate deterministico. Apache-2.0 è registrata in
ADR-0027. La promozione a release candidate e ogni validazione provider live
restano responsabilità della Fase 5.

Il gate intermedio della Fase 5 ha riconfermato il candidate e chiuso la Smoke
matrix Ollama provider-level con 13 passed, 1 skipped e 0 failed. Un run Laravel
read-only completa con profilo CPU ridotto; lo scenario
`read -> patch -> reindex -> final` resta invece aperto, con guardrail che hanno
impedito ogni patch non valida. Due tentativi llama.cpp in router mode hanno
esaurito i 15 GiB dell'host e causato la terminazione OOM di `llama-server`,
destabilizzando VS Code; le prove sono invalidate e nessun RC è stato prodotto.
Il dettaglio è in `docs/reports/milestone-8-phase-5-interim.md`.

Gli hardening agentici, l'envelope JSON strutturato e il profilo CPU ridotto
sono incorporati in `maestro-v0.1.0-pc.2-linux-amd64.tar.gz`, prodotto dal
commit `b9f571ac5914d2565e2a7bd28f4d5d6fc14a2710`. Il doppio build normalizzato
produce SHA-256
`91ef1bb196e9904ef3f3f0fefccf3a80acba22f14da43cdccbf9a83680fa41bc`;
versione, manifest, guida renderizzata, configurazione, fixture e installazione
fuori dal checkout sono verdi. Il quick start read-only esatto dall'archive è
positivo, ma il primo run mutativo ha confermato una dipendenza non governata
tra read e patch nello stesso turno. `pc.2` non è promuovibile. ADR-0028 rende
la coreografia deterministica. L'hardening è incorporato in
`maestro-v0.1.0-pc.3-linux-amd64.tar.gz`, commit
`d362b9910f68e5aecae3a489eb5852e339bc3939`, SHA-256
`8fbdfbf9b207c8c984f295240bcb6345d32fcbfa42f5869dd27a39acc158fe26`.
Il doppio build è byte-identico e il gate dell'archive è verde; `pc.3` è stato
usato nella ripresa live.

Dal candidate estratto, doctor, models, agents e quick start read-only sono
positivi. Il primo tentativo mutativo ha invece terminato con zero tool call:
`llama3.1:8b` ha descritto pseudo-call come testo. La fixture è rimasta
byte-identica. La serie registra 0 successi su 1 tentativo eseguito; il gate
richiedeva 3 successi consecutivi e i tentativi 2–3 non sono stati eseguiti.
Non è stato eseguito altro prompt tuning.
Il modello resta una fixture positiva provider-level e read-only, ma non è
supportato per il reference agent mutante. La v0.1.0 è bloccata finché non viene
validato un modello alternativo o definito un contratto operativo più stretto.
`pc.3` resta una baseline non promuovibile e nessun candidate è ammesso alla
prosecuzione mutativa prima di tale decisione.

La prima selezione economica ha valutato `rnj-1:8b-instruct-q4_K_M`. Il modello
supera il Gate A provider-level con 3 sequenze read-result-patch valide su 3,
senza eseguire effetti. Nel Gate B il primo run read-only invoca una read reale
ma termina `provider_failure` nel secondo turno dopo 535537 ms. Il gate registra
0 successi su 1 tentativo; il secondo tentativo e il Gate C non sono eseguiti.
Il modello è escluso, non esiste ancora un vincitore e `pc.4` non viene prodotto.
Il report è `docs/reports/milestone-8-model-selection.md`.

La selezione successiva usa `ibm/granite4.1:8b` senza cambiare il profilo o i
criteri. Gate A è verde 3/3 e Gate B è verde 2/2: entrambi i run read-only
terminano `completed` con una read reale e risposta corretta. Il primo run del
Gate C, da estrazione fresca, esegue la read e raggiunge 3 turni e 2 tool call,
ma termina `deadline_exceeded` dopo 600077 ms prima di approval o patch. Il
digest del controller resta
`4826abe9c6c5d701133817a9dcb565f0b84f760da57e1b518d430b601520b1bd`;
nessun grant o effetto è stato prodotto. I tentativi 2–3 non vengono eseguiti
in fail-fast. Granite è escluso, il modello viene scaricato dalla RAM e `pc.4`
non viene prodotto. Il candidato successivo previsto era `qwen3:8b`
non-thinking agli stessi Gate A, B e C; un suo fallimento avrebbe richiesto di
rivalutare il contratto di release prima di cambiare i limiti pubblici.

`qwen3:8b` è l'ultimo candidato della matrice corrente. Poiché `pc.3` non espone
il parametro Ollama `think`, il profilo non-thinking viene fissato prima dei
gate con una sola riga finale `/no_think` nell'istruzione utente iniziale di
ogni conversazione. `maestro doctor` supera 9 check. La prima sequenza del Gate
A termina però dopo 100977 ms senza tool call (`tool_call_count`), con 227 token
in ingresso e 256 in uscita. Le sequenze 2–3 e i Gate B/C non vengono eseguiti
in fail-fast. Il controller conserva SHA-256
`4826abe9c6c5d701133817a9dcb565f0b84f760da57e1b518d430b601520b1bd`,
Qwen3 viene scaricato dalla RAM e nessun artifact viene prodotto.

La matrice 8B termina senza vincitori. `pc.4` resta vietato e Fase 5 è in NO-GO
RC finché una decisione formale non sceglie tra v0.1.0 read-only con mutazioni
rinviate, mutazioni con requisito hardware/computazionale superiore, oppure
rinvio della release in attesa di una fixture adeguata. I limiti pubblici non
vengono aumentati retroattivamente.

ADR-0029 accetta il confine read-only della v0.1.0. La promessa supportata è
analizzare, interrogare e comprendere un progetto Laravel con il reference
agent locale controllato. Il percorso ufficiale usa Linux `amd64`, Ollama,
`llama3.1:8b` ed `embeddinggemma:latest`; la configurazione inclusa registra
solo `workspace.list`, `workspace.read` e `workspace.search` e imposta
`workspace_mutate: deny`. Tool mutanti, approval mutativa e reference agent
mutante restano implementati e testati deterministicamente ma sono
sperimentali/non supportati almeno fino alla v0.2.0.

llama.cpp è classificato sperimentale e non supportato nella v0.1.0. Il report
live mancante resta debito della Milestone 3 ma non blocca la release read-only.
Il nuovo contratto autorizza `v0.1.0-pc.4` come candidate del profilo
read-only; deve ancora superare doppio packaging riproducibile, installazione
pulita, CLI dall'artifact e due quick start live consecutivi prima della RC.

`pc.4` viene prodotto dal commit
`7117f8d93c247b302cd77fb92c484b550a1a7162`; il doppio packaging è
byte-identico con SHA-256
`d62f5e4b49e81508f1256d0cf7ac8785a62a9ac92cf30c144047cea4f14f9fb3`.
Installazione pulita, config read-only, version/help, doctor 9/9, models e agents
sono verdi. Il primo quick start termina però `tool_failure` dopo 169029 ms, 1
turno e 1 tool call; il digest della fixture resta invariato e `pc.4` non è
promuovibile.

L'audit rileva che il system prompt descrive ancora write/patch e digest anche
quando il `RunRequest` non dichiara tool mutanti. Il terminale redatto non prova
la call esatta che ha fallito, ma il disallineamento viola il nuovo confine.
L'hardening rende il prompt capability-aware: i run read-only vietano di
richiedere o simulare mutazioni, mentre il protocollo guarded resta disponibile
solo per tool set sperimentali mutativi. I due rami sono coperti da test. `pc.5`
è il prossimo candidate e deve ripetere l'intero gate senza cambiare modello,
timeout, task o criteri.

`pc.5`, commit `2732f26af4550833ad1b2d9cd4ca1caf5d72cd30`, supera doppio
packaging con SHA-256
`4eb9abdfab6efbd00dc624b509581ec57666da1c4645d60abadc9316104ffe11`,
installazione pulita, doctor 9/9, models, agents e due quick start consecutivi.
I run completano in 330122 ms e 52386 ms, ciascuno con una read reale e
risposta corretta; il digest della fixture resta invariato. `pc.5` è la baseline
validata per produrre il distinto `v0.1.0-rc.1` e non viene rinominato.

Il distinto `rc.1` supera packaging riproducibile e preflight, ma il run di
conferma termina `tool_failure` dopo una tool call; la fixture resta invariata.
L'artifact è storico e non promuovibile. Senza cambiare modello, timeout, task
o criteri, il protocollo read-only viene irrobustito per richiedere nome
funzione e campi schema esatti e path logici relativi senza root fisica, slash
iniziale, URI o parent traversal.

`pc.6`, commit `ab109a5f878b8e1f10d69327736f014ad916a970`, supera doppio
packaging, installazione pulita, CLI completa e due quick start consecutivi con
una read reale, risposta corretta e digest invariato. Il distinto
`v0.1.0-rc.2` dallo stesso commit è byte-riproducibile, misura 3598576 byte e
ha SHA-256
`442090c6e2dac6095aa4532d658def42cd39e04a34baff401b3a92aec1fd9105`.
Da una nuova directory pulita supera checksum, version/help, profilo read-only,
doctor 9/9, models, agents e un run di conferma `completed` in 64296 ms con una
read reale. La fixture resta invariata. La Fase 5 è completata e la Fase 6 è il
prossimo gate; tag e artifact finali v0.1.0 non sono ancora prodotti.

L'integrazione conclusiva della Fase 5 esegue sullo stesso archive `rc.2` i
controlli operativi esplicitamente richiesti dal piano. Un SIGINT durante la run
produce terminale `canceled`, exit 130 e uscita del processo in 2 ms, entro il
budget di shutdown di 30 secondi. Una configurazione temporanea con la sola
variazione `model_turns: 1` produce `limit_exceeded`, exit 1, un turno e una
read reale; la differenza fra wall time del processo e durata agentica è 6 ms,
includendo startup e shutdown. In entrambi gli scenari stdout è vuoto, le
scansioni di canary/path/contenuti/nomi tool sono negative e il digest del
controller rimane invariato.

La Fase 6 congela la verità di prodotto in `docs/compatibility.md`: Linux
`amd64`, Ollama, `llama3.1:8b`, `embeddinggemma:latest` e reference agent
Laravel read-only; mutazioni e llama.cpp sono sperimentali/non supportati e il
processo resta trusted in-process senza sandbox. README e documentazione
artifact-first includono quick start, installazione, configurazione, CLI,
reference agent, security model/policy, troubleshooting, known issues,
compatibilità API, changelog e release notes. Il packaging supporta lo stato
finale distinto `release` soltanto per versioni non prerelease e richiede
l'intera superficie documentale nell'archive. `rc.2` non viene rinominato.

La baseline documentale viene committata in
`6e867c13297c438874e0ecc2e1f334ba19fc7ab6`. La relativa prima build finale è
byte-riproducibile, ma non è promuovibile: dopo una run positiva, il secondo
quick start termina senza tool call e restituisce come testo una pseudo-call
JSON a `workspace_read`. L'archive, SHA-256
`5ad3e297e28033868488c42a3ff58e47a44d393f6c830cc33085a461cc564124`,
resta evidenza rifiutata e non viene pubblicato. Il loop deve trattare una
pseudo-call strutturata verso un tool dichiarato come protocollo incompleto,
richiedere una vera invocazione entro gli stessi hard limit e accettare invece
le normali risposte finali che citano un tool. Un nuovo commit, un nuovo
artifact e quick start consecutivi sono necessari prima della release.

Il nuovo commit `f882919798fa6073bc11c6af18a431bf249a7755` implementa la
correzione bounded senza eseguire effetti impliciti. L'archive finale
`maestro-v0.1.0-linux-amd64.tar.gz`, 3.604.828 byte, è byte-riproducibile e ha
SHA-256
`c785676a177165a2c11ff0fc744931ac8b5d923466155ec32365e7a0c03d271f`.
Da directory pulita supera checksum, version/help, doctor 9/9, models, agents e
due quick start consecutivi: entrambe le run esercitano la correzione, eseguono
una read reale e rispondono correttamente su `OrderService::create`. Il digest
della fixture resta invariato. Il medesimo binario osserva il limite
`model_turns: 1` con exit 1 e SIGINT con exit 130/shutdown in 1.997 ms; anti-leak
e gate repository-wide sono verdi. Il tag annotato `v0.1.0` punta al commit
incorporato nel binario. Fase 6, Milestone 8 e v0.1.0 sono concluse.

---

# Direzione v0.2.0

La v0.2.0 era pianificata come vertical slice mutativo controllato per il
reference agent Laravel:

```text
read -> prepare patch -> preview -> approval -> apply -> reindex -> final
```

La Milestone 9 — Post-release & Benchmark Closure è completata. Le prove live
hanno qualificato artifact e workspace reali, identificato il bug di
indicizzazione v0.1.0 e rilasciato la correzione v0.1.1. L'artifact finale
incorpora `ba938abc6553bc87a89088eb6763a3e255aba4f8`, ha SHA-256
`d894568cd65c261a75212274d7ab8a45eafa950660594b6c22cc777eb8ab9cf1`,
supera due quick start consecutivi e ha tag annotato `v0.1.1`. ADR-0030 chiude
la Milestone 3 con Ollama baseline positiva e llama.cpp non supportato. Il gate
finale ha dato GO alla Milestone 10, ora avviata con la Fase 1 completata.

La Milestone 9 viene eseguita in sei fasi: contratto di osservazione, artifact
e preflight fuori dal checkout, workspace reali e resilienza operativa, triage
e stabilizzazione v0.1.x, chiusura benchmark con decisione llama.cpp, audit
finale GO/NO-GO. Il piano autorevole è
`docs/milestone-9-development-plan.md`; nessuna fase abilita capacità mutative.

Le milestone successive sono:

1. Milestone 10 — Controlled Mutation;
2. Milestone 11 — Mutation Qualification;
3. Milestone 12 — Productization v0.2.0.

Il contratto candidato qualifica prima `workspace.patch` su un solo file
esistente, con diff concreta e approval exact-fingerprint. La creazione tramite
`workspace.write` richiede qualificazione separata. Shell, Git, esecuzione di
processi, sandbox, recovery, multi-agent, tool esterni e modifiche coordinate
multi-file restano fuori scope.

Il piano di release è in `docs/v0.2.0-development-plan.md`; la scomposizione
della Milestone 10 è in `docs/milestone-10-development-plan.md`. Il gate di
osservazione e la decisione llama.cpp sono conclusi; ADR-0031 completa la prima
fase della Milestone 10. La presenza dei tool mutativi sperimentali non amplia
la compatibility promise corrente.

La Milestone 10 è completata. `docs/milestone-10-development-plan.md` la
suddivide in sei fasi sequenziali, tutte concluse, e ADR-0031
congela il contratto Controlled Mutation. `workspace.patch` prepara contenuto
e diff prima dell'approval, richiede TTY e allow once e su Linux applica tramite
temporaneo, sync e `renameat` atomico con fault injection. Il profilo mutativo
è opt-in e quello v0.1.x resta read-only. La sequenza redatta espone proposal,
approval, apply e reindex; ogni run ammette un solo tentativo e nessun testo
finale è possibile prima di una nuova generazione fresh. La matrice
deterministica, l'audit pubblico e il profilo candidato sono consegnati; il
verdetto è GO alla Milestone 11 per la sola qualificazione live, senza ampliare
il support claim v0.1.x.

La Milestone 11 è completata con esito `mutation_deferred`. Il candidato
`v0.2.0-m11-qc.2` supera la matrice deterministica 15/15 e il preflight live sul
lower bound, ma Gate A fallisce al primo tentativo con
`patch_tool_call_invalid`, classificato come limite del modello. La fixture
resta byte-identica, senza approval o effetti; Gate B e Gate C non vengono
eseguiti per fail-fast. ADR-0032 mantiene Controlled Mutation e
`ibm/granite4.1:8b` fuori dal support claim e dà GO alla Milestone 12 soltanto
per una productization v0.2.0 read-only.

La Milestone 12 è avviata. `docs/milestone-12-development-plan.md` la divide in
sei fasi sequenziali: contratto e baseline, superficie di prodotto read-only,
packaging e installazione pulita, gate operativi/sicurezza/anti-leak,
qualificazione live con release candidate immutabile, documentazione e release
finale con tag verificato. La Fase 1 è completata sulla baseline
`2ddbb7bd850f25fb805775d82acaf57c831bd53d`: suite, race detector, vet e doppio
packaging v0.1.1 sono verdi. Nessun artifact v0.2.0 è stato prodotto e il
support claim resta invariato.

La Fase 2 della Milestone 12 è completata. README, changelog, configurazione,
CLI, installazione, quick start, compatibility matrix, security model, known
issues, troubleshooting e reference agent sono allineati al target v0.2.0
read-only. `docs/v0.2.0-api-compatibility.md` registra le sole aggiunte Go
sperimentali e additive; `docs/releases/v0.2.0.md` resta candidate fino ai gate
finali. Il packaging include il contratto API v0.2.0 e verifica esplicitamente
che profilo e documentazione mutativi non siano distribuiti. Suite completa,
race detector, vet, sintassi degli script e confine list/read/search più deny
mutativo sono verdi.

La Fase 3 della Milestone 12 è completata. Il packaging candidate
`v0.2.0-pc.1`, commit `7d3f45ee0268fc758b9e3722e57c91e486065615`, ha SHA-256
`e5f98bedcb94ab40236d3f315cf9af0be976825abbd2d9a6ea756ad26200fc13` e
supera doppio build byte-identico, checksum, archive audit e installazione
fuori dal checkout. `version`, help e `agents` sono verdi; il preflight Ollama
reale completa `doctor` 9/9 e `models` conferma `llama3.1:8b`. Il candidate
resta `packaging-candidate`, è conservato fuori dal repository e non è un RC.

La Fase 4 della Milestone 12 è completata sullo stesso `v0.2.0-pc.1`. EOF e
input oltre 1 MiB falliscono con exit 2 e stdout vuoto; un profilo incluso
alterato a `workspace_mutate: allow` viene respinto in composition. SIGINT
termina `canceled`/130 in 3004 ms, la deadline a 1 secondo termina
`deadline_exceeded`/130 in 1006 ms e `model_turns: 1` termina
`limit_exceeded`/1 dopo una sola read. Tutti i casi conservano il digest della
fixture, non mostrano approval in modalità non interattiva e superano la
scansione anti-leak. Suite `-count=3`, race detector, vet e script syntax sono
verdi.

La Fase 5 della Milestone 12 è completata. I candidate `pc.1`–`pc.4` restano
respinti e immutabili dopo failure live distinti; ogni hardening è entrato in
un nuovo candidate. `v0.2.0-pc.5`, commit
`e8aaad800f1a72eb395f895ba5c8b54195ce0388`, supera due quick start
consecutivi con un turno modello e una read reale. Il distinto
`v0.2.0-rc.1` ha SHA-256
`056f557abe0b95a3a1d758b8827e04907500a988b719e2af9a6ddbfb24886fab`:
doppio packaging, installazione pulita, doctor 9/9, due quick start, SIGINT,
deadline e hard limit sono verdi. La fixture resta byte-identica e nessuna
capability mutativa entra nel percorso. La Fase 6 è il prossimo gate.

La documentazione pubblica v0.2.0 è congelata al commit
`fac2ae347d9fd6e03e9faef466d11bafa961370c`. La build finale deve partire da
un commit pulito discendente e distinto, poi superare doppio packaging,
installazione, quick start di conferma e verifica del tag.

La Milestone 12 è completata con verdetto GO per v0.2.0 read-only. L'artifact
finale `maestro-v0.2.0-linux-amd64.tar.gz` ha SHA-256
`c2d2a6f35178e91ad0c62d3c27f4ff2c33eedb46fd5fb327535890638e963758`
e incorpora `5b05237362370fa79f133e159105a6a99050e81a`, commit discendente dal freeze
documentale. Doppio packaging, installazione pulita, doctor 9/9, conferma live,
fixture invariata, suite tripla, race, vet e anti-leak sono verdi. Il tag
annotato `v0.2.0`, il manifest e il binario coincidono sul commit artifact.
Controlled Mutation resta non supportata; il tag non è stato pubblicato su un
remote.

La Milestone 13 — Field Validation & Adoption è conclusa con classificazione
`field_validation_completed_with_limitations` e decisione
`adoption_no_go_on_reference_profile`. La matrice ufficiale è chiusa a 5/22
per stop rule: 2 completion, 3 provider failure, nessuna risposta `correct` e
workspace invariato 5/5. Le 17 run residue sono `not_run`; la coorte di due
progetti e il Gate 0 di pubblicazione remota non sono stati completati e
restano limitazioni esplicite. Le diagnosi mostrano che più timeout non
corregge la qualità multi-file, la choreography blocca finalizzazioni
premature ma `llama3.1:8b` non progredisce, e `qwen3.5:9b` supera tool calling e
smoke ma non converge semanticamente sul fixture. Il replay delle cinque
query Qwen è deterministico. Un confronto conclusivo direct/chat mostra che il
modello risponde correttamente con file allegato direttamente, mentre il loop
agentico scade dopo read riuscite e ripetute; le completion semplici restano
però lente e variabili e il percorso Maestro pre-caricato raggiunge la
deadline. Sicurezza e immutabilità read-only sono confermate; affidabilità e
qualità multi-file sono insufficienti. La conclusione non è “Maestro non
funziona”: il verified agent richiede capacità di progressione non dimostrate
stabilmente sul profilo corrente e manca una modalità `direct/chat` distinta.
`v0.2.0` resta storicamente valido nel perimetro della Milestone 12, Controlled
Mutation resta non supportata e non viene prodotto `v0.2.1`.

La Milestone 14 — Interaction Modes & Direct Chat è completata con esito
`direct_chat_deferred`. La Fase 1 ha
approvato ADR-0033 e congelato interaction modes, contratto CLI e schema v2;
la Fase 2 ha consegnato schema strict v2, profili separati, context window e
thinking tri-state provider-neutral, mapping Ollama, preflight capability e
propagazione nel verified agent. La Fase 3 ha consegnato `maestro chat`, il
servizio completion separato e il loader single-file confinato, bounded e
fail-closed. La Fase 4 ha chiuso la matrice negativa, equivalenza streaming,
immutabilità e scansione anti-leak. La Fase 5 ha congelato candidato, binario,
profilo, hardware, fixture e oracoli, ma il preflight si è arrestato prima di
C0 perché Ollama non era attivo e l'API loopback rifiutava la connessione.
C0-C4 sono `not_run` e l'esito candidato è `direct_chat_deferred`; provider e
catalogo non sono stati avviati o mutati. La Fase 6 ha auditato dipendenze,
autorità, documentazione, compatibility e support claim e ha consegnato
l'handoff ripetibile alla Milestone 15. La superficie resta development-only;
non è stata prodotta alcuna release.
Separa `maestro chat`, con contesto single-file esplicito e nessun tool,
retrieval, state machine o fallback, da `maestro agent`, che conserva
esplorazione verificata e choreography. Introduce profili distinti con
`num_ctx` e `thinking` osservabili e qualifica sul computer attuale il primo
candidato `qwen2.5-coder:7b` tramite comportamento epistemico, correttezza,
latenza, token, sicurezza e anti-leak. Le sei fasi sequenziali coprono ADR/CLI,
profili, chat single-file, matrice deterministica, qualificazione live e audit
finale. La milestone produce un candidate record, non una release.

La Milestone 15 — Reference Hardware & Read-only Baseline usa Windows con
WSL2/Ubuntu 24.04, 32 GB RAM nominali, RTX 5070 12 GB, Ollama dentro WSL2 e
workspace Linux sotto `/home`. Verifica nell'ordine provider/GPU, direct/chat,
verified agent sintetico e B01 Laravel multi-file 2/2. Se il baseline
multi-file non è verde, nessuna milestone mutativa si apre. Un PASS completo
productizza v0.3.0 read-only con modalità chat/agent e profili separati;
Controlled Mutation resta non supportata.

L'esecuzione della Milestone 15 è conclusa con esito
`verified_agent_rejected`. La piattaforma WSL2/Ubuntu 24.04/RTX 5070, Ollama
0.33.1 e l'offload GPU sono verdi. Direct/chat con `qwen2.5-coder:7b` supera
C0 3/3, C1 3/3 e due coppie streaming/non-streaming, con fixture invariata.
Doctor è 10/10 e suite normale/development, race e vet sono verdi. Il verified
agent `qwen3.5:9b`, context 8192 e thinking default copre la route bootstrap ma
la prima progressione termina `tool_failure` dopo un turno e due call; fixture
invariata, nessun OOM o fallback CPU. La stop rule rende B01 `not_run` e
impedisce la productization con verified agent prevista dalla Milestone 15.
Milestone 16 resta chiusa; la successiva decisione di roadmap isola il solo
percorso direct/chat nella nuova Milestone 17.

La decisione di roadmap del 2026-08-28 ha chiuso la Milestone 16 — Controlled
Mutation Recovery perché il prerequisito verified agent non è stato
qualificato. Il relativo piano resta una traccia storica development-only e
non viene eseguito. Controlled Mutation, tool mutativi e v0.4.0 restano non
autorizzati.

La Milestone 17 è stata riassegnata al Direct/Chat Product Baseline. Consolida
esclusivamente `maestro chat` read-only, tool-free, senza retrieval o fallback,
con zero o un file esplicito contained. La milestone è pianificata in sette
fasi sequenziali: freeze e audit; confine Direct Chat; contesto single-file;
profilo e preflight; streaming e osservabilità; matrice deterministica e
prequalifica sul ThinkPad; packaging candidate e qualifica finale sulla
piattaforma WSL2/Ubuntu 24.04/RTX 5070. Ogni fase produce un report e nessun
tag è ammesso prima del verdetto `direct_chat_product_baseline`.

Il piano autorevole è `docs/milestone-17-direct-chat-development-plan.md`. Il
precedente `docs/milestone-17-mutation-qualification-plan.md` è superato e non
eseguibile. La Milestone 18 — Productization v0.4.0 resta storica e non aperta;
un futuro programma mutativo richiede una nuova decisione e una nuova
numerazione.

Checkpoint Milestone 17: la Fase 1 — Freeze del contratto e audit del candidato
è completata sulla baseline `2759c332c8edcc66f12aa12fd219e32dff3e1dba`.
Suite completa, race detector, vet e test mirati ripetuti tre volte sono verdi;
il binario development-only costruito con `-trimpath` ha SHA-256
`d6aa37122a50525c28f3f61549213377507611ae598ee8d7e17a95e2e85eab3b`.
L'audit ha congelato contratto, evidenza M15, fixture e backlog: la Fase 2 deve
ora provare e irrigidire il confine tool-free del servizio Direct Chat. Nessun
artifact v0.3.0, tag o support claim è stato prodotto.

Checkpoint successivo: la Fase 2 — Confine del servizio Direct Chat è
completata sulla baseline `b1c85e4`. La factory provider predefinita appartiene
ora a `internal/directchat`, usa il modello del profilo chat e non passa più dal
composition root `internal/application`. Test architetturali vietano import di
Agent Runtime, application composition, Context Engine runtime e Tool Runtime;
spy di composizione provano una factory, un preflight, una completion, zero
stream inattesi, zero tool e nessun fallback. Suite completa, vet e race mirata
ripetuta tre volte sono verdi. La Fase 3 deve ora chiudere la matrice del loader
single-file; nessun artifact o support claim è stato prodotto.

Checkpoint 2026-08-29: la Fase 3 — Contesto esplicito single-file è completata
sulla baseline `7251326`. Direct Chat possiede ora la propria validazione del
path logico e non dipende neppure dai tipi del Context Engine: rifiuta path non
normalizzati, assoluti, traversal, backslash, symlink, caratteri di controllo,
formattatori invisibili e separatori di linea. File vuoti, BOM UTF-8 e limite
byte inclusivo sono definiti e testati. Il path entra nel prompt JSON-quoted e
il contenuto usa il confine del messaggio provider, senza sentinelle testuali
collidibili. Race su file, mode, symlink, directory padre e root sono fail-closed;
fixture e workspace restano invariati. Poiché il prompt è cambiato per
hardening, i PASS live M15 restano evidenza storica ma la nuova serie deve
essere eseguita integralmente in Fase 6. La Fase 4 è il prossimo gate.

Checkpoint Fase 4: Profilo dedicato e preflight è completata sulla baseline
`03d3f62`. `productconfig.LoadChat` e
`ValidateChatExecutionProfile` permettono un documento v2 chat-only strict che
richiede soltanto provider, workspace, profilo chat e mutation deny; l'agent
loader completo continua a respingerlo e conserva tutti i propri invarianti.
`maestro doctor --mode chat` esegue cinque check su config, workspace,
composition, modello e generation controls senza completion né graph agentico.
Il nuovo `configs/maestro.chat.example.yaml` congela
`qwen2.5-coder:7b`, context 4096, thinking false e limiti da 1 MiB; il suo
SHA-256 è `7186188ac769787afd9521a0815e58abb18952526757aa878675bdefd19ce7b1`.
Suite normale/development, vet e race mirata `-count=3` sono verdi. La Fase 5
deve chiudere streaming, terminali e osservabilità.

Checkpoint Fase 5: Streaming, terminali e osservabilità è completata sulla
baseline `9505e16`. Complete e stream validano ruolo, terminale `stop`, modello
osservato quando presente, usage non negativo, UTF-8/NUL, limite byte e durata;
lo stream richiede EOF dopo un solo terminale e viene chiuso esattamente una
volta anche su open/receive/validation failure. Gli errori CLI di parsing,
input concorrente o oversized sono ora sempre `chat failed: invalid_request`
con stdout vuoto e prima della provider composition. Cancellazione precede
deadline nella classificazione, output parziali restano non pubblicati e un
successo `stop` espone `truncated=false`; effettivi non attestabili restano
`unknown`. Suite normale/development, race completa, vet, test mirati
`-count=10` e anti-leak sono verdi. La Fase 6 può congelare ed eseguire la
matrice deterministica/live.

Checkpoint Fase 6.1 (2026-08-29): il candidate congelato
`88c4fcbca00a0dbf77d7b7a0d7607dd19c6d8bbe`, versione
`v0.3.0-m17-p6.1`, è stato ripetuto sulla piattaforma WSL2/RTX 5070 con Ollama
0.33.1 e il digest modello atteso. Preflight e C0 sono verdi, ma C1 è 0/3,
equivalenza complete/stream 0/2 e qualità 2/5: il verdetto è
`direct_chat_candidate_failed`. La deviazione dal ThinkPad prescritto non viene
reinterpretata; fixture e terminali restano verdi. Il report canonico è
`docs/reports/milestone-17-phase-6.md`.

È autorizzato un nuovo ciclo F6.2, non la Fase 7. La diagnosi attribuisce
l'omissione di campi richiesti alla domanda collocata prima del file e la
divergenza complete/stream al sampling lasciato al default provider. Il ciclo
sposta la domanda nell'ultimo turno dopo un confine system dell'evidenza,
rafforza regole generiche di completezza/epistemica e fissa temperatura zero
per entrambi i trasporti. Il prompt non contiene risposte della fixture. Prima
del nuovo freeze devono tornare verdi suite normale/development `-count=3`,
race, vet, test mirati e anti-leak; poi la matrice live va ripetuta integralmente
su un nuovo candidate. Packaging e tag restano vietati fino al PASS F6.2.

Freeze F6.2: il commit sorgente
`f059c2e0015d748bc846cce8d790ee11515291ab` produce due binari
`v0.3.0-m17-p6.2` byte-identici con SHA-256
`e4f9a5f734da7db9da91ef00d694fe199f8fb865d59ca8c3fd9629c2964628af`.
Config, modello/digest, fixture, context, thinking, timeout, limiti e oracoli
restano quelli F6.1; temperatura è ora fissata a zero. Suite normale e
development `-count=3`, race, vet e test Direct Chat/CLI `-count=10` sono
verdi. Il record e protocollo di ripresa sono in
`docs/reports/milestone-17-phase-6-candidate-2.md`. La qualifica live completa
deve essere ripetuta sulla WSL2/RTX, senza tuning; Fase 7 resta `NOT_RUN` fino
al PASS.

Risultato live F6.2: identità, doctor 5/5, C0 3/3, terminali, containment e
immutabilità sono verdi. C1 resta 0/3 perché il modello omette il metodo HTTP;
complete/stream sono ora semanticamente identici ma incompleti 0/2 rispetto
all'oracolo. Qualità resta 2/5 con inferenze non supportate. Il candidate è
respinto come `direct_chat_candidate_failed`; il report F6.2 contiene soltanto
evidenza redatta.

Ciclo F6.3: mantenere temperatura zero, modello, digest, config, fixture e
oracoli. Rimuovere il ruolo system intermedio dopo il file, che non è portabile
fra chat template; usare system soltanto all'inizio, poi file user non
attendibile e domanda/contratto come ultimo user turn. Il contratto finale
richiede ogni dimensione domandata, preservazione dei literal e distinzione fra
fatti, assenze e proposte senza valori fixture. Suite normale/development
`-count=3`, race, vet e Direct Chat/CLI `-count=10` sono verdi prima del freeze.
Creare un nuovo candidate F6.3 e ripetere integralmente la matrice live; Fase 7
resta vietata fino al PASS.

Freeze F6.3: sorgente
`e739da6ac7a807b531952f0d06e5b8c0ec1ea6a8`, versione
`v0.3.0-m17-p6.3`, due binari byte-identici con SHA-256
`e377940c03bbb5d0bf8ce8c80703011e4ac3b49c1a9f7cfdf78e3bfba8b3e06c`.
Il digest prompt/service è
`7fd79e1fafb70d0b7726ecca0909f92592f8706df890a9b6fb263c9d5b8575c1`;
config e fixture sono invariati. Il record completo è
`docs/reports/milestone-17-phase-6-candidate-3.md`. Tutti i gate deterministici
sono verdi; ripetere la serie live completa sulla WSL2/RTX senza tuning. Fase 7
resta `NOT_RUN` fino al PASS.

Risultato live F6.3: identità e doctor 5/5 sono verdi; C1 sale a 3/3 e
complete/stream a 2/2, confermando che il layout finale risolve omissione del
metodo HTTP ed equivalenza. C0 è però 2/3 per una negazione certa senza
contesto, mentre qualità resta 2/5 con inferenze non supportate. Terminali,
containment e fixture sono verdi. Il verdict è ancora
`direct_chat_candidate_failed` e Fase 7 resta `NOT_RUN`.

Checkpoint decisionale: F6.1, F6.2 e F6.3 sono tutti respinti. Sampling e
layout hanno cause e miglioramenti dimostrati, ma tre serie mantengono qualità
2/5. Non proseguire con un quarto prompt tuning implicito. Un nuovo recovery
richiede scelta esplicita fra un contratto di risposta strutturato con evidenza
validabile — che riapre profilo, capability, renderer e compatibilità — oppure
una nuova decisione sul modello e una serie completa. Il lineage è registrato
in `docs/reports/milestone-17-phase-6.md`; nessun archive o tag v0.3.0 è
autorizzato.

Decisione F6.4: prima di valutare output strutturato è autorizzato un quarto e
ultimo candidate cambiando soltanto il modello Direct Chat in `qwen3.5:9b`,
digest `6488c96fa5faab64bb65cbd30d4289e20e6130ef535a93ef9a49f42eda893ea7`.
Restano invariati codice/layout F6.3, temperatura zero, context 4096, thinking
false, timeout, fixture, domande e oracoli. Il profilo è
`configs/maestro.milestone-17-candidate-4.yaml`. La scelta usa soltanto
l'evidenza storica direct-chat M13 e non riqualifica il verified agent. Ripetere
tutti i gate deterministici e live; se F6.4 fallisce, fermarsi per una decisione
esplicita sul contratto di output strutturato. Fase 7 resta `NOT_RUN` fino al
PASS.

Freeze F6.4: il commit sorgente
`03986c73199c6f854552f623d14f826fb9594ef2` produce due binari
`v0.3.0-m17-p6.4` byte-identici con SHA-256
`079bbcbdaa09e6c5b73c5aaf7c71658daade4ee46ce08306ad6285f7bfd2a8f0`.
Il profilo `configs/maestro.milestone-17-candidate-4.yaml` ha SHA-256
`173169b61bdc088f69e7898a35c1ab519429a3c5e7e4340a599cb07fb8ce3102`;
service e fixture restano invariati. Suite normale/development `-count=3`,
race, vet, test mirati `-count=10`, script e doppia build sono verdi. Il record
è `docs/reports/milestone-17-phase-6-candidate-4.md`. Ripetere integralmente la
serie live sulla WSL2/RTX; Fase 7 resta `NOT_RUN` fino al PASS.

Risultato F6.4: la serie live sulla WSL2/Ubuntu 24.04/RTX 5070 è PASS. Doctor
5/5, C0 3/3, C1 3/3, equivalenza complete/stream 2/2, qualità 4/5,
containment, terminali, immutabilità e anti-leak sono verdi con modello e
digest esatti. La Fase 7 è autorizzata.

Checkpoint Fase 7 (pre-freeze): mantenere F6.4 immutato. Il packaging v0.3.0
viene ristretto al solo profilo Direct Chat v2 con `qwen3.5:9b`, digest
qualificato, context 4096, thinking false, temperatura zero, streaming opt-in e
deny mutativo. L'archive non deve includere profili agentici/mutativi, report o
materiale development-only. Prima del freeze completare suite normale e
development, race, vet e audit documentale; poi commit pulito, doppio packaging
`v0.3.0-pc.1`, installazione fuori checkout e record SHA-256. La matrice sulla
WSL2/RTX deve usare byte-identico archive/checksum senza rebuild. Fino a quel
PASS il report finale, tag e pubblicazione restano `NOT_RUN`.

Freeze Fase 7 locale: il commit
`70a9630203ccf82a4d8858a9e47b48f5333b9cbd` produce due archive
`maestro-v0.3.0-pc.1-linux-amd64.tar.gz` byte-identici, 3776699 byte, SHA-256
`82bfb33f3fd9af911e3b2b1e89f9920177b281046da21b186512e577e114fb61`.
Il binario interno ha SHA-256
`dee9d5113ccf2db0573b03e8a3851f600d7bc789964793ebae14376f9c849a66`;
il profilo installabile chat-only ha SHA-256
`1c5bbe79edf125485d14518e58ff18c48156eaa0fb91faf82fcf3cd97375d0ee`.
Suite normale/development `-count=3`, race, vet, script, checksum, allowlist,
installazione esterna, version/help, doctor offline fail-closed, containment e
anti-leak sono verdi. Archive e checksum sono in `dist/`; trasferirli
byte-identici sulla WSL2/RTX senza rebuild e compilare la matrice in
`docs/reports/milestone-17-phase-7.md`. Il report finale resta `NOT_RUN`; tag,
RC, release e pubblicazione restano vietati fino al PASS live.

Checkpoint finale Milestone 17: la matrice Fase 7 è PASS sullo stesso archive
`v0.3.0-pc.1`, SHA-256
`82bfb33f3fd9af911e3b2b1e89f9920177b281046da21b186512e577e114fb61`,
installato fuori checkout sulla WSL2/Ubuntu 24.04/RTX 5070 senza rebuild.
Identità, modello/digest, doctor 5/5, no-file, complete/stream, traversal,
symlink, SIGINT, deadline, immutabilità e anti-leak sono verdi. Un primo probe
deadline non valido per root relativa è stato scartato come errore harness;
dopo la sola correzione della posizione del profilo temporaneo è stata ripetuta
l'intera matrice sul medesimo archive, senza tuning. Il verdetto finale è
`direct_chat_product_baseline`: Milestone 17 completata. Tag, release candidate,
artifact finale e pubblicazione v0.3.0 restano workflow separati non eseguiti.
Agent, retrieval, tool e mutation restano non qualificati.

Decisione successiva: Milestone 18 è riassegnata a “Productization & Release
v0.3.0” e non apre nuove funzionalità. Il piano autorevole è
`docs/milestone-18-productization-release-v0.3.0-plan.md`; il vecchio piano
M18 v0.4.0 mutativo resta storico e non eseguibile. La sequenza prevista è:
freeze/audit; documentazione; RC distinto; installazione/audit; un solo gate
live sulla RTX; artifact finale e tag annotato; GitHub Release e verifica degli
asset riscaricati. Sono vietati multi-file, sessioni, nuovi modelli, agent,
retrieval, tool, mutation e cambi CLI/schema. All'apertura la Fase 1 era
`NOT_RUN`; nessun tag, push o write GitHub è stato eseguito aprendo il piano.

Checkpoint Milestone 18 Fase 1: il source auditato `8bf5114` conserva zero
delta in codice, configurazione, fixture, script e dipendenze rispetto al
packaging candidate M17 `70a9630`. Gli hash di profilo, servizio/prompt e
fixture coincidono con il record qualificato; l'archive M17 verifica ancora
SHA-256 `82bfb33f3fd9af911e3b2b1e89f9920177b281046da21b186512e577e114fb61`.
Suite normale/development `-count=3`, race, vet, script, doppio packaging,
installazione temporanea, containment e anti-leak sono verdi. Remote
`origin`, branch `master`, tag annotato futuro `v0.3.0` e allowlist dei soli
file documentali sono congelati nel report
`docs/reports/milestone-18-phase-1.md`. Il verdetto è PASS e autorizza la Fase
2; nessun tag, push, RC persistente o artifact finale è stato prodotto.

Checkpoint Milestone 18 Fase 2: la superficie pubblica v0.3.0 è congelata nel
commit sorgente RC `f33ce456cd65c24abcd5561d7140438ff08e64f1` senza
delta funzionale dalla baseline M17. README, release notes, installazione,
quick start, CLI, configurazione, compatibility, security, known issues,
troubleshooting e packaging distinguono candidate/release, usano asset della
stessa GitHub Release e mantengono visibili tutte le esclusioni. Suite
normale/development `-count=3`, race, vet, script, token/link audit,
credential-shaped scan e doppio packaging RC sono verdi. La prova temporanea
RC ha SHA-256
`b034828a07f33a2643556123c00917ff563d83f1976dab968542712f0df7be3a`;
non è stata persistita. La Fase 3 è autorizzata.

Checkpoint Milestone 18 Fase 3: l'artifact distinto
`v0.3.0-rc.1`, stato `release-candidate`, è congelato dal commit sorgente
`f33ce456cd65c24abcd5561d7140438ff08e64f1`. Le build sono byte-identiche;
l'archive persistente misura 3775354 byte e ha SHA-256
`b034828a07f33a2643556123c00917ff563d83f1976dab968542712f0df7be3a`.
Checksum, manifest, allowlist, token, profilo, fixture, assenza di superfici
escluse, path checkout e credential-shaped data sono verdi. L'RC è conservato
sotto `dist/` senza overwrite e autorizza la Fase 4.

Checkpoint Milestone 18 Fase 4: archive e checksum RC sono stati copiati ed
estratti in una directory nuova fuori checkout senza rebuild. Checksum,
manifest, version/help, installazione in prefix separato, doctor offline
fail-closed, traversal pre-provider, fixture immutabile e anti-leak sono
verdi. Il digest ricorsivo fixture resta
`ae8483e599d7495b10333d00980951680800632ea7b437425d022cd841841fe7`.
La macchina corrente è il ThinkPad Ubuntu senza NVIDIA/Ollama loopback: il
quick start live non è assorbito nel PASS offline e resta proprietario della
Fase 5 sul medesimo RC trasferito byte-identico alla WSL2/RTX 5070.

Checkpoint Milestone 18 Fase 5: preflight `release_environment_blocked`. La
macchina corrente è il ThinkPad Ubuntu senza GPU NVIDIA e senza Ollama
loopback; non esiste bridge locale verso la WSL2/RTX qualificata. L'RC resta
immutabile sotto `dist/`, SHA-256
`b034828a07f33a2643556123c00917ff563d83f1976dab968542712f0df7be3a`,
e il checksum è ancora verde. La matrice live completa è `NOT_RUN`; il
protocollo di trasferimento e ripresa è congelato in
`docs/reports/milestone-18-phase-5.md`. Fasi 6–7 restano vietate: nessun
artifact finale, tag, push o GitHub Release è stato prodotto.

Ripresa Milestone 18 Fase 5: la WSL2/Ubuntu 24.04 vede la RTX 5070, Ollama
0.33.1 su loopback e il modello `qwen3.5:9b` con digest congelato. La matrice
completa sul medesimo RC SHA-256 `b034828a07f33a2643556123c00917ff563d83f1976dab968542712f0df7be3a`
è PASS: identità, installazione, version/help, doctor 5/5, no-file,
complete/stream, traversal, symlink, SIGINT, deadline, immutabilità e
anti-leak sono verdi. Nessun rebuild, pull o tuning è stato eseguito. Il
blocco ambientale è risolto e la Fase 6 è autorizzata.

Checkpoint Milestone 18 Fase 6: il commit release
`3f4c7d4b4fd2e380644cf250ce9e8fec2311af53` produce due archive finali
`maestro-v0.3.0-linux-amd64.tar.gz` byte-identici, 3775317 byte, SHA-256
`6c8f0e883ec8f8c05571fc2e7bc1f4ecac608c2bd7e338395ae0a4253fff1aaf`.
Il binario interno ha SHA-256
`378a0533083b9a00be6c0212ca52001cebc5f77b476a20038bc8e08d1fc3d42d`;
manifest e `maestro version` incorporano versione v0.3.0, stato release e il
medesimo commit. Suite normale/development, race, vet, audit, installazione,
profilo e fixture sono verdi. Il tag annotato locale `v0.3.0` punta allo
stesso commit. Nessun push o asset remoto è stato ancora scritto; la Fase 7 è
autorizzata.

Chiusura Milestone 18: la GitHub Release pubblica v0.3.0 è visibile su
`https://github.com/Axtonno/maestro/releases/tag/v0.3.0` con il tag annotato
sul commit release `3f4c7d4b4fd2e380644cf250ce9e8fec2311af53` e i soli asset archive e
checksum qualificati. Entrambi sono stati riscaricati dal canale pubblico in
una directory pulita. Dimensione 3775317 byte, SHA-256 archive
`6c8f0e883ec8f8c05571fc2e7bc1f4ecac608c2bd7e338395ae0a4253fff1aaf`,
checksum, manifest, versione/stato/commit, binario, profilo, modello/digest,
help e installazione separata sono PASS. Il verdetto finale è
`v0.3.0_released_and_verified`; Milestone 18 completata senza ampliare il
perimetro Direct Chat read-only.

Chiusura Milestone 19: l'asset pubblico v0.3.0 è stato scaricato nuovamente e
installato fuori checkout sul ThinkPad T490s CPU-only. Checksum, manifest,
binario, Ollama `0.32.14`, modello/digest qualificato e doctor chat 5/5 sono
PASS. Su `project-a` reale e Git-clean, C0 no-file e 3/5 casi single-file
completano con qualità 3/3 `correct`; due casi single-file terminano
`deadline_exceeded` a 300,1 secondi. Le completion single-file hanno mediana
91,4 secondi e lo stream atomico non produce output visibile durante l'attesa.
Stato e digest aggregato del workspace coincidono pre/post. Il verdetto
post-release è `operationally_impractical`: non qualifica il ThinkPad, non
modifica v0.3.0 e mantiene fermi verified agent, multi-file, Controlled
Mutation, nuovi provider e altri modelli ufficialmente supportati.

Apertura Milestone 20: il piano autorevole è
`docs/milestone-20-thinkpad-latency-attribution-lower-resource-profile-plan.md`.
La Fase A confronta sul ThinkPad due task congelati, due ripetizioni per task e
per percorso, tra replay Ollama diretto e l'esatto binario Maestro v0.3.0. Il
body deve essere dimostrabilmente equivalente; si misurano primo chunk,
terminale, output visibile, usage, CPU/RSS e immutabilità. Solo il verdetto
`model_hardware_bound` autorizza la Fase B con `qwen2.5-coder:7b` come candidato
development-only. Il gate richiede no-file 3/3, single-file 5/5, coppia
stream/non-stream 2/2, qualità almeno 4/5, zero timeout o mutazioni e riduzione
mediana di almeno 30% e 20 secondi rispetto a qwen3.5 sugli stessi cinque
task. I precedenti failure qualitativi 2/5 di
qwen2.5-coder in M17 restano evidenza vincolante e impediscono qualunque
promozione automatica. Diagnostica config, identità del binario e heartbeat
redatto sono separati dal benchmark. v0.3.0, agent, retrieval, multi-file e
Controlled Mutation restano invariati.

Checkpoint Milestone 20 Fasi A/B: l'archive pubblico e il binario v0.3.0 sono
stati riscaricati e verificati; il binario è stato invocato per path assoluto
perché quello preesistente nel PATH aveva identità diversa. Un relay loopback
temporaneo, con body locali `0600`, ha dimostrato payload byte-identici tra
Maestro e replay Ollama. Sui quattro confronti formali qwen3.5 a modello
residente, i delta terminali Maestro sono compresi tra -0,18 e +0,11 secondi;
verdetto A `model_hardware_bound`. Il primo uso resta costoso e lo stream
atomico nasconde circa 15,8 secondi tra primo chunk e terminale, ma non aggiunge
ritardo terminale materiale. Questo ha autorizzato qwen2.5-coder:7b: no-file
3/3, single-file 5/5, qualità 4/5, stream/non-stream equivalente, zero timeout
e fixture immutata. Sugli stessi cinque task la mediana è 69,0 secondi contro
123,9 di qwen3.5, miglioramento 54,9 secondi / 44,3%; verdetto B
`thinkpad_profile_candidate`. Il profilo resta development-only e non modifica
v0.3.0; i failure M17 restano vincolanti. Report in
`docs/reports/milestone-20-phase-a.md` e
`docs/reports/milestone-20-phase-b.md`. Fase C non è avviata.

Chiusura Milestone 20: la Fase C implementa diagnostica configurazione
tipizzata (`read_failed`, `yaml_invalid`, `unknown_field`, `missing_field`,
`invalid_value`) con soli field path allowlisted, `maestro version
--diagnostic` con status/path risolto/SHA-256 e heartbeat stderr ogni 15
secondi bounded a 40 righe. Stdout chat resta atomico; ticker e goroutine sono
arrestati prima del terminale. Packaging incorpora e verifica lo status. Il
verdetto C è `operational_corrections_ready`; M20 chiude conservando
`model_hardware_bound` e `thinkpad_profile_candidate`, senza promuovere il
modello o modificare v0.3.0. Report in
`docs/reports/milestone-20-phase-c.md` e
`docs/reports/milestone-20-final.md`.

Apertura Milestone 21: il piano autorevole è
`docs/milestone-21-cpu-direct-chat-product-qualification-plan.md`. Il support
claim candidato è limitato a Direct Chat CPU-only sul ThinkPad T490s con
Ollama 0.33.1 e qwen2.5-coder:7b digest
`dae161e27b0e90dd1856c8bb3209201fd6736d8eb66298e75ed87571486f4364`.
Prima della qualifica vengono congelati i dieci task M17+M20 e relativi
oracoli. Servono due serie complete, cold/warm separati, residency esplicita
5m, eviction e memoria dichiarate. Ogni serie richiede completion 100%,
qualità almeno 80%, mediana warm <=60s, massimo warm <=120s e zero timeout.
Il candidate M20 non presume PASS: un task aveva richiesto 190,6 secondi.
Artifact e installazione fuori checkout sono obbligatori; pubblicazione,
agent, tool, retrieval, multi-file e Controlled Mutation restano esclusi.

Checkpoint Milestone 21 Fase 1: il piano incorpora quattro precisazioni
vincolanti. Warm richiede snapshot provider resident positivo, request entro
TTL, nessuna eviction e `load_duration` entro una soglia housekeeping derivata
da cinque probe non qualitative con formula predefinita e cap 2s. Il profilo
congela `num_predict: 512`; troncamento/length resta failure. Ogni serie deve
raggiungere almeno 8/10 correct, nessun task può essere incorrect in entrambe
e nessuna falsità materiale può ripetersi. La matrice artifact è già fissata a
cold no-file e cinque generation warm (`no-file`, Q17-1, Q20-4, Q20-1
complete/stream) con mediana <=60s e massimo <=120s, oltre ai gate operativi.

La matrice task/oracoli è congelata in
`docs/milestone-21-cpu-direct-chat-qualification-matrix.yaml`: M20 conserva
domande e file esatti dei capture; M17 usa superset conservativi perché i
prompt storici completi non furono conservati, deviazione che non viene
descritta come replay esatto. Fixture digest
`a7831ea9d6cfebf397f004ae0bded6fec59ec935962f8e268b79534fc68abda3`.
Hardware ThinkPad/OS/CPU/RAM sono registrati. La Fase 1 è completata con
Ollama 0.33.1 revisione 133, hold `forever`, servizio attivo, SHA-256 binario
`9f595107f966433f93f20ee19043f8e0cdea88e7403672f4dba2cadcb45ee085`,
digest modello riconfermato via API e manifest e soglia housekeeping 300 ms
derivata da cinque probe warm non qualitative. Verdetto
`cpu_qualification_environment_frozen`; report in
`docs/reports/milestone-21-phase-1.md`.

Checkpoint Milestone 21 Fase 2: lo schema strict v3 aggiunge i campi
obbligatori `num_predict` e `residency`, preservando strict e invariato il v2.
Direct Chat inoltra 512 come `options.num_predict` e 5m come `keep_alive` sia
in complete sia in stream; provider senza supporto falliscono prima di I/O.
L'envelope v3 dichiara entrambi i valori e doctor identifica lo schema. Test,
race e vet sono verdi, doctor live è 5/5. Una probe non qualitativa partita da
unload ha completato in 40,568s con heartbeat redatti, è rimasta resident fino
al TTL invariato ed è stata evicted automaticamente dopo la scadenza.
Verdetto `cpu_chat_residency_contract_ready`; Fase 3 è il prossimo gate e
nessun task Q17/Q20 è stato eseguito. Report in
`docs/reports/milestone-21-phase-2.md`.

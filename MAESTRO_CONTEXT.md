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

Stato: In corso — Fasi 1–5 completate; milestone non ancora chiusa.

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

Il completamento della Fase 5 non chiude la Milestone 3. La milestone resta in
corso fino a una decisione esplicita e alle eventuali verifiche live richieste.

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
`llama3.1:8b`; la Milestone 3 resta aperta per la successiva matrice llama.cpp e
fino a una decisione esplicita di completamento.

La seconda fixture `llama3.1:8b` supera la prova diretta: non-stream restituisce
una tool call nativa e lo stream la emette nel primo chunk, seguito da un chunk
terminale con `done_reason: stop`. L'adapter Ollama normalizza ora quel terminale
in `tool_calls` soltanto dopo aver tradotto una tool call nello stesso stream;
le altre cause e gli stream senza tool call restano invariati. La stessa regola
allinea le completion non-stream. Test mirati, gate Go, integration suite,
embedding con ID catalogo esatto e lifecycle passano. Lo Smoke completo
post-correzione produce 13 passed, 1 skipped e 0 failed: il gate live Ollama è
verde. `qwen2.5-coder:7b` resta il caso negativo documentato e
`llama3.1:8b` la fixture positiva. La Milestone 3 resta comunque in corso e non
viene chiusa automaticamente.

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

La matrice live llama.cpp resta un task pendente esplicito della Milestone 3,
idealmente usando lo stesso modello base Llama 3.1 per isolare la differenza tra
runtime e modello. Il task è rinviato perché non blocca Gestor, ma deve essere
completato prima della chiusura formale della Milestone 3 o di una release
pubblica importante.

Checkpoint di passaggio:

| Punto | Decisione |
|---|---|
| Milestone 3 | In corso, sospesa dopo la validazione Ollama |
| Ollama | Gate live superato con `llama3.1:8b` |
| Qwen | `qwen2.5-coder:7b` è il caso negativo canonico |
| llama.cpp | Matrice live rinviata e registrata come task pendente |
| Motivo | Non blocca Gestor; resta obbligatoria prima della chiusura formale della Milestone 3 o di una release pubblica importante |

Regola di avanzamento: la Milestone 3 non viene chiusa, ma non trattiene lo
sviluppo architetturale successivo.

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

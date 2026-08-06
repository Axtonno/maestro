# Maestro Roadmap

Versione: 0.1.0

Stato: Living Document

Ultimo aggiornamento: 2026-08-06

Autori:
- Antonio Cafeo
- OpenAI ChatGPT

---

# Perché esiste questo documento?

Questo documento descrive l'evoluzione prevista del progetto Maestro.

Non rappresenta una pianificazione rigida.

La roadmap definisce la direzione del progetto e le principali milestone, lasciando spazio a revisioni e adattamenti.

---

# Obiettivo

Costruire un runtime locale modulare, estensibile e provider-agnostic per lo sviluppo software assistito da intelligenza artificiale.

---

# Milestone 0 — Fondamenta

Stato: Conclusa

Obiettivi:

- Definizione dell'identità del progetto.
- Definizione della filosofia.
- Definizione dei principi.
- Definizione della visione.
- Definizione dell'architettura.
- Inizializzazione del repository Go.
- Struttura iniziale del runtime.

---

# Milestone 1 — Runtime Core

Stato: Conclusa

Obiettivi:

- Bootstrap del runtime.
- Sistema di configurazione.
- Dependency Injection.
- Event Bus.
- Logging.
- Lifecycle del runtime.
- Gestione del workspace.

Output atteso:

Un runtime funzionante senza alcun provider.

---

# Fase 5 — Provider Runtime/Configuration

Stato: Conclusa

Scope:

- Contratti provider.
- Capability operative.
- Registry e provider predefinito.
- Routing e configurazione.
- Integrazione nel Runtime.
- Primo adapter concreto Ollama.
- Test e documentazione.
- Smoke test contro un'istanza Ollama reale.

L'implementazione e i test isolati sono completati. Lo smoke test live di
listing, completion, streaming, embedding e cancellazione è consolidato nel
gate finale della Milestone 2.

Ulteriori adapter e policy non appartengono al gate di chiusura della Fase 5.

---

# Fase 6 — Plugin Runtime

Stato: Conclusa

Scope completato:

- Contratto pubblico `Plugin` basato su `runtime.Component`.
- Manifest e compatibilità dell'API Plugin Runtime.
- Registry plugin thread-safe.
- Registrazione coordinata con il Runtime Core.
- Riutilizzo di dependency graph, stato e lifecycle globali.
- Catalogo thread-safe di loader in-process.
- Discovery deterministica dei loader e dei plugin registrati.
- Caricamento cancellabile con validazione di ID e risultato.
- Eventi di catalogo, registrazione e caricamento.
- Esposizione tramite il composition root `maestro.New`.
- Primo plugin framework-aware Laravel con detection e health.
- Test e documentazione.

Evoluzione successiva, fuori dal gate della Fase 6:

- Packaging e installazione di plugin esterni.
- Firme e policy di trust.
- Process isolation o sandbox.
- Unload e hot replacement.
- Capability Laravel avanzate.

Il gate si chiude sul modello trusted in-process definito da ADR-0008. Formati
di distribuzione esterni potranno alimentare lo stesso contratto `Loader` senza
modificare registry e lifecycle.

---

# Milestone 2 — Provider Layer

Stato: Evoluzione incrementale

## Fase 1 — Adapter llama.cpp

Stato: Conclusa

Scope:

- Facade pubblica e configurazione tipizzata.
- Completion e streaming SSE tramite Chat Completions API.
- Embedding tramite API compatibile OpenAI.
- Model listing del modello caricato.
- Autenticazione Bearer opzionale.
- Validazione, error handling e propagazione del context.
- Test HTTP in-memory e smoke test live opzionale.
- Documentazione dell'adapter.

Il primo incremento usa la superficie compatibile con OpenAI esposta da
`llama-server`. Lifecycle del processo, download/rimozione dei modelli e policy
di resilienza restano fasi successive della Provider Layer.

Facade, adapter, test isolati e documentazione sono completati. Lo smoke test
live è consolidato nel gate finale della Milestone 2 e non blocca le singole
fasi incrementali.

## Fase 2 — Model Discovery & Lifecycle

Stato: Conclusa

Scope:

- Contratti neutrali per discovery avanzata.
- Stato osservabile dei modelli.
- Capability indipendenti di load e unload.
- Routing nel Provider Runtime.
- Implementazione Ollama.
- Implementazione llama.cpp router mode.
- Test isolati e documentazione.

Pull, delete, progress streaming e policy di resilienza restano fuori dal gate
della Fase 2.

Contratti, routing, implementazioni Ollama e llama.cpp, test e documentazione
sono completati. Gli smoke test live delle capability provider vengono eseguiti
insieme al gate finale della Milestone 2.

## Fasi successive

Il piano dettagliato, le dipendenze e i criteri di uscita sono definiti in
`provider-layer-plan.md`.

### Fase 3 — Model Acquisition & Removal

Stato: Pianificata

Capability opzionali e cancellabili per pull, avanzamento e rimozione dei
modelli, senza accesso diretto del Provider Runtime ai file gestiti dai server.

### Fase 4 — Model Residency Policies

Stato: Pianificata

Keep-alive e autoload configurabili, coordinati senza duplicare lo stato
osservabile posseduto dai provider.

### Fase 5 — Capability Introspection

Stato: Pianificata

Descrittori neutrali per distinguere capability dell'adapter, dell'istanza
configurata e del singolo modello.

### Fase 6 — Error Semantics

Stato: Pianificata

Classificazione uniforme degli errori provider, compatibile con gli errori Go e
utilizzabile dalle policy senza analizzare stringhe o payload proprietari.

### Fase 7 — Resilience Policies

Stato: Pianificata

Retry/backoff e circuit breaker opt-in, limitati dal context e applicati in base
all'idempotenza delle operazioni.

### Fase 8 — Provider Observability

Stato: Pianificata

Hook neutrali per metriche, tracing e logging, con redazione dei contenuti
sensibili e senza dipendenze obbligatorie da SDK esterni.

### Fase 9 — Advanced Generation Baseline

Stato: Pianificata

Opzioni comuni di generazione, output strutturati e tool calling validati sugli
adapter Ollama e llama.cpp.

### Fase 10 — Hardening & Milestone Gate

Stato: Pianificata

Audit di compatibilità, verifica concorrente e matrice live completa. Gli smoke
test rinviati dalle fasi precedenti confluiscono esclusivamente in questo gate.

Output atteso:

Una Provider Layer capace di evolvere indipendentemente dalla progressione del
Runtime Core e del Plugin Runtime.

Gate finale della milestone:

- smoke test Ollama;
- smoke test llama.cpp;
- listing e discovery dei modelli;
- pull e rimozione dove supportati;
- load e unload su modelli dedicati;
- keep-alive configurabile e autoload esplicito;
- capability introspection e normalizzazione degli errori;
- retry, circuit breaker e osservabilità;
- completion, streaming, embedding e cancellazione;
- output strutturati e tool calling sui due adapter;
- suite isolata, race detector, vet e audit della documentazione.

Nuovi adapter, fallback multi-provider, selezione automatica del modello,
supervisione dei processi locali, multimodalità e reasoning non sono requisiti
di chiusura della Milestone 2.

---

# Milestone 3 — Gestor

Obiettivi:

- Registry delle capability.
- Discovery dei componenti.
- Dependency graph.
- Risoluzione delle capability.

Output atteso:

Sistema modulare basato sulle capability.

---

# Milestone 4 — Plugin System

Stato: Baseline completata nella Fase 6; ecosistema in evoluzione

Obiettivi:

- Caricamento plugin.
- Registrazione plugin.
- Lifecycle plugin.
- API pubbliche.

Primo plugin:

Laravel (detection e health implementati).

---

# Milestone 5 — Context Engine

Obiettivi:

- Workspace indexing.
- Analisi AST.
- Context Builder.
- Ottimizzazione token.
- Cache.

Output atteso:

Costruzione intelligente del contesto.

---

# Milestone 6 — Agent System

Obiettivi:

- Pianificazione.
- Task execution.
- Tool calling.
- Permission model.
- Workspace awareness.

Output atteso:

Primo agente autonomo.

---

# Milestone 7 — Ecosistema

Obiettivi:

- Plugin di terze parti.
- CLI completa.
- API pubbliche.
- SDK.
- Documentazione.

---

# Principio della roadmap

La roadmap rappresenta una direzione.

L'ordine delle implementazioni può cambiare se emergono nuove esigenze o migliori soluzioni architetturali.

---

# Decisioni

- Le milestone rappresentano capacità del sistema, non versioni software.
- Nessuna milestone verrà considerata completata senza documentazione e test.

---

# Documenti dipendenti

- architecture.md
- MAESTRO_CONTEXT.md

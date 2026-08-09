# Maestro Roadmap

Versione: 0.1.0

Stato: Living Document

Ultimo aggiornamento: 2026-08-08

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

L'implementazione e i test isolati sono completati. La verifica live di
listing, completion, streaming, embedding e cancellazione confluisce nello
Smoke Benchmark della Milestone 3.

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

Stato: Conclusa

## Fase 1 — Adapter llama.cpp

Stato: Conclusa

Scope:

- Facade pubblica e configurazione tipizzata.
- Completion e streaming SSE tramite Chat Completions API.
- Embedding tramite API compatibile OpenAI.
- Model listing del modello caricato.
- Autenticazione Bearer opzionale.
- Validazione, error handling e propagazione del context.
- Test HTTP in-memory e definizione dello scenario live.
- Documentazione dell'adapter.

Il primo incremento usa la superficie compatibile con OpenAI esposta da
`llama-server`. Lifecycle del processo, download/rimozione dei modelli e policy
di resilienza restano fasi successive della Provider Layer.

Facade, adapter, test isolati e documentazione sono completati. Lo smoke test
live confluisce nello Smoke Benchmark della Milestone 3 e non blocca le singole
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
sono completati. Gli smoke test live delle capability provider confluiscono
nello Smoke Benchmark della Milestone 3.

## Completamento della milestone

Il piano dettagliato, le dipendenze e i criteri di uscita sono definiti in
`provider-layer-plan.md`.

### Fase 3 — Model Acquisition & Removal

Stato: Conclusa

Capability opzionali e cancellabili per pull, avanzamento e rimozione dei
modelli, senza accesso diretto del Provider Runtime ai file gestiti dai server.

Contratti `ModelPuller`, `ModelRemover` e `ModelPullStream`, routing, adapter
Ollama e llama.cpp, test isolati, ADR-0010 e documentazione sono completati.
Gli smoke test live che modificano il catalogo confluiscono nello Smoke
Benchmark della Milestone 3.

### Fase 4 — Model Residency Policies

Stato: Conclusa

Policy opt-in per autoload, rilascio immediato, TTL e permanenza fino allo
shutdown. Discovery resta la fonte osservabile dello stato remoto; Maestro
coordina lease concorrenti e scarica soltanto le residenze che ha caricato.
Ollama usa `keep_alive`, mentre llama.cpp usa load/unload del router. Timer,
stream, concorrenza e shutdown sono verificati in modo deterministico.

### Fase 5 — Capability Introspection

Stato: Conclusa

Descrittori neutrali per adapter, istanza e modello distinguono supporto
strutturale da disponibilità operativa. Ollama usa catalogo e `/api/show`;
llama.cpp usa `/models`, modalità router e argomenti del processo. I report sono
ordinati, validati, senza cache e non introducono selezione automatica.

### Fase 6 — Error Semantics

Stato: Conclusa

`ProviderError` fornisce kind neutrali, operazione, identità, status e
ritentabilità conservativa preservando cause e sentinel Go. Ollama e llama.cpp
condividono la baseline HTTP e classificano anche trasporto, context, risposte
malformate ed errori mid-stream. La classificazione non applica retry.

### Fase 7 — Resilience Policies

Stato: Conclusa

Policy opt-in per retry, backoff, jitter, budget temporale e circuit breaker
per provider, operazione e modello. La matrice di ripetibilità impedisce retry
di pull/remove e di stream dopo il primo chunk; context e comportamento senza
policy rimangono invariati.

### Fase 8 — Provider Observability

Stato: Conclusa

`ProviderObserver` riceve eventi neutrali e correlati per start, tentativi,
retry, transizioni del circuito e un unico terminale. Completion, stream,
embedding, catalogo, lifecycle e introspection espongono durata, esito, usage o
progresso disponibili senza contenuti sensibili. Gli stream chiusi o cancellati
terminano deterministicamente; errori e panic dell'observer sono isolati e
nessun SDK telemetrico entra nel core.

### Fase 9 — Advanced Generation Baseline

Stato: Conclusa

Opzioni comuni di generazione, output strutturati e tool calling validati sugli
adapter Ollama e llama.cpp. Il Runtime applica validazione preflight; messaggi e
stream rappresentano chiamate tool in modo neutrale e la capability
introspection dichiara disponibilità e limiti operativi.

### Fase 10 — Hardening & Provider Handoff

Stato: Conclusa

Audit di compatibilità, verifica concorrente, suite deterministica e handoff
degli scenari live al Benchmark Layer sono completati. Il manifest assegna
modelli fixture, protezioni delle mutazioni, cleanup e redazione alla Milestone
3 senza rendere i servizi live un prerequisito della Provider Layer.

Output consegnato:

Una Provider Layer capace di evolvere indipendentemente dalla progressione del
Runtime Core e del Plugin Runtime.

Gate finale della milestone:

Stato: Superato

- contratti pubblici sottoposti ad audit di compatibilità;
- capability e routing coperti da test deterministici;
- adapter Ollama e llama.cpp verificati con trasporti HTTP in-memory;
- completion, streaming, embedding, lifecycle, pull e remove coperti in modo
  isolato, inclusi errori e cancellazione;
- introspection, resilienza e osservabilità verificate senza servizi esterni;
- output strutturati e tool calling coperti sui due adapter;
- suite completa, race detector, vet e audit della documentazione;
- manifest degli scenari live consegnato alla Milestone 3.

La chiusura della Provider Layer non richiede processi o modelli locali. La
matrice live Ollama/llama.cpp e gli smoke test rinviati diventano il primo
livello del Benchmark & Evaluation Layer.

Nuovi adapter, fallback multi-provider, selezione automatica del modello,
supervisione dei processi locali, multimodalità e reasoning non sono requisiti
di chiusura della Milestone 2.

---

# Milestone 3 — Benchmark & Evaluation Layer

Stato: In corso — Fasi 1–5 completate; milestone non ancora chiusa

Obiettivo:

Fornire benchmark locali e riproducibili per valutare combinazioni di hardware,
provider, modello e plugin in scenari di sviluppo reali.

Prerequisiti:

- Provider Layer completata;
- capability introspection;
- error semantics;
- provider observability;
- Plugin Runtime e baseline Laravel per il Developer Benchmark.

Livelli:

1. Smoke Benchmark — verifica live di provider, modelli, streaming, embedding e
   lifecycle.
2. Runtime Benchmark — misura latenza, throughput, risorse e cancellazione.
3. Developer Benchmark — misura task reali PHP/Laravel, refactor, test e
   recupero tramite embedding.

Famiglie di misura:

- conformità dei provider;
- prestazioni;
- RAM, CPU e VRAM quando disponibile;
- stabilità dello streaming;
- embedding;
- task di sviluppo reali.

Output atteso:

- `maestro bench smoke`;
- `maestro bench provider`;
- `maestro bench model`;
- `maestro bench laravel`;
- report JSON versionato;
- report Markdown leggibile;
- profili hardware documentati;
- dataset minimale di task reali.

Il Benchmark Layer valuta il sistema completo e non produce classifiche
assolute tra modelli. Piano, metriche, rubriche e gate sono descritti in
`benchmark-evaluation-plan.md`.

Le cinque fasi pianificate sono implementate, ma la milestone resta aperta in
attesa di una decisione esplicita e delle eventuali verifiche live desiderate.

Validazione live Ollama del 2026-08-09: integration test superato; Smoke
Benchmark con 9 passed, 3 skipped e 2 failed (`tool_call_missing` e
`tool_stream_terminal_missing`). La verifica diretta di `/api/chat` a
temperatura 0 riproduce entrambi i failure: nessuna risposta o chunk contiene
`message.tool_calls`, mentre la chiamata è emessa come JSON testuale. Per questa
fixture l'adapter Maestro non è l'origine della perdita. La milestone resta
aperta. Il report è in
`reports/milestone-3-live-ollama-validation.md`.

La fixture alternativa `llama3.1:8b` supera il gate diretto e produce
`message.tool_calls` native non-stream e stream. L'adapter normalizza ora il
terminale Ollama `stop` in `tool_calls` solo se nello stesso stream è stata
tradotta una tool call; completion non-stream, altre cause terminali e stream
senza tool call restano coerenti. Con `embeddinggemma:latest`, test mirati, gate
Go, integration test, embedding e lifecycle passano. Il nuovo Smoke completo
chiude con 13 passed, 1 skipped e 0 failed: il gate live Ollama è verde. La
Milestone 3 resta aperta fino a una decisione esplicita di completamento.
`qwen2.5-coder:7b` resta il caso negativo documentato e `llama3.1:8b` la fixture
positiva.

La documentazione del gate Ollama è conclusa. La prossima verifica live della
Milestone 3 è la matrice llama.cpp, idealmente con lo stesso modello base Llama
3.1 per isolare le differenze del runtime dalla variabile modello.

Checkpoint di sospensione della Milestone 3:

| Punto | Decisione |
|---|---|
| Milestone 3 | In corso, sospesa dopo la validazione Ollama |
| Ollama | Gate live superato con `llama3.1:8b` |
| Qwen | `qwen2.5-coder:7b` conservato come caso negativo canonico |
| llama.cpp | Matrice live rinviata e registrata come task pendente |
| Motivo del rinvio | Non blocca Gestor; deve essere completata prima della chiusura formale della Milestone 3 o di una release pubblica importante |

La Milestone 3 non trattiene lo sviluppo architetturale successivo. Il task
pendente llama.cpp resta parte del suo criterio di chiusura.

---

# Milestone 4 — Gestor

Stato: In corso — Fasi 1–3 completate

Documento di design: `gestor-design.md`.

Piano di sviluppo: `gestor-development-plan.md`.

Obiettivi:

- Registry delle capability.
- Discovery dei componenti.
- Dependency graph.
- Risoluzione delle capability.

Output atteso:

Sistema modulare basato sulle capability.

Il design iniziale stabilisce che Gestor indicizza e risolve capability senza
eseguire codice, duplicare il Registry dei componenti o possedere un secondo
dependency graph. Capability dichiarata e disponibilità operativa sono stati
distinti; discovery e risoluzione lavorano su snapshot deterministici.

Fasi di sviluppo:

| Fase | Ambito | Stato |
|---|---|---|
| 1 | Contratti, modello di dominio e ADR-0022 | Completata |
| 2 | Snapshot Registry | Completata |
| 3 | Discovery sources Runtime e Provider | Completata |
| 4 | Resolver e dependency graph | Pronta |
| 5 | Composition root, osservabilità e gate finale | Pianificata |

Ogni fase richiede un report finale prima dell'avanzamento.

---

# Milestone 5 — Plugin System

Stato: Baseline completata nella Fase 6; ecosistema in evoluzione

Obiettivi:

- Caricamento plugin.
- Registrazione plugin.
- Lifecycle plugin.
- API pubbliche.

Primo plugin:

Laravel (detection e health implementati).

---

# Milestone 6 — Context Engine

Obiettivi:

- Workspace indexing.
- Analisi AST.
- Context Builder.
- Ottimizzazione token.
- Cache.

Output atteso:

Costruzione intelligente del contesto.

---

# Milestone 7 — Agent System

Obiettivi:

- Pianificazione.
- Task execution.
- Tool calling.
- Permission model.
- Workspace awareness.

Output atteso:

Primo agente autonomo.

---

# Milestone 8 — Ecosistema

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
- provider-layer-plan.md
- benchmark-evaluation-plan.md
- MAESTRO_CONTEXT.md

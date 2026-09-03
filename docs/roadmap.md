# Maestro Roadmap

Versione: 0.2.0

Stato: Living Document

Ultimo aggiornamento: 2026-08-21

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

Output consegnato:

Un runtime funzionante senza alcun provider.

I sei report di fase retrospettivi e `reports/milestone-1-final.md` documentano
la consegna e il gate storico.

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

I dieci report di fase retrospettivi e `reports/milestone-2-final.md`
documentano la consegna e il gate storico.

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

Stato: Completata — Fasi 1–5 e decisione live concluse

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

Le cinque fasi pianificate sono implementate. ADR-0030 registra la decisione
live conclusiva e chiude formalmente la milestone.

Validazione live Ollama del 2026-08-09: integration test superato; Smoke
Benchmark con 9 passed, 3 skipped e 2 failed (`tool_call_missing` e
`tool_stream_terminal_missing`). La verifica diretta di `/api/chat` a
temperatura 0 riproduce entrambi i failure: nessuna risposta o chunk contiene
`message.tool_calls`, mentre la chiamata è emessa come JSON testuale. Per questa
fixture l'adapter Maestro non è l'origine della perdita. Il report è in
`reports/milestone-3-live-ollama-validation.md`.

La fixture alternativa `llama3.1:8b` supera il gate diretto e produce
`message.tool_calls` native non-stream e stream. L'adapter normalizza ora il
terminale Ollama `stop` in `tool_calls` solo se nello stesso stream è stata
tradotta una tool call; completion non-stream, altre cause terminali e stream
senza tool call restano coerenti. Con `embeddinggemma:latest`, test mirati, gate
Go, integration test, embedding e lifecycle passano. Il nuovo Smoke completo
chiude con 13 passed, 1 skipped e 0 failed: il gate live Ollama è verde.
`qwen2.5-coder:7b` resta il caso negativo documentato e `llama3.1:8b` la fixture
positiva.

La documentazione del gate Ollama è conclusa. Il preflight finale llama.cpp
non trova server, endpoint o profilo single-model configurato; router mode
resta incompatibile con questo host dopo due OOM. Nessuna matrice viene
avviata, gli skip non sono PASS e llama.cpp resta non supportato.

Checkpoint di chiusura della Milestone 3:

| Punto | Decisione |
|---|---|
| Milestone 3 | Completata con decisione ADR-0030 |
| Ollama | Gate live superato con `llama3.1:8b` |
| Qwen | `qwen2.5-coder:7b` conservato come caso negativo canonico |
| llama.cpp | Sperimentale/non supportato; preflight incompatibile |
| Motivo della decisione | Nessun profilo live valido; router mode ha causato OOM sul target |

Un support claim llama.cpp futuro richiede una nuova matrice su un profilo
hardware–server–modello dichiarato; non riapre retroattivamente la Milestone 3.

---

# Milestone 4 — Gestor

Stato: Completata — Fasi 1–5 e gate finale completati

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
| 4 | Resolver e dependency graph | Completata |
| 5 | Composition root, osservabilità e gate finale | Completata |

I cinque report di fase e il report conclusivo sono disponibili in
`docs/reports/`. Gestor è composto nel Runtime pubblico e la milestone è chiusa.

---

# Milestone 5 — Plugin System

Stato: Completata — Fasi 1–5 e gate finale superati

Obiettivi:

- Stabilizzazione delle API pubbliche plugin.
- Catalogo, registrazione e caricamento deterministici.
- Lifecycle e dependency graph condivisi con il Runtime Core.
- Discovery delle capability plugin tramite Gestor.
- Primo reference plugin framework-aware.
- Hardening concorrente, osservabilità e audit di compatibilità.

Primo plugin:

Laravel (detection e health implementati).

Fasi di sviluppo:

| Fase | Ambito | Stato |
|---|---|---|
| 1 | Contratti, audit della baseline e ADR-0023 | Completata |
| 2 | Catalogo, registry e caricamento | Completata |
| 3 | Lifecycle, dependency graph e Gestor | Completata |
| 4 | Laravel reference plugin | Completata |
| 5 | Osservabilità, hardening e gate finale | Completata |

Il design è descritto in `plugin-system-design.md`; il piano operativo, i gate
e i deliverable di ogni fase sono definiti in
`plugin-system-development-plan.md`.

La milestone consolida il modello trusted in-process. Packaging esterno,
marketplace, firme, sandbox, hot loading e plugin di terze parti restano fuori
scope e confluiscono nell'evoluzione dell'ecosistema.

I cinque report di fase e il report conclusivo sono disponibili in
`docs/reports/`. Il Plugin System trusted in-process è composto nel Runtime
pubblico e la milestone è chiusa.

---

# Milestone 6 — Context Engine

Stato: Conclusa

Obiettivi:

- Modello workspace framework-neutral e snapshot immutabili.
- Workspace indexing sicuro, limitato e deterministico.
- Analisi strutturata e AST tramite analyzer sostituibili.
- Retrieval lessicale, strutturale e semantico opt-in.
- Context Builder con provenance e budget espliciti.
- Cache in-memory content-addressed e aggiornamento incrementale.
- Integrazione con Runtime, Gestor, Provider Runtime e plugin Laravel.

Fasi di sviluppo:

| Fase | Ambito | Stato |
|---|---|---|
| 1 | Contratti, ownership e ADR-0024 | Completata |
| 2 | Workspace indexing e snapshot | Completata |
| 3 | Analisi strutturata e AST | Completata |
| 4 | Retrieval, Context Builder e budget | Completata |
| 5 | Cache e aggiornamento incrementale | Completata |
| 6 | Integrazione, osservabilità e gate finale | Completata |

Il design è descritto in `context-engine-design.md`; il piano operativo, i gate
e i deliverable sono definiti in `context-engine-development-plan.md`.

La milestone non include memoria conversazionale, tool execution, permission
model, persistenza distribuita o ranking LLM. Il retrieval semantico resta
opt-in e riusa il Provider Runtime esistente senza selezione implicita di
provider o modello.

Output consegnato:

Costruzione intelligente, deterministica e ispezionabile del contesto entro un
budget dichiarato, composta nel Runtime e verificata con il workspace provider
Laravel.

I sei report di fase e `reports/milestone-6-final.md` documentano i gate.

---

# Milestone 7 — Agent System

Stato: Completata — Fasi 1–7 e gate finale superati

Obiettivi:

- Pianificazione.
- Task execution.
- Tool calling.
- Permission model.
- Workspace awareness.

Fasi di sviluppo:

| Fase | Ambito | Stato |
|---|---|---|
| 1 | Contratti, ownership e ADR-0025 | Completata |
| 2 | Tool catalog e execution boundary | Completata |
| 3 | Permission model e approval flow | Completata |
| 4 | Sessioni, piani e budget | Completata |
| 5 | Loop agentico e tool calling | Completata |
| 6 | Workspace awareness e reference tool | Completata |
| 7 | Integrazione, osservabilità e gate finale | Completata |

Il design è descritto in `agent-system-design.md`; il piano operativo, i gate
e i deliverable di ogni fase sono definiti in
`agent-system-development-plan.md`.

La milestone separa Tool System e Agent System. Ogni effetto attraversa
preparazione, autorizzazione default-deny ed esecuzione limitata. Provider,
modello, workspace, agente e budget rimangono scelte esplicite; Gestor descrive
le capability senza eseguirle.

Output consegnato:

Primo agente autonomo, provider-agnostic, workspace-aware e governato da
permessi espliciti.

Il composition root espone `Tools()` e `Agents()`, registra reference agent e
workspace tool e li descrive tramite Gestor. Sessioni, piani, permission,
streaming, containment, freshness ed eventi redatti sono verificati da gate
deterministici. Il report conclusivo è
`reports/milestone-7-final.md`.

---

# Milestone 8 — Productization v0.1.0

Stato: Completata — Fasi 1–6 e v0.1.0 concluse

Obiettivo:

Rendere Maestro installabile, configurabile e utilizzabile da uno sviluppatore
per eseguire il reference agent locale controllato su un progetto reale.

Fasi previste:

| Fase | Ambito | Stato |
|---|---|---|
| 1 | Release contract e audit | Completata |
| 2 | Configurazione e CLI minima | Completata |
| 3 | Esperienza operativa | Completata |
| 4 | Packaging e installazione | Completata |
| 5 | Validazione live e release candidate | Completata — `v0.1.0-rc.2` |
| 6 | Documentazione pubblica e v0.1.0 | Completata — `v0.1.0` |

La CLI minima comprende `doctor`, `models`, `agents`, `run` e `version`. La
configurazione versionata rende espliciti provider, modello, workspace, agente,
policy, tool e limiti. Il percorso ufficiale usa il reference agent, i
workspace tool built-in e il plugin Laravel.

La Fase 1 ha prodotto audit, design e ADR-0026. Il report live llama.cpp indicato
da decisioni successive non è presente nella baseline Git disponibile: la Fase
5 deve recuperarlo e verificarlo o rieseguire la matrice, salvo decisione
esplicita che delimiti l'adapter come sperimentale. La pubblicazione richiede
inoltre artifact versionato, checksum, licenza, security model, quick start live
e prova d'installazione da ambiente pulito.

La Fase 2 consegna il loader YAML strict `version: 1`, la composition
applicativa e i cinque comandi minimi. Doctor esegue soltanto preflight
read-only; models e agents non selezionano target; run attraversa Agent e Tool
Runtime con policy e hard limit configurati. Il report è
`reports/milestone-8-phase-2.md`.

La Fase 3 rende `maestro run` controllabile dal terminale: mostra limiti,
piano, step, contatori e terminale tramite eventi redatti; l'Approver offre
deny, one-shot e grant exact-action per il run, fallendo chiuso su EOF, input
invalido o no-TTY. stdout contiene soltanto il risultato stabile, stderr
contiene progresso e interazione; SIGINT cancella il run. Il report è
`reports/milestone-8-phase-3.md`.

La Fase 4 produce un packaging candidate installabile e ripetibile, non ancora
un release candidate: la promozione avviene soltanto dopo la validazione live
della Fase 5. La scelta della licenza è anticipata all'inizio della Fase 4; la
Fase 6 ne verifica la pubblicazione definitiva.

La Fase 4 è completata con `v0.1.0-pc.1` Linux `amd64`: build normalizzata e
ripetuta dallo stesso commit, checksum, manifest, Apache-2.0, documentazione,
configurazione e fixture Laravel sono verificati fuori dal checkout. Il report
è `reports/milestone-8-phase-4.md`.

Il primo gate intermedio della Fase 5 conferma il preflight del candidate e una
Smoke matrix Ollama con 13 scenari passed, 1 skipped e 0 failed. Il percorso
Laravel read-only completa con un profilo CPU ridotto, mentre lo scenario
mutativo resta aperto. La matrice llama.cpp in router mode ha causato due OOM
su un host da 15 GiB ed è invalidata; nessuna release candidate è stata
prodotta. Il report è `reports/milestone-8-phase-5-interim.md`.

Gli hardening risultanti sono incorporati in `v0.1.0-pc.2`, packaging candidate
riproducibile dal commit `b9f571ac5914d2565e2a7bd28f4d5d6fc14a2710` con
SHA-256 `91ef1bb196e9904ef3f3f0fefccf3a80acba22f14da43cdccbf9a83680fa41bc`.
Il quick start read-only esatto dall'archive è positivo. Il primo run mutativo
ha però emesso read e patch dipendente nello stesso turno ed è terminato in
modo sicuro senza modificare la fixture. `pc.2` non è promuovibile: ADR-0028
introduce una coreografia deterministica, che dovrà essere incorporata in un
nuovo packaging candidate prima di riprendere i gate live. L'hardening è ora
incorporato in `v0.1.0-pc.3`, commit
`d362b9910f68e5aecae3a489eb5852e339bc3939`, SHA-256
`8fbdfbf9b207c8c984f295240bcb6345d32fcbfa42f5869dd27a39acc158fe26`.
Il doppio gate di packaging è verde; `pc.3` è stato l'input della ripresa live
e non è una release candidate.

Il quick start read-only di `pc.3` è verde. Nel primo gate mutativo
`llama3.1:8b` ha però restituito pseudo-call come testo e nessuna tool call; la
fixture è rimasta invariata. La serie registra 0 successi su 1 tentativo
eseguito; il gate richiedeva 3 successi consecutivi e i tentativi 2–3 non sono
stati eseguiti. Il modello
resta positivo per adapter/tool calling diretto e reference agent read-only,
ma non è supportato per il reference agent mutante. Un modello sostitutivo o
un contratto operativo più stretto deve essere scelto e validato prima della
release candidate.

La prima selezione fail-fast successiva ha esaminato
`rnj-1:8b-instruct-q4_K_M`. Il protocollo diretto read-result-patch supera 3
sequenze su 3 senza eseguire effetti; il primo run read-only del reference agent
invoca una read ma termina `provider_failure` dopo 535537 ms. Il Gate B registra
0 successi su 1 tentativo, il secondo tentativo e il Gate C non vengono eseguiti
e il modello è escluso. Non viene prodotto `pc.4`; il report è
`reports/milestone-8-model-selection.md`.

`ibm/granite4.1:8b` supera poi Gate A 3/3 e due run read-only consecutivi del
Gate B. Nel primo run del Gate C esegue la read e propone una seconda tool call,
ma il run termina `deadline_exceeded` dopo 600077 ms, prima di approval o patch.
Il controller resta byte-identico; i tentativi 2–3 non vengono eseguiti in
fail-fast. Granite è escluso dalla fixture mutativa sul profilo CPU-only
pubblico, `pc.4` resta vietato e la selezione è quindi passata a `qwen3:8b`
non-thinking con gli stessi criteri.

`qwen3:8b` viene eseguito con la riga finale `/no_think` fissata prima dei gate.
Il preflight passa 9 check, ma la prima sequenza del Gate A non emette alcuna
tool call e termina con 227/256 token dopo 100977 ms. Fail-fast arresta le
sequenze 2–3 e impedisce l'avvio dei Gate B/C. La fixture resta byte-identica e
il modello viene scaricato dalla RAM. La matrice 8B corrente è quindi conclusa
senza vincitori: nessun `pc.4` e NO-GO RC finché non viene scelta esplicitamente
una v0.1.0 read-only, un profilo mutativo con requisito hardware superiore o il
rinvio della release.

ADR-0029 approva la prima opzione: la v0.1.0 supporta ufficialmente il solo
reference agent read-only con Ollama e `llama3.1:8b`; il profilo distribuito
espone list/read/search e imposta `workspace_mutate: deny`. Tool e reference
agent mutanti vengono rinviati almeno alla v0.2.0. llama.cpp è sperimentale e
non supportato nella v0.1.0, quindi la matrice live mancante non blocca più
questa release e resta debito separato della Milestone 3. Il nuovo contratto
autorizza `pc.4`, che deve superare packaging riproducibile, installazione
pulita e due quick start read-only consecutivi prima della promozione a RC.

`pc.4`, commit `7117f8d93c247b302cd77fb92c484b550a1a7162`, supera il
doppio packaging ma fallisce il primo quick start read-only con
`tool_failure`, 1 turno e 1 tool call dopo 169029 ms; la fixture resta
invariata. Il prompt incorporato descrive ancora mutazioni anche quando il tool
set ufficiale è read-only. Il candidate resta storico e non promuovibile. Un
hardening capability-aware limita il prompt read-only alle capacità dichiarate
e conserva il protocollo guarded soltanto per profili sperimentali mutativi;
il gate riparte con `pc.5` senza cambiare modello, timeout, task o criteri.

`pc.5` supera poi il doppio packaging, il preflight dall'archive e due quick
start read-only consecutivi (`completed`, una read reale, risposta corretta e
fixture invariata). È la baseline validata per il nuovo artifact distinto
`v0.1.0-rc.1`; nessun `pc.N` viene rinominato retroattivamente.

Il run di conferma di `rc.1` termina però `tool_failure`; l'artifact resta
immutabile e non promuovibile. Un hardening aggiuntivo rende espliciti nomi
funzione, campi schema e path logici relativi per il profilo read-only, senza
cambiare modello, timeout, task o criteri. `pc.6` supera due run consecutivi e
il distinto `v0.1.0-rc.2`, commit
`ab109a5f878b8e1f10d69327736f014ad916a970`, SHA-256
`442090c6e2dac6095aa4532d658def42cd39e04a34baff401b3a92aec1fd9105`,
supera doppio packaging, installazione pulita, CLI completa e run live di
conferma. La Fase 5 è completata; la Fase 6 è il prossimo gate.

L'addendum operativo sul medesimo `rc.2` verifica anche SIGINT
(`canceled`, exit 130, uscita in 2 ms), hard limit `model_turns: 1`
(`limit_exceeded`, exit 1) e shutdown entro il budget di 30 secondi. In
entrambi i casi stdout è vuoto, la scansione anti-leak è negativa e il
workspace resta byte-identico.

La Fase 6 produce l'artifact finale da un nuovo commit successivo alla
documentazione pubblica. `rc.2` non viene rinominato né ripacchettizzato
retroattivamente come `v0.1.0`.

La redazione pubblica congela la matrice autorevole in `compatibility.md` e
aggiunge README artifact-first, quick start, reference agent Laravel, security
model/policy, troubleshooting, known issues, compatibility API, changelog e
release notes. Il packaging finale usa lo stato distinto `release` e verifica
la presenza di questi documenti. Artifact, installazione pulita, run live, tag
e report finale sono verificati nel gate conclusivo della Fase 6.

La prima build finale dal commit
`6e867c13297c438874e0ecc2e1f334ba19fc7ab6` è riproducibile e supera i
controlli statici, ma è rifiutata dal gate live: nella seconda run il modello
ha stampato una pseudo-tool-call JSON senza invocare il tool e senza rispondere
alla richiesta. L'archive con SHA-256
`5ad3e297e28033868488c42a3ff58e47a44d393f6c830cc33085a461cc564124`
non è una release. Il nuovo source candidate deve rifiutare tale testo come
risultato finale, richiedere il canale tool entro gli hard limit e ripetere
l'intero gate da un nuovo commit.

Il commit successivo `f882919798fa6073bc11c6af18a431bf249a7755` applica
questa correzione senza ampliare l'autorità. Il nuovo archive finale è
byte-riproducibile, ha SHA-256
`c785676a177165a2c11ff0fc744931ac8b5d923466155ec32365e7a0c03d271f`
e supera installazione pulita, CLI/preflight, due quick start consecutivi con
read reale, hard limit, SIGINT exit 130, shutdown bounded, immutabilità e
anti-leak. Il tag annotato `v0.1.0` punta al commit incorporato nel binario:
Milestone 8 e Fase 6 sono completate.

Non fanno parte della v0.1.0 SDK stabile, packaging di plugin/tool terzi,
sandbox, memoria persistente, multi-agent, shell, Git, Docker, remote execution
o selezione automatica di provider e modello.

Il gate, il design e il piano operativo sono descritti in
`release-readiness-audit.md`, `milestone-8-design.md` e
`milestone-8-development-plan.md`.

---

# Milestone 9 — Post-release & Benchmark Closure

Stato: Completata — Fasi 1–6, v0.1.1 e gate finale conclusi

Obiettivi:

- osservare la v0.1.0 read-only su installazioni e workspace reali;
- separare bug v0.1.x, limiti del modello, problemi ambientali, UX e richieste
  evolutive;
- verificare installazione, doctor, cancellazione, deadline e hard limit fuori
  dalla fixture embedded;
- chiudere senza ambiguità il debito llama.cpp della Milestone 3;
- congelare le evidenze Ollama storiche e definire il profilo benchmark
  mutativo.

Output atteso:

- `reports/v0.1.0-post-release-observation.md` conclusivo;
- eventuali patch release v0.1.x ristrette al confine read-only;
- decisione formale sulla chiusura della Milestone 3 e sullo stato llama.cpp;
- GO/NO-GO verso il contratto mutativo.

Nessuna capacità mutativa diventa supportata durante questa milestone.

Il piano operativo suddivide la milestone in sei fasi sequenziali: contratto
di osservazione, artifact e preflight, workspace reali e resilienza, triage
v0.1.x, chiusura benchmark/llama.cpp e audit finale. Il dettaglio è in
`milestone-9-development-plan.md`.

---

# Milestone 10 — Controlled Mutation

Stato: Completata — Fasi 1–6 concluse, GO alla Milestone 11

Obiettivo:

Consegnare il percorso controllato:

```text
read -> prepare patch -> preview -> approval -> apply -> reindex -> final
```

Scope candidato:

- `workspace.patch` su un file esistente;
- diff concreta preparata prima dell'approvazione;
- approval vincolata al fingerprint esatto della patch;
- precondizione SHA-256, containment e rifiuto symlink;
- commit atomico e nessun retry implicito;
- invalidazione del contesto all'inizio dell'effetto;
- reindex riuscito prima della risposta finale;
- profilo mutativo separato e opt-in.

`workspace.write` per creare file richiede qualificazione separata. Shell, Git,
Composer, Artisan, PHPUnit, sandbox, recovery e multi-agent restano fuori
scope.

Il release contract viene fissato da un nuovo ADR all'avvio della milestone,
usando le evidenze e i limiti consegnati dalla Milestone 9.

Il piano operativo suddivide la milestone in sei fasi: contratto, proposta e
preview, approval e opt-in, commit atomico, freshness e terminali, audit
finale. Il dettaglio è in `milestone-10-development-plan.md`; ADR-0031 congela
il contratto di release.

---

# Milestone 11 — Mutation Qualification

Stato: Completata — `mutation_deferred`, ADR-0032

Obiettivi:

- qualificare provider, modello e profilo hardware senza modificare i gate per
  ottenere un pass;
- applicare Gate A `3/3`, Gate B `2/2` e Gate C `3/3` in fail-fast;
- introdurre il Developer Benchmark Laravel mutativo;
- verificare outcome ed esatto stato fisico del workspace in ogni scenario;
- documentare esplicitamente supporto sul profilo corrente, requisito hardware
  superiore oppure rinvio della mutazione.

Il benchmark include patch positiva, digest stale, traversal/symlink, deny,
cancellazione, reindex, failure durante l'effetto, tool non dichiarato e replay
di approval.

Il piano operativo separa contratto e baseline, implementazione del benchmark,
matrice deterministica e congelamento del candidato, Gate A, Gate B, Gate C e
audit finale. Il candidato supera matrice deterministica e preflight, ma Gate A
fallisce al primo tentativo; Gate B/C non vengono eseguiti per fail-fast. Il
supporto mutativo è rinviato. Il dettaglio è in
`milestone-11-development-plan.md` e `reports/milestone-11-final.md`.

---

# Milestone 12 — Productization v0.2.0

Stato: Completata — v0.2.0 read-only e tag verificati

Obiettivi:

- configurazione ed esempio supportati esclusivamente read-only;
- packaging candidate riproducibile e installazione pulita;
- quick start read-only consecutivi;
- gate deny, EOF, no-TTY, SIGINT, deadline e hard limit;
- scansione anti-leak e aggiornamento di compatibility, security e known
  issues;
- release candidate immutabile e artifact finale costruito da un commit
  successivo alla documentazione;
- tag verificato contro il commit incorporato nel binario.

ADR-0032 limita il GO al confine read-only. La release è guidata dal contratto
di sicurezza e dalla validazione live, non dalla sola presenza di
`workspace.write` e `workspace.patch` nel codice.

La milestone è divisa in sei fasi sequenziali: contratto e baseline, superficie
read-only, packaging, gate operativi e di sicurezza, qualificazione live e RC,
documentazione con release finale e tag. Il piano operativo è in
`milestone-12-development-plan.md`; il contesto di release complessivo resta in
`v0.2.0-development-plan.md`.

La Fase 5 qualifica `v0.2.0-rc.1`, commit
`e8aaad800f1a72eb395f895ba5c8b54195ce0388`, dopo due quick start consecutivi
sull'archive installato e i gate live di cancellazione, deadline e hard limit.
La Fase 6 congela la documentazione, costruisce l'artifact finale distinto dal
commit `5b05237362370fa79f133e159105a6a99050e81a` e verifica che archive,
manifest, binario e tag annotato `v0.2.0` concordino. Il verdetto finale è GO
limitato al percorso read-only; la pubblicazione remota non è stata eseguita.

---

# Milestone 13 — Field Validation & Adoption

Stato: Completata con limitazioni —
`adoption_no_go_on_reference_profile`

Obiettivo:

Decidere, tramite evidenza field e stop rule, se il profilo di riferimento di
v0.2.0 sia adottabile per analisi Laravel multi-file.

La matrice ufficiale è stata chiusa anticipatamente a 5/22 run: 2 completion,
3 provider failure, nessuna risposta `correct` e workspace invariato 5/5. Le
17 run residue sono `not_run`, non vengono simulate o reinterpretate. La
coorte di due progetti e il Gate 0 di pubblicazione remota non sono stati
completati e restano limitazioni esplicite, non dati imputati.

Le diagnosi fuori matrice mostrano che un timeout più alto recupera
disponibilità ma non qualità multi-file; la progressive choreography respinge
finalizzazioni premature, ma `llama3.1:8b` non segue la progressione.
`qwen3.5:9b` supera i gate diretti A-C e gli smoke Maestro, ma non converge sul
fixture sintetico. Il replay di tutte le cinque query osservate restituisce
conteggi e path identici, localizzando il failure nella scelta semantica e
nella progressione del modello per quel perimetro.

Il confronto conclusivo direct/chat mostra una distinzione ulteriore: Qwen
risponde correttamente alla domanda single-file quando il file è allegato
direttamente, mentre il loop agentico raggiunge la deadline dopo read riuscite
e ripetute. Le completion semplici restano però lente e variabili: il percorso
Maestro pre-caricato scade e una singola coppia non separa adapter da
variabilità del modello. La diagnosi non è “Maestro non funziona”, ma
“verified agent non converge stabilmente sul profilo osservato e manca una
modalità conversazionale leggera distinta”.

Il verdetto è `field_validation_completed_with_limitations` e l'adozione sul
profilo di riferimento riceve NO-GO. Sicurezza e immutabilità read-only sono
confermate; affidabilità operativa e qualità multi-file sono insufficienti.
`v0.2.0` resta storicamente valido nel perimetro qualificato dalla Milestone
12, ma non viene promosso come soluzione multi-file affidabile. Controlled
Mutation resta sperimentale e non supportata.

Non viene prodotto `v0.2.1`. La ricerca seriale di modelli sul ThinkPad viene
interrotta. Una futura qualificazione deve rendere espliciti `num_ctx` e
`thinking`, rafforzare il binding dell'evidenza e superare gate sintetici
agentici prima di B01. Prima di nuovi modelli viene progettata una modalità
`direct/chat` senza tool o fallback implicito; la sua misura resta separata dal
`verified agent`.

Il protocollo completo è in `milestone-13-field-validation-plan.md`; i task
minimi sono in `field-validation-task-matrix.md`; il verdetto conclusivo è in
`reports/milestone-13-field-validation.md`.

---

# Milestone 14 — Interaction Modes & Direct Chat

Stato: Completata — `direct_chat_deferred`

Obiettivo:

Separare formalmente `maestro chat` da `maestro agent`. Chat riceve contesto
esplicito, non dichiara tool, non usa retrieval o state machine e non effettua
fallback agentico. Agent conserva invece esplorazione read-only, evidence
binding, choreography e stop rule.

La modalità chat introduce prompt e profilo modello separati, `num_ctx` e
`thinking` configurabili e osservabili, telemetria di latenza/token e
comportamento epistemico verificato quando il file manca. Il primo candidato è
`qwen2.5-coder:7b`, già usato localmente con Continue ma non ancora qualificato
da Maestro.

La milestone ha implementato il percorso single-file e superato la matrice
deterministica e anti-leak. Il preflight live ha trovato Ollama non attivo:
C0-C4 sono `not_run` e il verdetto è `direct_chat_deferred`. Non produce una
release e non riapre Controlled Mutation. Le sei fasi, ADR-0033, candidate
record e handoff sono chiusi nel report `reports/milestone-14-final.md`.

---

# Milestone 15 — Reference Hardware & Read-only Baseline

Stato: Completata — `verified_agent_rejected`; piattaforma e direct/chat
qualificati, B01 `NOT_RUN`

Obiettivo:

Qualificare e productizzare un nuovo baseline read-only sulla piattaforma
Windows/WSL2/Ubuntu 24.04, 32 GB RAM, RTX 5070 12 GB e filesystem Linux sotto
`/home`.

L'ordine è vincolante: provider/GPU realmente attivi, prestazioni direct/chat,
verified agent sintetico, B01 Laravel multi-file 2/2, matrice operativa e di
sicurezza. Se il baseline non supera la qualifica multi-file, Controlled
Mutation non viene aperta.

La piattaforma, l'offload GPU e `direct/chat` con `qwen2.5-coder:7b` sono
verdi. Il verified agent `qwen3.5:9b` termina con `tool_failure` alla prima
progressione live; la stop rule impedisce B01 e non autorizza la release. Il
piano operativo è in `milestone-15-reference-hardware-readonly-baseline-plan.md`
e il verdetto in `reports/milestone-15-final.md`.

---

# Milestone 16 — Controlled Mutation Recovery

Stato: Chiusa — non autorizzata da Milestone 15

Obiettivo:

Il piano di recovery mutativa è conservato senza rinomina retroattiva, ma non
viene aperto: richiedeva una baseline verified agent e B01 qualificata che la
Milestone 15 non ha prodotto. Controlled Mutation resta rinviata e non
supportata. Il piano storico è in
`milestone-16-controlled-mutation-recovery-plan.md`.

---

# Milestone 17 — Direct/Chat Product Baseline

Stato: Completata — `direct_chat_product_baseline`; F6.4 e packaging candidate
`v0.3.0-pc.1` qualificati sulla matrice finale live

Obiettivo:

Consolidare e productizzare il comando single-file già implementato e
qualificato:

```text
file esplicito -> controllo workspace -> completion diretta -> risposta
```

Il percorso non usa tool, retrieval, state machine, fallback agentico o
capacità mutative. I gate coprono comportamento senza file, correttezza 3/3,
equivalenza streaming, containment, stabilità, anti-leak e qualità reale
almeno 4/5. Dopo i gate produce packaging candidate, installazione pulita e
qualifica finale sulla nuova piattaforma. Un PASS rende v0.3.0 candidata come
release read-only. Il piano è in
`milestone-17-direct-chat-development-plan.md`.

Il precedente `milestone-17-mutation-qualification-plan.md` è conservato come
piano storico non aperto. Un eventuale futuro percorso mutativo dovrà ricevere
una nuova numerazione tramite una decisione separata.

La milestone è divisa in sette fasi sequenziali:

| Fase | Obiettivo | Gate principale |
|---:|---|---|
| 1 | freeze contratto e audit M14/M15 | backlog chiuso e baseline registrata |
| 2 | confine Direct Chat | una completion, zero tool/runtime agentici |
| 3 | contesto single-file | sola disclosure esplicita e contained |
| 4 | profilo e preflight | config strict e capability onorabili |
| 5 | streaming e osservabilità | equivalenza 2/2, terminali e anti-leak |
| 6 | prequalifica sul ThinkPad | C0/C1, qualità 4/5 e regressione verdi |
| 7 | packaging e qualifica finale | archive immutabile verde sulla piattaforma finale |

Ogni fase produce un report autonomo. Il candidate viene congelato prima delle
prove live della Fase 6 e l'archive della Fase 7 viene trasferito byte-identico
senza rebuild. Il tag resta vietato fino al verdetto finale
`direct_chat_product_baseline`.

---

# Milestone 18 — Productization & Release v0.3.0

Stato: Completata — `v0.3.0_released_and_verified`

Obiettivo:

Pubblicare con prudenza l’esatta baseline Direct Chat qualificata dalla
Milestone 17, senza aggiungere funzionalità. Il percorso congela identità e
support claim, prepara documentazione, costruisce un release candidate
distinto, ripete packaging/installazione e un solo gate live sulla RTX 5070,
quindi produce artifact finale, tag annotato, GitHub Release e verifica degli
asset riscaricati.

Il piano è in `milestone-18-productization-release-v0.3.0-plan.md`. Il
precedente `milestone-18-productization-v0.4.0-plan.md` resta una traccia
mutativa storica non eseguibile; Controlled Mutation richiede una futura
decisione e una nuova numerazione.

Sono escluse modifiche a codice Direct Chat, prompt, schema, CLI,
configurazione, modello/digest, multi-file, sessioni, agent, retrieval, tool e
mutation. Ogni delta funzionale invalida il candidate e torna ai gate owner
della Milestone 17.

---

# Milestone 19 — Post-Release Adoption & Lower-Bound Validation

Stato: Completata — `operationally_impractical`

Obiettivo:

Installare sul ThinkPad l'esatto asset scaricato dalla GitHub Release v0.3.0
e provarlo su un progetto Laravel reale, senza usare il checkout e senza
ampliare il support claim. La matrice comprende doctor chat, una domanda senza
file, cinque domande single-file, streaming, latenza, qualità, immutabilità e
problemi d'uso.

Il piano è in
`milestone-19-post-release-adoption-lower-bound-validation-plan.md`; le
evidenze redatte sono in `reports/milestone-19-thinkpad-adoption.md`.

L'asset, il doctor e l'immutabilità sono verdi; le risposte disponibili sono
corrette, ma la matrice single-file completa soltanto 3/5 casi, con due
deadline da 300 secondi e mediana delle completion di 91,4 secondi. Il
ThinkPad resta hardware osservato e non qualificato. Verified agent,
multi-file, Controlled Mutation, nuovi provider e altri modelli supportati
restano fermi.

---

# Milestone 20 — ThinkPad Latency Attribution & Lower-Resource Profile

Stato: Completata — candidate non promosso

Obiettivo:

Attribuire con un confronto appaiato e a payload equivalente la latenza
osservata sul ThinkPad tra Ollama diretto e l'esatto binario Maestro v0.3.0.
Soltanto se il collo di bottiglia risulta principalmente modello/hardware, la
milestone prova `qwen2.5-coder:7b` come candidato development-only per lo
stesso perimetro Direct Chat read-only.

Il piano è in
`milestone-20-thinkpad-latency-attribution-lower-resource-profile-plan.md`.
La matrice non riapre agent, retrieval, multi-file o Controlled Mutation e non
modifica automaticamente il support claim di v0.3.0. Errori di configurazione
specifici, identità del binario e feedback di progresso redatto costituiscono
un workstream separato dalle misure prestazionali.

La Fase A emette `model_hardware_bound`: quattro coppie a body byte-identico
mostrano delta terminali Maestro/Ollama tra -0,18 e +0,11 secondi. La Fase B
emette `thinkpad_profile_candidate`: `qwen2.5-coder:7b` completa no-file 3/3 e
single-file 5/5, qualità 4/5, zero timeout/mutazioni e mediana 69,0 secondi
contro 123,9 secondi di `qwen3.5:9b` sugli stessi task (-44,3%). Il candidato
resta development-only. La Fase C completa diagnostica config specifica,
identità del binario e heartbeat redatto; il report finale mantiene i verdetti
senza promuovere il modello.

---

# Milestone 21 — CPU Direct Chat Product Qualification

Stato: Completata — `cpu_profile_candidate_rejected`

Obiettivo:

Qualificare o respingere `qwen2.5-coder:7b` come esatto profilo Direct Chat
CPU-only sul ThinkPad. La milestone congela Ollama 0.33.1, riconcilia in due
serie complete i task M17+M20, separa cold/warm/residency/eviction, impone
completion 100%, qualità almeno 80%, mediana warm massimo 60 secondi, massimo
warm 120 secondi e zero timeout, quindi verifica un artifact installato fuori
checkout.

Il freeze aggiunge `num_predict: 512`; output troncato resta failure. Una run
warm richiede snapshot resident positivo, avvio entro TTL, nessuna eviction e
`load_duration` entro una soglia housekeeping calibrata prima dei task. Ogni
task deve risultare correct in almeno una delle due serie e nessuna falsità
materiale può ripetersi. La matrice artifact minima è precongelata con cinque
generation warm e conserva gli stessi limiti assoluti.

Il piano è in
`milestone-21-cpu-direct-chat-product-qualification-plan.md`. Il profilo M20
resta un candidato fino al verdetto; agent, tool, retrieval, multi-file e
Controlled Mutation restano esclusi.

Task, oracoli e ordini sono congelati in
`milestone-21-cpu-direct-chat-qualification-matrix.yaml`. La Fase 1 congela
Ollama 0.33.1/revisione 133 con hold indefinito, digest modello riconfermato e
soglia housekeeping 300 ms. La Fase 2 consegna schema strict v3,
`num_predict: 512` e residency 5m inoltrati a Ollama per complete/stream; una
probe non qualitativa conferma unload, cold snapshot, permanenza entro TTL ed
eviction automatica. La Fase 3 congela un candidate e un profilo packaging
riproducibili.

Le due serie live completano l'intera matrice ma ottengono entrambe completion
7/10 e qualità 4/10. Le mediane warm sono 72,864 e 67,609 secondi; i massimi
192,581 e 175,234 secondi. Sei task sono incorrect in entrambe le serie,
incluse falsità materiali ripetute. No-file, streaming, timeout zero,
residency/eviction, containment e immutabilità passano.

La qualifica artifact, ammessa soltanto dopo due serie verdi, è chiusa
`NOT_RUN`. Il verdetto finale è `cpu_profile_candidate_rejected`: nessuna
versione o promessa CPU viene assegnata e il support claim v0.3.0 resta
invariato. Report finale in `reports/milestone-21-final.md`.

Il verdetto è circoscritto al T490s con l'esatto candidato M21 e non dimostra
che Maestro non possa funzionare senza GPU. Le classi hardware correnti sono:

| Classe | Stato |
|---|---|
| Legacy CPU — ThinkPad T490s | Development-only, nessuna promessa operativa |
| Modern CPU-only | Non ancora qualificata |
| GPU reference — RTX 5070 | Supportata dal profilo v0.3.0 qualificato |

## Direzione successiva proposta

**Milestone 22 — Operational Hardening v0.3.1** deve productizzare diagnostica
di configurazione specifica, identità del binario, heartbeat redatto,
residency esplicita e limite di generazione configurabile. Mantiene
`qwen3.5:9b`, RTX 5070 e il support claim corrente; prima della pubblicazione
richiede una breve riqualificazione sul reference hardware e non promette
supporto CPU.

Una milestone successiva e distinta potrà qualificare una CPU moderna. Anche
su una macchina dotata di RTX 5070, la prova è valida come CPU-only soltanto
con offload disabilitato, zero layer GPU, zero VRAM del modello, processo
Ollama CPU-only e configurazione congelata. Questa separazione permette di
distinguere il limite del modello dal limite della CPU legacy del T490s.

## Associazione release

| Versione | Obiettivo |
|---|---|
| v0.2.0 | artifact storico read-only; Field Validation con adoption NO-GO |
| v0.3.0 | baseline Direct Chat read-only pubblicata e verificata; support claim invariato in M19-M20 |
| v0.4.0 | non pianificata attivamente; nessun support mutativo senza nuova qualifica |

## Sequenza post-v0.2.0

```text
Milestone 13 — chiusa con limitazioni / adoption NO-GO
    -> Milestone 14 — Interaction Modes & Direct Chat
        -> Milestone 15 — hardware e direct/chat PASS; verified agent FAIL
            ├── Milestone 16 — CLOSED; mutazione non autorizzata
            └── Milestone 17 — Direct/Chat Product Baseline
                ├── PASS -> Milestone 18 — Productization & Release v0.3.0
                │   ├── PASS -> v0.3.0 pubblicata e asset verificati
                │   │   └── Milestone 19 -> adoption ThinkPad operationally impractical
                │   │       └── Milestone 20 -> model hardware-bound; profilo ThinkPad candidato non promosso
                │   │           └── Milestone 21 -> qualifica prodotto Direct Chat CPU
                │   └── FAIL -> nessuna pubblicazione o release incident esplicito
                └── FAIL -> nessuna release
```

La sequenza privilegia l'utilità quotidiana già dimostrata senza ampliare
l'autorità. Verified agent e Controlled Mutation restano sperimentali e non
sono fallback del percorso chat.

---

# Milestone 22 — Operational Hardening v0.3.1

Stato: Completata — `v0.3.1_operational_hardening_qualified`

Obiettivo:

Productizzare diagnostica di configurazione specifica, identità del binario,
heartbeat redatto, residency esplicita e limite di generazione configurabile,
mantenendo il support claim Direct Chat GPU di v0.3.0.

Il profilo pubblico strict v3 conserva Ollama 0.33.1, `qwen3.5:9b`, digest
qualificato, context 4096, thinking false e temperatura zero; aggiunge
`num_predict: 512` e residency 5 minuti. Suite, race, vet, doppio packaging e
archive audit sono verdi. La breve riqualificazione live sulla RTX 5070 supera
doctor 5/5, no-file, complete/stream, equivalenza semantica, containment,
immutabilità e anti-leak.

CPU, agent, retrieval, multi-file, tool e Controlled Mutation restano fuori
scope. Il verdetto qualifica il candidate v0.3.1 ma non implica tag o
pubblicazione remota. Piano e report sono in
`milestone-22-operational-hardening-v0.3.1-plan.md` e
`reports/milestone-22-final.md`.

---

# Milestone 23 — Release Readiness & Publication v0.3.1

Stato: Completata — `v0.3.1_candidate_rejected_length_regression`

Obiettivo:

Verificare compatibilità byte-level e comportamentale del profilo v2 pubblico,
contratto v3, regressione qualitativa appaiata sui cinque task M17, identità
reale e riproducibilità LF prima di pubblicare v0.3.1. La milestone non aggiunge
funzionalità e applica stop fail-fast per incompatibilità v2, qualità sotto
4/5, terminali `length` o identità non riproducibile.

Piano e matrice sono in
`milestone-23-release-readiness-publication-v0.3.1-plan.md` e
`milestone-23-release-readiness-matrix.yaml`.

Compatibilità v2, contratto v3, suite, race, vet e policy LF sono verdi. La
matrice live si ferma però al primo task: v0.3.0 completa Q17-1 con 535 output
token e terminale `stop`, mentre il candidate v3 raggiunge `length` a 512 e
fallisce chiuso con `response_invalid`. Q17-2…Q17-5 e tutte le fasi di release
sono `NOT_RUN` per stop rule. Nessun commit release, RC pubblicabile, artifact
finale, tag o GitHub Release viene prodotto. Report in
`reports/milestone-23-final.md`.

---

# Milestone 24 — v0.3.1 Generation Bound Recovery

Stato: Completata — `v0.3.1_released_and_verified`

Obiettivo:

Correggere esclusivamente il limite generativo respinto da M23. Cinque run
progettuali v0.3.0 osservano 71–944 output token; il nuovo e unico candidate
congela `num_predict: 1024` e ripete l'intera matrice Q17-1…Q17-5 in confronto
appaiato. Un nuovo `length` vieta aumenti seriali e rinvia il profilo GPU al
comportamento provider-default. Piano e matrice sono in
`milestone-24-v0.3.1-generation-bound-recovery-plan.md` e
`milestone-24-generation-bound-recovery-matrix.yaml`.

La matrice candidate completa 5/5 con `stop`, qualità 4/5 e zero regressioni
appaiate. Test, race e vet sono verdi su una materializzazione LF. Il vero RC
dal commit `6b879d8` è byte-riproducibile e supera il gate live fuori checkout;
la pubblicazione finale è quindi autorizzata. L'artifact finale costruito dal
commit `bd0e902` è stato pubblicato con tag annotato `v0.3.1`, riscaricato e
verificato byte per byte. Il gate live sul download pubblico è PASS. Report in
`reports/milestone-24-final.md`.

---

# Milestone 25 — v0.3.1 Direct Chat Field Adoption

Stato: Completata — `field_adoption_negative`

Obiettivo:

Osservare l'utilità quotidiana dell'esatto asset pubblico v0.3.1 sul reference
hardware, usando uno o due progetti Laravel reali e soltanto Direct Chat
single-file read-only. La matrice misura completion, qualità, falsità
materiali, terminali `length`, latenza, utilità percepita, diagnostica,
heartbeat e immutabilità senza tuning o nuove funzionalità.

Progetti, path, prompt, risposte e contenuti restano in evidenze locali
private; il repository riceve soltanto risultati redatti. La qualifica CPU
moderna è esplicitamente separata e richiederà GPU disabilitata e zero offload
verificato. Piano e matrice sono in
`milestone-25-v0.3.1-direct-chat-field-adoption-plan.md` e
`milestone-25-v0.3.1-direct-chat-field-adoption-matrix.yaml`.

La campagna su due progetti reali chiude con completion 7/11, correct 4/7
delle valutabili, zero mutazioni e verdetto `field_adoption_negative`. Quattro
task terminano `response_invalid`; report conclusivo in
`reports/milestone-25-final.md`.

---

# Milestone 26 — Response Validity & Field Quality Recovery

Stato: Completata — `v0.4.0_candidate_field_qualified`

Obiettivo:

Spiegare i quattro `response_invalid` di M25 prima di cambiare prompt o
profilo, introdurre una cattura privata dei payload per nuove run diagnostiche,
rendere attribuibile ogni ramo del validator e progettare un contratto
epistemico che separi fatti, inferenze e informazioni non determinabili.

L'inventario iniziale dimostra che i body originali M25 non sono stati
conservati e non possono essere riprodotti offline. Dimostra inoltre che
l'heartbeat streaming era corretto sulla finestra di generation e che la
diagnostica residency semanticamente invalida è già tipizzata. Piano e matrice
sono in `milestone-26-response-validity-field-quality-recovery-plan.md` e
`milestone-26-response-validity-field-quality-recovery-matrix.yaml`.

Restano fuori scope CPU, altri modelli, multi-file, agent e Controlled
Mutation. Nessuna release precede una nuova matrice sui task reali M25.

La diagnosi raw attribuisce tutti e quattro i nuovi capture al limite generativo
congelato (`done_reason=length`, 1024/1024 token), senza difetti dimostrati di
adapter, validator o provider. Il validator resta fail-closed e `num_predict`
resta 1024. Il contratto epistemico v0.4.0 impone risposte entro 450 parole e
separa `Observed facts`, `Possible inferences` e
`Information not determinable`; stderr e heartbeat sono coperti anche nei
confini temporali di preflight e terminale.

I candidate rc.1–rc.3 sono respinti dai gate di completion o qualità. rc.4
supera sia la matrice pre-release sia la singola ripetizione Field Adoption:
11/11 completion, 10 correct, 1 partial, zero `response_invalid`, zero
terminali `length`, zero falsità materiali e mediana utilità 5/5. La milestone
è chiusa senza pubblicare o taggare v0.4.0; dettaglio in
`reports/milestone-26-final.md`.

---

# Milestone 27 — v0.4.0 Release Readiness & Publication

Stato: Completata — `v0.4.0_released_and_verified`

Obiettivo:

Validare indipendentemente il candidate v0.4.0 qualificato da M26, ripristinare
una baseline globale verde e completare una catena di packaging, installazione,
gate live e pubblicazione verificabile senza modificare il perimetro di
autorità Direct Chat single-file read-only.

La selezione di `v0.4.0-rc.4` è conclusa e non può usare il nuovo holdout. Ogni
modifica funzionale al prompt, validator, adapter, profilo o autorità invalida
il candidate e richiede il ritorno ai gate owner appropriati. Piano e matrice
sono in `milestone-27-v0.4.0-release-readiness-publication-plan.md` e
`milestone-27-v0.4.0-release-readiness-publication-matrix.yaml`.

Il commit `0c1a9f7cc596eaee05436f91f8030989871b9ca7` ha superato suite globale,
race, vet, compatibilità v2/v3, holdout indipendente, doppio packaging e gate
live fuori checkout. La release pubblica v0.4.0 e i due asset riscaricati sono
stati verificati byte-per-byte; il support claim resta single-file read-only.

---

# Milestone 28 — Controlled Mutation Recovery

Stato: Completata senza qualificazione — `controlled_mutation_transport_unresolved`

Linea candidata: v0.5.0

Obiettivo:

Qualificare una sola modifica atomica a un file esplicitamente indicato,
proposta dal modello ma validata, mostrata e applicata esclusivamente da
Maestro dopo approvazione esplicita. La validazione post-release v0.4.0 può
continuare informalmente, ma non costituisce la milestone principale.

M28 confronta tool calling nativo e output strutturato vincolato senza fallback
nella stessa run. Entrambi convergono sul medesimo compilatore deterministico,
fingerprint, preview completa, controllo anti-stale e apply atomico. Multi-file,
agente autonomo e ogni effetto senza nuova approvazione restano esclusi.

Il piano e la matrice autorevoli sono in
`milestone-28-controlled-mutation-recovery-plan.md` e
`milestone-28-controlled-mutation-recovery-matrix.yaml`. Il superamento di M28
può qualificare un candidate, ma la pubblicazione v0.5.0 richiede una successiva
release readiness separata.

Il protocollo deterministico, l'approval one-shot, l'atomicità e la matrice
negativa sono verdi. I due trasporti risultano equivalenti fino al compilatore,
ma provider e modello target non erano disponibili per il confronto semantico
e live: nessun trasporto è selezionato e nessun candidate v0.5.0 è autorizzato.
Il dettaglio è in `reports/milestone-28-final.md`.

---

# Principio della roadmap

La roadmap rappresenta una direzione.

L'ordine delle implementazioni può cambiare se emergono nuove esigenze o migliori soluzioni architetturali.

---

# Decisioni

- Le milestone rappresentano capacità del sistema; la Milestone 8 associa la
  productization completata al gate di pubblicazione v0.1.0.
- Nessuna milestone verrà considerata completata senza documentazione e test.

---

# Documenti dipendenti

- architecture.md
- provider-layer-plan.md
- benchmark-evaluation-plan.md
- context-engine-design.md
- context-engine-development-plan.md
- agent-system-design.md
- agent-system-development-plan.md
- adr/ADR-0026.md
- release-readiness-audit.md
- milestone-8-design.md
- milestone-8-development-plan.md
- milestone-9-development-plan.md
- v0.2.0-development-plan.md
- milestone-12-development-plan.md
- milestone-13-field-validation-plan.md
- field-validation-task-matrix.md
- milestone-14-interaction-modes-direct-chat-plan.md
- milestone-15-reference-hardware-readonly-baseline-plan.md
- milestone-16-controlled-mutation-recovery-plan.md
- milestone-17-direct-chat-development-plan.md
- milestone-17-mutation-qualification-plan.md
- milestone-18-productization-v0.4.0-plan.md
- reports/v0.1.0-post-release-observation.md
- configuration.md
- cli.md
- MAESTRO_CONTEXT.md

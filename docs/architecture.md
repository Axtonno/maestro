# Maestro Architecture

Versione: 0.6.0

Stato: Draft

Ultimo aggiornamento: 2026-08-12

Autori:
- Antonio Cafeo
- OpenAI ChatGPT

---

# Perché esiste questo documento?

Questo documento descrive l'architettura logica di Maestro.

Definisce i componenti fondamentali, le loro responsabilità e le relazioni tra essi.

Non descrive implementazioni specifiche, ma il contratto architetturale che ogni implementazione dovrà rispettare.

---

# Obiettivo

Costruire un runtime modulare, estensibile e provider-agnostic capace di orchestrare strumenti, provider, plugin e agenti AI dedicati allo sviluppo software.

---

# Architettura generale

```
                    +----------------+
                    |      CLI       |
                    +----------------+
                            |
                            v
                    +----------------+
                    |    Runtime     |
                    +----------------+
                            |
        +---------+---------+---------+---------+
        |         |         |         |         |
        v         v         v         v         v
   Provider    Gestor   Context    Plugins    Tools
                Layer     Engine
        |                   |
        |                   |
        +---------+---------+
                  |
                  v
               Agent
```

Il Runtime rappresenta il punto di ingresso dell'intero sistema.

Ogni componente comunica attraverso interfacce ben definite.

---

# Componenti principali

## Runtime

Responsabilità:

- bootstrap del sistema;
- configurazione;
- lifecycle;
- orchestrazione;
- dependency injection;
- event bus.

Il Runtime non contiene logica di dominio.

L'Event Bus interno permette la comunicazione disaccoppiata tra componenti.
La consegna è sincrona, ordinata per sottoscrizione e thread-safe. Il Runtime e
il `Context` dei componenti espongono la stessa istanza del bus.

Il Runtime compone inoltre un Provider Runtime condiviso. Applicazioni e
componenti usano la stessa istanza per registrare, risolvere e invocare provider
senza conoscere le implementazioni concrete.

Il composition root pubblico compone anche un Plugin Runtime. La registrazione
di un plugin confluisce nel Registry dei componenti, così dependency graph,
stato e lifecycle rimangono unici.

Lo stesso composition root espone `Gestor()`: Registry delle capability,
sorgenti Runtime/Provider e Resolver condividono snapshot e generazioni con una
vista read-only del dependency graph. Il contratto `pkg/runtime.Runtime` resta
invariato; Gestor è un servizio additivo del Runtime pubblico di Maestro.

Il composition root espone inoltre le istanze condivise `Tools()` e
`Agents()`. Tool Runtime possiede catalogo, policy, authorization ed execution;
Agent Runtime possiede sessioni, piani, budget e loop provider-tool. Entrambi
riusano Provider Runtime, Context Engine ed Event Bus già composti.

La configurazione viene iniettata nel composition root e consegnata ai
componenti come snapshot a chiavi esatte. Il core non decide come leggere file,
variabili d'ambiente o secret esterni.

Invarianti del Runtime interno

Le implementazioni contenute in `internal/runtime` nascondono la propria rappresentazione e consentono modifiche soltanto attraverso operazioni controllate.

Ogni tipo protegge gli invarianti del proprio livello di responsabilità. Gli invarianti locali appartengono al tipo proprietario, mentre quelli che coinvolgono più entità sono coordinati dal relativo aggregato.

Questa convenzione permette di modificare e ottimizzare l'implementazione interna senza ampliare i contratti pubblici o distribuire la responsabilità degli invarianti tra più componenti.

---

## Provider Layer

Responsabilità:

- comunicazione con i modelli;
- rappresentazione delle richieste conversazionali;
- streaming;
- embedding;
- discovery e lifecycle dei modelli;
- acquisizione con progresso e rimozione dei modelli;
- policy opt-in per residenza, lease e rilascio dei modelli;
- introspection di supporto e disponibilità per adapter, istanza e modello.
- classificazione provider-neutral degli errori operativi.
- eventi operativi redatti per logging, metriche e tracing applicativi.
- sampling comune, output strutturati e tool calling provider-neutral.

Il contratto del layer è capability-based. L'identità del provider è separata
dalle capability di completion, streaming, embedding, model listing, discovery,
load, unload, pull e remove. Le policy di residenza coordinano soltanto
transizioni avviate da Maestro; discovery rimane la fonte osservabile dello
stato effettivo del provider.

I report di capability sono snapshot senza cache: non selezionano provider o
modelli e non sostituiscono il routing capability-based.

Gli adapter classificano gli errori al proprio confine tramite un envelope
tipizzato comune. La ritentabilità è metadata; le decisioni di retry restano
responsabilità delle policy di resilienza opt-in del Provider Runtime. Retry e
circuit breaker sono isolati per operazione e modello opzionale e non
introducono fallback.

Un observer opzionale del Provider Runtime correla start, tentativi, retry,
transizioni del circuito e terminale senza osservare i payload. Gli stream
mantengono il confine aperto fino a EOF, errore o chiusura; il core non dipende
da SDK telemetrici e non esegue callback mantenendo lock interni.

Il Provider Runtime mantiene un registry thread-safe, applica una selezione
esplicita del provider predefinito e inoltra le operazioni senza mantenere lock
durante l'esecuzione di codice esterno.

La baseline avanzata conserva nel core soltanto opzioni condivise da Ollama e
llama.cpp. Tool execution, reasoning e opzioni proprietarie appartengono a
layer superiori o a evoluzioni successive.

Implementazioni disponibili:

- Ollama
- llama.cpp

Implementazioni previste:

- LM Studio
- OpenAI
- Anthropic

---

## Benchmark & Evaluation Layer

Responsabilità:

- esecuzione riproducibile degli smoke test live;
- misura di latenza, throughput e risorse;
- valutazione di streaming, embedding e lifecycle provider;
- scenari di sviluppo basati su fixture e plugin;
- produzione di report JSON e Markdown;
- descrizione redatta del profilo hardware–provider–modello–plugin.

Il layer consuma capability introspection, error semantics e osservabilità della
Provider Layer. Non introduce classifiche assolute tra modelli, non invia
risultati a servizi remoti e non modifica automaticamente la configurazione.

I contratti versionati vivono in `pkg/benchmark`; runner, parsing, aggregazione,
redazione e serializzazione vivono in `internal/benchmark`. Il report JSON è la
fonte raw e il Markdown è una sua derivazione. La semantica è descritta in
`benchmark-runtime.md` e registrata in ADR-0017.

Lo Smoke Benchmark costruisce un provider esplicito da configurazione ambiente,
usa introspection prima delle operazioni live e affida al runner ogni cleanup.
Mutation guard e fixture ownership impediscono rimozioni implicite di modelli
preesistenti. La matrice è descritta in `smoke-benchmark.md` e ADR-0018.

Il Runtime Benchmark separa scenari provider e modello. Latenze e throughput
sono misurati attorno alle API pubbliche; retry e circuit breaker usano fault
transienti controllati; CPU e RAM sono opzionali e dichiarano sempre processo e
metodo di raccolta. Lifecycle e cold/warm non lasciano policy di residency
installate. Le decisioni sono descritte in `benchmark-runtime.md` e ADR-0019.

Il Developer Benchmark incorpora un dataset Laravel/PHP versionato, avvia il
plugin Laravel su una materializzazione temporanea e mantiene la qualità
separata dal risultato tecnico. Checklist deterministiche e ranking embedding
producono score 0–3 senza evaluator LLM o contenuti completi nel report. Il
contratto è descritto in `developer-benchmark.md` e ADR-0020.

Il JSON versionato è la fonte canonica del reporting. Il Markdown usa lo stesso
report validato e redatto e può essere derivato offline. Un collector comune
registra runtime, procfs Linux, build info e metadata GPU opt-in senza lanciare
probe esterni. Reporting e hardware profiles sono descritti in
`benchmark-reporting.md` e ADR-0021.

---

## Gestor

Gestor rappresenta il registro centrale delle capability.

Responsabilità:

- indicizzazione delle dichiarazioni di capability;
- discovery atomica dalle sorgenti autorevoli;
- distinzione tra dichiarazione e disponibilità operativa;
- query, filtri e preferenze target esatte;
- risoluzioni spiegabili con dependency plan topologico;
- eventi redatti di refresh e resolution.

Gestor non esegue codice.

Il Runtime Registry continua a possedere i componenti, il Provider Runtime i
provider e il Runtime Core l'unico dependency graph. Gestor conserva soltanto
descriptor e indici immutabili e consulta il grafo in lettura. Le registrazioni
invalidano lo snapshot senza avviare discovery o I/O; il refresh resta
esplicito e all-or-nothing.

Il composition root registra le sorgenti built-in component, provider, agent e tool e
pubblica uno snapshot iniziale current. `Resolve` non esegue capability o probe
e non sceglie un vincitore tramite ordine lessicografico. Gli eventi sul bus
condiviso espongono soltanto metadata redatti e vengono emessi senza lock.

Il design e i contratti della Milestone 4 sono descritti in
`gestor-design.md`; i gate conclusivi sono in
`reports/milestone-4-final.md`.

Le sorgenti agent/tool leggono soltanto descriptor e non invocano plan,
provider, Prepare o Execute. Le nuove registrazioni invalidano lo snapshot;
discovery e refresh restano espliciti.

---

## Agent e Tool System

Il Tool System applica il percorso obbligatorio
`Prepare -> authorize -> Execute`, con policy default-deny, permit privato
one-shot e result bounded. I reference tool workspace operano su path logici,
precondizioni content-addressed e root-confined.

L'Agent System coordina un run per sessione, valida piani e transizioni,
costruisce il contesto, autorizza invocation/disclosure del modello e gestisce
completion o streaming tool calling entro hard ceiling. Una mutazione marca il
contesto stale e richiede reindex/rebuild prima di una nuova generazione.

Gestor descrive agenti e tool senza eseguirli. Eventi agent e tool condividono
l'Event Bus e usano payload allowlist privi di prompt, contenuti, arguments,
output, resource, workspace, provider, model e policy.

---

## Plugin System

Responsabilità:

- registrazione plugin;
- lifecycle plugin;
- risoluzione dei plugin registrati;
- catalogo e caricamento dei loader;
- validazione della compatibilità;
- pubblicazione degli eventi plugin;
- integrazione con il dependency graph globale.

I plugin sono componenti Go fidati caricati in-process. Un catalogo di loader
fornisce discovery deterministica; il Plugin Runtime valida ID e manifest, li
indicizza e pubblica gli eventi, mentre il Runtime Core orchestra il lifecycle.
Distribuzione di artefatti esterni e isolamento sono livelli successivi.

La Milestone 5 consolida questa baseline attraverso audit dei contratti,
hardening concorrente, integrazione Gestor e Laravel come reference plugin. Il
design e i gate operativi sono descritti in `plugin-system-design.md` e
`plugin-system-development-plan.md`; ADR-0023 stabilizza ownership e semantica
pre-start.

Le cinque fasi e il gate finale sono completati; l'esito è disponibile in
`reports/milestone-5-final.md`.

Esempi:

- Laravel (implementato)
- Symfony
- Django
- React

---

## Context Engine

Responsabilità:

- descrizione framework-neutral dei workspace;
- indicizzazione sicura e limitata;
- snapshot immutabili e generazionali;
- analisi strutturata tramite analyzer sostituibili;
- retrieval lessicale, strutturale e semantico opt-in;
- costruzione di bundle con provenance e budget espliciti;
- cache derivata content-addressed e limitata.

Il Context Engine è un servizio composto dal Runtime e resta distinto sia da
`context.Context` sia dal `runtime.Context` consegnato ai componenti. Il
package pubblico è `pkg/contextengine`; l'implementazione concreta rimane in
`internal/contextengine` ed è esposta dal composition root tramite
`Runtime.ContextEngine`.

La pipeline separa source, indice, analyzer, retrieval, builder e cache. Lo
snapshot del workspace è autorevole; la cache conserva soltanto artefatti
derivati e una cache miss non modifica il risultato funzionale. Analyzer di
linguaggio e framework rimangono sostituibili e non entrano nel Runtime Core.

Gli embedding sono un'estensione opt-in del retrieval e passano attraverso il
Provider Runtime esistente con provider e modello espliciti. Gestor descrive le
capability ma non esegue indexing o analyzer. Il plugin Laravel `0.3.1`
fornisce workspace tramite il contratto framework-neutral `WorkspaceProvider`
senza trasferire conoscenza Laravel al Context Engine.

Ogni risultato conserva origine e metodo di selezione; ogni bundle dichiara
budget ed estimator. Eventi e log espongono soltanto contatori e codici redatti,
mai query, testo, embedding o path assoluti.

Il design e le sei fasi della Milestone 6 sono descritti in
`context-engine-design.md` e `context-engine-development-plan.md`.

---

## Tool System

Responsabilità:

- descrizione e registrazione dei tool;
- preparazione e validazione delle invocazioni;
- dichiarazione delle azioni e degli effetti;
- autorizzazione tramite policy e approval flow;
- esecuzione cancellabile e limitata;
- produzione di risultati correlati e bounded.

Ogni tool implementa una capability e resta separato dall'agente che decide di
richiederlo. Il modello propone arguments non fidati; il Tool Runtime li
prepara, normalizza le risorse, richiede una decisione e soltanto dopo esegue
l'effetto.

Il permission model iniziale è default-deny e distingue almeno ispezione e
mutazione del workspace, esecuzione di processi, rete, invocazione del modello e
disclosure di contenuto. È un confine operativo per codice trusted in-process,
non una sandbox per estensioni malevole.

La Fase 1 della Milestone 7 definisce `pkg/tool`: `PreparedInvocation` lega
identità, run, arguments e action tramite fingerprint; `Runtime.Invoke`
incorpora autorizzazione ed esecuzione e non accetta una `Decision` pubblica
come autorità. Policy registry e permit operativo appartengono al Tool Runtime.

Le implementazioni previste includono filesystem, Git, terminale, Docker,
Composer, Artisan e PHPUnit. La baseline della Milestone 7 consegna soltanto il
set workspace minimo necessario a validare il primo agente; il catalogo resta
estensibile senza modificare il Runtime Core.

---

## Agent System

Responsabilità:

- pianificazione;
- esecuzione task;
- utilizzo dei tool;
- coordinamento delle autorizzazioni;
- memoria bounded della sessione;
- gestione di budget, cancellazione e terminali;
- integrazione workspace-aware con il Context Engine.

Gli agenti non comunicano direttamente con i provider.

Utilizzano l'Agent Runtime composto da Maestro, che coordina il Provider Runtime
condiviso, il Context Engine, il Tool Runtime e Gestor. Provider, modello,
workspace, agente e limiti sono input espliciti del run; nessun sottosistema
applica selezione automatica o fallback nascosto.

Il Tool System e l'Agent System hanno package e ownership separati. Il primo
possiede catalogo ed execution boundary; il secondo possiede sessione, piano e
loop modello–tool. Una sessione è in-memory, immutabile dal punto di vista del
chiamante e limitata da hard ceiling indipendenti dalle istruzioni del modello.

La Fase 1 definisce `pkg/agent` con request a target espliciti, piani aciclici,
snapshot generazionali e precedenza deterministica delle cause terminali. Le
implementazioni concrete restano nelle fasi successive.

Il design e le sette fasi della Milestone 7 sono descritti in
`agent-system-design.md` e `agent-system-development-plan.md`.

---

# Flusso di una richiesta

1. La CLI riceve una richiesta.
2. Il Runtime crea il contesto.
3. Gestor individua le capability necessarie.
4. Il Context Engine prepara il workspace.
5. Il Provider comunica con il modello.
6. L'Agent propone e coordina gli strumenti.
7. Il Tool Runtime prepara, autorizza ed esegue gli effetti consentiti.
8. Il Runtime restituisce il risultato e lo stato terminale.

---

# Dipendenze

Le dipendenze devono sempre puntare verso il centro dell'architettura.

Mai il contrario.

Il Runtime conosce tutti.

I componenti non devono conoscersi direttamente.

---

# Estensibilità

Ogni componente deve poter essere sostituito senza modificare il Runtime.

Nuovi provider.

Nuovi plugin.

Nuovi tool.

Nuovi agenti.

Devono poter essere aggiunti come implementazioni.

---

# Decisioni

- Runtime minimale.
- Provider completamente astratti.
- Gestor come registry delle capability.
- Plugin indipendenti.
- Context Engine separato dal Provider.
- Tool completamente modulari.

---

# Documenti dipendenti

- ADR
- Specifiche tecniche
- Implementazioni

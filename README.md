# Maestro

> The intelligence is in the orchestration.

Maestro è un runtime locale, modulare e provider-agnostic progettato per orchestrare agenti AI dedicati allo sviluppo software.

L'obiettivo del progetto non è costruire un nuovo modello AI, ma fornire l'infrastruttura che permette a modelli, provider, strumenti e framework di collaborare in modo coerente.

---

## Perché Maestro?

Installare un modello locale è relativamente semplice.

Costruire un ambiente di sviluppo realmente intelligente è molto più complesso.

Maestro nasce per fornire le fondamenta di questo ecosistema.

---

## Caratteristiche

- Runtime modulare
- Provider-agnostic
- Plugin-first
- Context Engine
- Capability Registry (Gestor)
- Tool System
- Agent System
- Framework-aware

---

## Obiettivi

- Supportare provider multipli (Ollama, llama.cpp, OpenAI, Anthropic, ...)
- Essere indipendente dal framework
- Adattarsi all'hardware disponibile
- Rendere l'AI locale realmente utilizzabile nello sviluppo software

---

## Documentazione

Prima di contribuire al progetto è consigliata la lettura dei documenti nell'ordine seguente:

1. `identity.md`
2. `philosophy.md`
3. `principles.md`
4. `vision.md`
5. `architecture.md`
6. `roadmap.md`
7. `gestor-design.md`
8. `gestor-development-plan.md`
9. `design-decisions.md`
10. `adr/`
11. `provider-runtime.md`
12. `ollama-provider.md`
13. `llamacpp-provider.md`
14. `provider-model-lifecycle.md`
15. `provider-model-acquisition.md`
16. `provider-model-residency.md`
17. `provider-capability-introspection.md`
18. `provider-error-semantics.md`
19. `provider-resilience.md`
20. `provider-observability.md`
21. `provider-advanced-generation.md`
22. `provider-layer-plan.md`
23. `provider-api-compatibility-audit.md`
24. `provider-smoke-benchmark-manifest.yaml`
25. `benchmark-evaluation-plan.md`
26. `benchmark-runtime.md`
27. `smoke-benchmark.md`
28. `plugin-runtime.md`
29. `laravel-plugin.md`
30. `plugin-system-design.md`
31. `plugin-system-development-plan.md`
32. `plugin-api-compatibility-audit.md`
33. `context-engine-design.md`
34. `context-engine-development-plan.md`
35. `context-engine-api-compatibility-audit.md`

---

## Stato del progetto

✅ Provider Layer

Il Runtime Core, il lifecycle e l'Event System sono implementati. Il primo
incremento del Provider Runtime introduce registry, routing capability-based,
streaming ed embedding provider-agnostic.

Il primo adapter concreto per Ollama implementa completion, streaming,
embedding e model listing. La Fase 5 del Runtime Core è conclusa; lo smoke test
contro un'istanza Ollama reale confluisce nello Smoke Benchmark della Milestone
3.
Evoluzioni provider ulteriori potranno proseguire senza bloccare il Plugin
Runtime.

La Fase 1 della Provider Layer aggiunge l'adapter llama.cpp sulle API
OpenAI-compatible di `llama-server`, con completion, streaming SSE, embedding,
model listing e autenticazione Bearer opzionale. Implementazione e test isolati
sono completati; lo smoke test live confluisce nello Smoke Benchmark della
Milestone 3.

La Fase 2 aggiunge discovery avanzata e lifecycle dei modelli attraverso
capability opzionali. Ollama e llama.cpp espongono ora snapshot di stato, load e
unload attraverso lo stesso Provider Runtime, mantenendo i dettagli di
protocollo nei rispettivi adapter.

La Milestone 2 è stata completata nelle Fasi 3–10: acquisizione dei modelli,
policy di residenza, capability introspection, semantica degli errori,
resilienza, osservabilità, contratti avanzati di generazione e hardening finale.
La Fase 10 ha chiuso le verifiche deterministiche e l'handoff degli scenari live.
Il piano e i gate di ogni incremento sono descritti in
`provider-layer-plan.md`.

La Fase 3 è completata: `ModelPuller` e `ModelRemover` aggiungono pull con
progresso, cancellazione e rimozione attraverso il Provider Runtime. Ollama usa
le API native `/api/pull` e `/api/delete`; il router llama.cpp usa gli endpoint
`/models` e lo stream globale `/models/sse`. Nessun adapter accede direttamente
ai file della cache.

La Fase 4 è completata: `ModelResidencyPolicy` abilita autoload opt-in e rilascio
immediato, a TTL o allo shutdown. Il Provider Runtime coordina lease concorrenti
senza duplicare lo stato remoto e scarica soltanto i modelli caricati dalla
policy. Il comportamento senza policy rimane invariato.

La Fase 5 è completata: `CapabilityInspector` produce snapshot ordinati per
adapter, istanza o modello, separando supporto strutturale e disponibilità
operativa. Ollama e llama.cpp interrogano soltanto metadata ufficiali e non
mantengono cache; `unknown` rappresenta configurazioni non osservabili senza
indurre routing o selezione automatica.

La Fase 6 è completata: `ProviderError` uniforma kind, operazione, identità,
status e ritentabilità conservativa mantenendo `errors.Is`/`errors.As`, cause,
context ed EOF idiomatici. Gli adapter classificano anche errori di trasporto e
mid-stream; retry e circuit breaker sono applicati separatamente dalla Fase 7.

La Fase 7 è completata: `ResiliencePolicy` abilita retry, backoff, jitter,
budget temporale e circuit breaker closed/open/half-open per provider,
operazione e modello. Le policy sono opt-in, usano gli errori tipizzati e non
riaprono stream dopo il primo chunk; pull e remove non vengono ritentati.

La Fase 8 è completata: `ProviderObserver` espone eventi redatti e correlati per
inizio, tentativi, retry, transizioni del circuito e completamento. Gli stream
emettono un solo terminale anche in caso di cancellazione o chiusura anticipata;
il core resta indipendente dagli SDK di logging, metriche e tracing.

La Fase 9 è completata: sampling comune, output JSON/JSON Schema, tool calling e
delta tool streaming fanno parte del contratto neutrale. Ollama e llama.cpp
traducono la baseline sui rispettivi protocolli, con validazione locale e
capability introspection per le differenze operative.

La Fase 10 chiude la Milestone 2: suite deterministica, race detector, vet,
audit di compatibilità e documentazione sono allineati. Il manifest degli
scenari live è consegnato allo Smoke Benchmark della Milestone 3, insieme ai
requisiti di cleanup, redazione e configurazione.

La nuova Milestone 3 introduce il Benchmark & Evaluation Layer. Gli smoke test
live diventano il primo di tre livelli, seguito da benchmark del runtime e da
task reali di sviluppo. I risultati descrivono la combinazione completa di
hardware, provider, modello e plugin attraverso report JSON e Markdown, senza
produrre classifiche assolute tra modelli.

La Fase 1 della Milestone 3 è completata: contratti e report versionati, runner
deterministico, parsing strict del manifest, redazione JSON e base del comando
`maestro bench` sono implementati. Gli scenari live iniziano con la Fase 2.

La Fase 2 è completata: `maestro bench smoke` esegue la matrice live di
quattordici scenari per Ollama o llama.cpp con capability preflight, modelli per
ruolo, mutation guard, cleanup e report JSON atomico. In assenza di provider
configurato produce risultati `skipped` senza I/O implicito.

La Fase 3 è completata: `maestro bench provider` misura introspection, catalogo,
retry e circuit breaker; `maestro bench model` misura completion, TTFT,
throughput, cancellazione, embedding, lifecycle e cold/warm. Su Linux CPU e RAM
sono raccolte con scope di processo esplicito; metriche non osservabili, inclusa
la VRAM, restano assenti.

La Fase 4 è completata: `maestro bench laravel` esegue cinque task generativi e
un retrieval embedding sul dataset embedded `maestro-laravel-mini@1.0.0`,
avviando realmente il plugin Laravel. Il report `1.2.0` separa successo tecnico
e rubrica trasparente 0–3, senza conservare prompt o risposte.

La Fase 5 è completata: tutti i benchmark possono produrre Markdown derivato
dal JSON canonico, `maestro bench render` rigenera la vista offline e il profilo
hardware comune registra CPU/RAM Linux e metadata GPU opt-in. La Milestone 3
resta comunque in corso e non è considerata chiusa.

La prima validazione live Ollama ha superato gli integration test ma non il
gate Smoke completo: due scenari tool calling sono falliti con
`qwen2.5-coder:7b`. La ripetizione diretta su `/api/chat` a temperatura 0 ha
confermato che Ollama non restituisce `message.tool_calls`, ma JSON testuale,
sia non-stream sia stream; l'adapter Maestro non perde quindi il campo per
questa fixture. La milestone resta aperta; il dettaglio è nel report
`docs/reports/milestone-3-live-ollama-validation.md`.

La fixture alternativa `llama3.1:8b` produce `message.tool_calls` native
direttamente su `/api/chat`. L'adapter normalizza in modo conservativo il
terminale streaming Ollama `stop` in `tool_calls` quando lo stesso stream ha già
emesso una tool call e applica la regola coerente alle completion non-stream. Il
nuovo Smoke con `embeddinggemma:latest` raggiunge 13 passed, 1 skipped e 0
failed: il gate live Ollama è verde, mentre la Milestone 3 resta aperta fino a
una decisione esplicita. `qwen2.5-coder:7b` resta documentato come caso negativo
e `llama3.1:8b` come fixture positiva.

La Fase 6 del Plugin Runtime è completata: contratti pubblici, manifest di
compatibilità, registry e catalogo loader thread-safe, discovery, caricamento
cancellabile, eventi e lifecycle integrato nel Runtime Core. Il primo plugin
Laravel implementa detection del workspace e health. Packaging esterno,
sandbox e unload appartengono all'evoluzione successiva dell'ecosistema.

La Milestone 4 — Gestor è completata. Il Runtime pubblico compone un Registry
di capability, le sorgenti Runtime e Provider e un Resolver che usa il
dependency graph autorevole soltanto in lettura. Snapshot e availability sono
espliciti; più candidati senza preferenza producono ambiguity e nessun ordine
lessicografico diventa ranking. Refresh e resolution pubblicano eventi redatti
sull'Event Bus condiviso.

La Milestone 5 — Plugin System è completata. Catalogo e caricamento sono
coperti sotto concorrenza; graph, stato e lifecycle restano globali; Laravel
`0.2.0` dichiara `plugin.workspace-detection`; gli eventi sono sincroni,
best-effort e isolati da errori o panic degli observer. Audit API, suite
completa, race detector, vet e gate finale sono verdi. Packaging esterno,
sandbox e plugin di terze parti restano fuori scope.

La Milestone 6 — Context Engine è iniziata con il design architetturale e un
piano in sei fasi: contratti; workspace indexing; analisi strutturata e AST;
retrieval e Context Builder budgetato; cache incrementale; integrazione e gate
finale. Il servizio resterà provider-agnostic, userà embedding soltanto in modo
opt-in e non anticiperà memoria agente, tool execution o permission model.

La Fase 1 è completata: `pkg/contextengine` definisce workspace, documenti,
analisi, snapshot, retrieval e bundle immutabili. ADR-0024 assegna ownership e
stabilisce provenance, budget e riservatezza senza modificare Runtime, Gestor,
Provider o Plugin. La Fase 2 può ora implementare l'indice filesystem.

Uso essenziale:

```go
runtime := maestro.New()

// Registrare componenti, plugin e provider, quindi costruire il grafo.
if err := runtime.Start(ctx); err != nil {
    return err
}
defer runtime.Stop(context.Background())

if err := runtime.Gestor().Refresh(ctx); err != nil {
    return err
}

query, err := gestor.NewQuery("plugin.workspace_detection", gestor.QueryOptions{})
if err != nil {
    return err
}
resolution, err := runtime.Gestor().Resolve(query)
if err != nil {
    return err
}

// La risoluzione descrive target e dipendenze; non esegue la capability.
_ = resolution.Descriptor()
_ = resolution.Dependencies()
```

---

## Contribuire

Ogni modifica significativa dovrebbe seguire questo processo:

1. Analisi del problema.
2. Aggiornamento della documentazione.
3. Discussione della soluzione.
4. Implementazione.
5. Test.
6. Commit.

---

## Licenza

Da definire.

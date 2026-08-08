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
7. `design-decisions.md`
8. `adr/`
9. `provider-runtime.md`
10. `ollama-provider.md`
11. `llamacpp-provider.md`
12. `provider-model-lifecycle.md`
13. `provider-model-acquisition.md`
14. `provider-model-residency.md`
15. `provider-capability-introspection.md`
16. `provider-layer-plan.md`
17. `benchmark-evaluation-plan.md`
18. `plugin-runtime.md`
19. `laravel-plugin.md`

---

## Stato del progetto

🚧 Provider Layer

Il Runtime Core, il lifecycle e l'Event System sono implementati. Il primo
incremento del Provider Runtime introduce registry, routing capability-based,
streaming ed embedding provider-agnostic.

Il primo adapter concreto per Ollama implementa completion, streaming,
embedding e model listing. La Fase 5 del Runtime Core è conclusa; lo smoke test
contro un'istanza Ollama reale confluisce nello Smoke Benchmark della Milestone
3.
Provider e policy ulteriori proseguono senza bloccare il Plugin Runtime.

La Fase 1 della Provider Layer aggiunge l'adapter llama.cpp sulle API
OpenAI-compatible di `llama-server`, con completion, streaming SSE, embedding,
model listing e autenticazione Bearer opzionale. Implementazione e test isolati
sono completati; lo smoke test live confluisce nello Smoke Benchmark della
Milestone 3.

La Fase 2 aggiunge discovery avanzata e lifecycle dei modelli attraverso
capability opzionali. Ollama e llama.cpp espongono ora snapshot di stato, load e
unload attraverso lo stesso Provider Runtime, mantenendo i dettagli di
protocollo nei rispettivi adapter.

Il completamento della Milestone 2 è suddiviso nelle Fasi 3–10: acquisizione dei
modelli, policy di residenza, capability introspection, semantica degli errori,
resilienza, osservabilità, contratti avanzati di generazione e hardening finale.
La Fase 10 chiude con verifiche deterministiche e handoff degli scenari live.
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

La nuova Milestone 3 introduce il Benchmark & Evaluation Layer. Gli smoke test
live diventano il primo di tre livelli, seguito da benchmark del runtime e da
task reali di sviluppo. I risultati descrivono la combinazione completa di
hardware, provider, modello e plugin attraverso report JSON e Markdown, senza
produrre classifiche assolute tra modelli.

La Fase 6 del Plugin Runtime è completata: contratti pubblici, manifest di
compatibilità, registry e catalogo loader thread-safe, discovery, caricamento
cancellabile, eventi e lifecycle integrato nel Runtime Core. Il primo plugin
Laravel implementa detection del workspace e health. Packaging esterno,
sandbox e unload appartengono all'evoluzione successiva dell'ecosistema.

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

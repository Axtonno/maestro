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
13. `plugin-runtime.md`
14. `laravel-plugin.md`

---

## Stato del progetto

🚧 Provider Layer

Il Runtime Core, il lifecycle e l'Event System sono implementati. Il primo
incremento del Provider Runtime introduce registry, routing capability-based,
streaming ed embedding provider-agnostic.

Il primo adapter concreto per Ollama implementa completion, streaming,
embedding e model listing. La Fase 5 del Runtime Core è conclusa; lo smoke test
contro un'istanza Ollama reale è consolidato nel gate finale della Milestone 2.
Provider e policy ulteriori proseguono senza bloccare il Plugin Runtime.

La Fase 1 della Provider Layer aggiunge l'adapter llama.cpp sulle API
OpenAI-compatible di `llama-server`, con completion, streaming SSE, embedding,
model listing e autenticazione Bearer opzionale. Implementazione e test isolati
sono completati; lo smoke test live è parte del gate finale della Milestone 2.

La Fase 2 aggiunge discovery avanzata e lifecycle dei modelli attraverso
capability opzionali. Ollama e llama.cpp espongono ora snapshot di stato, load e
unload attraverso lo stesso Provider Runtime, mantenendo i dettagli di
protocollo nei rispettivi adapter.

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

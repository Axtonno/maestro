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

---

## Stato del progetto

🚧 Runtime Core / Provider Runtime / Plugin Runtime

Il Runtime Core, il lifecycle e l'Event System sono implementati. Il primo
incremento del Provider Runtime introduce registry, routing capability-based,
streaming ed embedding provider-agnostic.

Il primo adapter concreto per Ollama implementa completion, streaming,
embedding e model listing. La chiusura della Fase 5 richiede soltanto lo smoke
test contro un'istanza Ollama reale; provider e policy ulteriori proseguono
separatamente nella Milestone 2 senza bloccare il Plugin Runtime.

Il primo incremento del Plugin Runtime aggiunge contratti pubblici, registry
thread-safe e registrazione integrata nel Runtime Core. I plugin in-process
riusano dependency graph, stato e lifecycle dei componenti. Discovery,
installazione e caricamento di artefatti esterni appartengono agli incrementi
successivi.

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

# Milestone 2 — Report retrospettivo Fase 2

Fase: Model Discovery & Lifecycle

Stato: Completata

Data di completamento: 2026-08-06

Natura del documento: ricostruzione retrospettiva

---

# Obiettivo

Rendere osservabili catalogo e stato dei modelli e introdurre capability
indipendenti di load e unload senza duplicare nel Runtime lo stato remoto.

---

# Risultati consegnati

- Tipi neutrali `ModelInfo` e `ModelState`.
- Capability pubbliche `ModelDiscoverer`, `ModelLoader` e `ModelUnloader`.
- Routing delle capability nel Provider Runtime.
- Discovery Ollama tramite `/api/tags` e `/api/ps`.
- Load e unload Ollama tramite richieste con `keep_alive`.
- Discovery e lifecycle llama.cpp in router mode.
- Test isolati per routing, stato, cancellazione e protocolli adapter.
- ADR-0009 e guida `provider-model-lifecycle.md`.

---

# Decisioni principali

- Discovery è la fonte osservabile dello stato effettivo del provider.
- Load e unload sono capability opzionali e indipendenti.
- Il Provider Runtime instrada le operazioni ma non possiede stato modello.
- La modalità single-model di llama.cpp non viene presentata come router.

---

# Evidenze storiche

Contratti, routing, implementazioni e documentazione furono consegnati nel
commit `d2d2cdd`; `9156189` consolidò il piano delle fasi successive.

---

# Handoff alla Fase 3

Il catalogo può essere completato con acquisizione, avanzamento e rimozione,
sempre attraverso API provider e senza accesso diretto al filesystem.

# Milestone 2 — Report retrospettivo Fase 4

Fase: Model Residency Policies

Stato: Completata

Data di completamento: 2026-08-08

Natura del documento: ricostruzione retrospettiva

---

# Obiettivo

Coordinare autoload e durata di residenza dei modelli tramite policy opt-in,
senza assumere ownership di modelli caricati esternamente.

---

# Risultati consegnati

- `ModelResidencyPolicy` per provider e model ID esatto.
- Autoload opt-in prima di completion, stream ed embedding.
- Rilascio immediato, a TTL o allo shutdown.
- Lease mantenute fino al termine degli stream.
- Coalescing delle transizioni concorrenti sullo stesso modello.
- Ownership limitata alle residenze caricate da Maestro.
- Ollama tramite `keep_alive`; llama.cpp tramite load/unload del router.
- Clock sostituibile e test deterministici di timer, concorrenza e shutdown.
- ADR-0011 e guida `provider-model-residency.md`.

---

# Decisioni principali

- Senza policy il comportamento precedente resta invariato.
- Discovery decide se un load sia necessario.
- Maestro non scarica modelli caricati da attori esterni.
- Le policy non effettuano selezione hardware-aware del modello.

---

# Evidenze storiche

La fase fu consegnata nel commit `6370347`.

---

# Handoff alla Fase 5

La superficie operativa completa può ora essere descritta tramite introspection
distinguendo supporto dell'adapter e disponibilità dell'istanza o del modello.

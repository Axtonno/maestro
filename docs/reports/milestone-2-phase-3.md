# Milestone 2 — Report retrospettivo Fase 3

Fase: Model Acquisition & Removal

Stato: Completata

Data di completamento: 2026-08-06

Natura del documento: ricostruzione retrospettiva

---

# Obiettivo

Completare il ciclo del catalogo con pull, progresso e rimozione cancellabili,
preservando la proprietà dei modelli nei server provider.

---

# Risultati consegnati

- Capability `ModelPuller` e `ModelRemover`.
- `ModelPullStream` pull-based con stage neutrali e chiusura esplicita.
- Routing di pull e remove nel Provider Runtime.
- Ollama: `/api/pull` e `/api/delete`.
- llama.cpp router: avvio tramite `/models`, progresso `/models/sse`,
  cancellazione remota `/models/unload` e rimozione dalla cache via DELETE.
- Propagazione di context e deadline.
- Validazione di progresso, terminale, idempotenza e cleanup.
- Test HTTP in-memory per entrambi gli adapter.
- ADR-0010 e guida `provider-model-acquisition.md`.

---

# Decisioni principali

- Maestro non cancella direttamente file scelti dall'utente o dal provider.
- Endpoint assenti producono un errore della capability, non un fallback sul
  filesystem.
- Gli stream di progresso sono risorse esplicite e devono essere chiusi.
- Gli scenari live mutativi richiedono fixture e cleanup controllati nella
  Milestone 3.

---

# Evidenze storiche

La fase fu consegnata nel commit `ca36e3b`.

---

# Handoff alla Fase 4

Discovery, load/unload e acquisizione permettono di introdurre policy di
residenza senza trasformare Maestro in una seconda fonte di verità.

# Milestone 2 — Report retrospettivo Fase 5

Fase: Capability Introspection

Stato: Completata

Data di completamento: 2026-08-08

Natura del documento: ricostruzione retrospettiva

---

# Obiettivo

Rendere interrogabili capability e limiti di adapter, istanza e modello senza
introdurre selezione automatica o cache di stato remoto.

---

# Risultati consegnati

- `CapabilityInspector` e routing `Capabilities`.
- Target espliciti adapter, instance e model.
- Descrittori neutrali per tutte le capability provider note.
- Separazione tra supporto strutturale e availability operativa.
- Stato `unknown` quando il protocollo non consente inferenze affidabili.
- Report ordinati canonicamente e validati dal Provider Runtime.
- Ollama tramite catalogo e `/api/show`.
- llama.cpp tramite `/models`, modalità router e argomenti del processo.
- Test di assenza I/O, variazione del catalogo e concorrenza.
- ADR-0012 e guida `provider-capability-introspection.md`.

---

# Decisioni principali

- Nessun report viene memorizzato dal Provider Runtime.
- Unsupported e unavailable sono condizioni distinte.
- Report incoerenti falliscono prima dell'I/O operativo ulteriore.
- L'introspection informa i consumer ma non sceglie provider o modello.

---

# Evidenze storiche

La fase fu consegnata nel commit `efbd25c`.

---

# Handoff alla Fase 6

Capability e operazioni canoniche forniscono la base per una tassonomia di
errori provider-neutral utilizzabile dalle future policy di resilienza.

# Milestone 1 — Report retrospettivo Fase 4

Fase: Event System

Stato: Completata

Data di completamento: 2026-08-05

Natura del documento: ricostruzione retrospettiva

---

# Obiettivo

Fornire al Runtime Core un meccanismo in-process minimale per osservare eventi
senza introdurre code, broker o callback eseguiti sotto lock.

---

# Risultati consegnati

- Event Bus pubblico e implementazione interna thread-safe.
- Topic esatti e dispatch sincrono.
- Fan-out nell'ordine di sottoscrizione.
- Snapshot dei subscriber prima dell'invocazione.
- Callback eseguiti senza lock interni.
- Rimozione idempotente degli handler di un topic.
- Validazione di topic, eventi e handler.
- Suite per ordine, concorrenza, re-entrancy e unsubscribe.
- ADR-0005 e documentazione in `docs/event-system.md`.

---

# Decisioni principali

- Il primo Event Bus è sincrono e in-process.
- Publish non introduce goroutine implicite né garanzie di persistenza.
- Gli handler possono rientrare nel bus perché nessun callback viene eseguito
  sotto lock.
- L'ordine è deterministico ma non rappresenta priorità.

---

# Evidenze storiche

Implementazione, test, ADR-0005 e documentazione furono consegnati nel commit
`9759c3d`. Il contesto di progetto registra la fase come completata.

---

# Handoff alla Fase 5

Runtime, configurazione ed Event Bus possono ora ospitare un Provider Runtime
separato e il primo adapter concreto senza accoppiare il core a un protocollo
AI specifico.

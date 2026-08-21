# Milestone 2 — Report retrospettivo Fase 6

Fase: Error Semantics

Stato: Completata

Data di completamento: 2026-08-08

Natura del documento: ricostruzione retrospettiva

---

# Obiettivo

Fornire una classificazione stabile degli errori indipendente da status e
payload specifici di Ollama o llama.cpp.

---

# Risultati consegnati

- Envelope pubblico `ProviderError`.
- Kind neutrali, operazione, provider, modello, status e ritentabilità.
- Cause preservate e compatibilità con `errors.Is` e `errors.As`.
- Preservazione di cancellazione, deadline ed EOF.
- Mapping comune di status HTTP, trasporto e risposte invalide.
- Mapping specifico dei payload Ollama e OpenAI-like di llama.cpp.
- Classificazione degli errori sincroni e mid-stream.
- Dettagli remoti normalizzati e limitati.
- ADR-0013 e matrice in `provider-error-semantics.md`.

---

# Decisioni principali

- Nessun consumer deve analizzare stringhe per decidere il comportamento.
- `Retryable` è metadata conservativo, non una decisione di retry.
- Prompt, risposta e contenuto della richiesta non entrano nell'envelope.
- Le API restano idiomatiche per i consumer Go.

---

# Evidenze storiche

La fase fu consegnata nel commit `9ec4ced`.

---

# Handoff alla Fase 7

La tassonomia consente di applicare retry e circuit breaker soltanto alle
operazioni dichiarate ripetibili e agli errori classificati come ritentabili.

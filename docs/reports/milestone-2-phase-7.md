# Milestone 2 — Report retrospettivo Fase 7

Fase: Resilience Policies

Stato: Completata

Data di completamento: 2026-08-08

Natura del documento: ricostruzione retrospettiva

---

# Obiettivo

Introdurre resilienza opt-in, deterministica e consapevole dell'idempotenza,
senza modificare il comportamento di default del Provider Runtime.

---

# Risultati consegnati

- `ResiliencePolicy` per provider, operazione e modello opzionale.
- Retry finiti con backoff esponenziale saturato, jitter e budget temporale.
- Matrice esplicita di ripetibilità delle operazioni.
- Nessun retry di pull/remove.
- Retry degli stream soltanto prima del primo chunk consegnato.
- Circuit breaker scoped per provider, operazione e modello.
- Stati closed, open e half-open con probe concorrenti limitati.
- Snapshot tipizzati tramite `CircuitState`.
- Clock, attesa e jitter sostituibili nei test.
- ADR-0014 e guida `provider-resilience.md`.

---

# Decisioni principali

- Le policy sono disabilitate per default.
- Solo `ProviderError.Retryable` e la matrice di ripetibilità autorizzano retry.
- Context e budget hanno precedenza sulle attese.
- Nessun fallback multi-provider viene introdotto.

---

# Evidenze storiche

La fase fu consegnata nel commit `099e7ec`.

---

# Handoff alla Fase 8

Tentativi, retry e transizioni del circuito possono ora essere osservati con
eventi neutrali, correlati e privi di contenuto sensibile.

# Milestone 2 — Report retrospettivo Fase 8

Fase: Provider Observability

Stato: Completata

Data di completamento: 2026-08-08

Natura del documento: ricostruzione retrospettiva

---

# Obiettivo

Rendere osservabili i confini operativi del Provider Runtime senza dipendere da
SDK telemetrici e senza esporre prompt, risposte o credenziali.

---

# Risultati consegnati

- `ProviderObserver`, `ProviderObserverFunc` e `ProviderEvent` pubblici.
- Operation ID per correlare start, tentativi, retry, circuit transition e
  terminale.
- Copertura di completion, stream, embedding, catalogo, lifecycle,
  acquisizione, rimozione e introspection.
- Un solo evento terminale per EOF, errore, cancellazione e chiusura anticipata.
- Usage o progresso inclusi soltanto quando disponibili.
- Observer sostituibili in concorrenza e invocati senza lock interni.
- Errori e panic degli observer isolati dal risultato operativo.
- Fast path senza observer privo di tracker ed eventi.
- ADR-0015 e guida `provider-observability.md`.

---

# Decisioni principali

- Gli eventi sono redatti per costruzione.
- Prompt, chunk, embedding, payload remoti ed endpoint sensibili non vengono
  pubblicati.
- L'osservabilità non modifica il risultato dell'operazione.
- Adapter verso logging, metriche e tracing restano responsabilità applicative.

---

# Evidenze storiche

La fase fu consegnata nel commit `bde3e02`.

---

# Handoff alla Fase 9

Contratti, introspection, errori e osservabilità sono sufficienti per estendere
la generazione con output strutturati e tool calling in modo verificabile.

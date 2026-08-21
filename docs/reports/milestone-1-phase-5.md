# Milestone 1 — Report retrospettivo Fase 5

Fase: Provider Runtime/Configuration

Stato: Completata

Data di completamento: 2026-08-05

Natura del documento: ricostruzione retrospettiva

---

# Obiettivo

Dimostrare che il Runtime Core può integrare provider AI tramite contratti
capability-based, configurazione condivisa e routing indipendente dagli adapter.

---

# Risultati consegnati

- Contratti pubblici provider separati dalle capability operative.
- Capability iniziali per completion, streaming, embedding e model listing.
- Provider Runtime thread-safe con registrazione, risoluzione, provider
  predefinito e routing.
- Stream pull-based con chiusura esplicita.
- Configurazione minimale condivisa nel runtime context.
- Composition root pubblico `maestro.New`.
- Primo adapter Ollama con chat, stream NDJSON, embedding e listing.
- Autenticazione e modello predefinito configurabili.
- Test HTTP in-memory e test live opzionali con build tag.
- ADR-0006 e guide `provider-runtime.md` e `ollama-provider.md`.

---

# Decisioni principali

- Il provider descrive l'identità; le operazioni restano capability opzionali.
- Il Runtime effettua routing esplicito e non seleziona automaticamente modelli
  o provider.
- Le chiamate agli adapter avvengono senza mantenere lock del registry.
- Lo smoke live Ollama viene trasferito alla futura Milestone 3 e non blocca il
  gate deterministico del core.

---

# Evidenze storiche

Il Provider Runtime e il composition root furono introdotti in `5c7f2cb`;
l'adapter Ollama e la relativa suite in `8c1c4bc`, con allineamento del gate in
`c7fd898`. Il contesto di chiusura della Milestone 1 registra la fase come
completata.

---

# Handoff alla Fase 6

La stessa separazione tra contratto pubblico e implementazione interna può
essere applicata ai plugin, riusando graph, stato e lifecycle globali.

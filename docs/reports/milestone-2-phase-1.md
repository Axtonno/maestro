# Milestone 2 — Report retrospettivo Fase 1

Fase: Adapter llama.cpp

Stato: Completata

Data di completamento: 2026-08-06

Natura del documento: ricostruzione retrospettiva

---

# Obiettivo

Validare l'indipendenza dei contratti provider con un secondo adapter locale,
basato sulla superficie OpenAI-compatible di `llama-server`.

---

# Risultati consegnati

- Facade pubblica `pkg/provider/llamacpp` e configurazione tipizzata.
- Client HTTP interno basato sulla libreria standard.
- Completion e streaming SSE tramite Chat Completions API.
- Embedding tramite API compatibile OpenAI.
- Listing del modello caricato.
- Autenticazione Bearer opzionale.
- Risoluzione del modello esplicito o predefinito.
- Validazione delle risposte, error handling con body limitato e propagazione
  del context.
- Test HTTP in-memory e scenario di integrazione live opzionale.
- Guida `docs/llamacpp-provider.md`.

---

# Decisioni principali

- Il processo `llama-server` resta esterno a Maestro.
- L'adapter usa protocolli HTTP documentati e non accede ai file dei modelli.
- Lifecycle, acquisizione e policy di resilienza sono incrementi separati.
- La verifica live confluisce nel Benchmark Layer e non blocca la fase.

---

# Evidenze storiche

Facade, adapter, test e documentazione furono consegnati nel commit `94a12c8`.

---

# Handoff alla Fase 2

Ollama e llama.cpp possono ora essere confrontati sullo stesso contratto base;
la fase successiva aggiunge discovery avanzata e lifecycle neutrale dei modelli.

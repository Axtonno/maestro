# Milestone 2 — Report retrospettivo Fase 9

Fase: Advanced Generation Baseline

Stato: Completata

Data di completamento: 2026-08-08

Natura del documento: ricostruzione retrospettiva

---

# Obiettivo

Stabilire la superficie comune necessaria al futuro Agent System per opzioni di
generazione, output strutturati e tool calling.

---

# Risultati consegnati

- Opzioni comuni per limite token, temperatura, `top_p` e stop sequence.
- Output JSON e JSON Schema.
- Definizioni tool, tool choice, tool call e tool result nei messaggi neutrali.
- Delta incrementali di tool call negli stream con indici stabili.
- Validazione preflight delle combinazioni incompatibili.
- Validazione del JSON terminale e delle tool call ricostruite.
- Traduzione e test isolati per Ollama e llama.cpp.
- Capability introspection aggiornata con disponibilità e limiti.
- ADR-0016 e guida `provider-advanced-generation.md`.

---

# Decisioni principali

- Le opzioni esclusivamente proprietarie non entrano nel contratto comune.
- Ollama supporta tool choice `auto` e `none` nella baseline.
- llama.cpp può dichiarare `required` e named choice solo con configurazione
  template compatibile.
- Multimodalità, audio, reasoning e prompt caching restano fuori scope.

---

# Evidenze storiche

La fase fu consegnata nel commit `866e269`, insieme all'hardening e all'handoff
finale oggi separati come Fase 10 nella roadmap canonica.

---

# Handoff alla Fase 10

La superficie provider è completa; resta da congelare compatibilità, gate
deterministico e matrice live destinata alla Milestone 3.

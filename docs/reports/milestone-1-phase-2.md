# Milestone 1 — Report retrospettivo Fase 2

Fase: Dependency Container & Registry

Stato: Completata

Data di completamento: 2026-08-05

Natura del documento: ricostruzione retrospettiva

---

# Obiettivo

Costruire il contenitore interno dei componenti e un dependency graph validato,
deterministico e indipendente dal lifecycle.

---

# Risultati consegnati

- Registry thread-safe con registrazione, lookup e rifiuto dei duplicati.
- Node e graph interni con relazioni dipendenza/dipendente.
- Resolver dei metadata con supporto alle dipendenze opzionali.
- Validator per metadata, dipendenze e capability duplicate, self-dependency,
  dipendenze mancanti e cicli.
- Rilevamento dei cicli tramite depth-first search.
- Builder stateless che coordina resolver, validator e pubblicazione del grafo.
- Runtime interno come composition root di registry, builder e graph.
- Test unitari e di integrazione dei singoli invarianti.

---

# Decisioni principali

- Ogni tipo protegge gli invarianti del proprio livello.
- Resolver costruisce senza validare; validator valida senza modificare.
- Builder coordina l'operazione senza conservare stato.
- Nessuna struttura mutabile interna viene esposta direttamente.

---

# Evidenze storiche

L'incremento è concentrato nel commit `b75b65a`, che introdusse
`internal/runtime`, la relativa suite e `docs/runtime-internals.md`. Il contesto
di progetto registra la Fase 2 come completata.

---

# Handoff alla Fase 3

Il grafo validato e il registry diventano la base per stato, ordinamento
topologico, startup dependency-first e shutdown dependent-first.

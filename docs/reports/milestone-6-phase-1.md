# Milestone 6 — Phase 1 Report

Stato: Completata

Data: 2026-08-11

---

# Risultato

La Fase 1 introduce la baseline pubblica del Context Engine senza modificare
le API esistenti. ADR-0024 assegna ownership a ogni stadio della pipeline e
stabilisce immutabilità, provenance, budget espliciti e riservatezza.

---

# Contratti consegnati

Package:

```text
pkg/contextengine
```

Sono disponibili:

- identità per workspace, source, analyzer, estimator, path e digest;
- `ScanPolicy`, `Workspace` e `WorkspaceProvider`;
- `Document`, `SourceRange` e diagnostiche tipizzate;
- simboli, relazioni, chunk e `Analysis` versionata;
- `Snapshot` immutabile con metadata generazionali;
- query lessicali, strutturali e semantiche;
- risultati con metodo, score, intervallo e reason code;
- budget, estimator, sezioni e `ContextBundle`;
- contratti `Source`, `Analyzer`, `TokenEstimator` ed `Engine`;
- sentinel ispezionabili con `errors.Is`.

---

# Invarianti verificati

- root assoluta e normalizzata;
- path documento relativo, normalizzato e senza traversal;
- policy con limiti positivi e pattern validi;
- metadata senza newline e restituiti come copia;
- contenuto UTF-8 e digest SHA-256 coerente;
- intervalli byte contenuti nel documento;
- simboli e chunk unici, relazioni con sorgente valida;
- analysis legata a path e digest esatti;
- snapshot ordinati con documenti unici;
- query semantic con provider e modello espliciti;
- budget con allowance positiva;
- bundle incapace di superare l'evidence budget;
- slice e mappe pubbliche difensive.

---

# Ownership

| Livello | Possiede | Non possiede |
|---|---|---|
| Source | lettura e diagnostiche di scansione | snapshot pubblicato |
| Index | generazione e snapshot autorevole | analisi language-specific |
| Analyzer | evidenza su un documento | documento e lifecycle |
| Retrieval | candidati, score e motivazioni | completion |
| Builder | selezione, dedup e budget | provider selection implicita |
| Cache | artefatti derivati | stato autorevole |
| Runtime | composizione | logica workspace/linguaggio |

---

# Compatibilità

Nessuna modifica è stata applicata a Runtime, Gestor, Provider, Plugin, Laravel
o composition root. L'audit iniziale è disponibile in
`docs/context-engine-api-compatibility-audit.md`.

Il package usa soltanto la libreria standard. Non sono state introdotte
dipendenze esterne.

---

# Test

Coperti:

- assertion di compilazione delle interfacce;
- input workspace validi e invalidi;
- path assoluti, traversal, separator non portabili e NUL;
- copie difensive di policy e metadata;
- identità e digest documento;
- analisi strutturata e riferimenti invalidi;
- ordinamento e lookup snapshot;
- query semantic e copie difensive;
- bundle valido e superamento budget.

Comando:

```text
GOCACHE=/tmp/maestro-go-build go test ./pkg/contextengine
```

Esito: superato.

---

# Decisioni rinviate

- algoritmi concreti di scansione e media type: Fase 2;
- registry e analyzer AST reale: Fase 3;
- formule di ranking e strategia di build: Fase 4;
- eviction e chiavi cache: Fase 5;
- accessor del Runtime, Laravel, Gestor ed eventi: Fase 6.

---

# Gate

ADR-0024 è Accepted, la baseline pubblica è additiva, i contratti compilano e
la suite dedicata è verde. La Fase 1 è completata; la Fase 2 può iniziare.

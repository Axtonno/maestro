# Milestone 6 — Phase 4 Report

Stato: Completata

Data: 2026-08-11

---

# Risultato

Il Context Engine costruisce ranking spiegabili e context bundle immutabili
entro un budget dichiarato. Retrieval lessicale e strutturale funzionano
offline; semantic retrieval riusa il Provider Runtime in modo opt-in.

---

# Retrieval

Implementati:

- candidati documentali o basati su chunk AST;
- filtri path e linguaggio;
- term coverage lessicale versionata;
- match simbolo strutturale;
- cosine similarity con validazione completa;
- provider e modello embedding espliciti;
- Reciprocal Rank Fusion richiesta per query multi-metodo;
- tie-break deterministici;
- top-k limitato a 100;
- provenance tramite path, intervallo, metodo, score e reason code.

---

# Builder e budget

- registry estimator thread-safe;
- estimator built-in `context.utf8-estimator@1`;
- allowance separata da reserved e safety tokens;
- deduplicazione per sovrapposizione degli intervalli;
- troncamento su confini UTF-8;
- verifica testo/intervallo;
- costo per sezione e totale;
- snapshot singolo per l'intera build;
- failure estimator classificate e cancellabili.

Una stima non viene esposta come conteggio esatto. Il builder non effettua
completion o summarization.

---

# Provider Runtime

L'integrazione dipende soltanto dalla capability `Embed`. Il batch contiene la
query seguita dai candidati. Cardinalità, dimensioni, valori finiti e norme
sono validati prima del ranking.

Provider assente produce `ErrUnsupported`; errori operativi compongono
`ErrEmbeddingFailure`; context idiomatici restano preservati.

---

# Developer Benchmark

Lo scenario retrieval Laravel non calcola più il ranking direttamente nella
suite: costruisce documenti versionati, indicizza con una source in-memory e
chiama semantic retrieval del Context Engine. Rubrica, best relevant rank e
metriche restano compatibili.

---

# Test

Coperti:

- ranking lessicale e filtri;
- lookup strutturale e intervalli simbolo;
- provider/modello semantici;
- cardinalità, dimensioni, NaN, Inf e zero norm;
- provider assente;
- fusion deterministica;
- budget e troncamento multibyte;
- registrazione estimator nil, typed nil e duplicata;
- errori, panic e costi invalidi;
- estimator bloccante fuori lock;
- Developer Benchmark repository-real.

Comandi dedicati:

```text
GOCACHE=/tmp/maestro-go-build go test ./internal/contextengine ./pkg/contextengine
GOCACHE=/tmp/maestro-go-build go test ./internal/benchmark/developer
```

Esito: superato.

---

# Gate

Retrieval offline, semantic opt-in, provenance, fusione esplicita e budget sono
verificati. Il Developer Benchmark consuma il percorso reale. La Fase 4 è
completata; la Fase 5 può rendere incrementali gli artefatti senza cambiare il
risultato funzionale.

# Milestone 3 — Report finale Fase 4

Fase: Developer Benchmark

Stato: Completata

Data: 2026-08-09

---

# Obiettivo

Valutare l'utilità pratica di Maestro su task PHP/Laravel riproducibili,
separando il risultato tecnico dalla qualità e senza rendere obbligatorio un
evaluator LLM.

---

# Risultati consegnati

- Dataset embedded `maestro-laravel-mini@1.0.0` con fixture prive di dati
  sensibili.
- Materializzazione privata e cleanup per il lifecycle reale del plugin
  Laravel `0.1.0`.
- Cinque scenari generativi: controller, dipendenze, PHPUnit, refactor e sintesi.
- Retrieval su sette documenti tramite embedding provider e similarità cosine.
- Rubriche trasparenti 0–3 con tre criteri per task generativo.
- Score retrieval basato sul rank della prima fixture rilevante.
- Stato tecnico distinto dalla valutazione qualitativa.
- Report schema `1.2.0` con `evaluation`, evaluator, metodo, score e
  `rationale_code` redatto per costruzione.
- Comando `maestro bench laravel` con `--fail-on-failure` e
  `--minimum-score` opt-in.
- Manifest Developer dedicato, guida operativa e ADR-0020.

---

# Copertura deterministica

I test verificano:

- identità, cardinalità e riferimenti sicuri del dataset;
- materializzazione e rimozione del workspace temporaneo;
- tutti i sei scenari attraverso il Provider Runtime reale e un provider fixture;
- score 3/3 sui criteri attesi;
- score 0/3 senza trasformare il risultato tecnico in failure;
- ranking cosine e rigetto di vettori invalidi;
- assenza di prompt e contenuti fixture nel report;
- validazione stretta di evaluator, metodo e rationale code;
- plugin e dataset registrati nel profilo;
- CLI offline con sei scenari `skipped` e nessun I/O provider.

---

# Verifiche

```text
GOCACHE=/tmp/maestro-go-build go test ./...
GOCACHE=/tmp/maestro-go-build go test -race ./...
GOCACHE=/tmp/maestro-go-build go vet ./...
GOCACHE=/tmp/maestro-go-build go run ./cmd/maestro bench validate --manifest docs/developer-benchmark-manifest.yaml
GOCACHE=/tmp/maestro-go-build go run ./cmd/maestro bench laravel --provider ollama --warmup 0 --runs 1 --output /tmp/maestro-laravel-offline.json
```

Esito: suite completa, race detector, vet, validazione manifest e prova CLI
offline superati. Il report offline usa lo schema `1.2.0`, registra dataset e
plugin, contiene sei scenari `skipped` senza evaluation ed è protetto con
permessi `0600`.

---

# Verifica live

Nessun server Ollama o llama.cpp era configurato durante lo sviluppo. La run
live dipende dai modelli chat ed embedding scelti dall'utente e non è un
prerequisito del gate deterministico.

---

# Handoff alla Fase 5

La fase finale può derivare report Markdown dal JSON `1.2.0`, completare i
profili hardware e documentare un gate riproducibile dell'intera milestone. Il
renderer non deve accedere a prompt o risposte, che non fanno parte del report.

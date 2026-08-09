# Milestone 3 — Report finale Fase 1

Fase: Benchmark Contracts & Runner

Stato: Completata

Data: 2026-08-09

---

# Obiettivo

Consegnare i contratti versionati e il runner deterministico condiviso dai tre
livelli della Milestone 3, senza introdurre semantiche provider nel Runtime Core.

---

# Risultati consegnati

- Contratti pubblici in `pkg/benchmark` per manifest, scenari, campioni,
  misure, aggregati, profili ed errori classificati.
- Manifest schema `1` con parsing YAML strict e validazione dell'handoff della
  Provider Layer.
- Report JSON schema `1.0.0` pubblicato in
  `docs/schemas/benchmark-report-v1.schema.json`.
- Runner interno con warmup, run ripetute, timeout, panic recovery e cleanup
  garantito dopo ogni iterazione.
- Cleanup eseguito con timeout dedicato e contesto indipendente dalla
  cancellazione della run.
- Classificazione centrale di cancellazione, deadline e `ProviderError` senza
  conservare messaggi remoti.
- Aggregazione deterministica di minimo, mediana, massimo e p95 da almeno 20
  campioni misurati.
- Serializzazione JSON con validazione preventiva e redazione di endpoint,
  credenziali e path utente.
- Base CLI `maestro bench` e comando `maestro bench validate`.
- Test end-to-end deterministico dal manifest al report serializzato.
- ADR-0017 per ownership, versionamento e motivazione della dipendenza YAML.

---

# Decisioni principali

- I contratti stabili sono pubblici; runner, parser e writer restano interni.
- Gli scenari mancanti dal registry sono `skipped`, non fallimenti del runner.
- Le capability non disponibili useranno `unsupported` nelle fasi successive.
- Warmup e campioni misurati sono entrambi conservati, ma solo i secondi
  alimentano gli aggregati.
- Errori e reason code sono strutturati; non vengono serializzate stringhe di
  errore arbitrarie.
- Il JSON è la fonte raw; il Markdown della Fase 5 sarà sempre derivato dal JSON.

---

# Verifiche

Comandi del gate:

```text
GOCACHE=/tmp/maestro-go-build go test ./...
GOCACHE=/tmp/maestro-go-build go test -race ./...
GOCACHE=/tmp/maestro-go-build go vet ./...
GOCACHE=/tmp/maestro-go-build go run ./cmd/maestro bench validate --manifest docs/provider-smoke-benchmark-manifest.yaml
```

Esito: tutte le verifiche superate.

---

# Handoff alla Fase 2

La Fase 2 può registrare gli scenari Smoke contro il runner senza modificare lo
schema del report. Il prossimo incremento deve:

- costruire runtime e adapter da configurazione ambiente;
- usare capability introspection per decidere tra esecuzione e `unsupported`;
- implementare i 14 scenari del manifest per Ollama e llama.cpp;
- applicare mutation guard e fixture model;
- produrre il primo report live tramite `maestro bench smoke`.

Nessun servizio live o modello locale era requisito della Fase 1.

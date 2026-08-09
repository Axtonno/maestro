# Milestone 3 — Report finale Fase 3

Fase: Runtime Benchmark

Stato: Completata

Data: 2026-08-09

---

# Obiettivo

Misurare in modo locale e riproducibile prestazioni, risorse, cancellazione e
resilienza della configurazione hardware–provider–modello, senza introdurre
classifiche automatiche o dipendenze obbligatorie da telemetria esterna.

---

# Risultati consegnati

- Manifest Runtime versionato con quattro scenari provider e sette modello.
- Comandi `maestro bench provider` e `maestro bench model`.
- Latenza di introspection, listing e discovery.
- Retry transitorio controllato con recupero sul provider live.
- Circuit breaker controllato con verifica del blocco prima dell'adapter.
- Latenza completion, TTFT, latenza stream e token/sec quando provider-reported.
- Latenza di cancellazione per stream generativo e pull.
- Latenza e throughput embedding su batch fisso di otto input.
- Load/unload e confronto cold/warm con cleanup della residenza.
- Sampler Linux procfs per CPU e RAM con scope `maestro_process` o
  `provider_process`.
- Omissione delle metriche non osservabili; nessuno zero sintetico per VRAM.
- Mutation guard e fixture ownership conservate sul pull cancellato.
- ADR-0019 e documentazione operativa aggiornata.

---

# Copertura deterministica

Gli undici scenari vengono attraversati tramite il Provider Runtime reale e un
provider fixture. I test verificano:

- separazione esatta dei manifest per i due comandi;
- completion, stream terminale, TTFT e usage;
- cancellazione cooperativa di stream e pull;
- embedding batch e dimensioni coerenti;
- load/unload e cold/warm;
- retry con un solo fault e due tentativi;
- apertura del circuito e chiamata bloccata senza raggiungere il provider;
- cleanup di stream, modello, fixture e policy;
- parsing procfs e omissione CPU quando il delta non è disponibile;
- CLI offline con quattro e sette scenari `skipped`, senza I/O live.

---

# Verifiche

```text
GOCACHE=/tmp/maestro-go-build go test ./...
GOCACHE=/tmp/maestro-go-build go test -race ./...
GOCACHE=/tmp/maestro-go-build go vet ./...
GOCACHE=/tmp/maestro-go-build go run ./cmd/maestro bench validate --manifest docs/runtime-benchmark-manifest.yaml
GOCACHE=/tmp/maestro-go-build go run ./cmd/maestro bench provider --provider ollama --warmup 0 --runs 1 --output /tmp/maestro-provider-offline.json
GOCACHE=/tmp/maestro-go-build go run ./cmd/maestro bench model --provider ollama --warmup 0 --runs 1 --output /tmp/maestro-model-offline.json
```

Esito: suite completa, race detector, vet, validazione manifest e run CLI
offline superati. I due report contengono rispettivamente quattro e sette
scenari `skipped` e sono stati scritti con permessi `0600`.

---

# Verifica live

Nessun server Ollama o llama.cpp era configurato durante lo sviluppo. Le run
live dipendono dalla configurazione hardware e modelli dell'utente e non fanno
parte del gate deterministico. I comandi offline non tentano endpoint impliciti.

---

# Handoff alla Fase 4

Il Developer Benchmark può riusare runner, profili, report, sampler e scenari
embedding. Il prossimo incremento deve aggiungere dataset PHP/Laravel
versionato, fixture locali, risultato tecnico separato e rubrica qualitativa
0–3 senza evaluator LLM obbligatorio.

# Milestone 3 — Report finale Fase 2

Fase: Smoke Benchmark

Stato: Completata

Data: 2026-08-09

---

# Obiettivo

Eseguire in modo esplicito, riproducibile e sicuro la matrice live consegnata
dalla Provider Layer, conservando la suite ordinaria indipendente da servizi e
modelli locali.

---

# Risultati consegnati

- Composition root Smoke per Ollama e llama.cpp basato sui facade pubblici.
- Configurazione derivata dalle variabili dichiarate nel manifest.
- Implementazione di tutti i 14 scenari live.
- Capability preflight distinto per supporto e availability.
- Modelli fixture separati per chat, embedding, lifecycle e acquisition.
- Mutation guard esatta, controllo di fixture preesistente e ownership del
  cleanup acquisition.
- Cleanup di stream, pull, lifecycle, resilience policy e observer.
- Comando `maestro bench smoke` con warmup, run, timeout e gate opzionale.
- Report JSON su stdout o file atomico protetto con permessi `0600`.
- Schema report `1.1.0` con profili modello per ruolo.
- ADR-0018 e documentazione operativa in `smoke-benchmark.md`.

---

# Copertura deterministica

La matrice completa viene eseguita nei test attraverso il Provider Runtime reale
e un provider fixture che implementa i contratti pubblici. Sono verificati:

- tutti i 14 scenari in stato `passed`;
- routing, resilience policy e provider observer reali;
- completamento, cancellazione e chiusura degli stream;
- terminale ed EOF del pull;
- cleanup load/unload e pull/remove;
- distinzione tra `unsupported` e `capability_unavailable`;
- provider non configurato senza I/O;
- fixture acquisition preesistente mai rimossa;
- mutation guard alterata rifiutata;
- report atomico, redatto e con permessi restrittivi;
- CLI offline con 14 scenari `skipped`.

---

# Verifiche

```text
GOCACHE=/tmp/maestro-go-build go test ./...
GOCACHE=/tmp/maestro-go-build go test -race ./...
GOCACHE=/tmp/maestro-go-build go vet ./...
GOCACHE=/tmp/maestro-go-build go run ./cmd/maestro bench validate --manifest docs/provider-smoke-benchmark-manifest.yaml
GOCACHE=/tmp/maestro-go-build go run ./cmd/maestro bench smoke --provider ollama --output /tmp/maestro-smoke-offline.json
```

Esito: tutte le verifiche deterministiche superate. L'ultimo comando è stato
eseguito senza configurare un provider e ha prodotto quattordici scenari
`skipped`, senza tentare I/O live.

---

# Verifica live

Nessun server Ollama o llama.cpp era configurato nell'ambiente di sviluppo. Le
run live non sono un requisito del gate deterministico e verranno eseguite dagli
utenti sulle proprie configurazioni hardware–provider–modello.

---

# Handoff alla Fase 3

Il Runtime Benchmark può riusare composition root, profili, runner e writer. Il
prossimo incremento deve aggiungere:

- `maestro bench provider` e `maestro bench model`;
- cold/warm run e time to first token;
- latenza totale, throughput e cancellazione misurata;
- aggregati su run ripetute;
- sampler opzionali per CPU, RAM e VRAM con perimetro dichiarato.

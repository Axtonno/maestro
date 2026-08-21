# Milestone 11 — Report Fase 2

Data: 2026-08-21

Stato: **COMPLETATA — Developer Benchmark mutativo implementato**

## Risultato

È disponibile un harness separato dal Developer Benchmark read-only. Il nuovo
package `internal/benchmark/mutation` carica il profilo strict, verifica la
fixture congelata, produce snapshot fisici, applica serie fail-fast e definisce
report JSON/Markdown redatti e versionati.

Il comando seguente valida profilo e fixture senza I/O provider:

```text
GOCACHE=/tmp/maestro-go-build go run ./cmd/maestro bench mutation \
  --profile docs/mutation-qualification-profile.yaml
```

Esito: **PASS**, profilo versione 1, 3 gate e 15 scenari fisici.

## Contratti introdotti

- loader YAML strict con campi sconosciuti rifiutati;
- Gate A `3/3`, B `2/2` e C `3/3` non modificabili;
- materializzazione privata da `maestro-laravel-mini@1.0.0`;
- verifica della sostituzione e dei digest iniziale/finale;
- snapshot ordinato che rifiuta symlink e file non regolari;
- runner con deadline per tentativo e stop al primo non-PASS;
- schema `mutation-qualification-report/1.0.0`;
- decoder JSON strict e Markdown derivato;
- pubblicazione atomica dei report con permessi `0600`;
- lifecycle, approval, freshness, cleanup e stato fisico rappresentabili senza
  payload sensibili.

## Redazione

Il report non possiede campi per prompt, risposte, arguments, risultati tool,
diff, contenuti o path fisici. Identità, reason code e lifecycle accettano
soltanto valori bounded. La root temporanea non entra nello snapshot.

## Verifica

```text
GOCACHE=/tmp/maestro-go-build go test ./...
GOCACHE=/tmp/maestro-go-build go test -race ./...
GOCACHE=/tmp/maestro-go-build go vet ./...
git diff --check
```

Esito: **PASS**.

## Gate

- profili incompleti o divergenti rifiutati: superato;
- fixture isolata e cleanup verificato: superato;
- fail-fast deterministico: superato;
- JSON/Markdown coerenti e redatti: superato;
- approver di test distinto dall'evidenza live: superato dal contratto del
  report e dalla separazione degli executor.

La Fase 2 è completata. La Fase 3 può collegare la matrice deterministica agli
invarianti del prodotto e congelare il binario candidato.

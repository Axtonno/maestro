# Milestone 7 — Phase 6 Report

Stato: Completata

Data: 2026-08-12

---

# Risultato

La Fase 6 rende il loop workspace-aware con binding per-run, cinque reference
tool filesystem controllati e checkpoint di freshness dopo le mutazioni.

# Capacità consegnate

- workspace target immutabile e coerente con la request;
- registry run/workspace thread-safe e temporaneo;
- listing, read e search bounded;
- write content-addressed e create-if-absent;
- patch exact-occurrence con digest precondition;
- path logici normalizzati e action workspace-scoped;
- containment tramite `os.Root` e rifiuto di ogni symlink;
- revalidation durante Execute;
- stale marker all'inizio effettivo della mutazione;
- reindex e rebuild a checkpoint;
- nuova generazione usata dallo step successivo;
- ultimo snapshot valido preservato su refresh failure;
- percorso generico con `WorkspaceProvider` Laravel.

# Invarianti verificati

- la root assoluta non entra in arguments o messaggi provider;
- traversal e path assoluti non raggiungono il filesystem;
- symlink interni ed esterni sono rifiutati;
- inspect e mutate restano effect distinti;
- un file cambiato non supera la precondizione;
- una create concorrente non viene sovrascritta;
- deny prima di Execute non marca il contesto stale;
- mutation failure conserva stale e generazione precedente;
- il bundle dopo refresh coincide con la generazione indicizzata.

# Test e gate

```text
GOCACHE=/tmp/maestro-go-build go test ./internal/tool ./pkg/tool ./internal/agent ./pkg/agent
GOCACHE=/tmp/maestro-go-build go test -race ./internal/tool ./pkg/tool ./internal/agent ./pkg/agent
GOCACHE=/tmp/maestro-go-build go test ./...
GOCACHE=/tmp/maestro-go-build go vet ./...
git diff --check
```

# Gate

Il reference tool set legge e modifica soltanto workspace e risorse
autorizzati, mentre il Context Engine resta la fonte autorevole dell'evidenza.
La Fase 7 può comporre i servizi e aggiungere discovery ed eventi redatti.

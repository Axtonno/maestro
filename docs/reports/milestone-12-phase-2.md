# Milestone 12 — Phase 2 Report

Data: 2026-08-21

Stato: **COMPLETATA**

## Esito

La superficie destinata all'archive è allineata al target v0.2.0 e mantiene lo
stesso confine di autorità read-only qualificato nella v0.1.1. README,
installazione, quick start, configurazione, CLI, reference agent, compatibility
matrix, security model, known issues e troubleshooting concordano su Linux
`amd64`, Ollama, `llama3.1:8b` e list/read/search.

La configurazione inclusa contiene esattamente:

```text
workspace.list
workspace.read
workspace.search
workspace_mutate: deny
```

`docs/v0.2.0-api-compatibility.md` registra le aggiunte Go sperimentali dalla
v0.1.1 senza stabilizzarle. CLI e schema YAML `version: 1` del percorso
supportato restano invariati. Changelog e note di release dichiarano
`mutation_deferred` e non promuovono Granite o `workspace.patch`.

## Confine dell'archive

Gli script di packaging ora includono il contratto API v0.2.0 e verificano
esplicitamente l'assenza di:

- `configs/maestro.mutating.example.yaml`;
- `docs/mutation-qualification.md`;
- `docs/mutation-benchmark.md`.

La documentazione pubblica può descrivere le capacità sperimentali come non
supportate, ma non include profili o procedure che le presentino come percorso
di prodotto.

## Verifiche

| Controllo | Esito |
|---|---|
| `go test ./...` | PASS |
| `go test -race ./...` | PASS |
| `go vet ./...` | PASS |
| `bash -n` sugli script di packaging | PASS |
| tool set esatto della configurazione inclusa | PASS, 3/3 read-only |
| `workspace_mutate: deny` | PASS |
| `git diff --check` | PASS |

## Gate

**PASS.** La superficie pubblica v0.2.0 è coerente, il profilo distribuito non
amplia l'autorità e i controlli automatici proteggono l'allowlist archive. La
Fase 3 può produrre il primo packaging candidate soltanto dopo il commit di
questa fase.

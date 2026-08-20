# Milestone 9 — Report Fase 6

Data: 2026-08-20

Stato: **COMPLETATA — audit verde, GO alla Milestone 10**

## Identità finale v0.1.1

| Campo | Valore |
|---|---|
| Artifact | `maestro-v0.1.1-linux-amd64.tar.gz` |
| Stato manifest | `release` |
| Commit incorporato | `ba938abc6553bc87a89088eb6763a3e255aba4f8` |
| SHA-256 | `d894568cd65c261a75212274d7ab8a45eafa950660594b6c22cc777eb8ab9cf1` |
| Dimensione | 3.607.969 byte |
| Tag | `v0.1.1`, annotato, punta al commit incorporato |
| Piattaforma | Linux `amd64` |

Due build indipendenti sono byte-identiche. L'archive supera checksum,
inventory, path safety, manifest, permessi, versione/help, installazione fuori
dal checkout, scansione credential-shaped e assenza del path di build.

## Gate finale

| Gate | Esito |
|---|---|
| `go test ./...` | PASS |
| `go test -race ./...` | PASS |
| `go vet ./...` | PASS |
| `git diff --check` sulla sorgente di release | PASS |
| Packaging riproducibile | PASS |
| `doctor` dall'archive | 9/9 PASS |
| Quick start finale 1 | `completed`, 2 turni / 1 tool, 333.391 ms |
| Quick start finale 2 | `completed`, 2 turni / 1 tool, 69.986 ms |
| Profilo tool | list/read/search; mutazioni deny |
| Workspace reali/fixture | invariati nelle prove osservate |
| Anti-leak | PASS |

Entrambi i quick start finali respingono un primo turno testuale e completano
soltanto dopo una vera tool call provider-level. Non sono emersi falsi PASS.

## Audit delle osservazioni

- il bug di indicizzazione Laravel v0.1.0 è corretto e coperto in v0.1.1;
- le regressioni candidate RC1–RC3 sono state escluse prima della release;
- pseudo-call, call malformate, timeout e deadline sono bounded e classificati;
- non restano bug read-only bloccanti o osservazioni senza destinazione;
- support matrix, security model, known issues, troubleshooting, release notes,
  roadmap e contesto descrivono lo stesso confine;
- ADR-0030 chiude la Milestone 3 con Ollama baseline positiva e llama.cpp
  sperimentale/non supportato;
- nessun tool mutante, approval mutativa o capability esterna è stato aggiunto
  al profilo v0.1.x.

## Verdetto

**GO alla Milestone 10 — Controlled Mutation.**

Il GO autorizza l'avvio del design e dell'implementazione della vertical slice
mutativa, non promuove capacità mutative nella v0.1.x. La Milestone 10 deve
restare limitata a `workspace.patch` su un file esistente, preview concreta,
approval exact-fingerprint, precondizione digest, applicazione atomica e
reindex. `workspace.write`, shell, Git, processi, sandbox, recovery e
multi-agent restano fuori scope salvo qualificazione separata.

La Milestone 11 conserva il gate modello/hardware fail-fast `3/3`, `2/2`,
`3/3` definito da ADR-0030. Nessun supporto mutativo è dedotto dal GO di questa
fase.

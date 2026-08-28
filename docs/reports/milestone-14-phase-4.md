# Milestone 14 — Phase 4 Report

Data: 2026-08-28

Stato: **COMPLETATA**

## Esito

La matrice deterministica, negativa e anti-leak del direct chat è verde. Lo
streaming provider è abilitato soltanto da profilo e flag espliciti e conserva
content, usage, finish reason, metadati e terminale del percorso completion.

## Hardening

- il loader rileva anche sostituzione della workspace root, oltre a cambi di
  identità, stat o contenuto del file;
- provider e stream typed-nil vengono rifiutati;
- deadline e cancel vengono verificati prima e dopo preflight/generation e
  mantengono reason ed exit code distinti;
- stream senza terminale, con terminale non-stop, tool delta, chunk successivi,
  errore di receive/close o output eccessivo falliscono chiusi;
- l'output streaming è atomico: nessun chunk appare su stdout prima della
  validazione del terminale e dei limiti;
- un `--file` esplicito vuoto e input positional/stdin concorrente sono
  rifiutati prima della composition;
- le configurazioni v2 unknown e duplicate possiedono regressioni dedicate;
- snapshot e canary verificano immutabilità e assenza di leak operativi.

## Evidenza

La matrice versionata è in `docs/direct-chat-deterministic-matrix.md`. Gli
oracoli eseguibili sono nei test di `internal/directchat`, `cmd/maestro` e
`internal/productconfig`. Il dependency contract del servizio rende
inaccessibili retrieval, index, state machine, sessione, approver e fallback.

## Gate verificato

| Controllo | Esito |
|---|---|
| containment e tipi fisici | PASS |
| mutation/replacement durante lettura | PASS |
| hard limit input/file/output | PASS |
| response e streaming malformati | PASS |
| fail-closed e zero fallback | PASS |
| equivalenza streaming/non-streaming | PASS |
| deadline, cancel e reason code | PASS |
| workspace byte-identico | PASS |
| canary anti-leak | PASS |
| regressione agent/run | PASS |
| `go test ./...` | PASS |
| `go test -race ./...` | PASS |
| `go vet ./...` | PASS |
| `git diff --check` | PASS |

## Gate

**PASS.** Il contratto deterministico è idoneo al preflight live della Fase 5.
La qualifica del modello resta separata e deve rispettare C0–C4 e stop rule.

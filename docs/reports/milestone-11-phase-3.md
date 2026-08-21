# Milestone 11 — Report Fase 3

Data: 2026-08-21

Stato: **COMPLETATA — matrice deterministica verde e candidato congelato**

## Risultato

I quindici scenari di `mutation_matrix` sono collegati ai package che
possiedono gli invarianti e i fault seam reali. L'executor ha eseguito i test
focalizzati dal candidato e ha prodotto il report canonico JSON e la vista
Markdown senza I/O provider.

| Evidenza | Valore |
|---|---|
| Versione candidato | `v0.2.0-m11-qc.2` |
| Commit | `7e8ba62da22ad1942f3688b880922eacbec0889f` |
| Binario SHA-256 | `9870772b25f482eb4a5e539cea86e44aa19740e929c5789eab091d10c70101a3` |
| Dimensione | 9.310.464 byte |
| Piattaforma | Linux `amd64` |
| Profilo SHA-256 | `a64b7557ccd24f32bb4fb7cee7d64b630e16ec017c0776b2549d86bcd8480cac` |

Il checkout era pulito al momento del build. Il binario è stato costruito con
`-mod=readonly`, `-trimpath`, `-buildvcs=false`, build ID vuoto e identità
incorporata tramite ldflags.

## Matrice

| Gruppo | Scenari | Esito |
|---|---:|---|
| Positive path | exact patch | PASS |
| Precondizioni | stale digest, traversal, symlink | PASS |
| Approval | deny, EOF, no-TTY, input invalido | PASS |
| Commit/fault | cancellazione pre/post commit, fault filesystem | PASS |
| Freshness | refresh failure post-commit | PASS |
| Autorità | tool non dichiarato, replay, secondo tentativo | PASS |

Le failure pre-commit conservano SHA-256
`4826abe9c6c5d701133817a9dcb565f0b84f760da57e1b518d430b601520b1bd`.
Gli scenari post-commit registrano
`509b566bd04a17d567248a721885ac5af0d623f9f505288548c7c302628bac5d`,
contesto stale e terminale non riuscito. Il solo positive path raggiunge lo
stesso digest con contesto fresh e terminale completed.

## Report

- `reports/milestone-11-deterministic-qc2.json`;
- `reports/milestone-11-deterministic-qc2.md`.

Entrambi hanno permessi `0600`. La scansione non trova root fisiche, directory
temporanee, testo della patch o nomi dei campi payload esclusi.

## Verifica

```text
GOCACHE=/tmp/maestro-go-build go test -count=3 ./...
GOCACHE=/tmp/maestro-go-build go test -race ./...
GOCACHE=/tmp/maestro-go-build go vet ./...
git diff --check
```

Esito: **PASS**.

## Gate

- matrice completa verde: superato;
- fault pre-commit byte-identici e cleanup: superato;
- stato post-commit accurato: superato;
- profilo read-only invariato: superato;
- candidato pulito, identificato e riproducibile: superato;
- modifiche successive invalidano il candidato: regola attiva.

La Fase 3 è completata. Preflight e Gate A devono usare esclusivamente questo
binario, commit e profilo.

## Ripetizione del congelamento

Il primo candidato `qc.1` aveva matrice deterministica verde, ma il preflight
confrontava letteralmente la descrizione CPU e non normalizzava i marker
`(R)/(TM)` esposti da `/proc/cpuinfo`. Il candidato è stato invalidato senza
eseguire Gate A. La correzione è coperta da test; suite, race detector, vet e
matrice 15/15 sono stati ripetuti integralmente sul commit `7e8ba62...`.

`qc.1` e i suoi report restano evidenza storica non qualificabile. `qc.2` è il
solo candidato autorevole per le fasi live.

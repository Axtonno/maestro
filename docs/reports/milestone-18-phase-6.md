# Milestone 18 — Fase 6: artifact finale e tag annotato

Data: 2026-08-29

Stato: **COMPLETATA — PASS LOCALE**

## Identità finale

| Campo | Valore |
|---|---|
| commit release | `3f4c7d4b4fd2e380644cf250ce9e8fec2311af53` |
| tag locale | `v0.3.0` annotato, target commit release |
| archive | `maestro-v0.3.0-linux-amd64.tar.gz` |
| dimensione | `3775317` byte |
| SHA-256 archive | `6c8f0e883ec8f8c05571fc2e7bc1f4ecac608c2bd7e338395ae0a4253fff1aaf` |
| SHA-256 binario | `378a0533083b9a00be6c0212ca52001cebc5f77b476a20038bc8e08d1fc3d42d` |
| versione/stato | `v0.3.0` / `release` |
| SHA-256 profilo | `1c5bbe79edf125485d14518e58ff18c48156eaa0fb91faf82fcf3cd97375d0ee` |
| SHA-256 fixture route | `7e224d7e57bf0be6d2618e668d8515b07d332f94b7960de4640e4388b31bbc39` |

## Gate

Il delta dal commit RC `f33ce456…` è esclusivamente documentale: nessun file
sotto codice, configurazione, fixture, script o dipendenze è cambiato. I gate
sono stati eseguiti da un worktree Linux temporaneo e pulito del commit release
per preservare i line ending Git nativi.

| Gate | Esito |
|---|---|
| `go test ./...` | PASS |
| `go test -tags maestro_development ./...` | PASS |
| `go test -race ./...` | PASS |
| `go vet ./...` | PASS |
| syntax script | PASS |
| doppio packaging `v0.3.0` | PASS — byte-identico |
| checksum e allowlist | PASS |
| installazione pulita, version e help | PASS |
| manifest/binario/tag sul medesimo commit | PASS |
| profilo e fixture invariati | PASS |
| scansione credential-shaped | PASS |

L'archive e il checksum persistenti sotto `dist/` sono le sole coppie
autorizzate alla pubblicazione. Il tag è stato creato e verificato soltanto in
locale; non è ancora stato pushato e non esiste ancora una GitHub Release
creata da questo workflow.

## Gate finale

Verdetto Fase 6: **PASS LOCALE**. La Fase 7 è autorizzata a verificare lo stato
remoto, pushare commit e tag, pubblicare esattamente i due asset qualificati e
riscaricarli da zero. Ogni collisione remota arresta il workflow senza
overwrite.

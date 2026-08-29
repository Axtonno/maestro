# Milestone 18 — Fase 2: documentazione pubblica e metadata

Data: 2026-08-29

Stato: **COMPLETATA — PASS**

## Source freeze

| Campo | Valore |
|---|---|
| commit sorgente RC | `f33ce456cd65c24abcd5561d7140438ff08e64f1` |
| branch | `master` |
| delta funzionale da M17 `70a9630` | zero |
| profilo chat SHA-256 | `1c5bbe79edf125485d14518e58ff18c48156eaa0fb91faf82fcf3cd97375d0ee` |
| servizio/prompt SHA-256 | `7fd79e1fafb70d0b7726ecca0909f92592f8706df890a9b6fb263c9d5b8575c1` |
| fixture route SHA-256 | `7e224d7e57bf0be6d2618e668d8515b07d332f94b7960de4640e4388b31bbc39` |

Il commit congela la superficie pubblica destinata al release candidate. Il
successivo commit di governance può aggiungere questo report e aggiornare lo
stato del piano, ma non modifica i file inclusi nell'archive; il release
candidate della Fase 3 deve essere costruito dal commit esatto sopra.

## Superficie finalizzata

- README e guide installano dalla coppia archive/checksum della medesima
  GitHub Release e verificano manifest, versione, commit e stato;
- quick start usa soltanto asset estratti, profilo chat-only incluso e fixture
  distribuita, senza checkout;
- release notes e security policy sono temporalmente neutre: un candidate non
  viene presentato come release pubblicata;
- CLI e schema v2 restano esplicitamente sperimentali nella serie 0.x;
- compatibility, security, known issues e troubleshooting mantengono il solo
  support claim Direct Chat single-file read-only;
- packaging documenta build distinte per `packaging-candidate`,
  `release-candidate` e `release`, rifiuto overwrite e verifica temporanea;
- agent, retrieval, tool, mutation, altri modelli/provider e piattaforme non
  qualificate restano esclusi in tutte le superfici pubbliche.

I soli token sorgente `@MAESTRO_VERSION@` e `@MAESTRO_STATUS@` sono confinati
ai template `docs/installation.md` e `docs/quick-start.md`; il gate di
packaging li risolve completamente nell'archive.

## Gate

| Gate | Esito |
|---|---|
| link e file locali della superficie pubblica | PASS |
| token di packaging confinati e risolti | PASS |
| comandi installazione/quick start coerenti | PASS |
| claim supportati/non supportati coerenti | PASS |
| scansione credential-shaped della superficie | PASS |
| `go test -count=3 ./...` | PASS |
| `go test -tags maestro_development -count=3 ./...` | PASS |
| `go test -race ./...` | PASS |
| `go vet ./...` | PASS |
| `bash -n scripts/*.sh` | PASS |
| doppio packaging RC dal source freeze | PASS — byte-identico |
| checksum, allowlist, installazione e anti-leak | PASS |

La verifica temporanea `v0.3.0-rc.1`, stato `release-candidate`, ha prodotto
due archive byte-identici con SHA-256
`b034828a07f33a2643556123c00917ff563d83f1976dab968542712f0df7be3a`.
Questa esecuzione prova il source freeze ma non persiste né promuove l'RC; la
Fase 3 produce l'artifact immutabile sotto `dist/` senza overwrite.

## Gate finale

- documentazione installabile senza checkout: **PASS**;
- nessun claim di pubblicazione anticipato: **PASS**;
- zero modifica a codice/config/fixture/script: **PASS**;
- source RC congelato: **PASS**;
- blocker aperti della superficie pubblica: **zero**.

Verdetto Fase 2: **PASS**. La Fase 3 — Release candidate riproducibile è
autorizzata. Nessun tag, push, GitHub Release o asset remoto è stato creato.

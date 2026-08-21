# Milestone 12 — Phase 3 Report

Data: 2026-08-21

Stato: **COMPLETATA**

## Candidate

| Campo | Valore |
|---|---|
| Versione | `v0.2.0-pc.1` |
| Stato manifest | `packaging-candidate` |
| Commit | `7d3f45ee0268fc758b9e3722e57c91e486065615` |
| Piattaforma | Linux `amd64` |
| SHA-256 | `e5f98bedcb94ab40236d3f315cf9af0be976825abbd2d9a6ea756ad26200fc13` |

Il candidate è stato costruito da worktree pulito e conservato fuori dal
repository. Non viene rinominato o promosso implicitamente.

## Packaging riproducibile

`verify-package-candidate.sh` ha costruito due volte l'archive con input
normalizzati. Archive e checksum risultano byte-identici. La successiva build
conservata produce lo stesso SHA-256 del gate.

L'audit dell'archive verifica:

- path relativi sicuri e permessi attesi;
- binario, manifest, licenza Apache-2.0, NOTICE e attribution;
- fixture Laravel versionata senza dipendenze, `.env` o symlink;
- documentazione pubblica e contratto API v0.2.0;
- configurazione inclusa con list/read/search e deny mutativo;
- assenza di profilo mutante e documentazione operativa di qualificazione;
- assenza di token irrisolti, path del build workspace e credenziali-shaped.

## Installazione pulita

L'archive è stato estratto in una directory temporanea esterna al checkout. Le
verifiche sono state eseguite sul binario estratto:

| Controllo | Esito |
|---|---|
| checksum SHA-256 | PASS |
| `maestro version` | PASS, versione e commit esatti |
| root help | PASS |
| `maestro agents` | PASS, `agent.reference` presente |
| `maestro doctor` | PASS, 9/9 con Ollama locale |
| `maestro models` | PASS, `llama3.1:8b` presente |

Il probe live è read-only: non ha avviato Ollama, scaricato modelli o cambiato
la configurazione del provider.

## Gate

**PASS.** `v0.2.0-pc.1` è un packaging candidate riproducibile, installabile e
coerente con il confine read-only. Non è un release candidate. Lo stesso
archive immutabile passa alla matrice operativa e di sicurezza della Fase 4.

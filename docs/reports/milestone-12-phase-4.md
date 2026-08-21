# Milestone 12 — Phase 4 Report

Data: 2026-08-21

Stato: **COMPLETATA**

## Oggetto della prova

Tutti i controlli artifact-level sono stati eseguiti sullo stesso archive
immutabile `v0.2.0-pc.1`, commit
`7d3f45ee0268fc758b9e3722e57c91e486065615`, SHA-256
`e5f98bedcb94ab40236d3f315cf9af0be976825abbd2d9a6ea756ad26200fc13`.

Il digest iniziale del controller candidato è
`4826abe9c6c5d701133817a9dcb565f0b84f760da57e1b518d430b601520b1bd`.

## Matrice operativa

| Scenario | Exit | Terminale/esito | Turni/tool | Wall time | stdout |
|---|---:|---|---:|---:|---:|
| EOF senza istruzione | 2 | `read instruction: EOF` | 0/0 | bounded | 0 byte |
| stdin oltre 1 MiB | 2 | `instruction exceeds 1048576 bytes` | 0/0 | bounded | 0 byte |
| `workspace_mutate: allow` | 1 | `configuration_invalid` in composition | 0/0 | bounded | solo righe doctor |
| SIGINT dopo 3 secondi | 130 | `canceled` | 1/0 | 3004 ms | 0 byte |
| deadline run 1 secondo | 130 | `deadline_exceeded` | 1/0 | 1006 ms | 0 byte |
| hard limit `model_turns: 1` | 1 | `limit_exceeded` | 1/1 read | 179755 ms | 0 byte |

Nel caso hard limit il messaggio CLI sintetico è `execution_failed`, mentre
l'evento terminale precedente resta `limit_exceeded` e autorevole, come già
documentato nei known issues.

SIGINT, deadline e hard limit sono stati eseguiti senza TTY. Nessuna approval è
stata mostrata o concessa; la configurazione inclusa non registra tool mutanti.
I test deterministici coprono inoltre deny, EOF, input invalido e no-TTY
dell'approver terminale.

## Stato fisico e redazione

Dopo ogni scenario il controller conserva esattamente il digest iniziale. Non
sono osservate scritture, proposal o tentativi mutativi.

Le evidenze stdout/stderr sono state scandite contro:

- canary distinti per SIGINT, deadline e hard limit;
- nomi e contenuti applicativi della fixture;
- nomi dei workspace tool;
- path fisici temporanei;
- pattern di private key, token GitHub/OpenAI e access key AWS.

La scansione è negativa. I log conservano soltanto limiti, run ID, contatori,
terminal reason e failure sintetiche.

## Gate deterministici

| Controllo | Esito |
|---|---|
| `go test -count=3 ./...` | PASS |
| `go test -race ./...` | PASS |
| `go vet ./...` | PASS |
| syntax check script | PASS |
| `git diff --check` | PASS |

## Gate

**PASS.** Deny, EOF, no-TTY, SIGINT, deadline, hard limit, immutabilità e
anti-leak hanno evidenza sul candidate. `v0.2.0-pc.1` può entrare nella
qualificazione live read-only della Fase 5 senza cambiare artifact o criteri.

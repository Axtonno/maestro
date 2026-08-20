# Milestone 10 — Report finale

Data: 2026-08-20

Stato: **COMPLETATA — GO alla Milestone 11**

## Esito

La Milestone 10 consegna deterministicamente il vertical slice Controlled
Mutation:

```text
read -> prepare patch -> preview -> approval -> apply -> reindex -> final
```

La stessa proposta immutabile produce diff, permission fingerprint e contenuto
applicato. `workspace.patch` modifica al massimo una singola occorrenza in un
file PHP esistente sotto `app/`, dopo approval TTY one-shot, e usa un unico
punto di commit atomico Linux. La risposta finale è possibile soltanto dopo
nuova generazione indicizzata e bundle fresh.

| Fase | Risultato |
|---|---|
| 1 — Contratto | ADR-0031, threat boundary e profilo congelati |
| 2 — Proposta | preview autorevole, bounded e fingerprint-bound |
| 3 — Approval | deny/once TTY, exact fingerprint, profilo opt-in |
| 4 — Commit | temporaneo, fsync, recheck, rename atomico, fault matrix |
| 5 — Freshness | un tentativo, terminali accurati, reindex obbligatorio |
| 6 — Audit | matrice integrata verde, API/doc allineate, profilo M11 |

## Proprietà dimostrate

- nessuna approval mutativa run-scoped, automatica o non interattiva;
- preview e apply vincolati allo stesso fingerprint;
- stale digest, traversal, symlink, proposta cambiata e replay rifiutati;
- ogni fault pre-commit lascia il target byte-identico e tenta cleanup;
- i lettori osservano soltanto vecchio o nuovo contenuto completo;
- esiti post-commit non dichiarano il workspace invariato;
- apply/refresh falliti non producono testo finale né retry;
- proposal/approval/apply/reindex sono osservabili senza leakage;
- il profilo read-only v0.1.x resta byte-invariato.

## Contratto consegnato alla Milestone 11

Il candidato è Linux `amd64`, Ollama, `ibm/granite4.1:8b`, con lower bound
Intel Core i5-8365U, 8 CPU logiche, 15 GiB RAM e 4 GiB swap. Il file
`docs/mutation-qualification-profile.yaml` congela Gate A `3/3`, Gate B `2/2`
e Gate C `3/3`, fail-fast, insieme alla matrice fisica e agli obblighi di
redazione.

La Milestone 11 può concludere soltanto:

- supporto sul lower bound;
- supporto su requisito hardware superiore provato;
- ulteriore rinvio della mutazione.

## Confini invariati

`workspace.write`, create/delete/rename, multi-file, shell, Git, Composer,
Artisan, PHPUnit, processi, sandbox, rollback generale, crash recovery,
plugin/tool terzi e multi-agent restano fuori scope. Trusted in-process e
privilegi dell'utente restano limiti espliciti.

## Verdetto

**Milestone 10 completata. GO alla Milestone 11 per la qualificazione live;
Controlled Mutation non è ancora una capability supportata.**

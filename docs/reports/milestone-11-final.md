# Milestone 11 — Mutation Qualification Final Report

Data: 2026-08-21

Stato: **COMPLETATA — `mutation_deferred`**

## Esito per fase

| Fase | Esito | Evidenza principale |
|---:|---|---|
| 1 | completata | contratto, profilo e criteri congelati |
| 2 | completata | harness mutativo e report v1 |
| 3 | completata | matrice 15/15 e candidato `qc.2` |
| 4 | conclusa con FAIL | preflight PASS, Gate A 0/1 |
| 5 | conclusa senza run | Gate B bloccato dal fail-fast |
| 6 | conclusa senza run | Gate C bloccato dal fail-fast |
| 7 | completata | audit, ADR-0032 e handoff |

## Evidenza autorevole

Il candidato `v0.2.0-m11-qc.2` incorpora il commit
`7e8ba62da22ad1942f3688b880922eacbec0889f` e ha SHA-256
`9870772b25f482eb4a5e539cea86e44aa19740e929c5789eab091d10c70101a3`.
Il profilo congelato ha SHA-256
`a64b7557ccd24f32bb4fb7cee7d64b630e16ec017c0776b2549d86bcd8480cac`.

La matrice deterministica supera 15 scenari su 15. Il preflight live supera
piattaforma, lower bound hardware, configurazione, composizione, capability
provider/modello e detection Laravel. Questi risultati dimostrano che harness,
runtime e ambiente possono eseguire la prova, ma non qualificano il comportamento
generativo richiesto.

Gate A fallisce al primo campione: dopo una read call valida, il modello emette
una patch call con arguments non esatti. Il reason code è
`patch_tool_call_invalid`, la failure class è `model` e il terminale è
`provider_failure`. Il Tool Runtime non viene invocato; approval, tentativi
mutativi e differenze del workspace restano a zero. I restanti due campioni
Gate A, i due Gate B e i tre Gate C non vengono eseguiti per la stop rule.

## Verdetto vincolante

Dei tre esiti ammessi è selezionato esclusivamente il terzo: **Controlled
Mutation rinviata**. Non esiste evidenza per dichiarare supporto sul lower bound
o su hardware superiore. `workspace.patch`, il profilo mutante e
`ibm/granite4.1:8b` restano fuori dalla compatibility promise.

La Milestone 12 riceve un **GO limitato a una v0.2.0 read-only**. Non deve
produrre esempi mutanti supportati né presentare la presenza del codice come
qualificazione. Una futura riapertura richiede nuovo congelamento, nuovo
candidato e Gate A `3/3`, Gate B `2/2`, Gate C `3/3` completi.

## Catena dei report

- `reports/milestone-11-phase-1.md`;
- `reports/milestone-11-phase-2.md`;
- `reports/milestone-11-phase-3.md`;
- `reports/milestone-11-deterministic-qc2.json` e `.md`;
- `reports/milestone-11-phase-4.md`;
- `reports/milestone-11-gate-a.json` e `.md`;
- `reports/milestone-11-phase-5.md`;
- `reports/milestone-11-phase-6.md`;
- `reports/milestone-11-phase-7.md`.

Il primo candidato e i report senza suffisso `qc2` sono storici e non
qualificabili. Nessun artifact di release è stato creato durante la milestone.

## Gate finale

Suite completa, race detector, vet, validazione strict del profilo, integrità
dei report, scansione anti-leak e `git diff --check` risultano **PASS**. Il
profilo read-only non ha acquisito tool o permission mutative. Tutti i campioni
eseguiti e non eseguiti sono contabilizzati e il failure Gate A resta l'unica
causa che attiva la catena fail-fast.

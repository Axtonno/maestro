# Milestone 35 — Report finale

Data: 2026-09-05. Stato: **COMPLETATA — PROFILO MUTATIVO QUALIFICATO**.

Verdetto: **`mutation_specific_model_qualified`**.

Il confronto congelato ha selezionato `qwen2.5-coder:14b`, digest
`9ec8897f747e246e970bc5cfdda85d22f1123dc2e3d34978a010a75968716849`.
È l'unico dei tre profili con 12/12 output conformi, 9/9 proposte positive e
3/3 astensioni corrette. La selezione ha usato 36 generazioni senza retry,
repair, fallback o tuning.

La qualifica ha poi usato 30 casi nuovi, 15 development e 15 holdout. Sei
casi sono stati respinti dall'host prima del provider; gli altri 24 hanno
prodotto una sola generazione ciascuno. Le approval sono passate da TTY reale
e ogni applicazione ha operato su un workspace temporaneo con sentinella.

| Gate | Development | Holdout | Totale |
| --- | ---: | ---: | ---: |
| Output provider conformi | 12/12 | 12/12 | 24/24 |
| Proposte positive corrette | 10/10 | 10/10 | 20/20 |
| Astensioni necessarie corrette | 2/2 | 2/2 | 4/4 |
| Target host-bound conservati | 10/10 | 10/10 | 20/20 |
| Preview esatte | 10/10 | 10/10 | 20/20 |
| Approval allow/deny raggiunte | 10/10 | 10/10 | 20/20 |
| Apply attesi completati | 7/7 | 7/7 | 14/14 |
| Terminali corretti | 15/15 | 15/15 | 30/30 |
| Scritture stale/fuori selezione | 0/0 | 0/0 | 0/0 |
| Mutazioni errate/non approvate | 0/0 | 0/0 | 0/0 |
| Failure con effetti | 0 | 0 | 0 |

Sviluppo, holdout e aggregato superano tutti i gate. Gli hash del report
coincidono con il preflight:

| Artefatto | SHA-256 |
| --- | --- |
| Matrice | `e7e75d0daf3be372ac4bf311f8d02944d634573e54a62a3c96d63ee207d5db8c` |
| Schema | `bc3432a8f19867eec8e153adaa4434b688974cf34d24b6bd770e887e0dd7557d` |
| Prompt | `594659d52ec6142a5ef79c36dc0db4899e7ef1bb3f99d05017410f68bc1ba732` |
| Report di qualifica | `4dbbd1baf63f005d42eb30d06b997b4720462ab60a0858c55ecb6cde07ecc72d` |

Le matrici M33, M34 e M35 non vengono ripetute. La decisione assegna
`qwen3.5:9b` a Direct Chat e il profilo congelato `qwen2.5-coder:14b` alla
futura Controlled Mutation. Il support pubblico resta quello di v0.4.0;
package, tag e release v0.5.0 restano non autorizzati fino alla productization.

Evidenze: `milestone-35-selection-runs.json`,
`milestone-35-qualification-preflight.md` e
`milestone-35-qualification-runs.json`. Decisione in `../adr/ADR-0040.md`.
M36 è aperta nel piano `../milestone-36-controlled-mutation-productization-plan.md`.

`go test ./...`, `go test -race ./...`, `go vet ./...` e i contract test M35
passano nella copia Linux isolata con fixture Git LF.

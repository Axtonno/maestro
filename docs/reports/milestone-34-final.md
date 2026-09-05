# Milestone 34 — Report finale

Data: 2026-09-05. Stato: **COMPLETATA — PROFILO MUTATIVO RESPINTO**.

Verdetto: **`qwen3.5_9b_host_bound_mutation_profile_rejected`**.

Selezionato il ramo B. Il prompt host-bound non contiene istruzioni residue
di ricerca/unicità; i tre payload falliti contengono byte sufficienti e
richieste determinabili. Selezione e splice attesi sono corretti. Il vero
adapter, esercitato con HTTP in memoria, conserva messaggi, schema e opzioni.
Metadata del modello e renderer della versione dichiarata non mostrano
istruzioni aggiuntive incompatibili. Evidenze, fonti e limiti nel
[report di attribuzione](milestone-34-attribution.md).

La causa interna delle astensioni non è osservabile: il verdetto riguarda
il profilo operativo M33, non l'impossibilità universale del modello.

| Verifica | Esito |
| --- | --- |
| Hash prompt/schema/matrice M33 | PASS |
| Selezioni e risultati attesi dei tre casi | 3/3 |
| Richieste serializzate senza istruzioni aggiuntive | 3/3 |
| Identità modello e versione rispetto a M33 | Coincidenti |
| Conflitto concreto che autorizzi ramo A | Non rilevato nell'audit |
| Nuove generazioni o repliche M33 | 0 |
| Nuovo candidate M34 | Non creato: ramo B |

`go run ./scripts/m34audit`, `go test ./...`, `go vet ./...` e `git diff --check`
passano. Suite completa e vet eseguiti nella copia Linux isolata con fixture
Git LF; il contract test verifica ramo, soglie e identità delle evidenze.
Le soglie M34 restano invariate; i gate live del
ramo A non sono eseguiti e non sono dichiarati PASS.

Stop al tuning di `qwen3.5:9b` sul profilo host-bound respinto. Aperta
**Milestone 35 — Mutation-Specific Model Selection**, con piano e gate propri,
senza selezionare ancora un modello. `qwen3.5:9b` resta per Direct Chat.
M33 e i suoi artefatti restano invariati; Controlled Mutation e v0.5.0 non
sono autorizzate. Decisione in `../adr/ADR-0039.md`.

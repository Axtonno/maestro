# Milestone 35 — Selezione del modello mutativo

Data: 2026-09-05. Verdetto: `mutation_specific_model_selected`.

Modello selezionato: `qwen2.5-coder:14b`, digest
`9ec8897f747e246e970bc5cfdda85d22f1123dc2e3d34978a010a75968716849`.

| Profilo | Output | Positivi | Astensioni | Latenza aggregata | Eleggibile |
| --- | ---: | ---: | ---: | ---: | --- |
| `qwen2.5-coder:7b` | 12/12 | 9/9 | 1/3 | 6.276 ms | No |
| `qwen2.5-coder:14b` | 12/12 | 9/9 | 3/3 | 12.084 ms | Sì |
| `granite-code:8b-instruct` | 12/12 | 7/9 | 0/3 | 14.913 ms | No |

Il 7B propone nei casi con valore mancante e richiesta contraddittoria.
Granite altera in modo errato CRLF e cancellazione e propone in tutte le
astensioni necessarie. Gli output sono formalmente conformi in ogni profilo.

La latenza non determina la scelta: è tie-breaker e viene considerata solo
fra candidati che superano tutti i gate. Il 14B è l'unico eleggibile.

La selezione usa 36 generazioni, una per coppia modello/caso. Nessun retry,
repair, fallback o tuning. Matrice, prompt e schema coincidono con i digest
del preflight. Questa scelta autorizza la qualifica M35, non Controlled
Mutation né un candidate v0.5.0.

Evidenze: `milestone-35-candidate-inventory.md`,
`milestone-35-model-metadata.json`, `milestone-35-selection-preflight.md` e
`milestone-35-selection-runs.json`.

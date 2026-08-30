# Milestone 20 — Report finale

Data: 2026-08-30

Stato: **COMPLETATA**

Verdetti conservati:

```text
model_hardware_bound
thinkpad_profile_candidate
operational_corrections_ready
```

## Decisione

Maestro non introduce overhead terminale materiale rispetto a Ollama nel
confronto appaiato sul ThinkPad. `qwen2.5-coder:7b` è un candidato plausibile
per Direct Chat CPU-only: completa la matrice M20, conserva il workspace e
riduce la mediana appaiata del 44,3% rispetto a `qwen3.5:9b`.

Il modello **non viene promosso**. L'evidenza qualitativa M17 (2/5) e M20
(4/5) deve essere riconciliata in una matrice unica, precongelata e ripetuta.
Inoltre M20 usa Ollama 0.32.14, mentre la qualifica v0.3.0 usa 0.33.1; cold,
warm, residenza, memoria ed eviction non costituiscono ancora una promessa di
prodotto.

## Esiti per fase

| Fase | Esito | Evidenza |
|---|---|---|
| A — attribuzione | `model_hardware_bound` | delta Maestro/Ollama da -0,18 a +0,11 s su body byte-identici |
| B — lower-resource | `thinkpad_profile_candidate` | no-file 3/3, single-file 5/5, qualità 4/5, mediana 69,0 s vs 123,9 s |
| C — operabilità | `operational_corrections_ready` | config diagnostics, binary identity e heartbeat redatto implementati |

Report dettagliati:

- `milestone-20-phase-a.md`;
- `milestone-20-phase-b.md`;
- `milestone-20-phase-c.md`.

## Limiti che impediscono la promozione

1. M17 e M20 usano task/oracoli diversi e producono qualità differente;
2. Ollama non è allineato tra piattaforma qualificata e ThinkPad;
3. i costi cold e warm sono osservati ma non ancora contrattualizzati;
4. M20 non imponeva massimo warm assoluto: un task qwen2.5 ha richiesto 190,6
   secondi;
5. il candidate C non è ancora congelato in un artifact installabile.

Questi limiti non negano il risultato M20. Impediscono soltanto di trasformare
un candidato development-only in support claim.

## Handoff

La Milestone 21 è aperta come `CPU Direct Chat Product Qualification`. Deve
usare Ollama e modello/digest congelati, candidate con le correzioni C,
matrice combinata M17+M20 eseguita due volte, gate cold/warm separati, soglie
assolute, misure di memoria/eviction e installazione da artifact.

Verified agent, tool calling, retrieval, multi-file e Controlled Mutation
restano esclusi. Nessuna release o pubblicazione è autorizzata dalla chiusura
di M20.

## Verifica deterministica finale

| Gate | Esito |
|---|---|
| `go test ./...` | PASS |
| `go test -race ./...` | PASS |
| `go vet ./...` | PASS |
| `bash -n scripts/package-candidate.sh scripts/verify-package-candidate.sh` | PASS |
| `git diff --check` | PASS |

Una build development con versione `v0.3.0-m20.c`, status `development` e
commit `uncommitted` ha eseguito `version --diagnostic`. SHA-256 dichiarato e
calcolato sul file coincidono:

```text
4e181b5c2bb0fbd807e8e64f63c2e0b6108be0dbeadbe124dd880e527b23ae81
```

Questa build in `/tmp` è soltanto uno smoke della diagnostica, non un artifact
qualificato o distribuibile.

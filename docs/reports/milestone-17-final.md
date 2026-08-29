# Milestone 17 — Final Report

Data: 2026-08-29

Verdetto: **`direct_chat_product_baseline`**

## Decisione

La Milestone 17 è completata. Il packaging candidate immutabile
`v0.3.0-pc.1`, commit
`70a9630203ccf82a4d8858a9e47b48f5333b9cbd`, SHA-256
`82bfb33f3fd9af911e3b2b1e89f9920177b281046da21b186512e577e114fb61`,
ha superato doppio packaging, audit locale, installazione pulita e matrice
finale live sulla piattaforma WSL2/Ubuntu 24.04/RTX 5070.

Il verdetto autorizza la preparazione separata di tag, release candidate e
release artifact v0.3.0 nel solo perimetro qualificato. Nessuna di queste
azioni è stata eseguita durante la milestone e il packaging candidate non deve
essere rinominato o promosso retroattivamente.

## Catena di evidenza

| Gate | Esito |
|---|---|
| Fasi 1–5: contratto, boundary, loader, profilo e streaming | PASS |
| F6.4 deterministica e live | PASS — C0 3/3, C1 3/3, stream 2/2, qualità 4/5 |
| packaging riproducibile e archive audit | PASS |
| identità archive/manifest/binario/config | PASS |
| installazione pulita fuori checkout | PASS |
| doctor chat finale | PASS 5/5 |
| no-file, complete e stream | PASS |
| traversal e symlink evasivo | PASS |
| SIGINT e deadline | PASS |
| immutabilità e anti-leak | PASS |

La matrice finale usa Ollama 0.33.1 e `qwen3.5:9b`, digest
`6488c96fa5faab64bb65cbd30d4289e20e6130ef535a93ef9a49f42eda893ea7`,
context 4096, thinking disabilitato e temperatura zero. Complete e stream
restituiscono semanticamente `POST /orders`, `OrderController::store`, con
terminale `completed` e finish `stop`.

## Support claim autorizzato

- Linux `amd64`;
- provider Ollama locale su loopback;
- modello e digest esatti sopra;
- Direct Chat tool-free, read-only, con zero o un file esplicito contained;
- complete e streaming opt-in con output atomico;
- profilo strict v2, doctor dedicato, limiti e failure redatti.

## Non garanzie

Restano non qualificati e non supportati:

- `maestro agent`, alias `run` e verified/reference agent;
- retrieval, indicizzazione, selezione multi-file e memoria persistente;
- tool calling, plugin di terze parti e fallback agentico;
- Controlled Mutation, write, patch e approval;
- modelli o digest diversi, llama.cpp ed endpoint remoti;
- sandbox, shell, Git, Docker e orchestrazione multi-agent.

Il verdetto non modifica la chiusura della Milestone 16 e non autorizza alcuna
mutazione. La pubblicazione v0.3.0 richiede ancora il workflow separato di
release e i relativi controlli di identità; l'evidenza autorevole della matrice
installabile è `docs/reports/milestone-17-phase-7.md`.

# Milestone 15 — Fase 2: direct/chat

Data: 2026-08-28

Stato: **PASS — `direct_chat_candidate`**

## Candidate congelato

| Campo | Valore |
|---|---|
| modello | `qwen2.5-coder:7b` |
| digest catalogo | `dae161e27b0e` |
| context / thinking | 4096 / disabilitato |
| timeout | 5 minuti |
| profilo seed M14 | `cd3221714cbd3c255f7a140cf1540fbe59e2cee19ce44d103a630f6c0955040f` |
| binario candidate | `8ee35ba7c5b3f0b186220244e806569b774c31bc8d42d246a968c57eb4ce2ed3` |

## Risultati

| Gate | Esito |
|---|---|
| C0 senza file | PASS 3/3; contesto insufficiente dichiarato, zero endpoint inventati |
| C1 single-file | PASS 3/3; `POST /orders`, `OrderController::store`, nessun endpoint aggiuntivo |
| C2 stream/non-stream | PASS 2/2; ground truth e terminali equivalenti |
| C3 operatività | PASS; `completed`, `stop`, latenze warm 275–3.936 ms |
| C4 sicurezza | PASS; zero tool, retrieval, fallback o mutazioni |

Il digest redatto del fixture prima e dopo C2 coincide:
`a7831ea9d6cfebf397f004ae0bded6fec59ec935962f8e268b79534fc68abda3`.

L'envelope CLI ha rappresentato alcuni effettivi come `unknown`; i log Ollama
hanno osservato context 4096 e `thinking=0`. Il report non conserva prompt,
risposte complete, contenuti sorgente o path fisici.


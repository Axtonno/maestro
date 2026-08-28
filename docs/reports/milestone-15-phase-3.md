# Milestone 15 — Fase 3: verified agent sintetico

Data: 2026-08-28

Stato: **FAIL — `verified_agent_rejected`**

## Candidate congelato

| Campo | Valore |
|---|---|
| modello | `qwen3.5:9b` |
| digest | `6488c96fa5faab64bb65cbd30d4289e20e6130ef535a93ef9a49f42eda893ea7` |
| context / thinking | 8192 / default |
| binario progressive | `4c126c863ae42848dfd61572395f1a2790c3b9bab9c5cb9af023cd3f0b048b3b` |
| profilo M15 | `820856f57c7ccd9546985e5ec380cc635acd298562448fcea5c1464aabc2e7a5` |
| autorità | soli list/read/search; `workspace_mutate: deny` |
| doctor | 10/10 PASS |

## Prima progressione live

La route è stata coperta dalla read bootstrap verificata. Nel primo stato
guidato il modello ha emesso una seconda tool call che il runtime ha chiuso
con `tool_failure`.

| Campo | Valore |
|---|---|
| terminale | `execution_failed` / `tool_failure` |
| durata | 31.092 ms |
| turni / tool call | 1 / 2 |
| token in/out | 3.177 / 225 |
| GPU / context | 100% GPU / 8192 |
| workspace | invariato |

Il digest fixture pre/post coincide:
`a7831ea9d6cfebf397f004ae0bded6fec59ec935962f8e268b79534fc68abda3`.
Non sono stati osservati OOM, reset provider, timeout o fallback CPU.

Il gate richiede progressione completa 2/2. Dopo il primo failure non è più
raggiungibile; si applica la stop rule senza retry opportunistico e senza B01.


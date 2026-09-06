# Milestone 36 — Preflight residency v2

Data: 2026-09-05. Esito: PASS, senza cambi di residency o generazioni.

Il profilo v2 conserva integralmente modelli, richieste, tre cicli, sequenza e
soglie del v1. Corregge il collector PID/RSS affinché osservi il processo
Ollama Linux in WSL e rende la sua presenza obbligatoria per il PASS.

| Artefatto | SHA-256 |
| --- | --- |
| Profilo residency v2 | `230e3e524c628a495eb4bd6215e204acfe86481bde3c1e2d6c6e3ad3670cc8d7` |
| Prompt mutativo | `594659d52ec6142a5ef79c36dc0db4899e7ef1bb3f99d05017410f68bc1ba732` |
| Schema mutativo | `bc3432a8f19867eec8e153adaa4434b688974cf34d24b6bd770e887e0dd7557d` |

`go test ./scripts/m36residency`: PASS. Il report v2 è separato e viene creato
con `O_EXCL`; il report v1 non viene modificato o promosso a evidenza finale.

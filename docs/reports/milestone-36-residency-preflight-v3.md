# Milestone 36 — Preflight residency v3

Data: 2026-09-05. Esito: PASS, senza generazioni.

Il v3 sostituisce l'eviction automatica con unload esplicito del modello
uscente e attesa di `/api/ps` vuoto. Le soglie temporali, i modelli, le
richieste e i tre cicli restano invariati rispetto al v2.

| Artefatto | SHA-256 |
| --- | --- |
| Profilo residency v3 | `e1343bdcab53dd4b00ed36f34bbd7eb5ffbee6cca67cbffcda6a469f8e5e89b1` |
| Prompt mutativo | `594659d52ec6142a5ef79c36dc0db4899e7ef1bb3f99d05017410f68bc1ba732` |
| Schema mutativo | `bc3432a8f19867eec8e153adaa4434b688974cf34d24b6bd770e887e0dd7557d` |

Il collector distingue il processo `ollama serve` dagli helper di discovery e
dai worker `llama-server`. Il report è separato, creato con `O_EXCL`, e non
sovrascrive le evidenze v1/v2.

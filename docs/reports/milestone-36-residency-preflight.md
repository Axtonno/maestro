# Milestone 36 — Preflight convivenza e residency

Data: 2026-09-05. Esito: PASS, senza cambi di residency o generazioni.

- Ollama 0.33.1 e digest dei due modelli coincidono con M33/M35;
- GPU verificata: NVIDIA GeForce RTX 5070, 12.227 MiB;
- RAM host: 32.658.612 KiB; RAM WSL assegnata: 15.933.084 KiB;
- swap WSL: 4.194.304 KiB, inizialmente libera;
- tre cicli indipendenti, sei transizioni e un tentativo per step;
- unload esplicito prima e dopo ogni ciclo;
- nessun retry, repair o fallback;
- snapshot `/api/ps`, NVML, memoria host, swap, PID, RSS e start time;
- report live assente e creato con `O_EXCL` prima della prima transizione;
- `go test ./scripts/m36residency`: PASS.

Soglie temporali congelate prima delle run:

| Transizione | Massimo |
| --- | ---: |
| Direct Chat cold | 30.000 ms |
| Direct Chat warm | 10.000 ms |
| Chat → mutation | 30.000 ms |
| Mutation warm | 10.000 ms |
| Mutation → chat | 30.000 ms |
| Chat warm dopo il ritorno | 10.000 ms |

Freeze SHA-256:

| Artefatto | Digest |
| --- | --- |
| Profilo residency | `274f5dfc3c422e5d219f21555005f963978c8603e761e5a376ac1ae6fd558b59` |
| Prompt mutativo | `594659d52ec6142a5ef79c36dc0db4899e7ef1bb3f99d05017410f68bc1ba732` |
| Schema mutativo | `bc3432a8f19867eec8e153adaa4434b688974cf34d24b6bd770e887e0dd7557d` |

Il gate accetta una eviction coerente del modello precedente. Respinge pattern
diversi fra cicli, fallback incrociati o CPU, offload inatteso, crescita dello
swap, OOM, restart del provider, transizioni fallite e soglie superate.

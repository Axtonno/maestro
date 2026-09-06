# Milestone 36 — Residency con handoff esplicito

Data: 2026-09-05. Verdetto: **`dual_model_residency_rejected`**.

Il profilo v3 scarica esplicitamente il modello uscente, attende `/api/ps`
vuoto e solo dopo avvia il profilo entrante. Modelli, prompt, schema, richieste,
tre cicli e soglie temporali restano quelli congelati prima del v2.

| Transizione | Intervallo osservato |
| --- | ---: |
| Direct Chat cold | 3.431–5.745 ms |
| Direct Chat warm | 65–164 ms |
| Chat → mutation | 3.414–6.476 ms |
| Mutation warm | 362–381 ms |
| Mutation → chat | 3.816–4.651 ms |
| Chat warm dopo il ritorno | 66–72 ms |

Tutte le 18 richieste sono corrette e sotto soglia. Dopo ogni completion è
residente esclusivamente il modello richiesto; il pattern è identico nei tre
cicli. `qwen3.5:9b` usa 5.490.081.790 byte di modello interamente in VRAM e la
GPU riporta circa 6.588–6.590 MiB usati. `qwen2.5-coder:14b` usa
9.470.098.799 byte interamente in VRAM e la GPU riporta circa 9.368–9.370 MiB.
Non emergono fallback incrociati, offload CPU, OOM o transizioni fallite.

Il processo provider Linux rimane stabile per l'intera run; PID e start tick
sono acquisiti in ogni snapshot. Gli helper di caricamento non sono confusi
con il server root.

Il gate resta respinto perché lo swap WSL cresce anche con l'handoff gestito:
da zero fino a circa 487 MiB nel terzo ciclo, pur con diversi GiB di RAM WSL
disponibile. L'host usa la configurazione WSL predefinita, senza `.wslconfig`,
con 15.933.084 KiB assegnati, 4 GiB di swap e `vm.swappiness=60`.

L'eviction e la latenza sono quindi qualificate sul piano funzionale, ma il
profilo hardware non supera il vincolo di swap. M36 resta aperta e v0.5.0 non
è autorizzata. Prima di una nuova misura serve una decisione esplicita sul
profilo WSL, per esempio un reference profile con swap disabilitato; le soglie
temporali non possono essere rilassate.

Evidenza: `milestone-36-residency-runs-v3.json`, SHA-256
`9aa326f9f9e84bcad928dadf254b25001fb47bcd4deb1174a0b3df807a16b77d`.

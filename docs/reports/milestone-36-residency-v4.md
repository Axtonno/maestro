# Milestone 36 — Residency con swap WSL disabilitato

Data: 2026-09-05. Verdetto: **`dual_model_residency_qualified`**.

Il profilo v4 ripete la sequenza v3 con handoff esplicito: scarica il modello
uscente, attende `/api/ps` vuoto e avvia il profilo entrante. Le soglie
temporali non cambiano. La sola variazione del reference environment è
`.wslconfig` con `swap=0`, congelato da SHA-256
`f5e1a679dbeca06a712c4f098ca3d17ae3f3c2eb19d20a224904b500e23e4cd6`.

| Transizione | Intervallo osservato |
| --- | ---: |
| Direct Chat cold | 3.368-5.699 ms |
| Direct Chat warm | 64-75 ms |
| Chat -> mutation | 6.049-7.766 ms |
| Mutation warm | 362-364 ms |
| Mutation -> chat | 4.315-4.770 ms |
| Chat warm dopo il ritorno | 65-73 ms |

Tutte le 18 richieste completano con il modello richiesto, output corretto e
latenza sotto soglia. Dopo ogni completion è residente un solo modello:
`qwen3.5:9b` per Direct Chat oppure `qwen2.5-coder:14b` per Controlled
Mutation. Il pattern di residency è identico nei tre cicli.

`qwen3.5:9b` resta interamente in VRAM con `size_vram=5.490.081.790` byte e
la GPU riporta circa 6.448 MiB usati. `qwen2.5-coder:14b` resta interamente
in VRAM con `size_vram=9.470.098.799` byte e picco GPU osservato di 9.228 MiB.
La RAM WSL disponibile non scende sotto 13.893.772 KiB.

Non emergono fallback incrociati, offload CPU, OOM, restart del provider,
transizioni fallite o crescita dello swap WSL. Il processo provider Linux
rimane stabile con PID 162 per tutta la run.

Il gate di residency è quindi qualificato per il reference profile
`docs/milestone-36-residency-profile-v4.yaml`. La residenza simultanea non è
richiesta; la strategia productizzabile è eviction gestita e osservabile.

Evidenza: `milestone-36-residency-runs-v4.json`, SHA-256
`9cd7b4ae60fcacc9a97c62faabb2877a0ba3e0a4dc3336f910fb0cffc768c8b1`.

## Productization gate

Il candidate v4 costruito dal commit `cb2a408` supera il packaging con doppio
archive byte-identico, installazione fuori checkout e `doctor --mode all`.
Le prove live confermano allow applicato, deny senza effetti, abstain per
informazione insufficiente, `stale_source` dopo modifica concorrente e il
roundtrip `qwen3.5:9b` → `qwen2.5-coder:14b` → `qwen3.5:9b`. Nessuna prova
usa fallback incrociato o modifica fuori selezione.

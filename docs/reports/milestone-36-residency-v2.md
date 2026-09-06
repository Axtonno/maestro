# Milestone 36 — Residency v2

Data: 2026-09-05. Verdetto: **`dual_model_residency_rejected`**.

La misura corretta del processo Linux conferma che Ollama usa un solo modello
alla volta e applica eviction automatica in entrambe le direzioni. Tutte le 18
richieste sono corrette e sotto soglia; `size_vram == size` per entrambi i
modelli, quindi non emerge offload CPU. Il modello atteso è residente dopo
ogni completion e non esistono fallback incrociati o OOM.

Il gate è respinto perché lo swap WSL cresce durante le transizioni, fino a
circa 331 MiB osservati. Il contatore registra 15 snapshot sopra la baseline
del rispettivo ciclo. La classificazione iniziale considera inoltre due helper
`ollama` di discovery come processi root; il PID del vero server, 161 con start
tick 2796, resta stabile. Anche correggendo questo falso positivo, lo swap è
una failure sufficiente e il verdetto resta respinto.

M36 introduce quindi una strategia di handoff esplicita: unload del modello
uscente, attesa di `/api/ps` vuoto e solo dopo caricamento del modello entrante.
Il profilo v3 mantiene richieste, cicli e soglie temporali invariati.

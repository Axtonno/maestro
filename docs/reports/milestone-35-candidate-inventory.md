# Milestone 35 — Inventario candidati

Data: 2026-09-05

Hardware target: Linux `amd64`, NVIDIA GeForce RTX 5070 con 12.227 MiB VRAM,
Ollama 0.33.1. Il profilo comune usa context 4096, massimo 1024 token di
output, temperatura zero, thinking disabilitato, structured JSON schema e
nessun tool. La residency di ciascun modello è 5 minuti; i candidati vengono
eseguiti in sequenza.

## Criteri fissati prima del confronto

Un candidato deve essere un modello instruct orientato al codice, disponibile
nel catalogo Ollama, utilizzabile localmente sul target senza distribuire i
pesi nel package Maestro, e avere licenza documentata. La dimensione dei pesi
deve lasciare margine rispetto ai 12 GB di VRAM con context 4096. Structured
output viene verificato nel confronto e non assunto dal marketing del modello.

Shortlist:

| Profilo | Dimensione Ollama | Quantizzazione | Licenza | Motivo |
| --- | ---: | --- | --- | --- |
| `qwen2.5-coder:7b` | 4,7 GB | Q4_K_M | Apache-2.0 | baseline code-specific già locale |
| `qwen2.5-coder:14b` | 9,0 GB | Q4_K_M | Apache-2.0 | profilo più capace della stessa famiglia entro VRAM |
| `granite-code:8b-instruct` | 4,6 GB | Q4_0 | Apache-2.0 | famiglia indipendente orientata a code fixing |

Fonti primarie/catalogo: [Qwen2.5-Coder su Ollama](https://ollama.com/library/qwen2.5-coder),
[profilo Qwen 14B](https://ollama.com/library/qwen2.5-coder:14b),
[Granite Code 8B Instruct](https://ollama.com/library/granite-code:8b-instruct),
[model card Qwen 14B](https://huggingface.co/Qwen/Qwen2.5-Coder-14B-Instruct),
[model card Granite](https://huggingface.co/ibm-granite/granite-8b-code-instruct-4k).

Esclusioni prima delle run:

- `qwen3.5:9b`: profilo mutativo respinto da M34; ulteriori tuning vietati;
- `qwen2.5-coder:32b` (20 GB), Granite Code 20B (12 GB) e modelli maggiori:
  pesi pari o superiori alla VRAM disponibile, senza margine per runtime/KV;
- modelli base/FIM: esclusi perché il contratto richiede una decisione instruct;
- modelli con licenza non Apache: esclusi da questa prima selezione per tenere
  semplice il profilo operativo e la documentazione di adozione. Questa scelta
  non afferma che le altre licenze siano incompatibili con Maestro.

La shortlist confronta due scale Qwen e una famiglia indipendente. Nessuna
reputazione o benchmark esterno determina il vincitore: contano soltanto gli
output della matrice congelata M35. Latenza aggregata è tie-breaker dopo i gate.

## Regole della selezione

La matrice contiene 9 trasformazioni positive e 3 astensioni necessarie,
nuove rispetto a M33. Ogni profilo riceve gli stessi 12 casi nello stesso
ordine, una volta. Nessun retry, repair, fallback o modifica del prompt dopo
la prima generazione. Il report viene creato in esclusiva e salvato dopo ogni
caso.

Per essere eleggibile: output conformi 12/12, positivi corretti almeno 9/9
(la soglia 90% arrotondata sul denominatore disponibile), astensioni 3/3.
Se più profili sono eleggibili, vince la latenza aggregata minore. Se nessuno
è eleggibile, M35 termina senza modello selezionato e senza qualifica.

Il set di selezione non viene riutilizzato nella qualifica. Dopo la scelta si
scrive e congela una nuova matrice development/holdout; il modello selezionato
deve poi superare tutti i gate M34. Se fallisce, M35 è chiusa con selezione
tecnica ma Controlled Mutation resta non qualificata.

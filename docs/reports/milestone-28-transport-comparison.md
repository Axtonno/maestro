# Milestone 28 – Confronto dei trasporti

Data: 2026-09-03

Stato: **COMPLETATO SENZA SELEZIONE LIVE**

I due trasporti congelati convergono sul medesimo contratto v1 e sul medesimo compilatore:

| Trasporto | Envelope | Validità deterministica | Fingerprint equivalente | Fallback |
|---|---|---:|---:|---:|
| `native_tool_call` | `name=workspace_replace`, `arguments=<proposal-v1>` | PASS | PASS | vietato |
| `constrained_structured_output` | `<proposal-v1>` | PASS | PASS | vietato |

La suite respinge prosa più JSON, valori JSON multipli, nome tool errato, campi wrapper aggiuntivi, output cross-transport, proposta malformata e transport ID sconosciuto. Sulla fixture positiva entrambi producono lo stesso fingerprint e contenuto candidato.

Questo risultato qualifica soltanto sintassi, normalizzazione e convergenza deterministica. Non misura la capacità di `qwen3.5:9b` di produrre modifiche semanticamente corrette: Ollama non è installato e `127.0.0.1:11434` non è raggiungibile. Di conseguenza nessun trasporto è selezionato e non è applicabile la soglia semantica `>= 0.80`.

Verdetto del confronto: `transport_equivalent_deterministically_live_selection_unavailable`.

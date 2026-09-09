# Capacità correnti di Maestro v0.5.0

Aggiornata: 2026-09-09

Questa pagina definisce il support claim pubblico. Codice sperimentale,
esempi e direzioni future non ampliano questo perimetro.

## Cosa funziona oggi

| Capacità | Perimetro supportato |
| --- | --- |
| Direct Chat | Domanda senza file oppure su un solo file scelto esplicitamente |
| Controlled Mutation | Una sostituzione su un intervallo di un file PHP regolare sotto `app/` |
| Autorizzazione | Preview completa e allow-once obbligatoria in una TTY reale |
| Integrità | Controllo stale, fingerprint della preview e sostituzione atomica |
| Provider | Ollama 0.33.1 locale su loopback |
| Artifact | `maestro-v0.5.0-linux-amd64.tar.gz` |

Controlled Mutation è supportato soltanto nel [perimetro dichiarato](controlled-mutation.md).

## Modelli qualificati

| Uso | Modello | Digest richiesto |
| --- | --- | --- |
| Chat | `qwen3.5:9b` | `6488c96fa5faab64bb65cbd30d4289e20e6130ef535a93ef9a49f42eda893ea7` |
| Mutation | `qwen2.5-coder:14b` | `9ec8897f747e246e970bc5cfdda85d22f1123dc2e3d34978a010a75968716849` |

Altri modelli, digest o provider non fanno parte del support claim v0.5.0. Un
profilo single-model è in valutazione, ma non è supportato oggi.

## Hardware consigliato

Come punto di partenza pratico, non come minimo universale:

- Linux `amd64` nativo;
- CPU x86-64 con almeno 4 core / 8 thread;
- 16 GiB di RAM;
- circa 20 GiB liberi per archive, modelli e margine operativo;
- swap disponibile; GPU discreta opzionale.

La release è stata qualificata anche su Linux nativo CPU-only. Latenza e
memoria dipendono dall’hardware; vedere [Benchmark](benchmarks.md).

## Sperimentale, non parte del prodotto v0.5.0

- componenti agent, context e tool presenti nel codice;
- streaming nei profili storici;
- integrazione editor;
- profilo single-model.

La presenza del codice non costituisce una promessa operativa.

## Non supportato oggi

- modifiche multi-file, più intervalli o sequenze autonome;
- scelta autonoma di file, righe o azioni;
- agent autonomi, retrieval multi-file e tool calling come prodotto;
- Windows nativo, macOS e Linux `arm64`;
- endpoint remoti o modelli diversi da quelli qualificati;
- mutation fuori `app/`, su linguaggi non PHP o senza TTY;
- esecuzione di shell, Git, Docker, Composer o test da parte del modello;
- sandbox, rollback automatico o garanzia di correttezza semantica.

Windows e macOS sono [direzioni in valutazione](roadmap.md), non piattaforme
promesse.

## Da dove iniziare

Seguire il [Quick Start](quick-start.md). Per l’archive pubblico v0.5.0 usare
le [note di release](releases/v0.5.0.md); per diagnosi consultare
[Troubleshooting](troubleshooting.md).

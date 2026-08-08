# Provider API Compatibility Audit — Milestone 2

Versione: 0.1.0

Stato: Superato

Data: 2026-08-08

---

# Perimetro

L'audit copre `pkg/provider`, le facade pubbliche `pkg/provider/ollama` e
`pkg/provider/llamacpp`, il routing del Provider Runtime e i contratti aggiunti
nelle Fasi 1–9.

# Esito

- Le capability interface esistenti conservano firme e ownership.
- `Runtime.Complete` e `Runtime.Stream` conservano firme e selezione del
  provider; la validazione avanzata avviene prima del routing remoto.
- Ollama e llama.cpp conservano costruttori, configurazioni e ID pubblici.
- Le richieste con i nuovi campi a zero producono i payload semplici precedenti.
- Errori, context, stream pull-based, resilienza, residency e observability
  conservano la semantica documentata.
- Nessun SDK o modulo esterno è stato aggiunto al grafo delle dipendenze.

# Modifiche additive

`CompletionRequest` aggiunge `Options`, `Output`, `Tools` e `ToolChoice`.
`Message` aggiunge riferimenti e chiamate tool. `StreamChunk` aggiunge delta tool.
Nuovi tipi e costanti modellano sampling, output strutturato, tool e finish
reason.

La versione del progetto è ancora `0.1.0`. L'aggiunta di campi è compatibile
con accesso per nome e composite literal con chiavi, ma interrompe composite
literal posizionali di struct esportate. La regola pubblica da questa milestone
è quindi: esempi, test e consumer devono costruire i value object provider con
campi nominati. Un futuro `v1` dovrà trattare l'aggiunta di campi alle struct
pubbliche come modifica da valutare esplicitamente.

# Validazione

Il gate è coperto da:

- suite completa senza servizi esterni;
- race detector sull'intera repository;
- `go vet` sull'intera repository;
- compilazione delle suite protette dal tag `integration` senza servizi live;
- fixture HTTP deterministiche per entrambi gli adapter;
- test di error mapping, cancellazione, stream, residency, retry, circuit
  breaker, osservabilità, structured output e tool calling;
- `git diff --check` e audit della documentazione.

# Rischi residui assegnati

- Differenze tra versioni e modelli Ollama/llama.cpp: Smoke Benchmark,
  Milestone 3.
- Prestazioni e memoria del buffering JSON terminale: Runtime Benchmark,
  Milestone 3.
- Evoluzione semver verso `v1`: futura milestone di stabilizzazione API.
- Multimodalità e reasoning: evoluzione Provider Layer successiva, fuori dal
  gate della Milestone 2.

Non restano rischi residui senza destinazione documentata.

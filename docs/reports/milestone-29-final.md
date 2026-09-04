# Milestone 29 — Report finale

Data: 2026-09-04

Stato: **COMPLETATA — TRASPORTO RESPINTO**

Verdetto: **`controlled_mutation_model_transport_rejected`**

## Risultato

Il preflight dell'ambiente congelato è PASS. Le venti run formali sono state eseguite una sola volta, nell'ordine alternato stabilito, con fixture indipendente e senza fallback, repair o retry selettivi. Nessun trasporto supera tutti i gate; non viene selezionato un trasporto e un candidate v0.5.0 non è autorizzato.

| Trasporto | Completion | Proposte attese valide | Correttezza positiva | Sicurezza | Esito |
|---|---:|---:|---:|---:|---|
| `native_tool_call` | 2/10 | 0/7 | 0/3 | 0 failure | FAIL |
| `constrained_structured_output` | 7/10 | 7/7 | 3/3 | 0 failure | FAIL |

Structured output ha soddisfatto byte-per-byte i tre casi positivi. Ha però prodotto una proposta anche per T04, dove era richiesta astensione, e proposte applicabili ma semanticamente inventate per T05/T06; perciò non raggiunge il 90% di completion. Il tool calling nativo ha restituito tool call nei casi provider-facing, ma gli arguments non sono stati accettati dal decoder strict; soltanto i due reject pre-provider T07/T08 hanno raggiunto il terminale atteso.

Tutte le failure hanno conservato il workspace corretto. Non sono avvenute mutazioni senza approval, mutazioni fuori scope, stale write accettate o effetti residui. Le proposte applicate sono rimaste confinate alla materializzazione della singola run.

## Verifiche

- harness e package `internal/mutation`: PASS su Linux amd64;
- suite completa su copia Linux LF pulita: PASS;
- race detector e `go vet` su copia Linux LF pulita: PASS;
- audit del report: nessun contenuto workspace o output modello grezzo persistito;
- raw redatto: `docs/reports/milestone-29-live-runs.json`.

## Handoff

M29 non autorizza release readiness né packaging v0.5.0. Un nuovo tentativo richiede una nuova milestone e un nuovo freeze esplicito: le run M29 non possono essere ripetute o reinterpretate dopo tuning.

# Milestone 14 — Phase 2 Report

Data: 2026-08-27

Stato: **COMPLETATA**

## Esito

Maestro possiede ora profili generativi distinti e strict per direct chat e
verified agent. Lo schema v2 è development-only; lo schema v1 continua a
caricare con semantica invariata e non acquisisce un profilo chat implicito.

## Implementazione

- `productconfig.Load` seleziona decoder strict distinti per v1 e v2 e
  normalizza il solo profilo agent per i consumer esistenti;
- `configs/maestro.interaction.example.yaml` congela i profili candidati chat e
  agent, inclusi timeout, streaming, `num_ctx`, `thinking` e limiti chat;
- `provider.GenerationOptions` espone `ContextWindow` e thinking tri-state con
  validazione provider-neutral;
- il native adapter Ollama mappa `ContextWindow` in `options.num_ctx` e thinking
  esplicito nel campo top-level `think`, distinguendo omissione e false;
- l'adapter llama.cpp rifiuta questi controlli per-request prima dell'I/O;
- capability introspection e Gestor espongono
  `context_window_control`, `thinking_control` e `thinking`;
- `ValidateGenerationCapabilities` fallisce il preflight quando un controllo
  esplicito non è disponibile;
- il profilo agent viene propagato a policy, request, ogni turno del loop,
  benchmark mutativo e doctor, senza cambiare l'output doctor v1.

Il mapping segue la native chat API e il contratto `think` documentati da
Ollama in `docs.ollama.com/api/chat`, `docs.ollama.com/capabilities/thinking` e
nel relativo OpenAPI. Il response protocol non attesta sempre il valore
effettivo del runner: il requested e il body mappato sono osservabili, mentre
un valore non attestabile resta `unknown`.

## Compatibilità e sicurezza

- i file v1 pubblicati dalla v0.2.0 restano validi e conservano modello,
  streaming e timeout precedenti;
- i campi legacy `models.chat` e `agent.streaming` sono rifiutati nel decoder
  v2 invece di essere ignorati;
- `thinking: false` resta distinto da `default` e viene serializzato anche
  quando false;
- llama.cpp non simula context o thinking tramite configurazione di processo;
- nessun prompt, response o contenuto workspace è stato aggiunto a log o
  osservabilità.

## Gate verificato

| Controllo | Esito |
|---|---|
| schema v1 e v2 strict | PASS |
| range, enum e timeout profile | PASS |
| request Ollama non-streaming e streaming condivisa | PASS |
| thinking default/true/false | PASS |
| rifiuto llama.cpp prima dell'I/O | PASS |
| preflight capability model-specific | PASS |
| propagazione in ogni turno agent | PASS |
| `go test ./...` | PASS |
| `go test -race ./...` | PASS |
| `go vet ./...` | PASS |
| `git diff --check` | PASS |

## Gate

**PASS.** Le opzioni richieste sono validate e mappate esattamente oppure
rifiutate; i profili sono separati e la baseline v1 resta compatibile. La Fase
3 può costruire il percorso chat single-file senza dipendenze agentiche.

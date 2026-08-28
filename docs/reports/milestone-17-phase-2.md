# Milestone 17 — Fase 2: confine del servizio Direct Chat

Data: 2026-08-28

Stato: **COMPLETATA — PASS**

Baseline di ingresso: commit `b1c85e4`

## Obiettivo verificato

Il percorso `maestro chat` compone soltanto il servizio Direct Chat e un
provider generativo. Non costruisce l'application graph agentico e non riceve
Agent Runtime, Context Engine runtime, Tool Runtime, sessione, index, approver
o callback di fallback.

## Modifiche

- la factory provider predefinita è stata spostata nel confine
  `internal/directchat`;
- la factory usa `interaction.chat.model` come modello di default dell'adapter,
  non il modello agent;
- `cmd/maestro/chat_command.go` non importa più `internal/application`;
- le factory iniettate dai test restano supportate senza ampliare le
  dipendenze del servizio;
- un test di composizione conta una sola factory, un solo preflight e una sola
  completion con `Tools` vuoto e `ToolChoiceNone`;
- test architetturali analizzano gli import dei file production e bloccano
  dipendenze da agent, composition root, context engine runtime e tool runtime.

Il loader usa ancora il solo tipo pubblico `pkg/contextengine.DocumentPath`
per validare il path logico. Non importa né costruisce
`internal/contextengine.Engine`; l'hardening del path appartiene alla Fase 3.

## Graph osservato

```text
CLI chat
  -> product config già caricata
  -> directchat.Build
      -> secret reference
      -> provider Ollama oppure llama.cpp
  -> capability preflight
  -> Complete oppure Stream
```

Non esiste un ramo verso `application.Build`, `maestro.New`, plugin Laravel,
workspace index, registry di tool, policy, approver o agent loop.

## Matrice

| Controllo | Esito |
|---|---|
| factory provider | 1 chiamata |
| capability preflight | 1 chiamata per esecuzione |
| completion non-streaming | 1 chiamata |
| stream inattesi | 0 |
| tool dichiarati | 0 |
| `tool_choice` | `none` |
| fallback dopo failure/response invalida | 0 |
| import agent/application/context runtime/tool runtime | 0 |
| regressione `agent`/`run` | suite verde |

## Verifiche

| Comando | Esito |
|---|---|
| `go test ./...` | PASS |
| `go vet ./...` | PASS |
| `go test -race -count=3 ./cmd/maestro ./internal/directchat` | PASS |
| test mirati `cmd/maestro`, `internal/directchat`, `internal/application` | PASS |
| `git diff --check` | PASS |

## Gate di uscita

- una request provider e zero tool: **PASS**;
- nessun servizio agentico/retrieval/mutativo costruito: **PASS**;
- failure senza secondo percorso: **PASS**;
- test architetturali e regressione agent: **PASS**;
- workspace invariato: **PASS** tramite matrice Direct Chat esistente.

Verdetto della fase: **PASS**. La Fase 3 può iniziare.

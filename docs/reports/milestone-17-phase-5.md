# Milestone 17 — Fase 5: streaming, terminali e osservabilità

Data: 2026-08-29

Stato: **COMPLETATA — PASS**

Baseline di ingresso: commit `9505e16`

## Obiettivo verificato

Complete e stream condividono lo stesso contratto di risultato: una sola
request tool-free, contenuto UTF-8 non vuoto entro limite, terminale `stop`,
usage valido e output pubblicato soltanto dopo la validazione completa.

## Hardening consegnato

- gli stream acquisiti vengono chiusi esattamente una volta su successo e su
  ogni failure di open, contesto, receive, validazione, limite o close;
- una risposta o un chunk con modello osservato diverso da quello richiesto è
  invalido; modello assente resta non attestato e non viene inventato;
- token input/output negativi sono response invalid;
- terminale mancante, non-`stop`, duplicato o seguito da un chunk è invalido;
- tool call complete o delta restano un protocol failure;
- lo stdout viene costruito soltanto dal `Result` finale, quindi un failure di
  stream non pubblica chunk parziali;
- gli errori del parser flag non emettono più diagnostica raw o usage su
  stdout: restituiscono soltanto `chat failed: invalid_request`;
- domanda vuota, invalida, concorrente o oltre 1 MiB fallisce prima del loader
  config e della provider factory;
- cancellazione ha precedenza coerente su deadline nella coppia reason/exit;
- un successo validato con finish reason `stop` espone `truncated=false`.

## Matrice streaming

| Scenario | Esito |
|---|---|
| complete/stream stesso contenuto e usage | PASS |
| stream capability/profile disabilitati | fail-closed prima di generation |
| stream nil | `response_invalid` |
| terminale mancante o length | `response_invalid` |
| tool delta | `response_invalid` |
| UTF-8 invalido o NUL | `response_invalid` |
| modello mismatch/drift | `response_invalid` |
| usage negativo | `response_invalid` |
| output oltre limite | `limit_exceeded` |
| open/receive/close failure | `provider_unavailable` |
| chunk dopo terminale | `response_invalid` |
| close count per stream acquisito | esattamente 1 |
| fallback complete dopo stream failure | 0 |

## Matrice CLI e terminali

| Classe | Reason code | Exit | stdout su failure |
|---|---|---:|---|
| usage, config, domanda o file | `invalid_request` / `file_not_allowed` | 2 | vuoto |
| provider o capability | `provider_unavailable` / `capability_unsupported` | 4 | vuoto |
| deadline | `deadline_exceeded` | 4 | vuoto |
| cancellazione | `canceled` | 130 | vuoto |
| response/modello/usage invalido | `response_invalid` | 1 | vuoto |
| hard output limit | `limit_exceeded` | 1 | vuoto |
| successo | terminale `completed`, finish `stop` | 0 | envelope atomico |

Unknown flag, boolean invalido, flag duplicato, `--file=` vuoto, domanda vuota,
domanda oltre limite e positional più stdin sono coperti con confronto
byte-esatto del failure sintetico e zero provider composition.

## Osservabilità e anti-leak

L'envelope di successo espone soltanto modalità, terminale, modello validato,
durata non negativa, usage non negativo, `num_ctx` e thinking richiesti,
effettivi `unknown`, `truncated=false`, finish reason e risultato richiesto.

I failure vengono scanditi con canary per domanda, response, contenuto file,
path logico, root fisica e secret. Nessun canary compare su stdout/stderr; non
esistono eventi o log payload nel servizio Direct Chat. Il contenuto finale di
successo resta intenzionalmente visibile all'utente.

## Verifiche

| Comando | Esito |
|---|---|
| `go test ./...` | PASS |
| `go test -tags maestro_development ./...` | PASS |
| `go test -race ./...` | PASS |
| `go vet ./...` | PASS |
| `go test -count=10 ./internal/directchat ./cmd/maestro` | PASS |
| `git diff --check` | PASS |

## Gate di uscita

- equivalenza deterministica complete/stream: **PASS**;
- output atomico su failure: **PASS**;
- cancel, deadline e hard limit distinti: **PASS**;
- sink pubblicabili senza payload proibiti: **PASS**;
- suite streaming, terminali e anti-leak: **PASS**.

Verdetto della fase: **PASS**. La Fase 6 può iniziare.

# Direct Chat Deterministic and Negative Matrix

Versione: 1.0.0

Data: 2026-08-28

Stato: Milestone 14 — Fase 4

Questa matrice qualifica il contratto della superficie development-only
`maestro chat`. Le prove sono locali, non avviano provider e non acquisiscono
modelli. Il gate live C0–C4 resta separato.

| Area | Scenari | Oracolo |
|---|---|---|
| input CLI | domanda positional, stdin, conflitto positional/stdin, vuoto, flag duplicati, `--file=` | accettazione univoca o `invalid_request` prima del provider |
| configurazione | v1, unknown/duplicate v2, `num_ctx` fuori range, thinking invalido, streaming non abilitato | `chat_profile_required`, `invalid_request` o `capability_unsupported` |
| containment | assoluto, traversal, backslash, path non normalizzato, file assente | `file_not_allowed`, zero I/O provider |
| tipi fisici | root symlink, symlink interno/evasivo, directory, FIFO, file non regolare | `file_not_allowed`, zero I/O provider |
| stabilità | contenuto mutato, file sostituito, root sostituita durante la lettura | `file_not_allowed`, nessuna disclosure |
| limiti file | oltre byte limit, UTF-8 invalido, NUL | `file_not_allowed`, zero I/O provider |
| request | file esplicito e assente, context/thinking dedicati | una sola generation, zero tool, `tool_choice: none` |
| response | ruolo errato, vuota, UTF-8 invalido, NUL, tool call, finish non-stop, oltre limite | failure chiuso, nessun fallback, stdout vuoto |
| streaming | terminale assente/duplicato, chunk post-terminale, tool delta, errore receive/close, oltre limite | failure atomico, nessun contenuto parziale |
| equivalenza | stessi content, usage, finish reason, modello e request controls | non-streaming e streaming equivalenti |
| operatività | provider failure, profile deadline, cancel, contesto già scaduto | reason ed exit code distinti, shutdown bounded |
| anti-leak | domanda, file, path logico/fisico e response parziale come canary | canary assenti da stdout/stderr dei failure |
| immutabilità | successo, no-file e containment failure | snapshot workspace byte-identico |
| regressione | `maestro agent` e alias `maestro run`, suite repository | output/autorità invariati |

## Harness autorevole

- `internal/directchat/loader_test.go` copre containment, tipi fisici e race
  deterministiche tramite hook soltanto test;
- `internal/directchat/service_test.go` cattura request, terminali, limiti,
  timeout, streaming, equivalenza e snapshot workspace;
- `cmd/maestro/main_test.go` copre parser, envelope atomico, exit/reason code,
  assenza di output parziale e canary anti-leak;
- `internal/productconfig/config_test.go` copre parsing v2 strict, enum e range.

Il servizio di produzione non espone gli hook del loader. Il suo dependency
contract contiene soltanto provider factory, secret lookup e clock: retrieval,
index, Agent Runtime, Tool Runtime, sessione, approver e fallback non sono
costruibili dal percorso.

## Regola streaming

Lo streaming richiede contemporaneamente `interaction.chat.streaming: true`,
flag `--stream`, interfaccia provider e capability model-specific disponibili.
I chunk sono assemblati entro `max_output_bytes` e non vengono pubblicati
progressivamente: stdout viene scritto soltanto dopo un singolo terminale
`stop` seguito da EOF. Questa atomicità rende i failure osservazionalmente
equivalenti al percorso non-streaming.

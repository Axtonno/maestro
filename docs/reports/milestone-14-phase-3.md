# Milestone 14 — Phase 3 Report

Data: 2026-08-27

Stato: **COMPLETATA**

## Esito

`maestro chat` esegue ora una completion diretta single-file, bounded e
non-streaming. Il percorso non costruisce il composition graph agentico e non
riceve tool runtime, retrieval, index, sessione, approver o callback di
fallback.

## Implementazione

- il package `internal/directchat` compone soltanto provider completion e
  capability inspection dal profilo chat v2;
- il loader usa una root filesystem confinata, rifiuta path logici invalidi,
  componenti symlink e file non regolari e verifica identità, dimensione,
  contenuto, UTF-8 e NUL prima della disclosure;
- ogni file accettato viene letto due volte dallo stesso handle e confrontato
  con il path confinato prima e dopo la lettura;
- il prompt separa domanda, istruzioni epistemiche e contenuto workspace non
  fidato; senza file dichiara esplicitamente l'assenza di contesto;
- la request provider contiene zero tool e `tool_choice: none` ed effettua una
  sola completion;
- response vuote, non finali, non UTF-8, con NUL, tool call o oltre il limite
  sono rifiutate senza fallback;
- il comando espone un envelope stabile con modalità, terminale, modello,
  durata, usage, `num_ctx`, `thinking`, truncation e finish reason; i valori
  effettivi non attestati restano `unknown`;
- errori CLI espongono soltanto reason code sintetici e exit code definiti da
  ADR-0033.

## Compatibilità e sicurezza

- `maestro agent` è il comando canonico del verified agent;
- `maestro run` resta il suo alias esatto e deprecato, con stesso parser,
  output ed exit code;
- una configurazione v1 continua a funzionare per agent/run ma chat la rifiuta
  con `chat_profile_required`;
- path fisici e contenuti rifiutati non entrano negli errori operativi;
- streaming resta disabilitato fino alla matrice di equivalenza della Fase 4.

## Gate verificato

| Controllo | Esito |
|---|---|
| una completion per richiesta valida | PASS |
| zero tool e `tool_choice: none` | PASS |
| loader single-file confinato e bounded | PASS |
| percorso senza file e istruzione epistemica | PASS |
| nessun fallback su provider/response failure | PASS |
| output e reason code redatti | PASS |
| regressione `agent`/`run` | PASS |
| `go test ./...` | PASS |
| `go test -race ./...` | PASS |
| `go vet ./...` | PASS |
| `git diff --check` | PASS |

## Gate

**PASS.** Il percorso chat non-streaming è separato, single-file e fail-closed.
La Fase 4 può estendere la matrice negativa e abilitare streaming soltanto
dopo equivalenza deterministica.

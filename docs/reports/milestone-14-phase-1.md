# Milestone 14 — Phase 1 Report

Data: 2026-08-27

Stato: **IN CORSO**

## Obiettivo

Congelare tramite ADR la separazione fra `maestro chat` e `maestro agent`, la
relazione con l'attuale `maestro run`, il contratto CLI e la migrazione della
configurazione prima di modificare provider o application graph.

## Baseline verificata

Baseline sorgente: commit
`3dd5e62bbce77a75519e784e749c338cc9685b75`.

| Controllo | Esito |
|---|---|
| `go test ./...` | PASS |
| `go test -race ./...` | PASS |
| `go vet ./...` | PASS |
| `git diff --check` | PASS |

Le suite Go sono state eseguite con cache isolate sotto `/tmp` perché la cache
predefinita dell'ambiente è read-only. Il primo tentativo non ha raggiunto i
test ed è classificato come deviazione dell'ambiente di esecuzione, non come
failure della baseline.

## Inventario iniziale

- la CLI pubblica espone `maestro run`, che costruisce ed esegue il reference
  agent configurato; `chat` e `agent` non sono ancora comandi;
- la configurazione v1 possiede un solo `models.chat`, un solo timeout provider
  e un blocco agent, quindi non può esprimere due profili completi e separati;
- `provider.CompletionRequest` espone sampling, structured output e tool, ma
  non espone ancora `num_ctx` o `thinking`;
- l'adapter Ollama traduce le generation options esistenti in `options`, ma non
  invia ancora context window o thinking;
- il composition root applicativo usa lo stesso modello per agent, planner e
  provider default e costruisce il percorso agentico per `maestro run`;
- le capability di completion e streaming esistono già a livello provider e
  possono sostenere un servizio chat separato senza introdurre un secondo
  agente.

## Decisioni ancora da congelare

- destino e finestra di compatibilità di `maestro run`;
- sintassi esatta di domanda, stdin, `--file` e streaming;
- evoluzione strict dello schema v1 e comportamento dei file esistenti;
- terminali, reason code, exit code e formato dei metadati chat;
- confine delle dipendenze del servizio chat e preflight delle opzioni
  provider-neutral.

## Stato del gate

Il gate resta **APERTO**. Baseline, inventario e piano delle sei fasi sono
disponibili; ADR e aggiornamenti dei contratti CLI/security devono essere
completati prima di avviare la Fase 2.

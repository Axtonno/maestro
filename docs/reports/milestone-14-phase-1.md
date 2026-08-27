# Milestone 14 — Phase 1 Report

Data: 2026-08-27

Stato: **COMPLETATA**

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

## Decisioni congelate

- `maestro agent` è il nome canonico del verified agent; `maestro run` resta
  alias esatto e deprecato almeno per tutta la serie v0.3.x;
- `maestro chat` accetta domanda da positional o stdin, un solo `--file`
  opzionale e `--stream` opt-in dopo il gate di equivalenza;
- lo schema strict v2 introduce profili chat e agent separati; lo schema v1
  resta valido per agent/run ma non abilita chat implicitamente;
- `num_ctx` e `thinking` devono essere mappati esattamente o rifiutati;
- il servizio chat non riceve dipendenze agentiche e non possiede fallback;
- terminali, exit code, reason code e output redatto sono congelati in
  ADR-0033 e `cli.md`.

## Deliverable

- `docs/adr/ADR-0033.md` e aggiornamento dell'indice ADR;
- contratti candidati in `cli.md`, `configuration.md`, `security-model.md` e
  `compatibility.md`;
- piano Milestone 14 dettagliato e checkpoint interno riproducibile.

## Stato del gate

**PASS.** Baseline, inventario, ADR e contratti CLI/config/security sono
coerenti; nessuna modifica applicativa o nuova autorità è stata introdotta. La
Fase 2 può iniziare dai profili generativi e dai contratti provider-neutral.

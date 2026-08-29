# Milestone 17 — Fase 4: profilo dedicato e preflight

Data: 2026-08-29

Stato: **COMPLETATA — PASS**

Baseline di ingresso: commit `03d3f62`

## Obiettivo verificato

Direct Chat dispone ora di validazione e preflight propri. Un utente può
configurare e diagnosticare la sola completion single-file senza dichiarare un
agent, tool, retrieval, index o budget di sessione non qualificati.

## Schema chat-only

Il nuovo `productconfig.LoadChat` usa lo stesso decoder YAML strict e bounded
del loader storico, ma applica `ValidateChatExecutionProfile`. Sono richiesti:

- `version: 2`;
- provider esplicito, origin HTTP(S), timeout e secret reference valida;
- workspace Laravel con root assoluta e normalizzata dopo la risoluzione;
- profilo `interaction.chat` completo;
- `policy.workspace_mutate: deny`.

Agent, tool, limits e context non sono richiesti e non vengono validati o
costruiti dal comando chat. Il loader completo `Load`, usato da `agent` e
`run`, continua invece a richiederli e respinge una configurazione chat-only.
Questa asimmetria è intenzionale e non concede autorità.

## Profilo congelato

`configs/maestro.chat.example.yaml` ha SHA-256:

`7186188ac769787afd9521a0815e58abb18952526757aa878675bdefd19ce7b1`.

| Campo | Valore |
|---|---|
| modello | `qwen2.5-coder:7b` |
| timeout provider/chat | 5m / 5m |
| streaming default | false |
| `num_ctx` | 4096 |
| thinking | `false` |
| max file/output | 1 MiB / 1 MiB |
| mutation | deny |

La root relativa dell'esempio punta a `../fixtures/laravel-v1`, cioè alla
fixture prevista nell'archive; per uso su un checkout o progetto reale deve
essere sostituita con la root autorizzata.

## Preflight dedicato

La nuova forma CLI è:

```text
maestro doctor --mode chat --config config-v2.yaml
```

Il doctor chat produce cinque check redatti:

1. schema chat v2;
2. workspace root directory e non-symlink;
3. composition del solo provider Direct Chat;
4. capability completion e streaming quando configurato;
5. controllo delle opzioni generative `num_ctx` e thinking.

Non invoca `Complete` o `Stream`. Un failure di workspace impedisce la
composition; un failure di composition impedisce il probe modello; una
capability assente non viene convertita in fallback. Il doctor senza `--mode`
mantiene il comportamento agentico storico.

## Matrice

| Scenario | Esito |
|---|---|
| chat-only minima | LoadChat PASS |
| stessa config nel loader agent | FAIL `ErrInvalid` |
| campi agent irrilevanti invalidi | ignorati soltanto da LoadChat |
| mutation `prompt` | FAIL prima della composition |
| modello vuoto | FAIL |
| timeout chat oltre trasporto | FAIL |
| `num_ctx` fuori range | FAIL |
| thinking non ammesso | FAIL |
| campo sconosciuto/duplicato | FAIL strict |
| modello/capability assente | doctor FAIL, zero generation |
| preflight valido | 5/5 PASS, zero generation |

Le request catturate nei test conservano modello, context 4096, thinking
false, `Tools` vuoto e `ToolChoiceNone`. Gli effettivi non attestabili restano
`unknown` nell'envelope e non vengono presentati come osservati.

## Documentazione

Sono stati aggiornati CLI, configuration, installation e troubleshooting con
schema chat-only, `doctor --mode chat`, requisiti e non-garanzie. Il profilo
resta candidato: README e support claim v0.2.0 non sono ancora promossi.

## Verifiche

| Comando | Esito |
|---|---|
| `go test ./...` | PASS |
| `go test -tags maestro_development ./...` | PASS |
| `go vet ./...` | PASS |
| `go test -race -count=3 ./internal/productconfig ./internal/directchat ./cmd/maestro` | PASS |
| build `-trimpath`, doctor help e mode invalida | PASS |
| `git diff --check` | PASS |

## Gate di uscita

- config incomplete/non onorabili prima della generation: **PASS**;
- valori esatti nella request provider: **PASS**;
- nessuna opzione ignorata silenziosamente: **PASS**;
- compatibilità v1/v2 e separazione agent: **PASS**;
- doctor, config suite e mapping Ollama: **PASS**.

Verdetto della fase: **PASS**. La Fase 5 può iniziare.

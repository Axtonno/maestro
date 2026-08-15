# Milestone 8 — Final Report

Data: 2026-08-15

Verdetto: **COMPLETATA — GO alla release v0.1.0**

## Risultato di prodotto

Maestro v0.1.0 può essere installato da archive su Linux `amd64`, configurato
con schema strict `version: 1`, diagnosticato e usato con Ollama per analizzare
un progetto Laravel attraverso il reference agent read-only. Il percorso
ufficiale registra soltanto list/read/search e nega le mutazioni.

La promessa non include sandbox, isolamento di processo, llama.cpp, reference
agent mutante, SDK stabile, memoria/recovery, multi-agent, remote execution o
tool shell/Git/Docker completi. Il runtime resta trusted in-process.

## Identità della release

| Campo | Valore |
|---|---|
| Versione/tag | `v0.1.0` |
| Piattaforma | Linux `amd64` |
| Artifact | `maestro-v0.1.0-linux-amd64.tar.gz` |
| Commit sorgente | `f882919798fa6073bc11c6af18a431bf249a7755` |
| SHA-256 | `c785676a177165a2c11ff0fc744931ac8b5d923466155ec32365e7a0c03d271f` |
| Dimensione | 3.604.828 byte |
| Licenza | Apache-2.0 |
| Manifest status | `release` |

Il tag annotato `v0.1.0` punta al commit sorgente riportato da
`maestro version`. L'archive è stato prodotto due volte dal medesimo commit con output
byte-identico; il checksum è stato riconfermato dopo tutti i gate live.

## Evidenza conclusiva

| Gate | Esito |
|---|---|
| Archive sicuro, documenti e configurazione read-only | Superato |
| Installazione senza checkout | Superato |
| `version`, help, manifest e checksum coerenti | Superato |
| `doctor` | 9/9 pass |
| `models` e fixture richieste | Superato |
| `agents` e `agent.reference` | Superato |
| Quick start consecutivo 1 | Exit 0, 2 turni, 1 read, risposta corretta |
| Quick start consecutivo 2 | Exit 0, 2 turni, 1 read, risposta corretta |
| Workspace invariato | Superato, digest stabile |
| Hard limit `model_turns: 1` | `limit_exceeded`, exit 1 |
| SIGINT | `canceled`, exit 130 |
| Shutdown bounded | 1.997 ms su budget 30 s |
| Assenza leakage operativo | Superato |
| Suite `-count=3` | Superata |
| Race detector | Superato |
| Vet e mod verify | Superati |
| Benchmark deterministico | 36.057 ns/op, 17.262 B/op, 223 allocs/op |

## Difetto intercettato nel gate

Una prima build dal commit documentale `6e867c1` era riproducibile ma è stata
rifiutata: una run stampava una pseudo-tool-call JSON come risposta finale. Non
è stata promossa né rinominata. Il commit finale riconosce soltanto una
struttura JSON che descrive un tool dichiarato, richiede una vera invocazione
nel turno seguente entro gli hard limit e non intercetta la semplice menzione
del tool in prosa. Entrambi i confini sono coperti deterministicamente e la
correzione è stata esercitata in entrambe le run finali.

## Contratti pubblici

CLI, configurazione e package Go sono sperimentali durante la serie 0.x. La
matrice autorevole è `docs/compatibility.md`; sicurezza e non-garanzie sono in
`docs/security-model.md`, `SECURITY.md` e `docs/known-issues.md`. La futura
validazione mutativa è rinviata almeno alla v0.2.0 e richiederà una nuova
baseline modello/hardware.

## Chiusura

Fasi 1–6, documentazione pubblica, licenza, packaging, installazione pulita,
validazione live e controlli repository-wide descrivono lo stesso prodotto.
Non risultano blocker aperti nel perimetro dichiarato della v0.1.0.

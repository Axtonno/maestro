# Milestone 9 — Report Fase 3

Data: 2026-08-15

Stato: **COMPLETATA — workspace reali e controlli operativi osservati**

## Risultato

La v0.1.0 pubblicata fallisce prima della prima tool call su entrambi i
workspace reali. La causa è deterministica: la policy generica del Context
Engine include asset generati e incontra file oltre 1 MiB o oltre 64 MiB
complessivi.

Il candidato v0.1.1 restringe la scansione alle aree sorgente Laravel,
indicizza entrambi i progetti e completa almeno un task con una read reale per
workspace. Il profilo resta list/read/search con mutazioni deny.

## Workspace e immutabilità

I path fisici sono redatti e i repository sono indicati come
`real-laravel-a` e `real-laravel-b`.

| Workspace | Digest codice/config prima e dopo | Stato Git prima e dopo |
|---|---|---|
| `real-laravel-a` | `0af800f2c8f9f7b2198cfb4389d40e631e4dd9f61ded3f10f5d53bafa603549b` | invariato, incluse modifiche locali preesistenti |
| `real-laravel-b` | `8d5e2e6cad69fde13300f426247f5f93420fab53dd447d39e11cb625499534cb` | invariato e pulito |

Nessuna prova presenta approval o action mutativa.

## Riproduzione del bug v0.1.0

`doctor` è 9/9 su entrambi i progetti, ma `maestro run` v0.1.0 termina
immediatamente `execution_failed`. Il workspace A contiene 1.734 file per
circa 99 MiB fuori da `.git`, `vendor` e `node_modules`, con 17 file oltre
1 MiB. Il workspace B contiene 2.826 file per circa 58 MiB e 7 file oltre
1 MiB. Gli asset più grandi appartengono a `public` o a output generati e non
sono necessari al task sorgente.

La regressione deterministica aggiunge un asset pubblico da 3 MiB e una view
sorgente oltre 1 MiB. Il candidato esclude l'asset, conserva la view sotto il
nuovo bound di 2 MiB e completa l'indicizzazione.

Classificazione: **bug v0.1.x**, perché quick start e configurazione promettono
un progetto Laravel reale senza dichiarare l'incompatibilità con asset comuni.

## Matrice live del candidato

| Workspace/scenario | Exit | Terminale | Turni/tool | Durata | Classe |
|---|---:|---|---|---:|---|
| A, task principale 1 | 1 | `tool_failure` | 1 / 1 | 178.731 ms | Limite modello: call invalida |
| A, task principale 2 | 4 | `provider_failure` | 2 / 1 | 308.228 ms | Limite modello/hardware: timeout |
| A, sorgente minimo | 0 | `completed` | 2 / 1 | 289.066 ms | PASS funzionale |
| B, task principale 1 | 0 | `completed` | 2 / 1 | 349.755 ms | PASS funzionale |
| B, task principale 2 | 1 | `tool_failure` | 1 / 1 | 14.182 ms | Limite modello: call invalida |

I due PASS leggono il file richiesto e rispondono coerentemente. I failure del
modello sono già compatibili con il contratto generativo fail-closed: nessuna
call invalida viene corretta o ritentata implicitamente.

## Controlli operativi

| Controllo | Esito |
|---|---|
| Deadline run 1 secondo | `deadline_exceeded`, exit 130, terminale a 1.000 ms |
| Hard limit `model_turns: 1` | `limit_exceeded`, exit 1, 1 turno e 1 tool call |
| SIGINT durante inferenza | `canceled`, exit 130, chiusura immediata dopo il segnale |
| Workspace fixture | invariato |
| Workspace reali | invariati |

Il primo tentativo del controllo hard limit termina prima in `tool_failure` a
causa di una call invalida. L'unica ripetizione produce una call valida e
dimostra il ceiling senza secondo turno.

## Privacy

I report conservano alias, contatori, terminali, durate e digest aggregati. Non
conservano path fisici, prompt completi, risposte complete, sorgenti reali,
secret o arguments del modello.

## Gate

- almeno un task valido su ciascun workspace reale: superato;
- indicizzazione dei progetti reali: superata dal candidato v0.1.1;
- failure classificati senza attribuzione per esclusione: superato;
- deadline, hard limit e SIGINT: superati;
- immutabilità e anti-leak: superati;
- bug v0.1.x: uno, riprodotto e consegnato alla Fase 4.

La Fase 3 è completata. La Fase 4 deve qualificare e produrre la patch release
v0.1.1 prima del gate conclusivo.

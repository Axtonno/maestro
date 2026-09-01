# Milestone 24 — v0.3.1 Generation Bound Recovery

Versione target: 0.3.1

Stato: In esecuzione

Data: 2026-09-01

Prerequisito: M23 chiusa con
`v0.3.1_candidate_rejected_length_regression`.

## Obiettivo

Correggere esclusivamente il dimensionamento del limite generativo v3,
conservando compatibilità v2, diagnostica, identità, heartbeat, residency,
support claim e confine read-only già qualificati.

## Raccolta progettuale pre-freeze

Prima del candidate sono state eseguite cinque completion dell'asset pubblico
v0.3.0 sugli esatti task M17. Questi dati dimensionano il profilo ma non sono
evidenza di qualifica:

| Task | Output token | Terminale |
|---|---:|---|
| Q17-1 | 721 | `stop` |
| Q17-2 | 71 | `stop` |
| Q17-3 | 441 | `stop` |
| Q17-4 | 796 | `stop` |
| Q17-5 | 944 | `stop` |

Il limite viene congelato una sola volta a `num_predict: 1024`. Non viene
scelto 536 o 640 per adattarsi a Q17-1 e non verrà aumentato serialmente dopo
le run.

## Gate

- completion candidate 5/5;
- qualità almeno 4/5;
- terminali `length` zero;
- regressioni rispetto alla corrispondente run v0.3.0 zero;
- claim materialmente falsi nei PASS zero;
- complete e stream coerenti su Q17-2;
- workspace invariato;
- v2 e payload storico invariati.

Per ogni task vengono registrati token, terminale, qualità, durata, heartbeat
e digest workspace. La matrice appaiata alterna asset v0.3.0 e candidate v3.

## Stop rule

Un solo `length`, una regressione qualitativa o un workspace mutato respinge
il candidate. In caso di nuovo `length`, M24 non aumenta il valore: la
direzione successiva è rendere `num_predict` opzionale nel profilo GPU e
preservare il comportamento provider-default.

## Fasi

| Fase | Ambito | Stato |
|---:|---|---|
| 1 | raccolta baseline e freeze 1024 | Completata |
| 2 | aggiornamento profilo, contratti e test | Completata |
| 3 | matrice appaiata completa | Completata: PASS 4/5, `length` 0 |
| 4 | commit reale e RC LF | Non avviata |
| 5 | installazione e gate live archive | Non avviata |
| 6 | artifact finale, tag e pubblicazione | Non avviata |
| 7 | verifica post-download e audit | Non avviata |

## Esito della matrice appaiata

Tutte le cinque completion candidate terminano `stop`, inoltrano
`num_predict: 1024` e residency `5m`, e lasciano invariato il fixture. Q17-1,
Q17-2, Q17-3 e Q17-4 sono PASS. Q17-5 resta FAIL sia nella baseline appaiata
sia nel candidate per proposte non dimostrate dal solo controller; non e'
quindi una regressione introdotta dal nuovo limite.

Il controllo streaming Q17-2 termina a sua volta `stop` e conserva i fatti
richiesti. Il gate qualitativo e' PASS: completion 5/5, qualita' 4/5,
terminali `length` zero, regressioni appaiate zero e claim falsi nei PASS zero.

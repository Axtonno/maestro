# Milestone 24 — v0.3.1 Generation Bound Recovery

Versione target: 0.3.1

Stato: Release qualificata; pubblicazione in corso

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
| 4 | commit reale e RC LF | Completata: `6b879d8`, archive byte-identici |
| 5 | installazione e gate live archive | Completata: PASS |
| 6 | artifact finale, tag e pubblicazione | In esecuzione |
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

## Release candidate reale

Il commit `6b879d8bdaa61aebf55b55304155bad4784510e0` è stato clonato in un
filesystem Linux LF pulito. Due build indipendenti di `v0.3.1-rc.2` sono
byte-identiche:

```text
fb141965d972e321df2e7d4349c22abcb4ccb4d6054a458220668f8deab56867
```

L'archive è stato estratto fuori dal checkout. Identità binaria, doctor,
no-file, complete, stream, containment, redazione, `num_predict: 1024`,
residency `5m` e immutabilità sono PASS. Il digest aggregato del fixture prima
e dopo è `33aae03f668100c4b8f219c07d897661a53877c839367a226d7a0397d3cadd55`.

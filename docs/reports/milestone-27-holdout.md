# Milestone 27 — Holdout indipendente single-file

Data: 2026-09-03

Stato: **PASS**

Il holdout è stato scritto e congelato dopo la selezione di `v0.4.0-rc.4`.
Non riusa file, domande o oracoli delle matrici M25–M26 e non è stato usato
per modificare il candidate. Manifest, fixture, oracoli e risposte complete
restano nell'evidenza privata con directory `0700` e file `0600` fuori dal
repository.

| Identità redatta | SHA-256 |
|---|---|
| manifest task | `e01bd46a0e2c885f158e3168c2ae7e39ebb5a3f91c40a9792bab720f26e09cd3` |
| oracoli | `96941f4fabf7662f2edafefe8bd9f40675b3d419d6e479d0fb6a4a12a9b3be14` |

Ogni task ammesso ha ricevuto una sola generation. Tre invocazioni precedenti
sono state respinte nel preflight, prima di qualsiasi I/O provider, durante la
messa a punto del solo harness privato (`workspace.framework`, `workspace.id`
e stdin ereditato); non hanno prodotto risposte e non costituiscono retry.

| Metrica | Risultato | Soglia |
|---|---:|---:|
| completion | 10/10 (100%) | almeno 85% |
| correct | 8/10 (80%) | almeno 80% valutabili |
| partial | 2/10 | informativa |
| `response_invalid` | 0 | 0 |
| terminali `length` | 0 | 0 |
| falsità materiali | 0 | 0 |
| tool call | 0 | 0 |
| mutazioni workspace | 0 | 0 |

I due partial omettono una parte richiesta pur restando aderenti al file; non
introducono fatti falsi. Tutte le risposte dichiarano `num_predict=1024`,
`thinking=false`, residency `5m`, `truncated=false` e terminale `stop`.
Il digest aggregato del workspace è identico prima e dopo:
`15120c12d6450ead46b012842573161aca26b764c5e5d1817ec0da3ca6cea545`.

# Milestone 21 — Audit finale e decisione

Data: 2026-08-31

Stato: **COMPLETATA — candidato respinto**

Verdetto: `cpu_profile_candidate_rejected`

## Decisione

L'esatto profilo `qwen2.5-coder:7b` congelato per il ThinkPad T490s non
diventa una promessa di prodotto. Entrambe le serie complete falliscono
completion, qualità, stabilità per task, mediana warm e massimo warm. Il
risultato non è ambiguo e non è un blocco ambientale.

| Gate | Serie 1 | Serie 2 | Esito finale |
|---|---:|---:|---|
| no-file | 3/3 | 3/3 | PASS |
| completion task | 7/10 | 7/10 | FAIL |
| qualità | 4/10 | 4/10 | FAIL |
| mediana warm | 72,864 s | 67,609 s | FAIL |
| massimo warm | 192,581 s | 175,234 s | FAIL |
| timeout | 0 | 0 | PASS |
| complete/stream | 2/2 | 2/2 | PASS |
| cold/residency/eviction | PASS | PASS | PASS |
| containment/immutabilità | PASS | PASS | PASS |

Sei task — Q17-1, Q17-3, Q17-4, Q17-5, Q20-2 e Q20-4 — falliscono in entrambe
le serie. Q17-3 ripete l'inferenza di ordine creato/oggetto non dimostrata;
Q17-5 ripete POST come fatto del controller. Il limite di 512 token protegge
la durata ma causa tre terminali `length` per serie e non rende il profilo
qualitativamente sufficiente.

## Cosa è stato dimostrato

- Ollama 0.33.1, modello e digest sono allineati e stabili;
- Maestro aggiunge overhead trascurabile rispetto alla generazione già
  attribuita al modello/hardware in M20;
- il contratto v3 inoltra correttamente `num_predict: 512` e residency 5m;
- warm housekeeping resta sotto 4 ms in tutte le run formali;
- cold start è misurato: 47,291/44,638 s iniziali e 44,792/42,074 s dopo
  eviction;
- no-file, streaming, heartbeat, diagnostica, identità, containment e
  read-only behavior sono affidabili;
- non si osservano timeout, mutation, tool, retrieval, fallback, OOM o swap
  thrashing.

Questi PASS confermano la plausibilità tecnica emersa in M20, ma non
soddisfano il contratto d'uso congelato. In particolare, un profilo che
risponde correttamente soltanto a 4/10 task e supera ripetutamente 120 secondi
non può essere promosso per pochi risultati positivi.

## Artifact e support claim

La Fase 6 richiedeva due serie verdi. Non essendoci, la qualifica artifact è
stata chiusa `NOT_RUN` e nessun archive successivo alle serie è stato creato o
installato. La riproducibilità del packaging verificata in Fase 3 resta
evidenza ingegneristica, non evidenza di prodotto.

Restano quindi invariati:

- la release v0.3.0 e il suo esatto support claim;
- lo stato development-only del profilo CPU M20/M21;
- l'esclusione di agent, tool calling, retrieval, indexing, multi-file e
  Controlled Mutation;
- l'assenza di una promessa generica per CPU o hardware senza GPU.

## Audit delle fasi

| Fase | Esito |
|---:|---|
| 1 — freeze ambiente e matrice | PASS |
| 2 — contratto residency/cold-warm | PASS |
| 3 — candidate e correzioni operative | PASS |
| 4 — serie live 1 | FAIL |
| 5 — serie live 2 | FAIL |
| 6 — artifact qualification | NOT_RUN per gate |
| 7 — audit e decisione | `cpu_profile_candidate_rejected` |

La milestone è chiusa senza modificare post-hoc soglie, task o oracoli. Un
eventuale nuovo tentativo richiede un nuovo profilo e una nuova milestone; non
può riutilizzare selettivamente le risposte M21 né presentare questo candidate
come qualificato.

# Milestone 23 — Release Readiness & Publication v0.3.1

Versione target: 0.3.1

Stato: Completata — `v0.3.1_candidate_rejected_length_regression`

Data: 2026-09-01

Prerequisito: Milestone 22 completata con
`v0.3.1_operational_hardening_qualified`.

## Obiettivo

Dimostrare che l'hardening M22 è compatibile con la configurazione pubblica
v0.3.0, non riduce la qualità Direct Chat e può essere pubblicato da
un'identità reale e riproducibile. Non sono ammesse nuove funzionalità.

## Contratto di compatibilità

| Configurazione | Comportamento richiesto |
|---|---|
| v2 pubblica v0.3.0 | continua senza migrazione; generation limit e residency restano demandati al provider |
| v3 distribuita | inoltra `num_predict: 512` e residency `5m` |
| versione ignota | failure specifico e fail-closed |
| configurazione invalida | diagnostica tipizzata e redatta |

Il payload v2 del candidate deve coincidere con quello dell'asset pubblico
v0.3.0, salvo metadata dimostrabilmente non comportamentali. Il file v2 non
viene riscritto.

## Baseline e matrice

La matrice congelata è
`milestone-23-release-readiness-matrix.yaml`. Riusa esattamente Q17-1…Q17-5
dal freeze M21, che conserva prompt e oracoli M17, e confronta in modo appaiato
l'asset pubblico v0.3.0 con il candidate v3.

Gate qualità candidate:

- completion 5/5;
- almeno 4/5 risposte correct;
- zero terminali `length`;
- zero falsità materiale nei PASS;
- complete e stream coerenti sul task Q17-2;
- workspace invariato.

Se v0.3.0 completa correttamente un task che il candidate tronca, il candidate
è respinto anche qualora raggiunga 4/5.

## Sequenza

| Fase | Ambito | Stato |
|---:|---|---|
| 1 | freeze contratto, matrice e stop rule | PASS |
| 2 | gate deterministici v2/v3 e policy LF | PASS |
| 3 | confronto live appaiato e regressione qualitativa | FAIL al primo task |
| 4 | commit reale, RC riproducibile e installazione | NOT_RUN per stop rule |
| 5 | gate live sull'esatto RC | NOT_RUN per stop rule |
| 6 | artifact finale, tag e audit locale | NOT_RUN per stop rule |
| 7 | GitHub Release e verifica post-download | NOT_RUN per stop rule |
| 8 | audit e decisione | Completata |

## Policy LF e identità

- build e packaging soltanto da checkout Linux pulito o materializzazione dei
  blob Git con LF controllato;
- packaging diretto dal working tree `/mnt/c` vietato;
- `.gitattributes` rende espliciti Go, YAML, script e fixture;
- doppio archive byte-identico;
- RC con commit reale del repository principale;
- installazione e gate live fuori checkout sull'esatto archive;
- artifact finale distinto dal candidate.

## Stop rule

```text
v2 incompatibile
  -> stop e rivalutazione versione/loader

qualità < 4/5 o terminale length
  -> stop e rivalutazione num_predict

identità o LF non riproducibili
  -> stop prima del tag

tutti i gate PASS
  -> pubblicazione v0.3.1
```

Una failure di prodotto non autorizza retry selettivo o tuning. Un errore
harness dimostrato permette soltanto di ripetere l'intera serie. Tag, push e
GitHub Release sono vietati prima dei rispettivi gate.

La stop rule è scattata su Q17-1: l'asset v0.3.0 termina `stop` con 535 token,
mentre il candidate v3 termina `length` a 512. Il report autorevole è
`reports/milestone-23-final.md`.


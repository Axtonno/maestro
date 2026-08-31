# Milestone 21 — Fase 6: artifact qualification

Data: 2026-08-31

Stato: **NOT_RUN — non autorizzata dal gate live**

Verdetto: `artifact_qualification_not_authorized`

## Decisione

Il piano congela l'accesso alla qualifica artifact dietro il prerequisito
«dopo due serie verdi». Le Fasi 4 e 5 sono entrambe rosse:

| Serie | Completion task | Qualità | Mediana warm | Massimo warm |
|---|---:|---:|---:|---:|
| 1 | 7/10 | 4/10 | 72,864 s | 192,581 s |
| 2 | 7/10 | 4/10 | 67,609 s | 175,234 s |
| soglia | 10/10 | almeno 8/10 | <=60 s | <=120 s |

Sei task risultano incorrect in entrambe le serie e due falsità materiali si
ripetono. I gate operativi superati — zero timeout, no-file, streaming,
residency, eviction, containment e immutabilità — non compensano i gate di
prodotto falliti.

## Conseguenze

Non sono stati:

- costruiti archive finali di qualifica dopo le serie;
- installati artifact in un nuovo prefix;
- eseguiti Q17-1, Q20-4 o la matrice minima da un artifact;
- assegnati versione di release, tag o support claim;
- reinterpretati task, soglie o risposte per ottenere accesso alla fase.

La variante di packaging `cpu-qualification` e la sua riproducibilità erano
state verificate deterministicamente in Fase 3, prima delle serie. Quel PASS
prova il meccanismo di packaging, non qualifica live il profilo e non produce
un artifact distribuibile.

Eseguire comunque la matrice minima sarebbe contrario al protocollo congelato
e non potrebbe cambiare il verdetto. La Fase 6 è quindi chiusa esplicitamente
come `NOT_RUN`, non come `environment_blocked` e non come failure del
packaging.

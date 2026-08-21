# Milestone 13 — Field Batch 1

Data: 2026-08-21

Stato: campagna sospesa dopo il primo blocco

## Ambito

Il blocco comprende `project-b`, task `FV-01`–`FV-05`, ripetizione 1.
Artifact, configurazione, prompt, modello e limiti erano congelati prima
dell'esecuzione. Le evidenze grezze restano locali con permessi `0600`.

## Risultati

| Task | Exit | Terminale | Durata ms | Turni/call | Qualità |
|---|---:|---|---:|---|---|
| FV-01 | 4 | `provider_failure` | 300020 | 1/1 | non valutata |
| FV-02 | 4 | `provider_failure` | 300040 | 1/1 | non valutata |
| FV-03 | 4 | `provider_failure` | 300037 | 1/1 | non valutata |
| FV-04 | 0 | `completed` | 233194 | 1/1 | `partial` |
| FV-05 | 0 | `completed` | 238162 | 1/1 | `incorrect` |

Completion rate del blocco: 2/5 (40%). Tutte le run hanno eseguito una sola
tool call read-only. I digest fisici pre/post sono identici in 5/5 casi;
nessuna mutazione del workspace è stata osservata.

## Seconda revisione indipendente

La revisione obbligatoria conferma:

- FV-04 `partial`: identifica correttamente route, controller e vincoli, ma
  omette action, DTO, persistenza, evento, job accodato, forwarding, rami di
  esito e confine sincrono/asincrono; nessuna falsità materiale;
- FV-05 `incorrect`: attribuisce materialmente al controller/action
  l'invocazione del forwarding service, che appartiene invece al job
  accodato, e presenta come certo il timing runtime della coda.

I verdict sono quindi conclusivi per queste run e non più provvisori.

## Diagnosi timeout

FV-01, FV-02 e FV-03 terminano rispettivamente a 300020, 300040 e 300037 ms,
dopo una read autorizzata riuscita. La configurazione congela
`provider.timeout: 5m`, mentre il limite complessivo della run è 10 minuti.
Il pattern costituisce evidenza positiva di timeout della singola inferenza
Ollama sul profilo osservato. Il runtime termina fail-closed con
`provider_failure`; non sono stati applicati retry o cambi di profilo.

La classificazione è `environment/provider_timeout`. Non viene reinterpretata
come bug di prodotto o limite del modello per esclusione.

## Decisione

La campagna resta sospesa. Le prime ripetizioni non vengono rieseguite e il
profilo congelato non viene modificato. Prima del blocco successivo occorre
registrare una decisione esplicita sul valore diagnostico di ulteriori run con
lo stesso timeout e verificare disponibilità del budget operativo.

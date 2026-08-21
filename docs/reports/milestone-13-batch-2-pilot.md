# Milestone 13 — Batch 2 Pilot

Data: 2026-08-21

Stato: esecuzione completata; gate non superato

## Profilo e metodo

Il pilot usa il nuovo profile ID
`m13-b2p1-ollama-llama31-8b-ro-t10-r15`, con timeout provider di 10 minuti e
limite run di 15 minuti. Artifact, modello, retrieval, tool e policy read-only
restano quelli qualificati; la mutazione del workspace è negata.

Un unico task multi-file congelato è stato eseguito due volte in sequenza con
configurazione e prompt byte-identici. Le run `B2P-01-R1` e `B2P-01-R2` sono
un gate pilota e non entrano nel denominatore ufficiale di 22 run. Il preflight
del nuovo profilo ha superato 9/9 controlli.

## Risultati

| Run | Exit/terminale | Durata ms | Turni/call | Token in/out | Qualità finale | Workspace |
|---|---|---:|---|---:|---|---|
| B2P-01-R1 | 0/`completed` | 444848 | 1/1 | 1934/583 | `incorrect` | invariato |
| B2P-01-R2 | 0/`completed` | 181850 | 1/1 | 1934/446 | `partial` | invariato |

Completion e immutabilità sono 2/2. La mediana della durata è 313349 ms e il
massimo 444848 ms. La differenza di 262998 ms fra ripetizioni identiche mostra
elevata variabilità; una delle due run avrebbe ancora superato il timeout di
5 minuti usato in Batch 1.

Entrambe le risposte si fermano dopo la pre-read del file di routing. R1
combina evidenza utile sulla route di capture con componenti inventati o
classificati erroneamente e omette la catena multi-file. R2 identifica route e
controller, ma sostituisce action, modelli, job e servizi concreti con ipotesi
generiche. Nessuna risposta copre i quattro gruppi obbligatori della rubrica
multi-file.

## Seconda revisione indipendente

La seconda revisione conferma R1 `incorrect`: oltre alle omissioni, la risposta
indica un entry point non pertinente al flusso, inventa un modello e un
servizio e attribuisce al controller dashboard un metodo inesistente.

Per R2 la seconda revisione assegna `partial`, in disaccordo con il primo
`incorrect`: route, controller e middleware costituiscono evidenza utile e
corretta e non emergono falsità materiali, ma action, modelli, eventi,
notifiche, job, servizi e relazioni sincrono/queued restano assenti. La
riconciliazione adotta `partial`, coerentemente con la definizione congelata.

La distribuzione finale è quindi 0 `correct`, 1 `partial` e 1 `incorrect`.
Nessuna risposta copre tutti i gruppi obbligatori e il gate 2/2 `correct`
resta fallito.

## Decisione

Il pilot supera affidabilità e sicurezza, ma fallisce qualità e ripetibilità
prestazionale. Il timeout ampliato evita i `provider_failure`, senza risolvere
la tendenza del modello a finalizzare dopo una sola lettura né rendere
affidabile l'analisi multi-file.

Non viene congelata né avviata la matrice ufficiale Batch 2. Batch 1 resta
immutabile, la campagna ufficiale rimane sospesa a 5/22 run e nessuna run viene
ripetuta. Il prossimo lavoro utile è una modifica diagnostica del protocollo
di retrieval/choreography o un nuovo modello candidato; aumentare soltanto il
timeout non è sostenuto da questo pilot.

# Milestone 13 — Timeout Diagnostic

Data: 2026-08-21

Stato: completato, fuori dalla matrice ufficiale

## Ambito

La diagnosi `D13` conserva Batch 1 senza rieseguirlo o reinterpretarlo. Le due
run sono esplorative, non contribuiscono al denominatore ufficiale di 22 run e
usano un profilo distinto: `provider.timeout: 10m` e
`limits.duration: 15m`. Artifact, provider, modello, tool e policy read-only
restano invariati. Le evidenze grezze restano locali.

## Baseline Ollama

Una richiesta nativa minimale a `llama3.1:8b` ha completato in 11,468 secondi:

| Misura | Valore |
|---|---:|
| Load duration | 9,306 s |
| Prompt evaluation | 18 token, 1,698 s, 10,60 token/s |
| Generation | 3 token, 0,444 s, 6,76 token/s |

Il modello e il provider rispondono quindi correttamente su un input minimo.
Questa misura non è direttamente equivalente al contesto Maestro e non viene
usata per stimarne tempi esatti.

## Run diagnostiche

| ID | Progetto | Scopo | Exit/terminale | Durata ms | Turni/call | Token in/out | Qualità | Workspace |
|---|---|---|---|---:|---|---:|---|---|
| D13-01 | project-b | copia diagnostica del task FV-01 fallito | 0/`completed` | 392888 | 1/1 | 1934/461 | `partial` | invariato |
| D13-02 | project-a | relazione semplice controller-servizio | 0/`completed` | 376319 | 1/1 | 2911/153 | `correct` | invariato |

D13-01 completa oltre il timeout ufficiale di 5 minuti ma entro quello
diagnostico. La risposta identifica elementi generali utili, ma omette action,
modelli, job, servizi e gran parte dei percorsi richiesti. D13-02 risponde
correttamente al nucleo della domanda semplice. L'aumento del timeout recupera
quindi la disponibilità, non garantisce la qualità sui task multi-file.

## Perché viene osservata una sola lettura

Per il reference agent un'istruzione che inizia con `Read <logical-path>`
attiva una pre-read deterministica prima del primo turno modello. Se il modello
restituisce poi contenuto senza una tool call nativa, il runtime conclude il
passo. Nei tre timeout di Batch 1 il primo turno non aveva prodotto una
risposta; nelle completion il modello ha invece scelto di finalizzare dopo il
solo contenuto della pre-read.

La singola lettura non deriva da un deny o da un limite di tool: è l'effetto
combinato del bootstrap previsto dal runtime e della decisione del modello di
non richiedere ulteriori file. Questo spiega anche le omissioni osservate nei
task multi-file.

## Verdetto

La classe primaria è `slow_but_convergent`: entrambe le richieste completano
entro 10 minuti e avrebbero oltrepassato la soglia ufficiale di 5 minuti. Il
confronto con la baseline minimale è compatibile con un costo elevato di
valutazione del contesto, ma l'interfaccia Maestro non espone i tempi Ollama
per fase delle singole run. L'attribuzione precisa fra prompt evaluation,
generation e altri overhead resta quindi `unresolved`.

Il profilo ufficiale di Batch 1 non è sostenibile sull'host osservato. Un
eventuale Batch 2 può essere progettato con un nuovo `profile_id` e limiti
almeno pari a quelli D13, ma non dovrebbe ancora essere esteso direttamente
alla matrice completa: prima serve un piccolo gate pilota che dimostri sia
completion ripetibile sia qualità accettabile su un task multi-file. Batch 1
resta immutabile e la campagna ufficiale resta sospesa fino al congelamento e
all'approvazione di tale profilo.

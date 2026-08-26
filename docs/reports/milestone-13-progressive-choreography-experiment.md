# Milestone 13 — Esperimento di progressive choreography

Data di conclusione: 2026-08-26

Stato: esperimento development-only completato; ipotesi non confermata con il
modello corrente

## Scopo

Il pilot Batch 2 aveva mostrato che un timeout più ampio rendeva le richieste
convergenti, ma non impediva al modello di fermarsi dopo la pre-lettura del
file di routing. Questo esperimento isola quindi l'ipotesi successiva: rendere
obbligatoria una progressione di evidenza
`route -> controller/action -> simboli referenziati -> eventi/job/servizi -> risposta`.

Il task multi-file, la rubrica B01 e il timeout provider di 10 minuti sono
rimasti congelati. Le due ripetizioni sono sequenziali e byte-identiche. Il
limite complessivo di 45 minuti serve esclusivamente a consentire più turni e
non rappresenta una nuova correzione basata sul timeout. Le run sono
diagnostiche e restano fuori dal denominatore ufficiale di 22.

## Candidate diagnostico

Il candidate `v0.2.1-dev.m13.1` introduce l'agente
`agent.progressive-reference`, disponibile soltanto con il build tag
`maestro_development`. Una build normale continua a registrare esclusivamente
l'agente di riferimento già esistente.

Il runtime considera la route coperta dopo la pre-lettura deterministica. Per
ogni stato successivo richiede una lettura riuscita prima di accettare
`covered`, oppure una ricerca riuscita e vuota prima di accettare
`unavailable`. Una risposta finale viene rifiutata finché gli stati non sono
chiusi. La telemetria registra soltanto stato, decisione, tipo di strumento,
esito e motivo di arresto, senza percorsi, argomenti o contenuto sorgente.

Le suite complete normali e development-tag, le corrispondenti suite race e
`go vet` sono passate prima del freeze. Discovery del candidate e preflight
del profilo hanno avuto esito positivo.

## Risultati

| Ripetizione | Exit/terminale | Durata ms | Turni/call | Ultimo stato | Workspace |
|---|---|---:|---:|---|---|
| 1 | 4/`provider_failure` | 1282981 | 3/1 | `controller_action` aperto | invariato |
| 2 | 4/`provider_failure` | 1345623 | 3/1 | `controller_action` aperto | invariato |

Il comportamento è stato identico nei punti decisivi:

1. la lettura deterministica ha coperto `route`;
2. al primo turno il modello ha tentato di finalizzare e il runtime lo ha
   respinto;
3. al secondo turno il modello non ha emesso una chiamata nativa agli
   strumenti;
4. il terzo turno è terminato con `provider_failure`.

La sicurezza è 2/2: i digest del workspace prima e dopo ogni ripetizione
coincidono. La completion è 0/2 e l'aderenza alla progressione è 0/2, perché
nessuna ripetizione ha completato il primo stato guidato dal modello. La
qualità semantica B01 resta non valutabile: non è stata prodotta una risposta
finale da classificare come corretta, parziale o errata.

## Decisione

La choreography impedisce correttamente la finalizzazione prematura, ma il
modello corrente non segue il protocollo abbastanza da proseguire oltre la
pre-lettura. Non sono state eseguite ricerche o letture applicative successive,
quindi il risultato non sostiene ancora una diagnosi di retrieval
insufficiente.

Si applica il ramo decisionale previsto per finalizzazione prematura o stati
ignorati: il prossimo esperimento utile deve qualificare un modello realmente
nuovo con un nuovo profilo, non aumentare ancora il timeout e non ripetere
queste run. Non viene avviata una nuova matrice ufficiale; la campagna resta
sospesa a 5/22.

Il candidate diagnostico non rende retroattivamente valido l'artifact v0.2.0
e non costituisce un artifact di release. Qualsiasi prosecuzione verso v0.2.1
dovrà superare nuovamente i gate pertinenti.

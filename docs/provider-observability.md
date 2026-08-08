# Maestro Provider Observability

Versione: 0.1.0

Stato: Implementato

Ultimo aggiornamento: 2026-08-08

---

# Scopo

La Fase 8 della Provider Layer rende osservabili i confini operativi del
Provider Runtime senza introdurre SDK di logging, metriche o tracing nel core.
Un observer opzionale riceve valori neutrali e redatti; la sua assenza conserva
il percorso precedente senza allocazioni per operazione.

Gli eventi descrivono metadata operativi. Non sono un log del protocollo remoto
e non contengono prompt, risposte, chunk, embedding, credenziali o messaggi di
errore proprietari.

---

# Contratto pubblico

Il Runtime espone un unico punto di configurazione:

```go
type ProviderObserver interface {
    ObserveProviderEvent(ProviderEvent) error
}

type ProviderObserverFunc func(ProviderEvent) error

type Runtime interface {
    SetObserver(ProviderObserver)
}
```

`SetObserver(nil)` disabilita l'osservazione. La sostituzione è thread-safe e
le operazioni già iniziate conservano lo snapshot dell'observer con cui sono
partite; quelle successive vedono il nuovo valore.

`ProviderObserverFunc` è l'adapter minimale per integrare librerie applicative:

```go
runtime.SetObserver(provider.ProviderObserverFunc(func(event provider.ProviderEvent) error {
    metrics.RecordProviderEvent(event)
    return nil
}))
```

`metrics` può usare OpenTelemetry, Prometheus, un logger o un collector
proprietario senza trasferire tale dipendenza a `pkg/provider`.

---

# Eventi e correlazione

Ogni operazione osservata riceve un `OperationID` non nullo, locale al processo.
Tutti gli eventi della stessa operazione condividono ID, provider, operazione e
modello.

| Evento | Significato | Campi specifici |
|---|---|---|
| `operation_started` | il confine operativo pubblico è iniziato | timestamp, target |
| `attempt_started` | sta per iniziare una chiamata al provider | tentativo, massimo |
| `retry_scheduled` | un errore ritentabile ha pianificato il prossimo tentativo | prossimo tentativo, backoff, classificazione |
| `circuit_transition` | il breaker ha cambiato stato | stato precedente e successivo |
| `operation_completed` | unico evento terminale | durata, esito, tentativi, usage o progresso disponibili |

L'ordine di una chiamata semplice è start, attempt, terminale. Retry e
transizioni del circuito sono inseriti nello stesso flusso e mantengono lo
stesso `OperationID`. Una richiesta respinta da un circuito già aperto può
terminare senza `attempt_started`, perché non esegue I/O remoto.

Il timestamp usa lo stesso clock sostituibile della resilienza. `Duration` è
valorizzata solo sull'evento terminale e non può essere negativa.

---

# Confini osservati

Sono osservate le invocazioni pubbliche di completion, completion streaming,
embedding, model listing, discovery, load, unload, pull, remove e capability
introspection. Risoluzione del provider e verifica strutturale della capability
sono preflight: un loro fallimento non apre un confine operativo.

Le chiamate interne di discovery/load/unload usate dall'autoload di residenza
non generano un secondo flusso. Il loro eventuale fallimento termina
l'operazione pubblica originaria, evitando duplicazioni tra facade, runtime e
adapter. Le invocazioni esplicite di `DiscoverModels`, `LoadModel` e
`UnloadModel` restano invece confini autonomi.

Configurazione di policy, lettura degli snapshot e shutdown non sono chiamate
operative al provider osservate in questa fase.

---

# Esiti terminali e stream

`operation_completed` è emesso al massimo una volta con uno dei seguenti esiti:

- `success`: risposta valida, stage pull completato oppure `io.EOF` naturale;
- `error`: errore operativo;
- `canceled`: `context.Canceled` preservato nella catena;
- `deadline_exceeded`: `context.DeadlineExceeded` preservato nella catena;
- `closed`: stream chiuso esplicitamente prima di un terminale naturale.

Gli stream mantengono il confine aperto dopo il ritorno da `Stream` o
`PullModel`. EOF, errore terminale, stage pull `completed` e `Close` competono
in modo idempotente per l'unico evento finale. La chiusura dopo EOF non produce
un secondo terminale.

Per completion ed embedding il terminale contiene `Usage`. Per gli stream
contiene il massimo valore cumulativo osservato per input e output token. Per
il pull contiene il massimo `CompletedBytes` e `TotalBytes` osservato.

---

# Errori, redazione e cardinalità

Gli eventi espongono soltanto `ErrorKind`, status HTTP quando disponibile,
`Retryable` e l'indicazione `CircuitOpen`. `Error()` e i dettagli remoti non
sono copiati nell'evento.

La redazione è strutturale: `ProviderEvent` non possiede campi per messaggi,
contenuto generato, input di embedding, vettori, header o payload remoti. Il
model ID è metadata operativo e può avere cardinalità elevata; un adapter di
metriche dovrebbe usarlo come attributo soltanto quando il proprio catalogo è
limitato. `OperationID` è adatto a log e span, non a label metriche. Provider,
operazione, kind, outcome, error kind e stati del circuito sono dimensioni
normalmente limitate.

Timestamp e ID non devono essere interpretati come identificatori persistenti
o distribuiti. La correlazione tra processi appartiene all'adapter di tracing.

---

# Concorrenza e isolamento

Gli observer devono essere sicuri per chiamate concorrenti. Il Runtime li
invoca sincronicamente, ma mai mentre mantiene lock del registry, della
residenza, della resilienza, del circuito o dello stream. Il core non crea
goroutine né code implicite; un consumer che richiede esportazione asincrona la
implementa nel proprio adapter.

Un errore restituito o un panic dell'observer viene isolato e non modifica
risposta, errore, retry, stato del circuito o lifecycle dello stream. Questa
garanzia protegge gli invarianti operativi, ma non rende gratuito un observer
lento: la latenza del callback resta nel percorso della chiamata.

Quando non è configurato alcun observer, il Runtime esegue un solo caricamento
atomico, non crea tracker, non genera ID e non alloca memoria per gli eventi.

---

# Limiti

Questa fase non introduce:

- backend o SDK obbligatori di telemetria;
- sampling, batching, esportazione o persistenza degli eventi;
- propagazione di trace context verso protocolli provider;
- osservazione dei payload o logging diagnostico proprietario;
- metriche di CPU, RAM o VRAM;
- correlazione distribuita tra processi.

Le misure live e di risorse appartengono alla Milestone 3 — Benchmark &
Evaluation Layer. ADR-0015 registra ownership, redazione e isolamento del
contratto.

# Maestro Provider Runtime

Versione: 0.1.0

Stato: Implementato

Ultimo aggiornamento: 2026-08-08

---

# Scopo

Il Provider Runtime costituisce il confine tra Maestro e le implementazioni che
comunicano con i modelli AI.

Il contratto pubblico vive in `pkg/provider` e non dipende dal Runtime Core.
L'implementazione concreta vive in `internal/provider`; il composition root la
condivide attraverso `runtime.Runtime` e il `runtime.Context` consegnato ai
componenti.

---

# Contratti pubblici

Un `Provider` espone soltanto un identificativo stabile. Le operazioni sono
capability indipendenti:

* `Completer` produce una risposta a partire da una sequenza di messaggi;
* `Streamer` apre uno stream incrementale;
* `Embedder` calcola embedding per uno o più input;
* `ModelLister` elenca i modelli visibili al provider;
* `ModelDiscoverer` restituisce snapshot arricchiti e stato dei modelli;
* `ModelLoader` carica un modello nelle risorse runtime del provider;
* `ModelUnloader` rilascia un modello caricato;
* `ModelPuller` acquisisce un modello con progresso pull-based;
* `ModelRemover` rimuove un modello dal catalogo gestito dal provider;
* `CapabilityInspector` descrive supporto e disponibilità per adapter, istanza
  o modello.

Un provider può implementare qualsiasi combinazione di queste capability. Il
Runtime non richiede metodi fittizi per operazioni non disponibili.

I tipi condivisi modellano messaggi, richieste, risposte, utilizzo dei token,
chunk dello stream, embedding, descrittori dei modelli, opzioni comuni di
generazione, output JSON e tool call/result. I dettagli specifici di un SDK non
entrano nei contratti del core.

---

# Registrazione e selezione

Il Provider Runtime espone:

* registrazione atomica, con errore per ID duplicati;
* risoluzione per ID;
* configurazione e risoluzione del provider predefinito;
* routing delle capability operative e dell'introspection.

Il routing comprende anche discovery avanzata, load, unload, pull e remove dei
modelli. Il Runtime non mantiene uno snapshot autorevole del modello o un
progresso proprio: risolve il provider, verifica la capability e inoltra
l'operazione.

Le policy di residenza sono l'unica eccezione di coordinamento locale. Per una
coppia provider–model configurata il Runtime conserva lease, timer e ownership
delle sole transizioni di load avviate da Maestro. Discovery rimane la fonte
osservabile dello stato remoto.

Un ID vuoto nelle operazioni di routing richiede il provider predefinito. Il
default deve essere esplicito: può essere configurato tramite
`providers.default` oppure impostato dopo la registrazione con `SetDefault`.
L'ordine di registrazione non modifica implicitamente la selezione.

Il registry è thread-safe. Il lock protegge soltanto provider e selezione del
default; non viene mantenuto durante le chiamate a codice del provider.

---

# Capability introspection

`Capabilities` inoltra `CapabilityRequest` soltanto a provider che implementano
`CapabilityInspector`. Il target adapter non effettua I/O; instance e model
producono un nuovo snapshot remoto. Il Runtime valida identità, ordine
canonico, supporto e disponibilità prima di restituire il report.

L'introspection non viene eseguita implicitamente prima delle operazioni e non
sceglie provider o modello. Contratti e semantica sono descritti in
`provider-capability-introspection.md`.

---

# Streaming

Lo streaming è pull-based.

`Streamer.Stream` restituisce uno `Stream`; il chiamante riceve i chunk tramite
`Recv`, interpreta `io.EOF` come completamento e invoca sempre `Close` quando
interrompe o conclude il consumo. `ModelPuller.PullModel` applica la stessa
ownership esplicita attraverso `ModelPullStream`, restituendo prima lo stage
terminale `completed` e poi `io.EOF`.

Il contesto passato all'apertura dello stream definisce cancellazione e
deadline. L'implementazione del provider è responsabile di propagarle al
trasporto sottostante.

Il Runtime non crea goroutine, canali, buffer o retry impliciti per lo streaming
ordinario. Una policy di resilienza esplicita può riaprire lo stream soltanto
prima del primo chunk. Quando è configurata una policy di residenza, il Runtime
mantiene inoltre il lease fino a EOF, errore terminale o `Close`.

Gli stream di generazione avanzata espongono `ToolCallDelta` indicizzati. Gli
arguments sono frammenti da concatenare per indice; ID e nome possono comparire
soltanto nel primo chunk. Gli adapter validano la call ricostruita prima del
terminale. Structured output viene consegnato incrementalmente e validato come
JSON al finish reason.

---

# Model residency

`SetModelResidencyPolicy` configura autoload opt-in e rilascio immediato, a TTL
o allo shutdown per un model ID esatto. Prima di completion, streaming o
embedding, il Runtime usa discovery e carica il modello soltanto se non risulta
già residente. Richieste concorrenti condividono la transizione.

Solo un load eseguito dal Runtime assegna ownership alla policy. Un modello già
residente non viene scaricato. `LoadModel` e `UnloadModel` restano comandi
espliciti indipendenti e non installano una policy.

`Shutdown` annulla i timer, attende i lease attivi e rilascia le residenze
possedute. Il Runtime Core lo invoca durante `Stop`. Contratto, semantica e limiti
sono descritti in `provider-model-residency.md`.

---

# Configurazione

Il package `pkg/runtime` fornisce `NewConfig`, una configurazione minimale a
chiavi esatte costruita da una copia della mappa ricevuta. La configurazione è
iniettata nel Runtime tramite `maestro.WithConfig` ed è disponibile ai
componenti durante il lifecycle.

La chiave riservata al Provider Runtime in questa fase è:

* `providers.default`: ID del provider predefinito, come `string` o
  `provider.ID`.

Parsing di file, variabili d'ambiente, secret storage e merge tra sorgenti non
appartengono al core. Potranno essere implementati da adapter che producono il
contratto minimale `runtime.Config`.

---

# Ownership ed errori

Richieste e risposte sono valori applicativi. Slice, mappe e payload in esse
contenuti non vengono copiati dal router: il chiamante e il provider non devono
modificarli concorrentemente durante un'operazione.

Gli errori pubblici distinguono provider non valido, registrazione duplicata,
provider non trovato, default assente, capability non supportata e stream non
valido. Gli errori operativi degli adapter usano `ProviderError`, che espone
kind, operazione, provider, modello, status e ritentabilità conservativa.
Cause, context e sentinel vecchi e nuovi restano ispezionabili tramite
`errors.Is`/`errors.As`; nessun consumer deve interpretare `Error()`.

---

# Osservabilità

`SetObserver` installa opzionalmente un `ProviderObserver` sul Runtime.
Completion, streaming, embedding, catalogo, lifecycle, acquisition, removal e
capability introspection emettono eventi redatti per start, tentativi, retry,
transizioni del circuito e un unico terminale.

Gli eventi della stessa invocazione condividono un operation ID locale. Gli
stream restano aperti dal punto di vista osservabile fino a EOF, errore, stage
pull completato o `Close`. Gli observer vengono invocati senza lock interni;
errori e panic non cambiano il risultato provider. Senza observer non vengono
creati tracker né eventi.

Contratto, ordering, redazione e indicazioni di cardinalità sono descritti in
`provider-observability.md`.

---

# Estensioni ancora escluse

Il Provider Runtime non introduce:

* fallback automatico tra provider;
* load balancing;
* persistenza delle conversazioni;
* caching di risposte, modelli o embedding;
* traduzione di opzioni specifiche dei singoli SDK;
* implementazioni concrete diverse dall'adapter Ollama.

Discovery e lifecycle dei modelli sono introdotti dalla Fase 2 della Provider
Layer e descritti in:

```
docs/provider-model-lifecycle.md
```

Acquisizione e rimozione dei modelli sono introdotte dalla Fase 3 e descritte
in:

```
docs/provider-model-acquisition.md
```

Le policy di residenza sono introdotte dalla Fase 4 e descritte in:

```
docs/provider-model-residency.md
```

La capability introspection è introdotta dalla Fase 5 e descritta in:

```
docs/provider-capability-introspection.md
```

La semantica degli errori è introdotta dalla Fase 6 e descritta in:

```
docs/provider-error-semantics.md
```

Le policy di resilienza opt-in sono introdotte dalla Fase 7 e descritte in:

```
docs/provider-resilience.md
```

L'osservabilità provider è introdotta dalla Fase 8 e descritta in:

```
docs/provider-observability.md
```

La baseline di generazione avanzata è introdotta dalla Fase 9 e descritta in:

```
docs/provider-advanced-generation.md
```

La prima implementazione concreta è descritta in:

```
docs/ollama-provider.md
```

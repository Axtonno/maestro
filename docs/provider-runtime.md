# Maestro Provider Runtime

Versione: 0.1.0

Stato: Implementato

Ultimo aggiornamento: 2026-08-05

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
* `ModelLister` elenca i modelli visibili al provider.

Un provider può implementare qualsiasi combinazione di queste capability. Il
Runtime non richiede metodi fittizi per operazioni non disponibili.

I tipi condivisi modellano messaggi, richieste, risposte, utilizzo dei token,
chunk dello stream, embedding e descrittori dei modelli. I dettagli specifici
di un SDK non entrano nei contratti del core.

---

# Registrazione e selezione

Il Provider Runtime espone:

* registrazione atomica, con errore per ID duplicati;
* risoluzione per ID;
* configurazione e risoluzione del provider predefinito;
* routing delle capability operative.

Un ID vuoto nelle operazioni di routing richiede il provider predefinito. Il
default deve essere esplicito: può essere configurato tramite
`providers.default` oppure impostato dopo la registrazione con `SetDefault`.
L'ordine di registrazione non modifica implicitamente la selezione.

Il registry è thread-safe. Il lock protegge soltanto provider e selezione del
default; non viene mantenuto durante le chiamate a codice del provider.

---

# Streaming

Lo streaming è pull-based.

`Streamer.Stream` restituisce uno `Stream`; il chiamante riceve i chunk tramite
`Recv`, interpreta `io.EOF` come completamento e invoca sempre `Close` quando
interrompe o conclude il consumo.

Il contesto passato all'apertura dello stream definisce cancellazione e
deadline. L'implementazione del provider è responsabile di propagarle al
trasporto sottostante.

Il Runtime non crea goroutine, canali, buffer o retry impliciti.

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
valido. `ErrInvalidRequest` e `ErrInvalidResponse` permettono inoltre agli
adapter di segnalare errori neutrali di validazione. Gli errori dei provider
vengono arricchiti con operazione e ID e restano ispezionabili tramite
`errors.Is`.

---

# Estensioni escluse dalla prima versione

La prima implementazione non introduce:

* fallback automatico tra provider;
* retry o circuit breaker;
* load balancing;
* persistenza delle conversazioni;
* caching di risposte, modelli o embedding;
* traduzione di opzioni specifiche dei singoli SDK;
* implementazioni concrete diverse dall'adapter Ollama.

Queste funzionalità verranno aggiunte sopra i contratti correnti quando i
relativi requisiti saranno concreti.

La prima implementazione concreta è descritta in:

```
docs/ollama-provider.md
```

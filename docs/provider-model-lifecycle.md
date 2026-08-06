# Maestro Provider Model Discovery & Lifecycle

Versione: 0.1.0

Stato: Implementato

Ultimo aggiornamento: 2026-08-06

---

# Scopo

La Fase 2 della Provider Layer introduce contratti neutrali per osservare i
modelli disponibili e governarne il caricamento in memoria.

Il Provider Runtime orchestra le capability. Gli adapter traducono le
operazioni nei protocolli concreti; nessuno stato modello viene duplicato nel
core.

---

# Capability pubbliche

Le operazioni sono separate in tre interfacce opzionali:

* `ModelDiscoverer` restituisce snapshot arricchiti dei modelli visibili;
* `ModelLoader` assicura che un modello venga caricato;
* `ModelUnloader` richiede il rilascio di un modello caricato.

Un provider può implementare qualsiasi combinazione. `ModelLister` rimane la
vista minimale e retrocompatibile basata su `[]Model`.

Il Provider Runtime espone il routing corrispondente e restituisce
`ErrUnsupportedCapability` quando il provider selezionato non implementa
l'operazione richiesta.

---

# ModelInfo e stato

`ModelInfo` contiene:

* identità neutrale `Model`;
* stato osservato;
* digest;
* dimensione su storage e memoria dedicata;
* context length;
* formato, famiglia, parameter size e quantizzazione;
* timestamp di modifica e scadenza in memoria.

I valori di `ModelState` sono:

* `unknown`;
* `available`;
* `downloading`;
* `loading`;
* `loaded`;
* `sleeping`;
* `failed`.

Lo stato è uno snapshot del provider, non una state machine posseduta da
Maestro. Può cambiare immediatamente dopo il ritorno della chiamata.

Campi che il provider non può determinare restano al valore zero. I dettagli
specifici non vengono esposti attraverso mappe opache.

---

# Lifecycle

`LoadModel` e `UnloadModel` ricevono richieste distinte, così future opzioni
neutrali possono evolvere senza condividere accidentalmente semantiche diverse.

Il completamento di `LoadModel` indica che il provider ha accettato e completato
la propria operazione sincrona di load. `UnloadModel` indica che la richiesta di
rilascio è stata completata secondo il protocollo del provider.

Il Runtime non introduce:

* polling implicito;
* retry;
* timeout propri;
* lock distribuiti;
* ref-count dei consumer;
* unload automatico;
* stato persistente.

Cancellazione e deadline appartengono al `context.Context` del chiamante.

---

# Ollama

La discovery unisce:

* `GET /api/tags` per i modelli disponibili;
* `GET /api/ps` per i modelli caricati e i dati runtime.

I modelli presenti in `/api/ps` sono marcati `loaded`. Gli altri modelli
installati sono `available`. Eventuali modelli in memoria non presenti nello
snapshot di `/api/tags` vengono aggiunti in ordine di risposta.

Il load usa una richiesta vuota a `POST /api/generate` con `stream: false` e
`keep_alive: -1`, mantenendo il modello caricato fino a unload esplicito. Lo
unload usa lo stesso endpoint con `keep_alive: 0`.

---

# llama.cpp

La discovery usa `GET /models`, che in router mode espone modelli disponibili,
metadati e stato. I valori del protocollo vengono tradotti in `ModelState`.

Load e unload usano rispettivamente:

* `POST /models/load`;
* `POST /models/unload`.

Questi endpoint richiedono `llama-server` in router mode. L'adapter implementa
le capability perché il protocollo le supporta, ma un processo avviato in
single-model mode può rifiutare le operazioni con un errore HTTP.

---

# Concorrenza e ownership

Ogni invocazione opera sullo snapshot restituito dal provider. Il Runtime non
mantiene lock mentre esegue codice dell'adapter.

Le slice restituite appartengono al chiamante. Gli adapter costruiscono valori
neutrali nuovi e non espongono DTO o collezioni mutabili del trasporto.

---

# Fuori scope

Questa fase non include:

* pull, download o delete;
* progress streaming;
* eventi di cambiamento stato;
* keep-alive configurabile per richiesta;
* autoload policy;
* caching della discovery;
* retry, backoff o circuit breaker;
* selezione automatica del modello.

Queste funzioni sono assegnate alle fasi successive descritte in
`provider-layer-plan.md`. Acquisizione e rimozione appartengono alla Fase 3,
residenza e autoload alla Fase 4, mentre errori e resilienza sono separati nelle
Fasi 6 e 7. Caching della discovery e selezione automatica del modello non sono
requisiti di chiusura della Milestone 2.

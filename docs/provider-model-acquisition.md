# Maestro Provider Model Acquisition & Removal

Versione: 0.1.0

Stato: Implementato

Ultimo aggiornamento: 2026-08-08

---

# Scopo

La Fase 3 della Provider Layer completa la gestione del catalogo con capability
neutrali per acquisire e rimuovere modelli.

Il Provider Runtime esegue soltanto risoluzione e routing. File, cache, stato
del download e cancellazione remota restano responsabilità degli adapter e dei
server provider.

---

# Contratti pubblici

`pkg/provider` espone due capability opzionali:

```go
type ModelPuller interface {
    Provider
    PullModel(context.Context, ModelPullRequest) (ModelPullStream, error)
}

type ModelRemover interface {
    Provider
    RemoveModel(context.Context, ModelRemoveRequest) error
}
```

Acquisizione e rimozione sono indipendenti: un provider può implementare una
soltanto delle due capability.

`ModelPullStream` segue la stessa ownership esplicita dello streaming di
completion:

```go
type ModelPullStream interface {
    Recv() (ModelPullProgress, error)
    Close() error
}
```

Il chiamante deve consumare fino a `io.EOF` oppure invocare `Close`.

---

# Progresso neutrale

Ogni snapshot contiene modello, stage, dettaglio informativo, digest e conteggi
in byte quando disponibili.

Gli stage pubblici sono:

* `unknown`;
* `resolving`;
* `downloading`;
* `verifying`;
* `finalizing`;
* `completed`.

`completed` viene restituito come ultimo evento valido. La chiamata `Recv`
successiva restituisce `io.EOF`. EOF del trasporto prima di `completed` produce
`provider.ErrInvalidResponse`.

`Detail` è destinato a log o interfacce utente. Può contenere uno status del
provider e non deve essere analizzato per prendere decisioni; il controllo di
flusso usa esclusivamente `Stage` ed errori.

Conteggi negativi, overflow e progresso superiore al totale dichiarato sono
rifiutati come risposte invalide.

---

# Routing

Il Provider Runtime aggiunge:

```go
PullModel(context.Context, provider.ID, provider.ModelPullRequest)
RemoveModel(context.Context, provider.ID, provider.ModelRemoveRequest)
```

L'ID vuoto usa il default configurato. Il Runtime verifica la capability,
mantiene ispezionabile `ErrUnsupportedCapability`, rifiuta stream nil e non
conserva lock mentre esegue codice del provider.

Il Runtime non mantiene una lista di download, non aggrega il progresso e non
modifica lo stato restituito da discovery.

---

# Ollama

`PullModel` usa `POST /api/pull` con `stream: true` e traduce lo stream NDJSON:

* risoluzione del manifest → `resolving`;
* download dei layer → `downloading`;
* verifica del digest → `verifying`;
* scrittura del manifest e pulizia → `finalizing`;
* `success` → `completed`.

`RemoveModel` usa `DELETE /api/delete` con il modello nel body JSON. Una
risposta HTTP 2xx senza body è valida.

La chiusura anticipata interrompe la richiesta HTTP. Cancellazione e deadline
rimangono riconoscibili tramite `errors.Is`.

---

# llama.cpp

La capability richiede una versione del router che esponga gli endpoint di
gestione del catalogo.

La sequenza di pull è:

1. `POST /models` avvia il download;
2. `GET /models/sse` fornisce gli eventi globali;
3. l'adapter filtra gli eventi sul modello richiesto;
4. i progressi paralleli dei file vengono aggregati in byte;
5. `download_finished` produce il refresh di `GET /models` e lo stage
   `completed`.

`download_failed` termina lo stream con errore. Una chiusura anticipata o la
cancellazione del context chiude SSE e invia `POST /models/unload` con un
context di cleanup limitato, come previsto dal protocollo del router.

`RemoveModel` usa `DELETE /models?model=...`. Soltanto i modelli presenti nella
cache del router possono essere rimossi; preset e directory esterne rimangono
sotto il controllo dell'operatore.

---

# Idempotenza e sicurezza

Il Provider Runtime non introduce deduplicazione o stato locale. Pull ripetuti
e rimozioni di modelli assenti conservano la semantica e gli errori del server,
ma non producono mutazioni dirette nel filesystem di Maestro.

Nessun adapter costruisce path di cancellazione o rimuove file. Anche quando il
provider non supporta la capability, Maestro restituisce un errore invece di
accedere alla cache sottostante.

---

# Test

La suite isolata copre:

* routing esplicito e tramite default;
* capability non supportate e stream nil;
* mapping delle richieste Ollama e llama.cpp;
* stage, digest e aggregazione del progresso;
* evento terminale e `io.EOF`;
* JSON, NDJSON e SSE malformati;
* conteggi invalidi;
* errori HTTP e messaggi intermedi;
* chiusura idempotente e cancellazione;
* cleanup remoto del download llama.cpp;
* rimozione senza accesso diretto al filesystem.

Gli smoke test live di pull e remove confluiscono nello Smoke Benchmark della
Milestone 3, perché modificano il catalogo locale e richiedono modelli dedicati.

---

# Fuori scope

Questa fase non introduce:

* cache o persistenza dei trasferimenti nel Runtime;
* resume gestito da Maestro;
* retry o circuit breaker;
* policy di keep-alive o autoload;
* selezione automatica del modello;
* supervisione dei processi provider;
* cancellazione o rimozione diretta di file.

Residenza e autoload sono completate nella Fase 4 e descritte in
`provider-model-residency.md`. La semantica degli errori è completata nella
Fase 6 e descritta in `provider-error-semantics.md`; la resilienza appartiene
alla Fase 7.

# Maestro Plugin Runtime

Versione: 0.1.0

Stato: Implementato

Ultimo aggiornamento: 2026-08-05

---

# Scopo

Il Plugin Runtime registra, scopre e carica estensioni opzionali senza
introdurre un secondo sistema di componenti o un secondo lifecycle.

I contratti pubblici vivono in `pkg/plugin`, l'implementazione concreta in
`internal/plugin`. Il composition root restituito da `maestro.New` espone la
stessa istanza tramite `Plugins()`.

---

# Plugin e manifest

Un `Plugin` è un `runtime.Component` che espone anche un `Manifest`.

I metadati del componente restano l'unica fonte per identità, versione,
dipendenze e capability. Il manifest dichiara la versione dell'API Plugin
Runtime richiesta dal plugin. La versione corrente è disponibile come
`plugin.RuntimeAPIVersion`; una versione vuota produce `ErrInvalidManifest` e
una versione differente produce `ErrIncompatible` prima della registrazione nel
Runtime Core.

`Metadata.ID` deve rimanere stabile dopo la registrazione.

---

# Registry e osservabilità

Il Plugin Runtime espone:

* `Register`, `Resolve` e `Has` per le istanze plugin;
* `Registered`, che restituisce gli ID registrati in ordine deterministico;
* `RegisterLoader` per aggiungere una factory al catalogo;
* `Available`, che restituisce gli ID dei loader in ordine di registrazione;
* `Load`, che costruisce e registra un plugin dal catalogo.

Gli snapshot restituiti non espongono le slice interne. Registry e catalogo sono
thread-safe. ID vuoti, blank o con spazi iniziali/finali, typed nil, duplicati e
collisioni con normali componenti vengono rifiutati con errori ispezionabili
tramite `errors.Is`.

Le operazioni riuscite pubblicano sullo stesso Event Bus del Runtime:

* `plugin.loader.registered`;
* `plugin.registered`;
* `plugin.loaded`.

Il payload pubblico contiene ID e, quando disponibile, l'istanza plugin.
L'istanza è un riferimento trusted in-process, non un payload serializzabile:
adapter di logging o telemetria non devono convertirla implicitamente in dati.

La pubblicazione avviene dopo lo stato corrispondente e fuori lock. Errori e
panic dell'Event Bus sono isolati dal Plugin Runtime e non trasformano
un'operazione già riuscita in un errore apparente. La consegna resta sincrona:
un subscriber lento applica backpressure al publisher secondo ADR-0005.

## Terminologia operativa

Il Plugin System distingue quattro concetti:

* `Available` indica che nel catalogo esiste un loader per l'ID;
* `Registered` e `Has` indicano che un'istanza è registrata anche nel Runtime
  Core;
* `plugin.loaded` indica che una singola operazione di factory e registrazione
  è riuscita;
* lo stato running appartiene esclusivamente allo `StateManager` globale.

Loaded non è uno stato persistente e `Load` non esegue il lifecycle. Il plugin
deve essere caricato o registrato prima di `Runtime.Start`; registrazione durante
startup, esecuzione o shutdown viene rifiutata dal Runtime Core.

---

# Loader e caricamento

Un `Loader` è una factory pull-based:

```go
type Loader interface {
    Load(context.Context) (Plugin, error)
}
```

`LoaderFunc` adatta direttamente una funzione. Il Plugin Runtime risolve il
loader sotto lock, lo esegue senza lock, verifica cancellazione, risultato non
nil, corrispondenza tra ID richiesto e prodotto, manifest e registrazione nel
Runtime Core.

Un loader costruisce soltanto l'istanza: non la registra, non ne avvia il
lifecycle e non acquisisce risorse di lunga durata. Queste ultime appartengono
alle capability `Initialize` e `Start`, così un errore di registrazione non
lascia risorse orfane.

Ogni `Load` è un tentativo indipendente. Chiamate concorrenti sullo stesso ID
possono invocare il loader più volte, ma il Registry globale consente una sola
registrazione riuscita. I loader devono quindi proteggere eventuale stato
mutabile della factory. Maestro non accorpa context o risultati distinti tramite
singleflight implicito.

Factory, registrar ed eventi vengono invocati senza lock interni del Plugin
Runtime. `Available` e `Registered` sono snapshot difensivi ordinati per
registrazione riuscita; l'ordine relativo di operazioni concorrenti non è
specificato e non costituisce selezione implicita.

Il catalogo descrive plugin disponibili nella build corrente. Non esegue
scansioni, download o selezioni implicite.

---

# Lifecycle e dipendenze

Il Plugin Runtime non invoca direttamente le capability lifecycle.

Dopo il caricamento, il Runtime Core inserisce il plugin nel dependency graph
globale. Startup e shutdown rispettano gli stessi ordinamenti dei componenti;
stato e failure sono osservabili tramite lo `StateManager` globale. Il plugin
riceve lo stesso `runtime.Context` con configurazione, logger, Event Bus,
Registry e Provider Runtime.

Dipendenze plugin -> componente, plugin -> plugin e componente -> plugin sono
archi dello stesso graph. Dipendenze richieste mancanti e cicli impediscono lo
startup prima del lifecycle; dipendenze opzionali assenti vengono ignorate.

La registrazione durante startup o running produce `runtime.ErrAlreadyStarted`;
durante shutdown produce `runtime.ErrInvalidState`. Gestor vede una
registrazione riuscita attraverso la source globale dei componenti: lo snapshot
diventa stale, il refresh resta esplicito e la risoluzione non esegue il plugin.

---

# Modello di fiducia

I loader e i plugin sono codice Go in-process e hanno i privilegi del processo
Maestro. La registrazione di un loader equivale quindi ad autorizzare codice
fidato già collegato all'applicazione.

Shared object, download, firme, sandbox, process isolation, permission model,
unload e hot replacement sono livelli di distribuzione e sicurezza successivi;
non modificano il contratto di registry e lifecycle completato in questa fase.

Le decisioni di stabilizzazione della Milestone 5 sono formalizzate in
`docs/adr/ADR-0023.md`; la matrice pubblica è disponibile in
`docs/plugin-api-compatibility-audit.md`.

---

# Primo plugin

Il primo adapter framework-aware è Laravel:

```
pkg/plugin/laravel
internal/plugin/laravel
```

Configurazione e detection sono descritte in:

```
docs/laravel-plugin.md
```

# Milestone 5 — Plugin System Design

Versione: 0.2.0

Stato: Implementato — Milestone 5 completata

Data: 2026-08-11

---

# Contesto

Maestro possiede già una baseline Plugin Runtime introdotta durante lo sviluppo
del Runtime Core. La baseline comprende:

- il contratto pubblico `plugin.Plugin` basato su `runtime.Component`;
- manifest con compatibilità esatta dell'API Plugin Runtime;
- registry e catalogo di loader thread-safe;
- caricamento cancellabile di plugin Go trusted in-process;
- registrazione nel Registry globale dei componenti;
- lifecycle e dependency graph condivisi con il Runtime Core;
- eventi di registrazione e caricamento;
- il plugin Laravel con detection e health del workspace.

La Milestone 5 non ricostruisce questi meccanismi. Li assume come baseline,
verifica che i loro contratti siano sufficienti e li porta a un gate di
qualità esplicito, preparando il confine che verrà consumato dal Context Engine.

---

# Obiettivo

Consegnare un Plugin System in-process stabile, deterministico e osservabile,
nel quale un'applicazione possa dichiarare loader fidati, caricare e registrare
plugin prima dell'avvio del Runtime e lasciare al Runtime Core la gestione di
dipendenze, stato e lifecycle.

Il primo caso di riferimento resta Laravel. La milestone deve dimostrare il
percorso completo:

```text
catalogo -> load -> registrazione -> Gestor -> dependency graph -> lifecycle
```

La chiusura richiede contratti pubblici sottoposti ad audit, test concorrenti ed
end-to-end, documentazione e report di fase.

---

# Confini della milestone

## Incluso

- audit e stabilizzazione delle API pubbliche in `pkg/plugin`;
- semantica di manifest, compatibilità, loader e registrazione;
- catalogo e listing deterministici;
- caricamento cancellabile e failure atomiche;
- comportamento concorrente di registry e loader;
- integrazione pre-start con il Runtime Core;
- dependency graph, stato e lifecycle globali;
- invalidazione e discovery delle capability tramite Gestor;
- eventi plugin sul bus condiviso;
- plugin Laravel come implementazione di riferimento;
- test, race detector, vet, documentazione e audit di compatibilità.

## Escluso

- download e installazione di artefatti esterni;
- marketplace e discovery remota;
- firme, provenance e aggiornamenti automatici;
- caricamento tramite shared object Go;
- process isolation, sandbox e permission model;
- unload, hot replacement e registrazione dopo l'avvio;
- plugin di terze parti e relativo SDK di distribuzione;
- funzionalità proprie del Context Engine;
- comandi Artisan e tool execution.

Packaging, trust distribuito e plugin di terze parti appartengono alla
Milestone 8. Il permission model operativo appartiene alla Milestone 7.

---

# Modello architetturale

## Un plugin resta un componente

`plugin.Plugin` estende `runtime.Component`; non possiede un lifecycle
alternativo. `runtime.Metadata` resta la fonte autorevole per identità,
versione, dipendenze e capability. Il manifest plugin dichiara soltanto i
requisiti di compatibilità propri del Plugin Runtime.

Questa separazione evita di duplicare:

- stato dei componenti;
- dependency graph;
- ordinamento di startup e shutdown;
- configurazione e contesto runtime;
- health e altre capability lifecycle.

## Catalogo, registry e stato sono distinti

Il catalogo contiene loader disponibili nella build corrente. Il registry
contiene istanze plugin già registrate. Lo stato operativo appartiene allo
`StateManager` del Runtime Core.

Quindi:

- `Available` non implica che un plugin sia istanziato;
- `Registered` non implica che il plugin sia inizializzato o running;
- `Load` costruisce e registra, ma non avvia il plugin;
- soltanto `Runtime.Start` esegue il lifecycle nel dependency order globale.

La milestone non introduce un secondo enum di stato plugin.

## Loader trusted in-process

Un loader è una factory pull-based collegata all'applicazione. Viene risolto
sotto lock ma eseguito senza lock del Plugin Runtime. Deve costruire una nuova
istanza senza registrarla, avviarla o acquisire risorse di lunga durata.

Il codice del loader e del plugin opera con gli stessi privilegi del processo
Maestro. La compatibilità del manifest non è un confine di sicurezza.

## Registrazione pre-start

La registrazione confluisce nel Registry globale. Di conseguenza eredita la
regola del Runtime Core: componenti e plugin possono essere registrati prima
dell'avvio, non mentre il Runtime è in startup, running o shutdown.

Questa regola rende il dependency graph uno snapshot coerente per l'intero
lifecycle. Hot loading e unload richiederebbero transazioni su grafo, stato e
risorse e restano fuori scope.

---

# Invarianti

- ID plugin e loader non vuoti e privi di whitespace non normalizzato;
- identità del plugin prodotto uguale all'ID richiesto al catalogo;
- manifest validato prima della registrazione nel Runtime Core;
- al massimo una registrazione riuscita per ID nel Runtime globale;
- nessun callback di loader o evento eseguito mantenendo lock interni;
- nessun plugin parzialmente visibile nel registry dedicato;
- listing restituiti come snapshot difensivi e deterministici;
- cancellazione precedente al load impedisce l'esecuzione del loader;
- cancellazione osservata dopo la factory impedisce la registrazione;
- un loader non avvia il lifecycle e non possiede risorse longeve;
- plugin e componenti normali condividono collision detection e dipendenze;
- Gestor indicizza il plugin una sola volta attraverso il Registry globale;
- nessuna selezione implicita di plugin tramite ordine di registrazione.

---

# Compatibilità pubblica

L'audit iniziale parte con una preferenza conservativa: estensioni additive e
nessuna modifica breaking alle interfacce `runtime.Runtime`, `plugin.Runtime` o
ai componenti già implementati.

Una nuova API pubblica è ammessa soltanto se rappresenta un confine di consumo
necessario e non può essere derivata in modo autorevole dai contratti esistenti.
In particolare, non verrà duplicata nel manifest l'identità già presente in
`runtime.Metadata`.

Le decisioni definitive dell'audit saranno registrate in ADR-0023.

---

# Integrazione con Gestor

I plugin entrano in Gestor come componenti del Registry globale. Le capability
custom dichiarate nei metadata devono essere namespaced e vengono indicizzate
dalla sorgente Runtime esistente.

Il Plugin Runtime non introduce una seconda source Gestor. Una registrazione
riuscita invalida lo snapshot attraverso il normale percorso del Runtime; il
refresh resta esplicito. Gestor risolve descriptor e dependency plan, ma non
carica plugin e non esegue capability.

---

# Laravel come reference plugin

Laravel deve dimostrare che un adapter framework-aware può:

- essere dichiarato nel catalogo tramite loader;
- validare una configurazione di workspace;
- rilevare il framework durante `Initialize`;
- pubblicare una vista thread-safe delle informazioni rilevate;
- ripetere la detection tramite `Health`;
- dichiarare capability coerenti e individuabili da Gestor;
- fallire senza lasciare risorse o stato plugin separato.

La milestone stabilizzerà soltanto la capability framework-aware necessaria a
descrivere il workspace. Route analysis, Eloquent, Blade, Artisan e costruzione
del contesto restano incrementi successivi.

---

# Osservabilità ed errori

Gli eventi plugin usano l'Event Bus sincrono condiviso e vengono pubblicati
fuori lock. L'audit verificherà cardinalità, ordine e payload dei topic
pubblici, mantenendo separati gli eventi di catalogo, registrazione e load.

Gli errori pubblici devono restare ispezionabili con `errors.Is` e distinguere
almeno input invalido, incompatibilità, duplicati, loader assente e fallimento
del load. Le cause provenienti da loader e Runtime Core devono essere
preservate. Gli errori non devono introdurre credenziali o contenuti del
workspace nei payload degli eventi.

---

# Strategia di verifica

La milestone usa esclusivamente test deterministici e non richiede rete,
processi esterni o un'applicazione Laravel installata. Le fixture locali devono
coprire:

- contratti e compatibilità;
- registrazione e caricamento concorrenti;
- cancellazione e failure della factory;
- collisioni con componenti normali;
- dipendenze tra plugin e componenti;
- startup, health e shutdown;
- invalidazione e refresh Gestor;
- eventi e assenza di callback sotto lock;
- workspace Laravel valido, assente, malformato e mutato;
- test end-to-end dal composition root pubblico.

Il gate finale richiede `go test ./...`, test con race detector sui package
coinvolti e `go vet ./...`.

---

# Decisioni formalizzate in ADR-0023

- il Plugin System della Milestone 5 resta trusted in-process;
- catalogo, registry e stato operativo sono tre viste distinte;
- il caricamento è pre-start e non equivale ad attivazione lifecycle;
- stato e lifecycle restano proprietà esclusiva del Runtime Core;
- Gestor scopre i plugin attraverso il Registry globale, senza source dedicata;
- il gate non include distribuzione esterna, sandbox o hot loading;
- l'evoluzione delle API è additiva salvo una motivazione esplicita emersa
  dall'audit.

ADR-0023 è Accepted. Il piano operativo è descritto in
`plugin-system-development-plan.md`.

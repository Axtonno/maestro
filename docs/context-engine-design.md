# Milestone 6 — Context Engine Design

Versione: 1.0.0

Stato: Implementato — ADR-0024 Accepted

Data: 2026-08-11

---

# Contesto

Maestro possiede già le fondamenta necessarie per introdurre il Context Engine:

- un Runtime Core con lifecycle, Registry, configurazione ed Event Bus;
- Gestor per discovery e risoluzione deterministica delle capability;
- Provider Runtime con embedding provider-neutral;
- Plugin System trusted in-process e un reference plugin Laravel che descrive
  un workspace framework-specific;
- un Developer Benchmark con fixture versionate e un primo retrieval cosine
  che usa direttamente gli embedding provider.

Queste fondamenta non definiscono ancora un modello workspace generico, un
indice, una rappresentazione strutturata del codice o un contratto per
costruire contesti entro un budget. Il Context Engine introduce questi confini
senza trasferire nel core logica di linguaggio, framework o provider.

---

# Obiettivo

Consegnare un servizio provider-agnostic capace di trasformare un workspace in
uno snapshot indicizzato e analizzato, recuperare evidenza rilevante per una
richiesta e costruire un bundle di contesto deterministico entro un budget
esplicito.

Il percorso completo della milestone è:

```text
workspace -> snapshot -> analisi -> retrieval -> selezione -> context bundle
```

Il risultato deve essere ispezionabile: ogni sezione del bundle conserva
origine, intervallo, motivazione di selezione e costo stimato. Nessun ranking,
provider, modello o strategia di troncamento viene scelto implicitamente.

---

# Confini della milestone

## Incluso

- modello pubblico framework-neutral del workspace;
- scansione filesystem limitata e cancellabile;
- path logici normalizzati e policy esplicite di inclusione/esclusione;
- snapshot immutabili, generazionali e pubblicati atomicamente;
- documenti testuali con identità e digest content-addressed;
- registry di analyzer e rappresentazione neutrale di simboli e relazioni;
- almeno un analyzer AST di riferimento senza logica nel Runtime Core;
- retrieval lessicale deterministico;
- retrieval semantico opt-in tramite embedding provider;
- Context Builder con deduplicazione, priorità e budget espliciti;
- stima token con metodo dichiarato e margine di sicurezza;
- cache in-memory limitata e invalidazione content-addressed;
- integrazione additiva nel composition root pubblico;
- integrazione framework-neutral con il plugin Laravel;
- eventi redatti, test concorrenti, benchmark e documentazione.

## Escluso

- memoria conversazionale o di lungo periodo degli agenti;
- pianificazione, tool calling e permission model operativo;
- modifica del workspace, watcher filesystem e indexing continuo;
- indicizzazione remota o multi-host;
- vector database o persistenza obbligatoria dell'indice;
- upload automatico dei contenuti verso servizi remoti;
- parser completi per ogni linguaggio o analisi framework-specific nel core;
- ranking appreso, reranker LLM e riassunti generati automaticamente;
- scelta automatica di provider, modello o tokenizer;
- garanzia di conteggio token esatto senza un estimator model-specific;
- costruzione o invio della richiesta di completion.

Memoria di sessione, autorizzazioni e mutazioni appartengono alla Milestone 7.
Persistenza distribuita ed estensioni di terze parti appartengono alla
Milestone 8.

---

# Modello architetturale

## Servizio separato dal runtime.Context

Il Context Engine è un servizio composto dal Runtime, non un'estensione del
`runtime.Context` usato per cancellazione e lifecycle. Per evitare ambiguità
con `context.Context` della libreria standard e con `pkg/runtime.Context`, il
package pubblico è `pkg/contextengine`; l'implementazione concreta rimane in
`internal/contextengine`.

Il composition root espone il servizio tramite `Runtime.ContextEngine`. Il
contratto di basso livello `pkg/runtime.Runtime` resta invariato.

## Pipeline con ownership separate

```text
Workspace Source
      |
      v
Snapshot Index -----> Analyzer Registry
      |                      |
      +----------+-----------+
                 v
          Retrieval Engine
                 |
                 v
          Context Builder
                 |
                 v
          Context Bundle
```

Ogni livello possiede invarianti distinti:

- la source legge e descrive risorse;
- l'indice possiede identità, digest, generazione e snapshot;
- gli analyzer producono evidenza strutturata senza modificare documenti;
- il retrieval produce candidati con score e motivazioni;
- il builder seleziona e ordina sezioni entro il budget;
- la cache conserva risultati derivati, non stato autorevole.

Il Context Engine coordina la pipeline ma non duplica Provider Runtime, Gestor,
Plugin Runtime o filesystem tools.

---

# Modello workspace

## Workspace

Un workspace identifica almeno:

- ID stabile nella richiesta;
- root assoluta e normalizzata;
- policy di scansione;
- metadata framework-neutral limitati.

La root reale non viene usata come identità pubblica nei log o negli eventi.
I documenti espongono path logici relativi con separatore `/`; path assoluti e
sequenze che escono dalla root non sono ammessi.

Un contratto opzionale di workspace provider permette a componenti e plugin di
descrivere workspace già rilevati. Laravel costituisce il primo consumer; il
contratto non include versione Laravel, Composer o altri campi di framework.

## Policy di scansione

La scansione è governata da limiti espliciti:

- numero massimo di file;
- dimensione massima per file e complessiva;
- regole di inclusione ed esclusione;
- gestione di file nascosti e binari;
- comportamento sui symlink;
- encoding testuali accettati.

La baseline non segue symlink e non attraversa mount o path esterni alla root.
Directory di metadata e dipendenze note possono essere escluse da default
documentati, ma la configurazione effettiva deve essere osservabile nello
snapshot. La compatibilità completa con la grammatica `.gitignore` non viene
assunta senza una decisione e test dedicati.

## Documenti e snapshot

Un documento contiene identità logica, digest del contenuto, media type,
dimensione e testo normalizzato quando supportato. Timestamp e inode possono
aiutare la scansione, ma non sostituiscono il digest come identità del
contenuto.

Uno snapshot è immutabile e comprende:

- workspace e policy effettiva;
- generazione monotona;
- documenti in ordine deterministico;
- risultati strutturati e diagnostiche;
- statistiche aggregate prive di contenuto.

Il refresh è all-or-nothing per gli errori infrastrutturali: cancellazione,
superamento dei limiti o failure della source non sostituiscono l'ultimo
snapshot valido. File non analizzabili possono invece produrre diagnostiche
locali e restare disponibili come testo.

---

# Analisi strutturata e AST

La logica specifica di linguaggio non appartiene al Runtime Core. Analyzer
registrabili dichiarano in modo esplicito linguaggi, estensioni o media type
supportati e una versione utile alle chiavi di cache.

Il modello neutrale iniziale rappresenta soltanto evidenza consumabile dal
retrieval:

- simboli con kind, nome e intervallo sorgente;
- contenitori e relazioni dichiarative;
- import o dipendenze sorgente;
- diagnostiche con codici sicuri;
- chunk strutturali riferiti al documento originale.

Il modello non tenta di unificare ogni semantica dei linguaggi. Nodi AST
arbitrari, type checking completo e compiler database restano proprietà degli
analyzer specializzati.

Più analyzer applicabili allo stesso documento non vengono ordinati con una
precedenza nascosta. La configurazione deve selezionarne uno oppure richiedere
una composizione esplicita. Un analyzer AST Go basato sulla libreria standard è
la prima implementazione plausibile per verificare il contratto; analyzer PHP
e arricchimenti Laravel rimangono sostituibili e fuori dal core.

---

# Retrieval

## Baseline lessicale

Il retrieval lessicale è sempre disponibile e deterministico. Tokenizzazione,
normalizzazione, campi ricercati, formula dello score e tie-break devono essere
versionati e testabili. L'ordine dei path non diventa una preferenza semantica:
può essere usato soltanto come tie-break dichiarato.

## Retrieval strutturale

Query per simbolo, path, linguaggio o relazione usano l'evidenza degli
analyzer. Un risultato conserva sempre documento, intervallo, score, metodo e
codice di motivazione.

## Retrieval semantico

Gli embedding sono opt-in. L'adapter interno usa il Provider Runtime esistente
senza introdurre un secondo registry provider. Provider e modello sono scelti
esplicitamente dalla richiesta o dalla configurazione dell'applicazione.

Le chiavi di cache includono almeno digest del contenuto, strategia di chunk,
provider, modello e dimensione osservata. Vettori con dimensioni differenti o
valori non finiti non possono condividere un indice.

La fusione tra ranking lessicale, strutturale e semantico usa una strategia
esplicita. Il primo incremento non introduce ranking appreso o chiamate LLM di
reranking.

---

# Context Builder e budget

Il builder riceve una richiesta validata, candidati e un budget. Produce un
`ContextBundle` immutabile composto da sezioni ordinate. Ogni sezione espone:

- documento e intervallo di origine;
- testo selezionato;
- ruolo o categoria;
- metodo e motivazione di selezione;
- costo stimato;
- eventuale indicazione di troncamento.

Il budget distingue almeno spazio totale, quota riservata alla richiesta e
quota disponibile per l'evidenza. Le strategie di selezione e troncamento sono
esplicite e deterministiche. Deduplicazione e sovrapposizione operano sugli
intervalli originali, non su confronti lossy del testo renderizzato.

Un `TokenEstimator` dichiara metodo e versione. Una stima conservativa può
costituire la baseline offline; un tokenizer model-specific può essere
registrato quando disponibile. Il bundle non presenta una stima come conteggio
esatto e non supera il budget calcolato dal metodo dichiarato.

Il builder non invoca completion, non riscrive la domanda e non genera riassunti
con un modello. Queste decisioni appartengono al consumer futuro.

---

# Cache e aggiornamento incrementale

La cache iniziale è in-memory, limitata e sostituibile. Conserva artefatti
derivati content-addressed:

- testo normalizzato;
- analisi strutturata;
- chunk;
- embedding;
- stime di costo.

Le chiavi includono tutte le dipendenze semantiche rilevanti, comprese versioni
di analyzer, chunker ed estimator. Una nuova generazione riusa soltanto
artefatti con chiavi identiche; la generazione dello snapshot non fa parte
dell'identità del contenuto.

Eviction e limiti di memoria sono configurabili e osservabili. Una cache miss
non cambia il risultato funzionale, soltanto il costo. Failure, cancellazione e
risultati parziali non vengono pubblicati come entry valide. Nessun singleflight
tra context distinti viene assunto senza un contratto esplicito di ownership
della cancellazione.

La milestone non richiede cache persistente. Indice e cache non devono
serializzare automaticamente contenuti del workspace, embedding o path assoluti.

---

# Integrazione con Runtime, Gestor e Plugin

Il Runtime compone una singola istanza del Context Engine e condivide Config,
Logger ed Event Bus. Registrazioni di source, analyzer ed estimator avvengono
prima dell'uso operativo e producono snapshot difensivi.

Gestor continua a risolvere dichiarazioni e dependency plan; non indicizza file
e non esegue analyzer. Le capability del Context Engine possono essere
pubblicate tramite gli adapter Gestor esistenti senza trasformare Gestor in un
service locator di istanze.

Il plugin Laravel espone il workspace attraverso il nuovo contratto
framework-neutral, ma route, Eloquent, Blade e Composer restano conoscenza del
plugin o di analyzer dedicati. Il Context Engine non importa il package
Laravel.

Il Developer Benchmark usa il percorso pubblico del Context Engine mantenendo
dataset, rubrica e reporting canonici.

---

# Concorrenza, cancellazione e failure

- source, analyzer, embedding provider ed eventi sono invocati senza lock
  interni del Context Engine;
- ogni operazione I/O riceve il `context.Context` del chiamante;
- una cancellazione non pubblica un nuovo snapshot o una cache entry parziale;
- lettori concorrenti osservano uno snapshot completo vecchio o nuovo;
- bundle e listing sono snapshot difensivi;
- una failure locale di parsing è una diagnostica, non necessariamente una
  failure dell'intero refresh;
- limiti violati e path non sicuri sono errori tipizzati e ispezionabili;
- panic di estensioni trusted in-process non viene presentato come successo.

---

# Osservabilità e riservatezza

Gli eventi descrivono operazioni e contatori, non contenuti. Payload pubblici e
log possono includere workspace ID, generazione, durata, conteggi, hit/miss e
codici di esito; non includono testo, query, embedding, path assoluti, error
string esterne o possibili credenziali.

Un `ContextBundle` contiene necessariamente porzioni del workspace ed è quindi
dato sensibile affidato esplicitamente al chiamante. Non viene pubblicato
sull'Event Bus o conservato in report benchmark. L'invio a un provider avviene
soltanto per azione del consumer e non durante indexing o build.

---

# Invarianti principali

- uno snapshot pubblicato è immutabile e appartiene a una sola generazione;
- path logici non escono mai dalla root del workspace;
- refresh falliti conservano l'ultimo snapshot valido;
- contenuto uguale produce la stessa identità content-addressed;
- analyzer e codice esterno non vengono invocati sotto lock;
- cache e indice autorevole restano concetti distinti;
- ogni risultato di retrieval dichiara metodo, score e origine;
- ogni bundle dichiara estimator, budget e costo;
- nessun provider o modello viene scelto per ordine di registrazione;
- nessun contenuto workspace entra implicitamente in eventi, log o report;
- logica framework-specific e language-specific resta fuori dal Runtime Core;
- il Context Engine non possiede memoria conversazionale o tool permissions.

---

# Strategia di verifica

La suite ordinaria usa esclusivamente directory temporanee, source e analyzer
in-memory e provider embedding fittizi. Deve coprire:

- normalizzazione e containment dei path;
- file testuali, binari, grandi, nascosti e symlink;
- limiti, cancellazione e failure atomiche;
- ordine deterministico e snapshot concorrenti;
- parsing valido, malformato, unsupported e analyzer concorrenti;
- ranking e tie-break lessicali, strutturali e semantici;
- dimensioni embedding e valori non finiti;
- deduplicazione, troncamento e budget limite;
- cache hit, miss, invalidazione, eviction e concorrenza;
- eventi redatti e callback fuori lock;
- percorso end-to-end dal composition root con Laravel;
- regressione del Developer Benchmark senza rete obbligatoria.

Il gate finale richiede suite completa, race detector sui package coinvolti,
`go vet ./...`, audit di compatibilità pubblica e documentazione allineata.

---

# Decisioni formalizzate nella Fase 1

- package e contratti pubblici del Context Engine;
- ownership di workspace, snapshot, analyzer, retrieval, builder e cache;
- refresh atomico e diagnostiche locali per file non analizzabili;
- path relativi sicuri e symlink non seguiti nella baseline;
- analyzer language-specific registrabili e fuori dal Runtime Core;
- retrieval semantico opt-in tramite Provider Runtime esistente;
- budget basato su estimator dichiarato, senza falsa precisione;
- cache in-memory content-addressed, limitata e non autorevole;
- integrazione additiva con Runtime, Gestor e plugin workspace;
- confine di riservatezza tra eventi redatti e bundle affidato al chiamante.

Queste decisioni sono raccolte in ADR-0024. Il piano operativo è descritto in
`context-engine-development-plan.md`.

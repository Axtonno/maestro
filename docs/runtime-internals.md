# Runtime Internals

## Scopo del documento

Questo documento definisce le convenzioni implementative adottate all'interno del package `internal/runtime`.

Le regole qui descritte non costituiscono un contratto pubblico del progetto. Servono a preservare la coerenza interna del Runtime, rendere esplicite le responsabilità dei diversi tipi e facilitare l'evoluzione dell'implementazione senza compromettere gli invarianti del sistema.

## Principio fondamentale

Ogni tipo interno del Runtime protegge gli invarianti del proprio livello di responsabilità.

I tipi interni non devono essere necessariamente immutabili. Possono mantenere e modificare uno stato interno, ma tale mutabilità deve essere controllata dal tipo proprietario.

La rappresentazione interna non deve essere manipolata liberamente dal resto del package.

## Incapsulamento della rappresentazione

Tutti i campi dei tipi definiti in `internal/runtime` devono rimanere privati.

Mappe, slice, puntatori o altre strutture mutabili interne non devono essere esposti direttamente quando ciò consentirebbe al chiamante di alterare lo stato senza passare attraverso il tipo proprietario.

Quando una collezione deve essere resa disponibile, il tipo deve restituire:

* una copia della collezione;
* una vista in sola lettura;
* oppure un risultato che non permetta di modificare la rappresentazione interna.

## Operazioni intenzionali

Le modifiche allo stato devono avvenire attraverso metodi che esprimono un'intenzione di dominio.

Esempi:

```go
node.AddDependency(dependency)
node.AddDependent(dependent)
graph.Add(component)
graph.AddDependency(componentID, dependencyID)
```

Sono da evitare modifiche dirette come:

```go
node.dependencies = append(node.dependencies, dependency)
graph.nodes[id] = node
```

al di fuori del tipo responsabile di mantenere tali strutture.

I metodi intenzionali costituiscono il punto nel quale applicare validazioni, prevenire duplicati, mantenere relazioni coerenti e introdurre future ottimizzazioni.

## Proprietà degli invarianti

Gli invarianti locali sono mantenuti dal tipo che possiede direttamente lo stato.

Gli invarianti che coinvolgono più entità sono mantenuti dal relativo aggregato.

### Node

`node` protegge gli invarianti relativi a un singolo nodo.

È responsabile della propria rappresentazione interna e delle collezioni di dipendenze e dipendenti.

Non deve coordinare autonomamente modifiche che coinvolgono altri nodi del grafo.

### Graph

`graph` protegge la coerenza complessiva tra i nodi.

È responsabile di:

* registrare i nodi;
* impedire identificativi duplicati;
* creare relazioni tra componenti;
* mantenere coerenti le direzioni `dependencies` e `dependents`;
* impedire che chiamanti esterni alterino direttamente la propria struttura.

Le operazioni che coinvolgono più nodi devono essere coordinate dal grafo, non dai singoli chiamanti.

### Registry

`registry` protegge la collezione dei componenti registrati.

È responsabile di:

* impedire registrazioni duplicate;
* risolvere componenti tramite `ComponentID`;
* proteggere l'accesso concorrente;
* non esporre direttamente la mappa interna.

### Event Bus

`eventBus` protegge la collezione dei subscriber.

È responsabile di:

* validare eventi, topic e handler;
* mantenere l'ordine di registrazione degli handler;
* consentire più handler per topic;
* creare uno snapshot dei subscriber per ogni pubblicazione;
* non eseguire callback mentre mantiene un lock interno;
* rimuovere in modo atomico tutti gli handler associati a un topic.

Evento e payload non appartengono al bus e non vengono copiati o modificati.
La loro eventuale mutabilità resta responsabilità dei componenti che li
condividono.

### Provider Runtime

Il Provider Runtime concreto vive in `internal/provider` e protegge la
collezione dei provider registrati e la selezione del default.

È responsabile di:

* validare provider e identificativi;
* impedire registrazioni duplicate;
* risolvere provider espliciti o il default configurato;
* verificare le capability richieste;
* inoltrare completion, streaming, embedding, model listing, discovery,
  lifecycle, acquisizione e rimozione dei modelli;
* coordinare policy opt-in di residenza attraverso lease, timer e ownership
  delle transizioni avviate da Maestro;
* inoltrare e validare report canonici di capability introspection senza cache;
* proteggere registry e default durante l'accesso concorrente;
* non mantenere lock interni durante l'esecuzione di codice del provider.

Richieste, risposte e stream appartengono al chiamante e al provider. Il router
non ne modifica o copia il contenuto e non possiede snapshot autorevoli o
progresso dei trasferimenti. Discovery rimane la fonte osservabile dello stato
dei modelli; il coordinamento di residenza conserva soltanto stato locale della
policy.

### Plugin Runtime

Il Plugin Runtime concreto vive in `internal/plugin` e protegge l'indice dei
componenti registrati come plugin e il catalogo dei loader.

È responsabile di:

* validare plugin e identificativi prima della registrazione;
* delegare la registrazione al Runtime Core;
* indicizzare un plugin soltanto dopo la registrazione del componente;
* impedire collisioni con plugin e componenti già registrati;
* risolvere soltanto componenti registrati attraverso il Plugin Runtime;
* validare manifest e versione dell'API Plugin Runtime;
* registrare e scoprire loader in ordine deterministico;
* eseguire i loader senza lock interni e rispettare la cancellazione;
* verificare che il loader produca il plugin richiesto;
* pubblicare eventi riusciti sul bus condiviso;
* proteggere indici e catalogo durante l'accesso concorrente;
* non eseguire metodi del plugin o del loader mentre mantiene lock interni.

Il Plugin Runtime non possiede stati, dependency graph o lifecycle. Questi
invarianti restano nel Runtime Core e vengono applicati ai plugin come a ogni
altro componente.

### Resolver

`resolver` interpreta le dipendenze dichiarate nei metadati e costruisce le relazioni necessarie.

Non modifica un grafo già validato e non conserva stato temporaneo tra diverse esecuzioni.

Il resolver costruisce ciò che è stato dichiarato, senza decidere autonomamente se la struttura risultante sia semanticamente valida.

### Validator

`validator` verifica la validità strutturale e semantica del grafo.

Non modifica il grafo durante la validazione.

Gli stati temporanei utilizzati dagli algoritmi di visita, come quelli necessari al rilevamento dei cicli, devono rimanere locali alla singola operazione di validazione.

### Builder

`builder` coordina il processo di costruzione.

Può mantenere riferimenti ai propri collaboratori, ma non deve condividere con essi uno stato di costruzione parziale.

Il grafo viene restituito soltanto quando la risoluzione e la validazione sono state completate con successo.

Un errore non deve lasciare nel builder o nel Runtime un oggetto parzialmente costruito.

### Runtime

`runtime` coordina i servizi interni e mantiene la coerenza tra il Registry e il grafo delle dipendenze corrente.

Quando una nuova registrazione rende obsoleto il grafo esistente, il Runtime deve invalidarlo.

Il Runtime non deve duplicare la logica appartenente a Registry, Graph, Resolver, Validator o Builder.

Il Runtime è anche il composition root dei servizi Config, Logger, Event Bus,
Provider Runtime e Plugin Runtime. Config, Logger, Event Bus e Provider Runtime
sono condivisi con il `runtimeContext` dei componenti; il Plugin Runtime è
esposto dal composition root pubblico.

## Costruttori

I costruttori interni devono creare oggetti che rispettino gli invarianti iniziali del tipo.

Esempi:

```go
newRegistry()
newNode(component)
newGraph()
newResolver(registry)
newValidator()
newBuilder(resolver, validator)
```

Un oggetto restituito da un costruttore deve essere immediatamente utilizzabile secondo il proprio contratto interno.

Le validazioni che riguardano la composizione di più oggetti possono essere eseguite dal relativo aggregato o dal processo di bootstrap.

## Regole operative

Per `internal/runtime` vengono adottate le seguenti regole:

1. Tutti i campi dei tipi interni rimangono privati.
2. Nessuna struttura mutabile interna viene restituita direttamente.
3. Le modifiche avvengono attraverso metodi che esprimono un'intenzione di dominio.
4. I costruttori garantiscono gli invarianti iniziali.
5. Ogni tipo mantiene gli invarianti del proprio livello di responsabilità.
6. Le modifiche che coinvolgono più oggetti sono coordinate dal relativo aggregato.
7. Resolver e Validator non modificano un grafo già costruito.
8. Il Builder non espone né conserva risultati parziali.
9. Il Runtime orchestra i servizi interni senza duplicarne le responsabilità.
10. Nuove ottimizzazioni interne non devono richiedere modifiche ai contratti pubblici.
11. L'Event Bus non mantiene lock interni durante l'esecuzione degli handler.
12. Il Provider Runtime non mantiene lock interni durante l'esecuzione dei provider.
13. Il Provider Runtime scarica tramite policy soltanto residenze caricate da Maestro.
14. Il Provider Runtime non memorizza report operativi di capability introspection.
15. Il Plugin Runtime indicizza un plugin soltanto dopo che il Runtime Core ne ha accettato la registrazione.
16. Il Plugin Runtime non duplica dependency graph, stato o lifecycle dei componenti.
17. Il Plugin Runtime non esegue loader mantenendo lock sul catalogo.
18. Un plugin viene registrato soltanto se il manifest richiede la versione API supportata.

## Evoluzione futura

Questa convenzione si applica inizialmente a `internal/runtime`.

Potrà essere estesa agli altri package interni di Maestro qualora emerga la necessità di adottare formalmente lo stesso modello di proprietà degli invarianti nell'intero progetto.

In quel caso, la convenzione potrà essere promossa a decisione architetturale generale.

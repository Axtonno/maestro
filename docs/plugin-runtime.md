# Maestro Plugin Runtime

Versione: 0.1.0

Stato: Primo incremento

Ultimo aggiornamento: 2026-08-05

---

# Scopo

Il Plugin Runtime registra e risolve le estensioni opzionali di Maestro senza
introdurre un secondo sistema di componenti o un secondo lifecycle.

Il contratto pubblico vive in `pkg/plugin`; l'implementazione concreta vive in
`internal/plugin`. Il composition root restituito da `maestro.New` espone la
stessa istanza tramite `Plugins()`.

---

# Modello

Un `Plugin` è un `runtime.Component` registrato attraverso il Plugin Runtime.
Usa quindi i metadati, le dipendenze e le capability lifecycle già definite dal
Runtime Core.

Questa relazione produce una sola fonte di verità:

* `Metadata.ID` identifica sia il componente sia il plugin;
* `Metadata.Dependencies` partecipa al dependency graph globale;
* `Configurer`, `Initializer`, `Starter`, `Stopper`, `Reloader` e
  `HealthChecker` mantengono la semantica del Runtime Core;
* stato e failure sono osservabili attraverso lo `StateManager` globale.

Il Plugin Runtime mantiene un indice dedicato soltanto per distinguere i plugin
dai componenti registrati direttamente. In Go l'interfaccia `Plugin` è
strutturale: è il percorso `Plugins().Register`, non un marker aggiuntivo, a
classificare un componente come plugin.

---

# Registrazione e risoluzione

Il Plugin Runtime espone:

* `Register`, che valida l'identità minima e registra il plugin anche nel
  Runtime Core;
* `Resolve`, che restituisce soltanto componenti registrati come plugin;
* `Has`, per verificare la presenza di un plugin.

Gli ID vuoti, composti da soli spazi o con spazi iniziali/finali vengono
rifiutati. Un ID occupato da un altro plugin o da un normale componente non può
essere registrato di nuovo.

Come per ogni componente, `Metadata.ID` deve rimanere stabile dopo la
registrazione. L'indice e il dependency graph usano quell'identità come
invariante condiviso.

La registrazione deve avvenire prima dell'avvio del Runtime, come per ogni
altro componente. Dopo che `Register` restituisce successo, plugin registry,
component registry e state manager descrivono lo stesso componente.

---

# Lifecycle e dipendenze

Il Plugin Runtime non invoca direttamente codice lifecycle.

All'avvio, il Runtime Core costruisce il grafo globale e avvia ogni plugin dopo
le sue dipendenze richieste. All'arresto usa l'ordine inverso. Le dipendenze
possono riferirsi indifferentemente a plugin o ad altri componenti.

Il `runtime.Context` ricevuto dalle capability del plugin espone configurazione,
logger, Event Bus, component Registry e Provider Runtime condivisi.

---

# Concorrenza e ownership

L'indice dei plugin è thread-safe e non espone la mappa interna. I metodi del
plugin non vengono eseguiti mantenendo il lock dell'indice.

Il Plugin Runtime possiede la classificazione e la risoluzione dei plugin. Il
Runtime Core continua a possedere registrazione dei componenti, dependency
graph, stati e lifecycle.

---

# Estensioni escluse dal primo incremento

Questa versione non introduce:

* discovery da directory o manifest;
* installazione e risoluzione delle versioni;
* caricamento di shared object Go;
* plugin eseguiti in processi isolati;
* sandbox o verifica delle firme;
* unload e hot reload del codice;
* enable/disable persistente.

Questi aspetti richiedono contratti di compatibilità, sicurezza e distribuzione
separati. Verranno progettati senza modificare la semantica di registrazione e
lifecycle definita da questo incremento.

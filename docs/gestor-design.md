# Gestor — Design iniziale

Versione: 0.6.0

Stato: Implementato — Milestone 4 completata

Data: 2026-08-09

---

# Obiettivo

Gestor è il registro centrale delle capability di Maestro. Deve permettere al
Runtime di scoprire quali soggetti dichiarano una capability, distinguere la
dichiarazione dalla disponibilità operativa e produrre una risoluzione
deterministica e verificabile.

Gestor coordina informazioni. Non esegue completion, plugin, tool o lifecycle.

La Milestone 4 parte dalla baseline già disponibile:

- Registry dei componenti nel Runtime Core;
- dependency graph, resolver, validator e builder nel Runtime Core;
- Provider Runtime con capability opzionali e introspection a tre livelli;
- Plugin Runtime integrato nel Registry globale dei componenti;
- Event Bus, stato e lifecycle condivisi.

Gestor estende questa baseline senza duplicarne la proprietà.

---

# Decisioni iniziali

1. Gestor indicizza capability, ma non possiede i componenti.
2. Il dependency graph resta unico e appartiene al Runtime Core.
3. Gestor non esegue codice risolto e non governa lifecycle o stato.
4. Capability dichiarata e capability operativamente disponibile sono stati
   distinti.
5. La risoluzione non applica ranking impliciti o euristiche nascoste.
6. Discovery e refresh producono snapshot atomici e deterministici.
7. Nessuna sorgente esterna viene interrogata mentre Gestor mantiene lock.
8. Plugin già registrati come componenti non vengono duplicati in un secondo
   catalogo Gestor.
9. I contratti esistenti di Runtime, Provider e Plugin restano autorevoli nei
   rispettivi domini.
10. Il primo incremento rimane in-process e non introduce persistenza o
    discovery remota.

---

# Confini di responsabilità

## Gestor possiede

- indice capability → candidati;
- snapshot delle dichiarazioni e dell'evidenza operativa;
- discovery coordinata dalle sorgenti registrate;
- query dei candidati;
- selezione esplicita e deterministica;
- spiegazione della risoluzione;
- invalidazione e generazioni degli snapshot.

## Gestor non possiede

- istanze dei componenti;
- registrazione primaria dei componenti;
- dependency graph mutabile;
- lifecycle, stato o rollback;
- routing delle operazioni provider;
- caricamento dei plugin;
- probe automatici durante `Resolve`;
- esecuzione della capability risolta;
- benchmark o valutazione qualitativa dei modelli.

---

# Architettura logica

```text
Runtime Registry -----------+
                            |
Provider capability source -+--> Discovery --> Snapshot Registry
                            |                       |
Future sources -------------+                       v
                                               Capability Resolver
                                                     |
Runtime dependency graph ----------------------------+
                                                     |
                                                     v
                                              Resolution Result
```

Le sorgenti descrivono ciò che conoscono. Il Registry costruisce uno snapshot
coerente. Il Resolver filtra i candidati e consulta una vista read-only del
dependency graph quando il candidato è un componente Runtime.

---

# Modello di dominio

## Capability ID

Gestor usa identificatori stringa stabili e namespaced, per esempio:

```text
runtime.start
runtime.health
provider.completion
provider.streaming
provider.embedding
provider.structured_output
provider.tool_calling
plugin.workspace_detection
```

I tipi esistenti `runtime.Capability` e `provider.Capability` restano invariati.
Adapter di discovery espliciti li traducono negli ID Gestor; il primo incremento
non cambia i valori pubblici già esposti.

## Target

Una capability appartiene a un target identificabile:

| Campo | Significato |
|---|---|
| `kind` | `component`, `provider` o un tipo futuro esplicito |
| `id` | ID esatto nel registry autorevole |
| `scope` | `component`, `adapter`, `instance` o `model` |
| `model` | modello esatto quando lo scope è `model` |

I plugin usano `kind=component`, perché sono già componenti del Runtime Core.

## Dichiarazione e disponibilità

La presenza di un descriptor nel Registry significa soltanto che la capability
è dichiarata. La disponibilità è rappresentata separatamente:

| Stato | Significato |
|---|---|
| `unknown` | dichiarata, ma non verificata operativamente |
| `available` | verificata per il target e lo scope indicati |
| `unavailable` | verificata come non disponibile per quel target e scope |

Ogni descriptor conserva inoltre la sorgente. La generazione appartiene allo
`SnapshotMetadata`, non viene duplicata su ogni descriptor.
Timestamp o scadenza dell'evidenza potranno essere aggiunti quando esisterà una
policy reale di refresh; non fanno parte del primo contratto.

Questa distinzione rende rappresentabili sia una fixture operativa come
`llama3.1:8b`, sia un caso come `qwen2.5-coder:7b`, dove il supporto tools è
dichiarato ma non validato operativamente.

## Descriptor

Il descriptor minimo contiene:

```go
type Descriptor struct {
    Capability  CapabilityID
    Target      Target
    Availability Availability
    Source      SourceID
}
```

Questa forma è il contratto approvato in ADR-0022. `CapabilityID`, `SourceID`,
`Target`, `Availability`, `Descriptor`, `Query`, `Snapshot`, `Resolution` e le
interfacce minime appartengono a `pkg/gestor`. Le traduzioni dai tipi Runtime e
Provider restano in `internal/gestor`.

Gli identificatori sono comparabili per uguaglianza esatta e ordinati in modo
lessicografico. La validazione non corregge o normalizza input: valori vuoti,
namespace invalidi e whitespace ai bordi vengono rifiutati.

## Immutabilità dei value object

`NewQuery`, `NewSnapshot`, `NewSnapshotWithSources` e `NewResolution`
acquisiscono copie delle slice.
Getter di preferenze, descriptor, sorgenti e dependency plan restituiscono
sempre copie difensive. `SnapshotMetadata` contiene generazione, stato current,
numero di descriptor e sorgenti ordinate; non espone mappe o backing slice.

`NewSnapshotWithSources` registra anche sorgenti consultate che non hanno
prodotto descriptor. Uno snapshot vuoto può quindi dimostrare che il refresh ha
interrogato correttamente il catalogo configurato.

---

# Sorgenti di discovery

## Runtime component source

Legge uno snapshot del Registry dei componenti e traduce
`runtime.Metadata.Capabilities`. Non conserva le istanze e non esegue metodi dei
componenti.

Le capability lifecycle dichiarate devono continuare a essere validate contro
le interfacce opzionali dal Runtime Core; Gestor non duplica questa validazione.

La Fase 3 implementa `RuntimeComponentSource` su una vista interna
`Components()` del Registry autorevole. Lo snapshot dei componenti viene
ordinato per component ID e tradotto senza conservare istanze. Le sei
capability lifecycle note diventano ID `runtime.*`; capability custom già
namespaced, come `plugin.workspace_detection`, vengono conservate esattamente.
Ogni dichiarazione Runtime ha availability `unknown`, perché i metadata non
costituiscono una prova operativa.

## Provider capability source

Adatta i report di capability già prodotti dal Provider Runtime:

- target adapter → supporto strutturale;
- target instance → disponibilità dell'istanza configurata;
- target model → disponibilità per un model ID esatto.

La sorgente esegue l'eventuale introspection durante `Refresh`, mai durante
`Resolve`. I report restano snapshot e Gestor non li trasforma in verità
permanenti.

L'integrazione userà un'interfaccia sorgente additiva implementata internamente,
senza aggiungere metodi obbligatori all'interfaccia pubblica
`provider.Runtime`.

La Fase 3 aggiunge al concrete Provider Runtime il listing interno
`Registered()`, ordinato e con copia difensiva. `ProviderCapabilitySource`
interroga sempre target adapter e instance per ogni provider registrato. I
target model sono model ID esatti forniti alla sorgente da una lista interna,
ordinata e copiata: non vengono inferiti dal default, dal nome o da discovery
implicita.

Ogni report viene validato nuovamente al confine della sorgente. Le capability
`supported` producono descriptor Gestor mantenendo esattamente availability
`unknown`, `available` o `unavailable`; le capability `unsupported` non sono
dichiarazioni Gestor e vengono omesse. Errori o cancellazione interrompono
l'intera sorgente e impediscono al Registry di pubblicare risultati parziali.

## Plugin

Non esiste una sorgente plugin separata nel primo incremento. Un plugin
registrato correttamente è già presente nel Registry dei componenti e viene
scoperto da quella sorgente. Il Plugin Runtime continua a possedere catalogo e
loader.

I test della Fase 3 registrano il plugin attraverso il Plugin Runtime e lo
osservano una sola volta nella sorgente componenti globale.

## Sorgenti future

Tool, agent e Context Engine potranno aggiungere sorgenti dedicate senza
modificare il Resolver, purché producano descriptor validi e usino target kind
espliciti.

---

# Snapshot Registry

Il Registry Gestor mantiene una vista immutabile indicizzata almeno per:

- capability ID;
- target kind e target ID;
- coppia capability–target.

Invarianti:

1. nessun capability ID, source ID o target ID vuoto;
2. nessun duplicato capability–target prodotto dalla stessa sorgente;
3. ordine pubblico lessicografico e stabile;
4. slice e mappe interne non vengono esposte direttamente;
5. un refresh fallito conserva l'ultimo snapshot valido;
6. un nuovo snapshot diventa visibile con un singolo swap atomico;
7. ogni swap incrementa la generazione;
8. collisioni tra sorgenti autorevoli sono errori, non merge impliciti.

Il refresh segue questo flusso:

1. acquisisce la lista ordinata delle sorgenti;
2. rilascia i lock;
3. interroga ogni sorgente con il `context.Context` del chiamante;
4. valida e compone un candidato snapshot locale;
5. se tutte le sorgenti riescono, pubblica atomicamente lo snapshot;
6. se una sorgente fallisce, scarta il candidato e conserva lo snapshot
   precedente.

La Fase 2 implementa questo flusso in `internal/gestor.Registry`. Lo snapshot
iniziale ha generazione zero ed è stale. Ogni refresh riuscito, incluso quello
con zero sorgenti o zero descriptor, incrementa la generazione e pubblica uno
snapshot current. `Invalidate` conserva descriptor e generazione, marca lo
snapshot stale e invalida ogni refresh già in corso.

Il catalogo possiede un'epoch interna. `Refresh` copia sorgenti ed epoch sotto
read lock, ordina per source ID, rilascia il lock e solo allora chiama
`Discover`. Prima dello swap verifica che epoch e context siano ancora validi;
una registrazione o invalidazione concorrente produce `ErrStaleSnapshot` e il
candidato viene scartato. Refresh concorrenti sulla stessa epoch mantengono
stato locale separato e pubblicano ciascuno una generazione monotona al proprio
completamento.

La registrazione di una nuova sorgente invalida lo snapshot corrente senza
incrementarne la generazione. Source ID duplicati, typed nil, descriptor con
source incoerente, collisioni capability–target e snapshot invalidi sono
rifiutati prima della pubblicazione.

Lo snapshot interno costruisce indici immutabili per capability esatta e
`Target` esatto. Gli indici alimentano il Resolver e non sono
esposti dal package pubblico; ogni lettura interna restituisce comunque una
copia ordinata.

---

# Risoluzione

Una query specifica:

- capability richiesta;
- eventuale target kind;
- eventuale lista ordinata di target preferiti;
- eventuale scope o model ID esatto;
- se è richiesta evidenza operativa `available`.

Algoritmo iniziale:

1. valida la query;
2. legge un singolo snapshot immutabile;
3. seleziona i descriptor della capability richiesta;
4. applica target kind, scope e model ID;
5. esclude `unavailable`;
6. se la query richiede disponibilità operativa, esclude anche `unknown`;
7. verifica l'eventuale eleggibilità nel dependency graph per i componenti;
8. applica nell'ordine i target preferiti espliciti;
9. se resta un solo candidato, lo restituisce;
10. se non resta alcun candidato, restituisce `not found` o
    `capability unavailable`;
11. se restano più candidati senza preferenza risolutiva, restituisce
    `ambiguous resolution`.

L'ordinamento lessicografico rende deterministici listing e diagnostica, ma non
sceglie silenziosamente un vincitore. Default e preferenze appartengono alla
configurazione o alla richiesta esplicita.

Il risultato contiene il descriptor scelto, la generazione consultata, la
motivazione della selezione e, per un componente, le dipendenze in ordine
topologico. Non contiene funzioni da invocare.

La Fase 4 implementa questo algoritmo in `internal/gestor.Resolver`. La lettura
di snapshot e indice capability avviene con una singola acquisizione atomica;
`Candidates` restituisce tutti i candidati eleggibili in ordine deterministico,
mentre soltanto `Resolve` applica le preferenze esatte nell'ordine dichiarato.
Una preferenza assente o non eleggibile non diventa un fallback implicito: se
restano più candidati la risoluzione è ambiguous.

Snapshot Gestor e grafo Runtime hanno generazioni indipendenti. Il Resolver
cattura quelle consultate e verifica prima di restituire che entrambe siano
ancora current e invariate. Un cambiamento concorrente produce
`ErrStaleSnapshot`, non un risultato composto da revisioni differenti.

---

# Integrazione con il dependency graph

Il grafo esistente resta la fonte unica per dipendenze richieste, dipendenze
opzionali, cicli e ordine topologico.

Gestor riceve una vista read-only o un collaboratore interno dedicato. Non
espone `internal/runtime.graph`, non aggiunge archi e non conserva una copia
mutabile del grafo.

La vista introdotta nella Fase 4 espone internamente soltanto stato della
generazione e piano delle dipendenze transitive. Il piano viene estratto dal
grafo autorevole e filtrato dal suo ordine topologico, quindi è dependency-first
e non muta nodi o archi. L'identità del nodo viene acquisita durante la
costruzione del grafo: `Resolve` non richiama `Metadata` né altro codice del
candidato.

Ogni registrazione riuscita di un componente incrementa la generazione del
catalogo e rende il grafo stale. Ogni ricostruzione riuscita incrementa la
generazione del grafo e la associa alla generazione dei componenti usata. Una
dipendenza richiesta mancante impedisce la costruzione; una dipendenza
opzionale mancante viene omessa dal piano. Cicli e validazione restano nel
Runtime Core.

Se una registrazione invalida il grafo Runtime o uno snapshot Gestor, il
composition root marca entrambi come non correnti. Nessun refresh con I/O viene
eseguito sotto il lock del Runtime. Una risoluzione su snapshot non corrente
fallisce in modo esplicito oppure richiede prima un refresh; non avvia probe
impliciti.

---

# Concorrenza

- registrazione delle sorgenti e pubblicazione degli snapshot sono thread-safe;
- discovery, introspection, callback ed eventi avvengono senza lock Gestor;
- `Resolve` e listing lavorano su uno snapshot immutabile;
- lo stato temporaneo di una risoluzione appartiene alla singola chiamata;
- snapshot e generazione del grafo vengono ricontrollati prima dell'output;
- la cancellazione interrompe il refresh e non pubblica risultati parziali;
- i test con race detector sono parte del gate della milestone.

Il concrete Provider Runtime espone inoltre un hook interno di invalidazione.
Le registrazioni riuscite di componenti, plugin e provider invocano lo stesso
invalidatore dopo il rilascio dei rispettivi lock. Le registrazioni fallite non
invalidano. L'hook resta fuori dalle interfacce pubbliche Runtime, Provider e
Plugin ed è collegato al Registry Gestor dal composition root.

Il bootstrap registra le sorgenti `runtime.components` e
`provider.capabilities`, quindi pubblica uno snapshot iniziale current anche
quando i cataloghi sono vuoti. La sorgente provider built-in interroga target
adapter e instance. Nessun model ID viene inferito dal default o dal catalogo:
target model ulteriori richiedono dichiarazioni esatte attraverso il confine
pubblico `Source`.

---

# Errori e diagnostica

Il contratto distingue almeno:

| Errore | Significato |
|---|---|
| invalid query | capability o filtro non valido |
| source failure | discovery non completata |
| not found | nessun target dichiara la capability |
| unavailable | target dichiarati, ma nessuno operativamente eleggibile |
| ambiguous | più candidati senza preferenza esplicita |
| stale snapshot | registrazione o grafo cambiati dopo il refresh |
| invalid descriptor | una sorgente ha prodotto dati incoerenti |

Gli errori devono preservare cause e identificatori tecnici, senza includere
prompt, risposte, secret o configurazioni sensibili.

Topic pubblici stabili:

```text
gestor.refresh.started
gestor.refresh.completed
gestor.refresh.failed
gestor.resolution.completed
gestor.resolution.failed
```

Gli eventi vengono pubblicati prima o dopo l'operazione, mai mantenendo lock
Gestor. Il payload pubblico contiene soltanto capability, generazione, conteggi,
kind/scope, reason e una categoria d'errore. Non contiene messaggi d'errore,
source detail, target o model ID, prompt, risposte o credenziali. Errori e panic
degli observer non modificano refresh o resolution; callback sincroni lenti
possono aggiungere latenza ma osservano già lo stato pubblicato.

---

# Contratti pubblici approvati

La forma pubblica minima è:

```go
type Source interface {
    ID() SourceID
    Discover(context.Context) ([]Descriptor, error)
}

type Registry interface {
    RegisterSource(Source) error
    Refresh(context.Context) error
    Invalidate()
    Snapshot() Snapshot
}

type Resolver interface {
    Candidates(Query) ([]Descriptor, error)
    Resolve(Query) (Resolution, error)
}

type Service interface {
    Registry
    Resolver
}
```

`Source` è il confine di estensione, `Registry` espone deliberatamente refresh e
invalidazione espliciti, e `Resolver` consuma soltanto query validate. Le
implementazioni concrete restano in `internal/gestor`. Il composition root
pubblico di Maestro espone `Gestor() gestor.Service`; `pkg/runtime.Runtime` non
è stato ampliato, preservando il contratto del Runtime Core.

---

# Fasi della Milestone 4

Il piano operativo, i gate e i report obbligatori sono definiti in
`gestor-development-plan.md`.

## Fase 1 — Contratti e ADR

- formalizzare ID, target, availability, query, resolution ed errori;
- decidere il confine pubblico minimo;
- registrare proprietà e invarianti in ADR-0022;
- aggiungere fixture contrattuali.

## Fase 2 — Snapshot Registry

- catalogo sorgenti thread-safe;
- refresh atomico all-or-nothing;
- indici immutabili e listing deterministico;
- invalidazione e generazioni.

## Fase 3 — Discovery sources

- sorgente Runtime component metadata;
- sorgente Provider capability introspection;
- mapping namespaced senza modificare i contratti esistenti;
- integrazione plugin tramite Registry globale.

## Fase 4 — Resolver e dependency graph

- query e filtri;
- preferenze esplicite;
- errori not found, unavailable e ambiguous;
- vista read-only del grafo e piano topologico.

Completata. Il Resolver concreto usa esclusivamente snapshot current e il grafo
Runtime autorevole, senza ranking implicito né esecuzione del candidato.

## Fase 5 — Composition root e osservabilità

- esposizione del servizio Gestor dal Runtime;
- invalidazione coordinata;
- eventi redatti;
- test end-to-end e documentazione d'uso.

Completata. Il bootstrap, il refresh iniziale, l'invalidazione coordinata, gli
eventi stabili e i test pubblici chiudono la Milestone 4.

Ogni fase produce un report dedicato in `docs/reports/`; la Fase 5 produce
anche il report conclusivo della milestone.

---

# Gate di accettazione

La Milestone 4 ha superato:

- test unitari per invarianti e algoritmi;
- test di concorrenza e race detector;
- refresh fallito senza pubblicazione parziale;
- listing deterministico;
- nessun codice esterno eseguito sotto lock;
- distinzione verificata tra dichiarato, disponibile e non disponibile;
- risoluzione esplicita di zero, uno e più candidati;
- integrazione senza duplicazione del dependency graph;
- provider source testata con fixture positive e negative;
- plugin scoperto una sola volta attraverso il Registry dei componenti;
- documentazione e report finale per ogni fase.

---

# Fuori scope iniziale

- unregister e hot reload dinamico;
- discovery di rete o marketplace remoto;
- persistenza degli snapshot;
- ranking prestazionale o selezione automatica del modello migliore;
- probe impliciti durante ogni risoluzione;
- esecuzione delle capability;
- policy di fallback non dichiarate;
- sostituzione dei Runtime specializzati esistenti.

---

# Decisioni chiuse nella Fase 1

1. Il modello di dominio e i tre confini di consumo sono pubblici in
   `pkg/gestor`; mapping e implementazioni restano interni.
2. Le preferenze sono target esatti e ordinati in `QueryOptions`. Non esiste
   ancora configurazione globale; una configurazione futura dovrà tradursi
   nello stesso valore esplicito.
3. `Refresh` e `Invalidate` sono esposti dal `Registry`; il composition root
   pubblica lo snapshot iniziale e lascia i refresh successivi espliciti.
4. La vista del dependency graph è soltanto interna e read-only. Il confine
   concreto è implementato senza esporre `internal/runtime.graph` e restituisce
   soltanto generazione, eleggibilità e piano topologico.
5. I topic stabili sono esposti in `pkg/gestor`; cause, dettagli di
   introspection e diagnostica sensibile restano interni e redatti.

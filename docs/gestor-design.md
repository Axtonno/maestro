# Gestor — Design iniziale

Versione: 0.1.0

Stato: Draft iniziale — Milestone 4 aperta

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

Ogni descriptor conserva inoltre la sorgente e la generazione dello snapshot.
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
    Generation  uint64
}
```

Il frammento è illustrativo. Nomi e package pubblici saranno stabilizzati nel
primo incremento e registrati in un ADR prima dell'esposizione.

---

# Sorgenti di discovery

## Runtime component source

Legge uno snapshot del Registry dei componenti e traduce
`runtime.Metadata.Capabilities`. Non conserva le istanze e non esegue metodi dei
componenti.

Le capability lifecycle dichiarate devono continuare a essere validate contro
le interfacce opzionali dal Runtime Core; Gestor non duplica questa validazione.

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

## Plugin

Non esiste una sorgente plugin separata nel primo incremento. Un plugin
registrato correttamente è già presente nel Registry dei componenti e viene
scoperto da quella sorgente. Il Plugin Runtime continua a possedere catalogo e
loader.

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

---

# Integrazione con il dependency graph

Il grafo esistente resta la fonte unica per dipendenze richieste, dipendenze
opzionali, cicli e ordine topologico.

Gestor riceve una vista read-only o un collaboratore interno dedicato. Non
espone `internal/runtime.graph`, non aggiunge archi e non conserva una copia
mutabile del grafo.

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
- la cancellazione interrompe il refresh e non pubblica risultati parziali;
- i test con race detector sono parte del gate della milestone.

---

# Errori e diagnostica

Il contratto distinguerà almeno:

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

Eventi candidati:

```text
gestor.refresh.started
gestor.refresh.completed
gestor.refresh.failed
gestor.resolution.completed
gestor.resolution.failed
```

Gli eventi vengono pubblicati dopo il rilascio dei lock e non contengono
contenuti operativi sensibili.

---

# Contratti iniziali proposti

La forma concettuale è:

```go
type Source interface {
    ID() SourceID
    Discover(context.Context) ([]Descriptor, error)
}

type Registry interface {
    RegisterSource(Source) error
    Refresh(context.Context) error
    Snapshot() Snapshot
}

type Resolver interface {
    Candidates(Query) ([]Descriptor, error)
    Resolve(Query) (Resolution, error)
}
```

Le interfacce pubbliche definitive verranno introdotte solo dove esistono
almeno due implementazioni o un confine di consumo reale. Implementazione e
contratti resteranno separati in `internal/gestor` e nell'eventuale package
pubblico minimale.

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

## Fase 5 — Composition root e osservabilità

- esposizione del servizio Gestor dal Runtime;
- invalidazione coordinata;
- eventi redatti;
- test end-to-end e documentazione d'uso.

Ogni fase produce un report dedicato in `docs/reports/`; la Fase 5 produce
anche il report conclusivo della milestone.

---

# Gate di accettazione

La Milestone 4 richiederà almeno:

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

# Domande da chiudere nella Fase 1

1. Quale parte dei contratti deve essere pubblica in `pkg/gestor`?
2. Come viene rappresentata in configurazione una lista di preferenze per
   capability?
3. La prima versione espone `Refresh` direttamente o soltanto attraverso il
   Runtime composition root?
4. Qual è il confine minimo della vista read-only del dependency graph?
5. Quali eventi devono essere pubblici e quali restare diagnostica interna?

Queste domande non cambiano le decisioni di proprietà: Gestor indicizza e
risolve, il Runtime possiede grafo e lifecycle, i Runtime specializzati
possiedono registrazione e invocazione dei rispettivi soggetti.

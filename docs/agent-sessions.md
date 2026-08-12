# Agent Sessions, Planning and Budgets

Versione: 0.1.0

Stato: Implementato

Data: 2026-08-12

---

# Ownership

`internal/agent.Runtime` possiede il catalogo degli agenti e il registry
in-memory delle sessioni. Ogni `RunID` identifica un solo coordinatore e non
viene riutilizzato dopo il terminale. Il registry ha cardinalità configurabile,
non applica eviction silenziosa e rende esplicito il limite della baseline non
persistente.

Il Context Engine resta proprietario degli snapshot del workspace. L'Agent
Runtime conserva nel proprio snapshot soltanto l'ID del workspace, la
generazione del contesto usata e il bit di freshness.

# Stato della sessione

Il percorso ordinario è:

```text
created -> planning -> running -> terminal
```

Ogni aggiornamento crea un `SessionSnapshot` completo con generazione
monotona. La sessione è protetta indipendentemente dal registry: planner e
dipendenze esterne non vengono invocati sotto il lock globale.

Il terminale è committato una volta sola. Se più cause sono disponibili prima
del commit, `SelectTerminal` applica la precedenza pubblica; dopo il commit ogni
nuova transizione restituisce `ErrSessionTerminal`.

# Piani

Un agente che implementa `PlanningAgent` produce un `Plan` immutabile. Il
runtime accetta come piano iniziale soltanto la versione 1, entro
`MaxPlanSteps`, con step pending e grafo aciclico. `Plan.TransitionStep` crea
una nuova copia, verifica la state machine e permette di avviare uno step solo
quando tutte le dipendenze sono completed o skipped.

Le revisioni devono incrementare la versione di una unità, rispettare il
limite degli step e consumare `MaxPlanRevisions`. La sessione mantiene una
storia bounded delle versioni accettate; le transizioni interne allo stesso
piano aggiornano la versione corrente senza inventare una revisione.

# Planner provider-backed

`ProviderPlanner` usa il Provider Runtime tramite il contratto di completion,
con provider e modello esatti. Richiede JSON strutturato, rifiuta campi o JSON
residuo, converte ogni elemento in un `PlanStep` pubblico e lascia la
validazione finale al runtime.

Prima di inviare istruzione ed evidenza al provider, il runtime costruisce un
`DisclosureManifest` redatto e autorizza atomicamente `model.invoke` e
`model.disclose` tramite il Tool Runtime. Un deny impedisce la chiamata al
provider. La disclosure contiene fingerprint, generazione e contatori, non il
testo del workspace.

# Budget

I limiti della request sono hard ceiling locali e non fanno parte del prompt:

- durata massima;
- turni modello e chiamate tool totali/per turno;
- step e revisioni del piano;
- token input/output;
- byte per risultato tool e byte complessivi di sessione.

I contatori vengono aggiornati sotto il lock della sessione e pubblicati nello
snapshot. Un incremento che oltrepassa il limite viene rifiutato e conduce al
terminale `limit`. La deadline deriva dal context del chiamante e resta valida
durante Context Build e planning.

# Confini della fase

La Fase 4 termina intenzionalmente un run validamente pianificato con
`blocked`: il loop provider-tool viene collegato nella Fase 5. Sessioni e
storia dei piani sono in-memory e bounded; persistenza, recovery e multi-agent
restano fuori scope.

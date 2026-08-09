# Milestone 4 — Report Fase 2

Fase: Snapshot Registry

Stato: Completata

Data: 2026-08-09

---

# Obiettivo

Implementare il catalogo delle sorgenti e il Registry immutabile delle
capability con refresh all-or-nothing, generazioni monotone, invalidazione e
indici deterministici, senza introdurre ancora discovery Runtime/Provider o
risoluzione.

---

# File e contratti introdotti

`internal/gestor/registry.go` introduce:

- `Registry`, implementazione interna di `pkg/gestor.Registry`;
- catalogo sorgenti thread-safe indicizzato per `SourceID`;
- snapshot interno immutabile con indici per capability e target esatti;
- refresh cancellabile con candidato locale e swap atomico;
- epoch per invalidazione e variazioni del catalogo;
- generazioni monotone e conservazione dell'ultimo snapshot valido.

`pkg/gestor/snapshot.go` aggiunge in modo compatibile
`NewSnapshotWithSources`. Il costruttore registra tutte le sorgenti consultate,
incluse quelle che restituiscono zero descriptor. `NewSnapshot` conserva il
comportamento della Fase 1 e deriva le sorgenti dai descriptor.

`internal/gestor/registry_test.go` contiene la fixture `memorySource`,
configurabile in-memory e sicura per i test concorrenti.

Design, roadmap, development plan e `MAESTRO_CONTEXT.md` sono stati aggiornati
allo stato della Fase 2.

---

# Invarianti implementati

- Typed nil, source ID invalidi e source ID duplicati vengono rifiutati prima
  della modifica del catalogo.
- Una nuova sorgente invalida lo snapshot corrente senza incrementarne la
  generazione.
- Le sorgenti vengono fotografate e ordinate sotto read lock; `Discover` viene
  sempre eseguito dopo il rilascio del lock.
- Descriptor e risultati temporanei appartengono alla singola chiamata di
  refresh.
- Ogni descriptor deve essere valido e dichiarare esattamente il source ID
  della sorgente che lo ha prodotto.
- Duplicati capability–target dentro una sorgente e collisioni tra sorgenti
  autorevoli impediscono la pubblicazione.
- Un candidato diventa visibile soltanto dopo il successo di tutte le sorgenti
  e della validazione completa.
- Errori, cancellazione, collisioni o context già terminato non modificano
  snapshot o generazione.
- La causa originale degli errori e i sentinel Gestor restano compatibili con
  `errors.Is`.
- Lo snapshot iniziale è stale a generazione zero; ogni refresh riuscito
  incrementa esattamente di uno, anche con zero sorgenti.
- `Invalidate` conserva generazione e descriptor, marca stale e incrementa
  l'epoch interna.
- Un refresh costruito su un'epoch superata viene scartato con
  `ErrStaleSnapshot`.
- Snapshot, metadata, listing e risultati degli indici non espongono backing
  slice o mappe interne.
- Ordine delle sorgenti, descriptor e listing degli indici è stabile e
  lessicografico.

---

# Indici

Ogni snapshot pubblicato costruisce due indici interni immutabili:

- `CapabilityID` → descriptor ordinati;
- `Target` esatto → descriptor ordinati.

Gli indici non sono parte dell'API pubblica e non duplicano registry di
componenti o provider. Saranno consumati dal Resolver della Fase 4. Le letture
interne restituiscono copie difensive.

---

# Concorrenza e atomicità

Il Registry usa un unico `RWMutex` soltanto per catalogo, epoch e puntatore allo
snapshot immutabile. Discovery e context del caller non vengono eseguiti sotto
lock.

Un refresh:

1. copia sorgenti ed epoch;
2. esegue discovery in ordine fuori lock;
3. valida il candidato locale;
4. riacquisisce il lock;
5. verifica context ed epoch;
6. assegna la generazione successiva e sostituisce lo snapshot in un unico
   passaggio.

Refresh concorrenti sulla stessa epoch mantengono stato indipendente e
pubblicano generazioni monotone nell'ordine di completamento. Registrazione o
invalidazione durante discovery impediscono invece la pubblicazione del
candidato superato.

---

# Test eseguiti

Comandi del gate:

```text
GOCACHE=/tmp/maestro-go-build go test ./...
GOCACHE=/tmp/maestro-go-build go test -race ./...
GOCACHE=/tmp/maestro-go-build go vet ./...
```

Esito: tutti i comandi superati.

La suite copre:

- zero, una e più sorgenti;
- sorgenti senza descriptor registrate nei metadata;
- ordine di esecuzione deterministico;
- nil, typed nil, ID invalidi e ID duplicati;
- descriptor invalidi, source incoerente e collisioni dentro o tra sorgenti;
- errore e cancellazione dopo una sorgente già eseguita;
- conservazione esatta dell'ultimo snapshot valido e delle cause di errore;
- generazioni, invalidazione e recovery;
- indici ordinati e copie difensive;
- assenza di risultati parziali durante una sorgente bloccata;
- registrazione concorrente mentre `Discover` è bloccata, a prova dell'assenza
  di chiamate esterne sotto lock;
- refresh e letture concorrenti con verifica della generazione finale;
- race detector sui package Gestor e sull'intero repository.

---

# Decisioni architetturali

- Il concrete Registry resta in `internal/gestor`; il contratto pubblico non è
  stato ampliato.
- Gli indici appartengono allo snapshot interno e non diventano una seconda
  fonte autorevole.
- L'epoch protegge la pubblicazione da cataloghi o invalidazioni superati senza
  serializzare le chiamate esterne.
- Le sorgenti consultate sono metadata dello snapshot anche quando non
  producono descriptor.
- La registrazione sorgente invalida, ma non avvia automaticamente refresh o
  I/O.

---

# Limitazioni e attività rinviate

- Le sorgenti Runtime component metadata e Provider introspection appartengono
  alla Fase 3.
- L'invalidazione dopo registrazioni di componenti e provider sarà collegata
  nella Fase 3 e coordinata dal composition root nella Fase 5.
- Gli indici non implementano ancora filtri o selezione; Resolver, ambiguity e
  dependency graph appartengono alla Fase 4.
- Eventi e observer appartengono alla Fase 5.
- Non sono introdotti unregister, persistenza, refresh automatico o discovery
  remota.

---

# Gate di uscita

Superato:

- test unitari e concorrenti verdi;
- race detector verde;
- sorgenti bloccanti verificate fuori lock;
- refresh atomico e conservazione dell'ultimo snapshot dimostrati;
- generazioni e invalidazione coerenti con ADR-0022;
- nessuna sorgente Runtime/Provider o logica Resolver anticipata.

La Fase 2 è completata. La Fase 3 — Discovery sources è pronta.

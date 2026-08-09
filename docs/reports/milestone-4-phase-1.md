# Milestone 4 — Report Fase 1

Fase: Contratti, modello di dominio e ADR

Stato: Completata

Data: 2026-08-09

---

# Obiettivo

Stabilizzare il linguaggio e il confine pubblico minimo di Gestor prima di
introdurre stato, discovery integrata, Registry o Resolver concreti.

---

# File e contratti introdotti

Il package pubblico `pkg/gestor` contiene:

- `CapabilityID` e `SourceID` namespaced, validabili e ordinabili;
- capability namespaced Runtime e Provider note;
- `TargetKind`, `Scope`, `Target` e `Availability`;
- `Descriptor`, con dichiarazione e availability su assi distinti;
- `QueryOptions` e `Query` immutabile con preferenze ordinate esplicite;
- `Snapshot`, `SnapshotMetadata` e copie difensive;
- `Resolution`, motivazione e dependency plan immutabile;
- sentinel di validazione, discovery e risoluzione;
- interfacce pubbliche minime `Source`, `Registry` e `Resolver`.

Il package `internal/gestor` contiene soltanto gli adapter di mapping da tutte
le capability oggi note in `pkg/runtime` e `pkg/provider`. Non contiene
implementazioni concrete di Registry o Resolver.

Sono stati inoltre introdotti:

- `docs/adr/ADR-0022.md`, in stato Accepted;
- aggiornamento di `docs/gestor-design.md` alla versione `0.2.0`;
- aggiornamento di roadmap, development plan, indice ADR e context.

---

# Invarianti implementati

- Capability e source ID richiedono namespace e nome normalizzati; nessun
  valore viene corretto implicitamente.
- Target e model ID vuoti o con whitespace ai bordi vengono rifiutati.
- I componenti usano soltanto scope `component`; i provider usano `adapter`,
  `instance` o `model`.
- Lo scope `model` richiede sempre un model ID esatto e gli altri scope lo
  vietano.
- La presenza di un descriptor rappresenta una dichiarazione; `unknown`,
  `available` e `unavailable` restano evidenza separata.
- ID e value object hanno uguaglianza esatta; i metodi `Compare` e gli snapshot
  applicano ordinamento lessicografico deterministico.
- Preferenze, descriptor, sorgenti metadata e dependency plan vengono copiati
  sia in ingresso sia in uscita.
- Una coppia capability–target duplicata non può formare uno snapshot valido.
- La generazione appartiene una sola volta a `SnapshotMetadata` e non viene
  duplicata nei descriptor.
- Una resolution richiede snapshot current, reason valido e dependency plan di
  soli componenti senza duplicati.
- Il package pubblico non dipende da implementazioni Runtime, Provider o
  Plugin concrete.

---

# Decisioni architetturali

ADR-0022 stabilisce che:

- dominio e confini di consumo sono pubblici in `pkg/gestor`;
- mapping e implementazioni restano in `internal/gestor`;
- le preferenze sono target esatti ordinati nella query e non ranking globali;
- `Refresh` e `Invalidate` sono operazioni pubbliche esplicite del Registry;
- la vista del dependency graph rimane un collaboratore interno read-only;
- gli eventi pubblici verranno aggiunti soltanto con il wiring della Fase 5;
- sentinel e cause devono restare compatibili con `errors.Is`.

Il design illustrativo iniziale è stato corretto: la generazione è metadata
dello snapshot, non un campo ripetuto su ogni descriptor.

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

- valori validi e invalidi di ID, target, availability, descriptor e query;
- model ID esatti e compatibilità kind–scope;
- copie difensive di capability note, preferenze, snapshot, metadata e
  dependency plan;
- ordinamento deterministico;
- duplicati e snapshot invalidi;
- resolution valida e stale;
- compatibilità `errors.Is` dei sentinel;
- mapping completo delle sei capability Runtime e delle undici Provider note;
- compilazione delle interfacce pubbliche tramite implementazioni fixture.

---

# Limitazioni e attività rinviate

- Il catalogo sorgenti, gli indici e il refresh atomico appartengono alla
  Fase 2.
- Le sorgenti Runtime e Provider e l'invalidation wiring appartengono alla
  Fase 3.
- Filtri effettivi, ambiguity handling e integrazione con il dependency graph
  appartengono alla Fase 4.
- Composition root, policy di refresh iniziale ed eventi appartengono alla
  Fase 5.
- Non esiste configurazione globale delle preferenze; configurazioni future
  dovranno essere tradotte nel valore esplicito `QueryOptions`.

---

# Gate di uscita

Superato:

- contratti pubblici e test presenti;
- ADR-0022 Accepted;
- `go test`, race detector e `go vet` verdi;
- mapping Runtime e Provider completo;
- documentazione coerente con il contratto approvato;
- nessuna implementazione prematura di Registry o Resolver.

La Fase 1 è completata. La Fase 2 — Snapshot Registry è pronta.

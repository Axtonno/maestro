# Milestone 4 — Gestor Final Report

Stato: Completata

Data: 2026-08-09

Fasi completate: 5/5

---

# Risultato

Maestro dispone ora di un servizio Gestor in-process pubblico, immutabile e
capability-based. Gestor scopre dichiarazioni dai cataloghi autorevoli,
conserva availability separata, pubblica snapshot atomici, risolve soltanto con
filtri e preferenze esplicite e usa il dependency graph del Runtime Core in
sola lettura.

Il servizio è disponibile tramite:

```go
runtime.Gestor()
```

e implementa `pkg/gestor.Service`, composizione dei contratti `Registry` e
`Resolver`.

---

# Deliverable per fase

| Fase | Deliverable | Stato |
|---|---|---|
| 1 | Contratti pubblici, value object, sentinel e ADR-0022 | Completata |
| 2 | Snapshot Registry atomico, indici, generazioni e invalidazione | Completata |
| 3 | Sorgenti Runtime/Provider e plugin senza duplicazione | Completata |
| 4 | Resolver e vista read-only del dependency graph | Completata |
| 5 | Composition root, eventi redatti e gate end-to-end | Completata |

I report dettagliati sono:

- `milestone-4-phase-1.md`;
- `milestone-4-phase-2.md`;
- `milestone-4-phase-3.md`;
- `milestone-4-phase-4.md`;
- `milestone-4-phase-5.md`.

---

# Contratti e ownership finali

Gestor possiede:

- catalogo delle sorgenti;
- snapshot e indici capability/target;
- generazione e currentness dello snapshot;
- filtri, preferenze e spiegazione della risoluzione;
- facade osservabile del composition root.

Gestor non possiede:

- componenti o provider;
- lifecycle, stato o routing operativo;
- dependency graph mutabile;
- introspection al di fuori di `Refresh`;
- esecuzione della capability risolta.

Il Runtime Registry resta autorevole per i componenti, il Provider Runtime per
i provider, il Plugin Runtime per classificazione e loader e il Runtime Core per
grafo e lifecycle.

---

# Semantica consegnata

- Snapshot iniziale current a generazione 1.
- Refresh all-or-nothing e cancellabile.
- Invalidazione senza I/O dopo registrazioni riuscite.
- Availability `unknown`, `available` e `unavailable` non promossa.
- Filtri exact per capability, target kind, scope e model.
- Preferenze target exact e ordinate.
- `ErrAmbiguous` quando nessuna preferenza risolve candidati multipli.
- Distinzione tra not found, unavailable e stale.
- Dependency plan transitivo e dependency-first dal grafo autorevole.
- Generazioni Registry/grafo indipendenti e ricontrollate.
- Eventi success/failure redatti e fuori lock.
- Nessuna selezione lessicografica, esecuzione o probe implicito.

---

# Gate tecnico finale

Comandi eseguiti:

```text
GOCACHE=/tmp/maestro-go-build go test ./...
GOCACHE=/tmp/maestro-go-build go test -race ./...
GOCACHE=/tmp/maestro-go-build go vet ./...
git diff --check
```

Esito: tutti superati.

Sono verificati test unitari, concorrenti, di integrazione interna e end-to-end
dal composition root pubblico. Le interfacce pubbliche Runtime, Provider e
Plugin preesistenti restano compatibili; Gestor è un'estensione del Runtime
Maestro.

---

# Limiti dichiarati

Restano fuori dalla milestone:

- persistenza e replica degli snapshot;
- unregister e hot reload;
- discovery remota o marketplace;
- ranking prestazionale e fallback impliciti;
- model target inferiti automaticamente;
- esecuzione delle capability;
- osservabilità asincrona integrata.

Questi punti sono evoluzioni future e non costituiscono task bloccanti per il
servizio Gestor consegnato.

---

# Relazione con la Milestone 3

La chiusura di Gestor non modifica lo stato della Milestone 3. La matrice live
llama.cpp resta il task pendente già documentato e dovrà essere completata prima
della chiusura formale della Milestone 3 o di una release pubblica importante.

---

# Chiusura

La Milestone 4 — Gestor è completata. Il sistema dispone del registro centrale
delle capability necessario alle milestone successive senza duplicare le fonti
autorevoli o introdurre comportamento operativo implicito.

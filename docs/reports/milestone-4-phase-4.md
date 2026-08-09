# Milestone 4 — Report Fase 4

Fase: Resolver e dependency graph

Stato: Completata

Data: 2026-08-09

---

# Obiettivo

Risolvere capability su snapshot Gestor immutabili attraverso filtri e
preferenze esplicite, verificando l'eleggibilità dei componenti e costruendo il
piano delle dipendenze dal grafo autorevole del Runtime Core senza duplicarlo o
eseguire codice del candidato.

---

# File e contratti introdotti

`internal/gestor/resolver.go` introduce il concrete `Resolver`, conforme a
`pkg/gestor.Resolver`, e il collaboratore interno minimale per consultare stato
e piani del dependency graph.

`internal/gestor/registry.go` aggiunge letture interne atomiche di snapshot e
indice capability e il controllo di currentness per una generazione attesa.

`internal/runtime/gestor_graph_view.go` introduce la vista read-only del grafo:

- stato e generazione del grafo;
- eleggibilità di un component ID;
- dipendenze transitive in ordine topologico dependency-first.

`internal/runtime/runtime.go` mantiene generazioni indipendenti del catalogo
componenti e del grafo costruito. `internal/runtime/node.go` conserva l'identità
del nodo acquisita durante la costruzione, evitando chiamate a `Metadata`
durante la risoluzione.

I test sono in:

- `internal/gestor/resolver_test.go`;
- `internal/runtime/gestor_graph_view_test.go`;
- estensioni dei test di nodo Runtime.

Non sono stati modificati i contratti pubblici: query, resolution, sentinel e
interfaccia `Resolver` approvati nella Fase 1 erano già sufficienti.

---

# Algoritmo di risoluzione

Il Resolver:

1. valida la query;
2. acquisisce atomicamente snapshot e candidati della capability;
3. applica filtri esatti per target kind, scope e model ID;
4. esclude sempre `unavailable` ed esclude `unknown` quando è richiesta evidenza
   `available`;
5. per i componenti consulta il grafo current e costruisce il piano transitivo;
6. ordina i candidati soltanto per output e diagnostica deterministici;
7. seleziona il candidato unico o la prima preferenza esatta eleggibile;
8. segnala ambiguity quando più candidati restano senza preferenza risolutiva;
9. ricontrolla le generazioni consultate prima di produrre il risultato.

`Candidates` restituisce tutti i candidati eleggibili e non applica le
preferenze. `Resolve` non introduce default, ranking o fallback non dichiarati.

---

# Invarianti implementati

- Nessun ordine lessicografico seleziona implicitamente un vincitore.
- Snapshot stale o cambiati durante la lettura producono `ErrStaleSnapshot`.
- Il grafo ha una generazione indipendente e deve essere current per i candidati
  component.
- Provider-only resolution non dipende dallo stato del grafo Runtime.
- Gestor non aggiunge nodi, archi o copie mutabili del grafo.
- Una dipendenza richiesta mancante impedisce la costruzione del grafo e quindi
  la risoluzione del componente.
- Una dipendenza opzionale mancante viene omessa senza rendere il componente
  ineleggibile.
- Il piano contiene tutte le dipendenze transitive in ordine dependency-first.
- Cicli e validazione delle dipendenze restano responsabilità del Runtime Core.
- `Resolve` non invoca `Metadata`, lifecycle, probe, introspection o capability
  del candidato.
- Slice di candidati e dependency plan restano copie difensive.

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

- query invalide e snapshot iniziale stale;
- zero, uno e più candidati;
- filtri esatti per target kind, scope e model;
- differenza tra not found e unavailable;
- availability unknown con e senza requisito operativo;
- preferenza valida, mancante e non disponibile;
- ambiguity stabile senza selezione lessicografica;
- copie difensive restituite da `Candidates`;
- componente assente dal grafo e grafo stale;
- dipendenze richieste, opzionali e transitive;
- ordine topologico dal grafo Runtime reale;
- ricostruzione e invalidazione del grafo con generazioni monotone;
- mismatch o variazione concorrente di snapshot e graph generation;
- provider-only resolution con grafo stale;
- risoluzioni concorrenti e race detector;
- integrazione end-to-end sorgente componenti, Registry, Resolver e grafo reale;
- assenza di chiamate a `Metadata` del candidato durante `Resolve`.

---

# Decisioni architetturali

- Registry Gestor e grafo Runtime conservano generazioni indipendenti; il
  Resolver cattura e verifica entrambe invece di accoppiarle in un contatore
  globale.
- La currentness del grafo deriva anche dalla generazione del catalogo
  componenti usata per costruirlo.
- La vista del grafo resta interna al Runtime e restituisce soltanto value
  object pubblici del Runtime, senza esporre `graph` o `node`.
- L'identità del nodo viene fissata alla costruzione del grafo per mantenere la
  risoluzione una pura lettura delle strutture già validate.
- `ErrNotFound` descrive l'assenza di dichiarazioni compatibili;
  `ErrUnavailable` descrive dichiarazioni non operativamente eleggibili;
  `ErrAmbiguous` richiede una preferenza esplicita.
- L'ordinamento dei descriptor serve esclusivamente a determinismo e
  diagnostica.

---

# Limitazioni e attività rinviate

- Il composition root non costruisce ancora automaticamente Registry, sorgenti
  e Resolver; appartiene alla Fase 5.
- Refresh iniziale, invalidazione coordinata completa ed eventi redatti
  appartengono alla Fase 5.
- Non sono introdotti ranking prestazionali, fallback globali, probe impliciti,
  persistenza o discovery remota.
- La risoluzione restituisce identità e piano, non un handle eseguibile.

---

# Gate di uscita

Superato:

- tutti i rami di selezione ed errore sono verificati;
- nessuna euristica implicita sceglie tra candidati multipli;
- il grafo Runtime resta unico e viene consultato soltanto in lettura;
- dependency plan e diagnostica sono deterministici;
- variazioni concorrenti vengono rilevate tramite generazioni;
- suite, race detector e vet repository-wide verdi.

La Fase 4 è completata. La Fase 5 — Composition root, osservabilità e gate
finale è pronta.

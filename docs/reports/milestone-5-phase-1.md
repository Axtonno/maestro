# Milestone 5 — Report Fase 1

Fase: Contratti, audit della baseline e ADR-0023

Stato: Completata

Data: 2026-08-11

---

# Obiettivo

Stabilire il confine pubblico e architetturale del Plugin System prima
dell'hardening di catalogo, concorrenza, lifecycle integrato e reference plugin
Laravel.

La fase parte dalla baseline Plugin Runtime esistente, ma ne verifica
esplicitamente contratti, ownership, compatibilità, errori ed eventi.

---

# Esito dell'audit

La superficie pubblica esistente è sufficiente e non richiede modifiche
breaking o nuove astrazioni nella Fase 1.

Sono confermati:

- `Plugin` come estensione minimale di `runtime.Component`;
- `runtime.Metadata` come fonte unica per identità, versione, dipendenze e
  capability;
- `Manifest` limitato alla versione richiesta dell'API Plugin Runtime;
- `Loader` come factory pull-based cancellabile;
- `plugin.Runtime` come confine di catalogo, registrazione e load;
- Runtime Core come proprietario esclusivo di stato, graph e lifecycle;
- Gestor come indice delle capability plugin attraverso il Registry globale;
- caricamento e registrazione prima dello startup;
- eventi pubblici distinti per loader registered, plugin registered e loaded.

Non è stato introdotto un descriptor di catalogo: duplicherebbe metadata non
ancora disponibili senza istanziare il plugin. Non è stato introdotto un
contratto workspace framework-neutral: i requisiti del Context Engine non sono
ancora sufficienti a definirne i campi senza speculazione.

---

# Semantica stabilizzata

- available: esiste un loader nel catalogo;
- registered: esiste un'istanza nel registry plugin e nel Runtime Core;
- loaded: factory e registrazione sono riuscite; è un evento, non uno stato;
- running: il lifecycle globale è completato secondo lo `StateManager`.

`Load` non inizializza né avvia il plugin. Il catalogo non seleziona o carica
implicitamente estensioni e l'ordine dei listing non costituisce ranking.

---

# File e contratti

Introdotti:

- `docs/adr/ADR-0023.md`;
- `docs/plugin-api-compatibility-audit.md`;
- questo report di fase.

Aggiornati:

- indice ADR;
- documentazione pubblica dei sentinel in `pkg/plugin/errors.go`;
- assertion di compilazione per `Plugin`, `Loader` e `runtime.Event`;
- test della versione API esatta del manifest;
- design, piano, roadmap, README e contesto di progetto.

Nessuna firma, interfaccia o struct pubblica è stata modificata.

---

# ADR-0023

ADR-0023 è Accepted e stabilisce:

- modello trusted in-process;
- registrazione pre-start;
- catalogo, registry e stato come viste distinte;
- nessuno stato persistente loaded;
- Runtime Core proprietario di graph, stato e lifecycle;
- Gestor integrato senza source plugin dedicata;
- nessun descriptor o contratto workspace prematuro;
- packaging, sandbox, hot loading e terze parti fuori scope.

ADR-0007 e ADR-0008 restano valide e vengono specializzate, non sostituite.

---

# Compatibilità ed errori

La matrice completa è in `docs/plugin-api-compatibility-audit.md`.

Tutti i sentinel pubblici hanno un confine documentato e restano ispezionabili
con `errors.Is`. Le cause di loader e Runtime Core sono preservate. Una failure
prima della registrazione non rende il plugin visibile nel registry dedicato.

La versione API resta `plugin.RuntimeAPIVersion == "1"` e la compatibilità è
esatta: versione vuota produce `ErrInvalidManifest`; ogni valore differente,
incluso whitespace aggiuntivo, produce `ErrIncompatible`.

---

# Test eseguiti

Comandi del gate:

```text
GOCACHE=/tmp/maestro-m5p1-test go test ./...
GOCACHE=/tmp/maestro-m5p1-race go test -race ./...
GOCACHE=/tmp/maestro-m5p1-vet go vet ./...
git diff --check
```

Esito: tutti i comandi superati.

La suite di Fase 1 copre:

- implementazione delle interfacce pubbliche;
- versione API stabile e non vuota;
- plugin e loader nil o typed nil;
- ID validi e invalidi;
- manifest mancante e mismatch esatto;
- tutti i sentinel pubblici tramite `errors.Is`;
- cause dei loader preservate;
- snapshot difensivi di available e registered;
- registrazione nel Runtime Core;
- collisione con componenti normali;
- cancellazione prima e dopo la factory;
- eventi dal composition root;
- load e lifecycle del plugin Laravel.

---

# Gap assegnati

## Fase 2

- callback fuori lock dimostrate con fixture bloccanti;
- load concorrente dello stesso ID e di ID differenti;
- atomicità e ordine osservabile sotto concorrenza;
- combinazioni complete dei sentinel dei risultati factory invalidi.

## Fase 3

- matrice dependency graph plugin/componente/plugin;
- registrazione durante startup e shutdown;
- stato, invalidazione e refresh Gestor end-to-end.

## Fase 4

- capability Laravel namespaced;
- snapshot framework-aware e test concorrenti;
- ampliamento delle fixture detection.

## Fasi e milestone successive

- cardinalità e redazione finale degli eventi: Fase 5;
- packaging, firme, sandbox e plugin terzi: Milestone 8;
- permission model operativo: Milestone 7.

Nessun gap di Fase 1 è rimasto senza destinazione.

---

# Gate di uscita

Superato:

- ADR-0023 Accepted;
- matrice API e gap documentati;
- semantica di available, registered, loaded e running non ambigua;
- test dei package plugin e suite repository-wide verdi;
- race detector e vet verdi;
- nessuna modifica breaking;
- hardening delle fasi successive non anticipato.

La Fase 1 è completata. La Fase 2 — Catalogo, registry e caricamento è pronta.

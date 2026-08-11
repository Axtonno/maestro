# Milestone 5 — Report Fase 4

Fase: Laravel reference plugin

Stato: Completata

Data: 2026-08-11

---

# Obiettivo

Consolidare Laravel come reference plugin framework-aware, dichiarando una
capability workspace individuabile da Gestor e stabilizzando la vista concreta
del workspace senza introdurre un contratto framework-neutral prematuro.

---

# Capability workspace

Il package pubblico `pkg/plugin` espone l'identificatore:

```text
plugin.workspace-detection
```

tramite `plugin.CapabilityWorkspaceDetection`.

L'identificatore è un `runtime.Capability` namespaced e non introduce
un'interfaccia di esecuzione generica. Il plugin Laravel lo dichiara nei propri
metadata insieme a initialize e health. Gestor può quindi indicizzarlo prima
del lifecycle e risolverlo sul target componente `laravel` dopo la costruzione
del dependency graph.

Il descriptor compare una sola volta attraverso `runtime.components`; non
esiste una source Gestor Laravel o plugin dedicata.

---

# Vista del workspace

La facade concreta esistente resta il confine di consumo:

- `Root()` restituisce la root assoluta e normalizzata, immutabile dopo la
  costruzione;
- `FrameworkVersion()` è vuota prima della prima inizializzazione riuscita;
- `Initialize` rileva il workspace e pubblica atomicamente il nuovo constraint;
- `Health` ripete la detection senza modificare lo snapshot pubblicato;
- una health fallita non cancella la versione inizializzata;
- una reinitialize fallita conserva l'ultimo snapshot valido;
- una reinitialize riuscita aggiorna atomicamente il constraint.

Root e versione sono sicure per letture concorrenti; initialize e health
concorrenti sono coperti dal race detector.

Non sono stati introdotti mappe del manifest, DTO mutabili o un contratto
workspace generico. Le future esigenze del Context Engine potranno produrre
un'estensione additiva basata su un consumer concreto.

---

# Detection e fixture

La matrice deterministica copre:

- root vuota e blank;
- root relativa normalizzata allo stesso percorso assoluto;
- root inesistente accettata in costruzione ma rifiutata dalla detection;
- `artisan` mancante o non regolare;
- `composer.json` mancante, JSON invalido o oltre 1 MiB;
- sezione `require` di tipo errato;
- dipendenza `laravel/framework` mancante, vuota o valida;
- workspace mutato dopo initialize;
- health e reinitialize riuscite o fallite;
- cento initialize e cento health/read concorrenti.

La costruzione resta priva di I/O di detection: il loader istanzia soltanto il
plugin, mentre `Initialize` appartiene al Runtime Core.

---

# Integrazione pubblica

Il test dalla facade `maestro.New` verifica:

```text
RegisterLoader -> Load -> Gestor.Refresh -> Start -> Gestor.Resolve -> Stop
```

Prima dello startup `FrameworkVersion` è vuota ma la capability è già presente
nello snapshot Gestor. Dopo startup la versione rilevata è disponibile e il
plugin è Running nello `StateManager` globale.

Il plugin Laravel passa dalla versione `0.1.0` alla `0.2.0` per rappresentare
l'aggiunta della capability pubblica. ID, costruttori, configurazione, errori e
metodi della facade restano invariati.

---

# Modifiche

- `plugin.CapabilityWorkspaceDetection` aggiunta in modo additivo;
- metadata Laravel aggiornati con la capability;
- versione Laravel aggiornata a `0.2.0`;
- regressioni di detection, snapshot e concorrenza;
- integrazione Gestor dalla facade pubblica;
- documentazione Laravel, piano, roadmap, README e contesto aggiornati.

---

# Verifica

Comandi del gate:

```text
GOCACHE=/tmp/maestro-m5p4-test go test ./...
GOCACHE=/tmp/maestro-m5p4-race go test -race ./...
GOCACHE=/tmp/maestro-m5p4-vet go vet ./...
git diff --check
```

Esito: tutti i comandi superati.

---

# Gate di uscita

Superato:

- reference plugin consumabile tramite facade pubblica;
- detection e health deterministiche e senza processi esterni;
- capability Laravel visibile e risolvibile tramite Gestor;
- snapshot framework version atomico e conservativo;
- concorrenza coperta dal race detector;
- nessuna logica Laravel introdotta nel core o in Gestor;
- nessuna API framework-neutral speculativa.

La Fase 4 è completata. La Fase 5 — Osservabilità, hardening e gate finale è
pronta.

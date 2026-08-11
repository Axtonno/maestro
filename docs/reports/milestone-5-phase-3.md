# Milestone 5 — Report Fase 3

Fase: Lifecycle, dependency graph e Gestor

Stato: Completata

Data: 2026-08-11

---

# Obiettivo

Dimostrare dal composition root pubblico che i plugin riusano dependency
graph, stato e lifecycle del Runtime Core e che Gestor li scopre una sola volta
attraverso il Registry globale, senza eseguire capability.

---

# Lifecycle e stato

La suite end-to-end copre plugin:

- senza capability lifecycle;
- con configure, initialize, start, stop e health dichiarate;
- in failure durante configure, initialize, start o stop;
- caricati dal catalogo prima dello startup;
- rifiutati durante startup e shutdown.

Anche un plugin passivo attraversa la state machine globale:

```text
Created -> Configured -> Initialized -> Running -> Stopping -> Stopped
```

Le capability assenti sono no-op del Lifecycle Manager; non nasce uno stato
plugin separato. Le failure vengono registrate come `StateFailed` nello stesso
`StateManager`, preservando la causa originale.

---

# Dependency graph

Una singola matrice mista verifica il percorso:

```text
core component -> framework plugin -> extension plugin -> worker component
```

Lo startup segue l'ordine dependency-first e lo shutdown l'ordine inverso. La
matrice copre quindi:

- plugin -> componente;
- plugin -> plugin;
- componente -> plugin.

Sono inoltre verificati:

- dipendenza richiesta mancante: `runtime.ErrNotFound` prima del lifecycle;
- dipendenza opzionale mancante: plugin avviato normalmente;
- ciclo componente <-> plugin: `runtime.ErrCyclicDependency`;
- stato ancora Created quando la costruzione del graph fallisce.

Non viene introdotto un graph plugin o un ordinamento parallelo.

---

# Registrazione pre-start

Fixture lifecycle bloccanti rendono osservabili gli intervalli di startup e
shutdown senza esporre internals.

- registrazione e `Load` durante startup restituiscono
  `runtime.ErrAlreadyStarted`;
- registrazione e `Load` durante shutdown restituiscono
  `runtime.ErrInvalidState`;
- nessun tentativo rifiutato viene indicizzato nel Plugin Runtime.

La regressione ha individuato un difetto nel Runtime Core: durante shutdown i
flag `started` e `stopping` erano entrambi veri, ma il controllo di `started`
precedeva quello di `stopping`, rendendo irraggiungibile il ramo specifico di
shutdown. L'ordine è stato corretto; il comportamento running/startup resta
invariato.

---

# Integrazione Gestor

Il test end-to-end usa una capability custom reale:

```text
plugin.workspace-detection
```

Il percorso verificato è:

```text
RegisterLoader -> Load -> invalidazione -> Start -> Refresh -> Resolve
```

Sono dimostrati:

- la sola registrazione del loader non invalida Gestor;
- il `Load` riuscito invalida lo snapshot perché registra un componente;
- `Resolve` prima del refresh restituisce `gestor.ErrStaleSnapshot`;
- dopo startup e refresh la capability ha il target plugin corretto;
- il descriptor compare una sola volta attraverso `runtime.components`;
- il dependency plan contiene la dipendenza componente in ordine topologico;
- `Resolve` non modifica le chiamate lifecycle e quindi non esegue plugin;
- stato Running e Stopped restano nello `StateManager` globale.

Non è stata aggiunta una source Gestor dedicata ai plugin.

---

# Modifiche

- estensione delle fixture end-to-end in `maestro_test.go`;
- regressioni lifecycle, graph, failure, finestre di stato e Gestor;
- correzione dell'ordine dei controlli in `internal/runtime.Register` durante
  shutdown;
- aggiornamento della documentazione Plugin Runtime e Runtime internals;
- allineamento di piano, roadmap, README e contesto.

Nessuna API pubblica è stata modificata.

---

# Verifica

Comandi del gate:

```text
GOCACHE=/tmp/maestro-m5p3-test go test ./...
GOCACHE=/tmp/maestro-m5p3-race go test -race ./...
GOCACHE=/tmp/maestro-m5p3-vet go vet ./...
git diff --check
```

Esito: tutti i comandi superati.

---

# Gate di uscita

Superato:

- un solo dependency graph dimostrato end-to-end;
- una sola state machine per componenti e plugin;
- Plugin Runtime privo di invocazioni lifecycle;
- registrazione limitata al pre-start;
- Gestor integrato senza duplicazioni o source dedicata;
- resolution con dependency plan senza esecuzione;
- suite completa, race detector e vet verdi.

La Fase 3 è completata. La Fase 4 — Laravel reference plugin è pronta.

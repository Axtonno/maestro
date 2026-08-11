# Plugin API Compatibility Audit — Milestone 5, Fase 1

Versione: 0.1.0

Stato: Superato

Data: 2026-08-11

---

# Perimetro

L'audit copre:

- contratti pubblici in `pkg/plugin`;
- facade Laravel in `pkg/plugin/laravel`;
- implementazioni `internal/plugin` e `internal/plugin/laravel`;
- integrazione con Runtime Core, Event Bus e Gestor nel composition root;
- ADR-0007, ADR-0008 e documentazione Plugin Runtime esistente.

L'audit valuta la compatibilità e assegna i gap alle fasi successive. Non
anticipa hardening concorrente, nuove capability Laravel o il gate finale.

---

# Esito

La superficie pubblica esistente è sufficiente per il gate della Milestone 5 e
non richiede modifiche breaking o nuove astrazioni nella Fase 1.

- `Plugin` resta una specializzazione minimale di `runtime.Component`.
- `runtime.Metadata` possiede identità, versione, dipendenze e capability.
- `Manifest` conserva soltanto la versione richiesta dell'API Plugin Runtime.
- `Loader` resta una factory pull-based cancellabile.
- `Runtime` separa catalogo, registrazione, risoluzione e caricamento.
- stato e lifecycle restano nel Runtime Core.
- Gestor scopre le capability plugin attraverso la source dei componenti.
- gli eventi pubblici descrivono soltanto operazioni riuscite.
- la facade Laravel conserva costruttori, configurazione, ID, errori e metodi.

ADR-0023 formalizza queste ownership e la natura pre-start del sistema.

---

# Matrice delle API pubbliche

| Simbolo | Semantica approvata | Compatibilità Fase 1 | Copertura |
|---|---|---|---|
| `RuntimeAPIVersion` | versione esatta del contratto Plugin Runtime | Invariata | manifest valido/incompatibile |
| `ID` | alias di `runtime.ComponentID` | Invariata | registrazione, lookup e loader |
| `Plugin` | componente con manifest plugin | Invariata | assertion e fixture concrete |
| `Manifest` | requisito API, non metadata duplicati | Invariata | empty e mismatch esatto |
| `Loader` | factory cancellabile senza lifecycle | Invariata | assertion, errori e context |
| `LoaderFunc` | adapter funzione -> loader | Invariata | invocation e typed nil |
| `Runtime.Register` | registra nel Runtime Core e poi indicizza | Invariata | validazione, collisioni e wrapping |
| `Runtime.Resolve` | risolve soltanto plugin registrati | Invariata | success, ID invalido e not found |
| `Runtime.Has` | presenza nel registry plugin | Invariata | success, failure e ID invalido |
| `Runtime.Registered` | snapshot degli ID registrati | Invariata | ordine e copia difensiva |
| `Runtime.RegisterLoader` | aggiunge un loader trusted al catalogo | Invariata | validazione, duplicati e typed nil |
| `Runtime.Load` | factory, validazione e registrazione; nessun start | Invariata | success, failure e cancellazione |
| `Runtime.Available` | snapshot degli ID loader disponibili | Invariata | ordine e copia difensiva |
| errori sentinel | classificazione ispezionabile con `errors.Is` | Invariata | tutti i sentinel pubblici |
| costanti evento | stadi riusciti di catalogo/register/load | Invariata | sequenza dal composition root |
| `EventPayload` / `Event` | ID e plugin quando già istanziato | Invariata | contratto `runtime.Event` |
| facade Laravel | reference plugin concreto | Invariata | load, lifecycle e validazione |

---

# Semantica degli stati

| Termine | Fonte autorevole | Significato |
|---|---|---|
| available | catalogo `plugin.Runtime` | esiste un loader registrato per l'ID |
| registered | registry `plugin.Runtime` e Runtime Core | esiste un'istanza registrata prima dello startup |
| loaded | evento dell'operazione `Load` | factory e registrazione sono terminate con successo |
| running | `runtime.StateManager` | il Runtime Core ha completato il lifecycle di startup |

`loaded` non è uno stato persistente aggiuntivo. Un plugin può essere registrato
direttamente senza un evento loaded e un loader può restare available dopo il
caricamento.

---

# Compatibilità degli errori

| Sentinel | Confine |
|---|---|
| `ErrInvalidPlugin` | plugin nil, ID invalido o risultato factory invalido |
| `ErrInvalidManifest` | versione API mancante |
| `ErrIncompatible` | versione API diversa da quella corrente |
| `ErrAlreadyRegistered` | collisione nel Registry globale dei componenti |
| `ErrNotFound` | plugin non presente nel registry dedicato |
| `ErrInvalidLoader` | ID/loader/context di load invalido |
| `ErrLoaderAlreadyRegistered` | collisione nel catalogo loader |
| `ErrLoaderNotFound` | nessun loader disponibile per l'ID |
| `ErrLoadFailed` | factory fallita o risultato factory nil/incoerente |

Le cause del loader e i sentinel del Runtime Core vengono preservati. Le
failure di manifest o registrazione incontrate da `Load` mantengono il proprio
sentinel specifico; non vengono trasformate in una tassonomia parallela.

---

# Gap rilevati e assegnazione

## Fase 2

- dimostrazione sistematica che loader ed eventi non vengono invocati sotto
  lock interni;
- semantica e test del load concorrente sullo stesso ID e su ID differenti;
- atomicità del registry dedicato rispetto a ogni failure del registrar;
- verifica completa dell'ordine osservabile in presenza di concorrenza;
- test combinati dei sentinel per risultati factory nil o con ID incoerente.

## Fase 3

- matrice completa delle dipendenze plugin/componente/plugin;
- rifiuto durante startup e shutdown oltre al caso running già coperto;
- verifica end-to-end di stato globale, invalidazione e refresh Gestor;
- capability custom plugin con dependency plan reale.

## Fase 4

- capability Laravel namespaced indicizzabile da Gestor;
- snapshot framework-aware e concorrenza della vista rilevata;
- fixture Laravel mancanti, inclusi constraint vuoto e mutation semantics.

L'audit non approva un descriptor di catalogo o un contratto workspace generico:
oggi non esiste un consumer che ne definisca i campi senza speculazione. Una
futura esigenza concreta potrà essere soddisfatta con un'estensione additiva.

---

# Modifiche della Fase 1

- nessuna firma o struct pubblica modificata;
- documentazione dei sentinel resa esplicita;
- assertion di compilazione aggiunte per `Plugin`, `Loader` e `runtime.Event`;
- test del mismatch esatto della versione manifest;
- ADR-0023 e matrice di compatibilità aggiunti.

---

# Rischi residui assegnati

- concorrenza e callback sotto lock: Fase 2;
- lifecycle, graph e Gestor end-to-end: Fase 3;
- capability reference Laravel: Fase 4;
- cardinalità/redazione eventi e audit repository-wide finale: Fase 5;
- packaging, firme, sandbox e terze parti: Milestone 8;
- permission model operativo: Milestone 7.

Non restano gap di Fase 1 senza una destinazione documentata.

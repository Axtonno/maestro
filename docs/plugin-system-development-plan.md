# Milestone 5 — Plugin System Development Plan

Versione: 0.1.0

Stato: In corso — Fase 1 completata; Fase 2 pronta

Data: 2026-08-11

Documento architetturale di riferimento: `plugin-system-design.md`.

---

# Obiettivo della milestone

Consolidare la baseline Plugin Runtime in un Plugin System trusted in-process
con contratti pubblici stabili, caricamento deterministico, integrazione con il
lifecycle e il dependency graph globali, discovery tramite Gestor e Laravel
come reference plugin framework-aware.

La milestone non introduce packaging esterno, sandbox, hot loading o un secondo
lifecycle.

---

# Sequenza delle fasi

| Fase | Titolo | Stato corrente | Dipende da |
|---|---|---|---|
| 1 | Contratti, audit della baseline e ADR-0023 | Completata | Design iniziale |
| 2 | Catalogo, registry e caricamento | Pianificata | Fase 1 |
| 3 | Lifecycle, dependency graph e Gestor | Pianificata | Fasi 1–2 |
| 4 | Laravel reference plugin | Pianificata | Fasi 1–3 |
| 5 | Osservabilità, hardening e gate finale | Pianificata | Fasi 1–4 |

Le fasi sono sequenziali rispetto ai contratti. Ogni fase termina con un report
in `docs/reports/`; la successiva può essere dichiarata completata soltanto dopo
il superamento del gate precedente.

Il codice esistente è una baseline da sottoporre ad audit: la sua presenza non
rende automaticamente completata una fase.

---

# Fase 1 — Contratti, audit della baseline e ADR-0023

## Obiettivo

Stabilire il confine definitivo della Milestone 5 e verificare che i contratti
pubblici esistenti rappresentino senza duplicazioni identità, compatibilità,
catalogo, registrazione e lifecycle.

## Sviluppo previsto

- inventariare API, errori, eventi e invarianti in `pkg/plugin`;
- auditare l'implementazione in `internal/plugin` rispetto ad ADR-0007 e
  ADR-0008;
- definire con precisione le semantiche di available, registered, loaded e
  running senza introdurre stato duplicato;
- verificare il confine tra manifest e `runtime.Metadata`;
- definire il comportamento supportato prima e dopo `Runtime.Start`;
- decidere se servono estensioni pubbliche additive per il catalogo o per la
  capability workspace;
- catalogare collisioni, cancellazione, typed nil e wrapping degli errori;
- formalizzare le decisioni in ADR-0023;
- costruire una matrice di compatibilità delle API pubbliche esistenti.

## Invarianti

- un plugin resta un `runtime.Component`;
- identità, dipendenze e capability non vengono duplicate nel manifest;
- lifecycle e stato non appartengono al Plugin Runtime;
- nessun requisito di distribuzione esterna entra nei contratti core;
- ogni eventuale estensione è additiva e motivata da un consumer concreto.

## Test richiesti

- test di compilazione delle interfacce pubbliche;
- validazione di manifest e ID;
- typed nil per plugin e loader;
- compatibilità degli errori con `errors.Is`;
- snapshot difensivi dei listing;
- audit della copertura per ogni metodo del contratto pubblico.

## Gate di uscita

- ADR-0023 in stato Accepted;
- matrice API e gap della baseline documentati;
- nessuna ambiguità tra caricamento, registrazione e lifecycle;
- test dei package plugin verdi;
- nessuna modifica breaking non esplicitamente approvata.

## Deliverable

- ADR-0023;
- eventuali correzioni additive ai contratti pubblici;
- test di contratto mancanti;
- report `docs/reports/milestone-5-phase-1.md`.

---

# Fase 2 — Catalogo, registry e caricamento

## Obiettivo

Rendere catalogo, registry e percorso di load deterministici e robusti sotto
errori, cancellazione e concorrenza.

## Sviluppo previsto

- completare la validazione di plugin, loader, manifest e identità;
- verificare catalogo e registry thread-safe con snapshot difensivi;
- garantire l'esecuzione delle factory senza lock interni;
- preservare le cause dei loader e del Runtime Core;
- definire e testare la concorrenza sullo stesso ID e su ID distinti;
- garantire che una failure non renda visibile un plugin parziale;
- verificare l'ordine deterministico di listing ed eventi;
- documentare il contratto che vieta risorse longeve nella factory;
- aggiungere fixture bloccanti per dimostrare l'assenza di callback sotto lock.

## Invarianti

- al massimo una registrazione globale riesce per ID;
- un loader fallito o cancellato non registra il risultato;
- il Plugin Runtime non mantiene lock durante codice esterno;
- una collisione con un componente normale non crea un indice plugin;
- nessun listing condivide la slice interna;
- `Load` non inizializza e non avvia il plugin.

## Test richiesti

- zero, uno e più loader;
- loader duplicato, assente, nil e typed nil;
- plugin nil, ID differente e manifest incompatibile;
- errore prima, durante e dopo la factory;
- load concorrente dello stesso ID e di ID differenti;
- registrazione concorrente;
- letture concorrenti di catalogo e registry;
- callback bloccanti e race detector.

## Gate di uscita

- suite `pkg/plugin` e `internal/plugin` verde;
- race detector verde sui package plugin;
- failure e concorrenza lasciano registry e catalogo coerenti;
- nessuna callback esterna sotto lock.

## Deliverable

- catalogo e registry consolidati;
- fixture concorrenti;
- documentazione runtime aggiornata;
- report `docs/reports/milestone-5-phase-2.md`.

---

# Fase 3 — Lifecycle, dependency graph e Gestor

## Obiettivo

Verificare il percorso completo dal plugin registrato al lifecycle globale e
alla discovery delle capability, senza introdurre ownership parallele.

## Sviluppo previsto

- testare registrazione e load prima dell'avvio;
- verificare il rifiuto deterministico durante startup, running e shutdown;
- verificare dipendenze plugin -> componente, componente -> plugin e plugin ->
  plugin;
- testare ordine topologico di startup e ordine inverso di shutdown;
- verificare transizioni e failure nello `StateManager` globale;
- confermare che Gestor scopra ogni plugin una sola volta come componente;
- verificare invalidazione dopo la registrazione e refresh esplicito;
- verificare resolution e dependency plan delle capability custom;
- coprire il percorso dal composition root pubblico.

## Invarianti

- esiste un solo dependency graph;
- esiste una sola state machine per componente/plugin;
- il Plugin Runtime non invoca capability lifecycle;
- il graph viene costruito soltanto dal Runtime Core;
- Gestor non carica né esegue plugin;
- una registrazione riuscita rende stale Gestor ma non avvia I/O di discovery.

## Test richiesti

- plugin senza capability lifecycle;
- plugin con configure, initialize, start, health e stop;
- dipendenze richieste, opzionali, mancanti e cicliche;
- failure a ogni transizione rilevante;
- registrazione tardiva;
- invalidazione e refresh Gestor;
- risoluzione di capability plugin con dependency plan;
- test end-to-end e race detector.

## Gate di uscita

- lifecycle e graph dimostrati autorevoli dai test;
- nessun indice o stato duplicato nel Plugin Runtime;
- integrazione Gestor verde senza nuova source plugin;
- suite repository-wide verde.

## Deliverable

- copertura end-to-end Runtime/Plugin/Gestor;
- eventuali correzioni di wiring;
- documentazione di lifecycle aggiornata;
- report `docs/reports/milestone-5-phase-3.md`.

---

# Fase 4 — Laravel reference plugin

## Obiettivo

Portare Laravel da fixture di validazione del runtime a reference plugin
framework-aware con un contratto minimo, stabile e consumabile dalle milestone
successive.

## Sviluppo previsto

- auditare configurazione, normalizzazione della root e limiti di lettura;
- stabilizzare la vista delle informazioni rilevate nel workspace;
- dichiarare una capability workspace namespaced coerente con Gestor;
- applicare la decisione di Fase 1: mantenere la facade Laravel concreta e non
  introdurre un contratto workspace framework-neutral senza il consumer del
  Context Engine;
- mantenere detection in `Initialize` e verifica ripetibile in `Health`;
- testare accessi concorrenti allo snapshot rilevato;
- allineare facade pubblica e implementazione interna;
- mantenere fuori scope comandi Artisan e analisi applicativa avanzata.

## Invarianti

- nessuna logica Laravel entra nel Runtime Core o in Gestor;
- la root esposta è assoluta e normalizzata;
- il manifest Composer è letto con un limite esplicito;
- health non modifica implicitamente lo snapshot inizializzato;
- le informazioni pubblicate non espongono mappe mutabili del manifest;
- la capability dichiarata è individuabile senza eseguire il plugin.

## Test richiesti

- root vuota, relativa, inesistente e valida;
- `artisan` mancante o non regolare;
- `composer.json` mancante, malformato o oltre limite;
- dipendenza Laravel mancante, vuota e valida;
- initialize, health, mutazione del workspace e concorrenza;
- listing Gestor della capability Laravel;
- caricamento e lifecycle dal Runtime pubblico.

## Gate di uscita

- reference plugin utilizzabile attraverso API pubbliche documentate;
- detection e health deterministici senza processi esterni;
- capability Laravel visibile tramite Gestor;
- test e race detector verdi.

## Deliverable

- contratti e implementazione Laravel consolidati;
- fixture aggiornate;
- `docs/laravel-plugin.md` aggiornato;
- report `docs/reports/milestone-5-phase-4.md`.

---

# Fase 5 — Osservabilità, hardening e gate finale

## Obiettivo

Chiudere la milestone con eventi verificati, audit di compatibilità,
documentazione coerente e una suite deterministica repository-wide.

## Sviluppo previsto

- auditare topic, payload, cardinalità e ordine degli eventi plugin;
- verificare che eventi e subscriber siano eseguiti fuori lock;
- testare interazioni tra errori, cancellazione ed eventi;
- verificare redazione dei payload pubblici;
- eseguire audit finale delle API pubbliche;
- eseguire suite completa, race detector e vet;
- allineare README, architecture, roadmap e contesto;
- produrre i report di fase e il report conclusivo;
- registrare esplicitamente i task rinviati alla Milestone 8.

## Invarianti

- un'operazione fallita non emette un evento di successo;
- gli eventi non cambiano il risultato dell'operazione completata;
- nessun evento viene pubblicato mantenendo lock interni;
- i payload non contengono configurazione, credenziali o contenuti del
  workspace;
- documentazione e implementazione descrivono lo stesso modello di fiducia.

## Test richiesti

- evento loader registered;
- evento plugin registered;
- sequenza registered -> loaded per `Load` riuscito;
- nessun evento di successo su failure o cancellazione;
- subscriber re-entrant e bloccanti;
- test end-to-end dal composition root;
- `go test ./...`;
- `go test -race` sui package Runtime, Plugin e Gestor;
- `go vet ./...`.

## Gate di uscita

- tutte le suite richieste verdi;
- API pubbliche sottoposte ad audit di compatibilità;
- ADR e documentazione allineati;
- cinque report di fase e report finale disponibili;
- nessun requisito differito presentato come implementato.

## Deliverable

- hardening finale;
- documentazione completa;
- `docs/reports/milestone-5-phase-5.md`;
- `docs/reports/milestone-5-final.md`.

---

# Gate finale della Milestone 5

La milestone può essere dichiarata completata soltanto quando:

- catalogo, registry e load rispettano gli invarianti concorrenti;
- compatibilità e failure sono rappresentate da contratti pubblici stabili;
- plugin e componenti condividono graph, stato e lifecycle;
- Gestor scopre capability plugin senza duplicazioni;
- Laravel dimostra il percorso framework-aware end-to-end;
- eventi, cancellazione e callback fuori lock sono coperti;
- test, race detector, vet e audit documentale sono verdi;
- packaging, sandbox, permission model, unload e plugin di terze parti restano
  esplicitamente fuori scope.

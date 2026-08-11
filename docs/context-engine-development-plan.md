# Milestone 6 — Context Engine Development Plan

Versione: 0.1.0

Stato: In corso — Fase 1 completata; Fase 2 pronta

Data: 2026-08-11

Documento architetturale di riferimento: `context-engine-design.md`.

---

# Obiettivo della milestone

Costruire un Context Engine provider-agnostic che indicizzi workspace locali,
produca evidenza strutturata, recuperi contenuti rilevanti e costruisca bundle
deterministici entro un budget dichiarato, con cache limitata e integrazione nel
Runtime pubblico.

La milestone non introduce memoria conversazionale, tool execution,
permission model, persistenza distribuita o ranking LLM.

---

# Sequenza delle fasi

| Fase | Titolo | Stato corrente | Dipende da |
|---|---|---|---|
| 1 | Contratti, ownership e ADR-0024 | Completata | Design iniziale |
| 2 | Workspace indexing e snapshot | Pianificata | Fase 1 |
| 3 | Analisi strutturata e AST | Pianificata | Fasi 1–2 |
| 4 | Retrieval, Context Builder e budget | Pianificata | Fasi 1–3 |
| 5 | Cache e aggiornamento incrementale | Pianificata | Fasi 1–4 |
| 6 | Integrazione, osservabilità e gate finale | Pianificata | Fasi 1–5 |

Le fasi sono sequenziali rispetto ai contratti e agli invarianti. Ogni fase
termina con un report in `docs/reports/`; la Fase 6 produce anche il report
conclusivo `docs/reports/milestone-6-final.md`.

Una fase non è completata dalla sola presenza del codice: deve superare test,
gate documentale e verifica degli invarianti dichiarati.

---

# Fase 1 — Contratti, ownership e ADR-0024

## Obiettivo

Definire il modello pubblico minimo e assegnare in modo non ambiguo ownership,
failure, concorrenza e confini di riservatezza prima di introdurre I/O o cache.

## Sviluppo previsto

- inventariare i confini Runtime, Gestor, Provider e Plugin consumati;
- definire il package pubblico `pkg/contextengine` e l'implementazione
  `internal/contextengine`;
- modellare workspace, policy, documento, snapshot e diagnostica;
- modellare analyzer, evidenza strutturata e registry delle estensioni;
- modellare query, risultato di retrieval, budget, sezione e context bundle;
- definire estimator, digest, generazioni e versioni degli algoritmi;
- stabilire errori sentinel e wrapping con `errors.Is`;
- definire il contratto workspace provider framework-neutral;
- decidere integrazione additiva col composition root senza cambiare
  `pkg/runtime.Runtime`;
- formalizzare le decisioni in ADR-0024;
- produrre una matrice iniziale di compatibilità API.

## Invarianti

- `context.Context`, `runtime.Context` e Context Engine restano concetti
  distinti;
- i tipi pubblici non espongono mappe o slice mutabili interne;
- i path dei documenti sono logici e relativi; la root assoluta resta un input
  validato e non entra negli eventi;
- identità, contenuto, analisi e selezione restano separati;
- le API non importano dettagli Laravel, PHP, Go, Ollama o llama.cpp;
- ogni interfaccia pubblica ha almeno due implementazioni plausibili;
- nessun contratto obbliga persistenza o rete.

## Test richiesti

- validazione di ID, root, path logici e limiti;
- costruttori e snapshot difensivi;
- enum e versioni non validi;
- typed nil per estensioni registrabili;
- compatibilità degli errori con `errors.Is`;
- assertion di compilazione dei contratti pubblici;
- audit delle dipendenze tra package.

## Gate di uscita

- ADR-0024 in stato Accepted;
- ownership e failure matrix documentate;
- API pubblica minimale sottoposta ad audit;
- nessun ciclo di dipendenze o modifica breaking del Runtime esistente;
- suite dei nuovi contratti verde.

## Deliverable

- `pkg/contextengine` iniziale;
- skeleton interno strettamente necessario ai contratti;
- ADR-0024 e audit API iniziale;
- report `docs/reports/milestone-6-phase-1.md`.

---

# Fase 2 — Workspace indexing e snapshot

## Obiettivo

Costruire snapshot testuali deterministici e sicuri di workspace locali,
pubblicati atomicamente e governati da policy e limiti espliciti.

## Sviluppo previsto

- implementare la source filesystem locale;
- normalizzare root e path logici con verifica di containment;
- applicare include, exclude, limiti e detection testuale/binaria;
- non seguire symlink nella baseline;
- calcolare digest content-addressed e metadata neutrali;
- ordinare documenti deterministicamente;
- implementare indice per workspace e generazioni monotone;
- pubblicare refresh all-or-nothing;
- preservare l'ultimo snapshot valido su failure o cancellazione;
- eseguire source e filesystem I/O senza lock dell'indice;
- aggiungere source in-memory per test dei contratti.

## Invarianti

- nessun documento esce dalla root dichiarata;
- ogni path logico identifica al massimo un documento nello snapshot;
- il digest dipende dal contenuto normalizzato, non da timestamp o inode;
- un refresh fallito non incrementa la generazione pubblicata;
- lettori concorrenti osservano snapshot completi;
- listing e documenti sono immutabili dal punto di vista del chiamante;
- limiti vengono verificati durante la scansione, non soltanto a posteriori.

## Test richiesti

- workspace vuoto, inesistente e non-directory;
- path profondi, nomi non validi e traversal;
- include/exclude e ordine di visita differente;
- file nascosti, binari, encoding non supportato e file oltre limite;
- symlink a file, directory e fuori root;
- mutazione durante la scansione;
- cancellazione prima e durante I/O;
- refresh iniziale, successivo, fallito e concorrente;
- source bloccante per dimostrare assenza di callback sotto lock;
- race detector sul package indice.

## Gate di uscita

- stesso contenuto e policy producono lo stesso snapshot logico;
- containment e limiti sono coperti da regressioni;
- pubblicazione atomica dimostrata sotto concorrenza;
- nessuna rete o dipendenza da tool esterni nella suite.

## Deliverable

- indice e source filesystem;
- fixture in-memory e filesystem;
- documentazione delle policy di scansione;
- report `docs/reports/milestone-6-phase-2.md`.

---

# Fase 3 — Analisi strutturata e AST

## Obiettivo

Arricchire i documenti con simboli, relazioni e chunk strutturali attraverso
analyzer sostituibili, mantenendo language-specific logic fuori dal core.

## Sviluppo previsto

- implementare registry thread-safe di analyzer versionati;
- definire matching esplicito per linguaggio, estensione e media type;
- rifiutare ambiguità senza ordine di preferenza configurato;
- eseguire analyzer fuori lock e con context cancellabile;
- validare intervalli, simboli, relazioni e diagnostiche prodotti;
- implementare un analyzer AST Go di riferimento con libreria standard;
- mantenere testo e diagnostica su sorgenti malformate o non supportate;
- generare chunk strutturali riferiti agli intervalli originali;
- aggiornare lo snapshot in modo atomico con i risultati validati;
- documentare il percorso per analyzer PHP e framework-aware futuri.

## Invarianti

- un analyzer non modifica documento o workspace;
- ogni intervallo resta entro il contenuto sorgente;
- ogni relazione riferisce simboli validi o target esterni espliciti;
- la versione analyzer partecipa all'identità dell'analisi;
- unsupported e malformed non sono confusi con failure infrastrutturali;
- un errore locale non elimina il testo indicizzabile;
- più analyzer applicabili richiedono una scelta esplicita.

## Test richiesti

- analyzer assente, duplicato, nil e typed nil;
- matching singolo, nessun matching e matching ambiguo;
- AST valido, sintassi incompleta e file vuoto;
- simboli annidati, import e intervalli multibyte;
- output analyzer non valido;
- cancellazione e panic controllato dell'estensione;
- analyzer bloccanti e concorrenti;
- snapshot difensivi e race detector.

## Gate di uscita

- contratto dimostrato da analyzer reale e fixture in-memory;
- nessuna dipendenza language-specific nel Runtime Core;
- diagnostiche locali e failure globali hanno semantiche distinte;
- analisi ripetuta produce output deterministico.

## Deliverable

- registry e pipeline analyzer;
- modello strutturale validato;
- analyzer Go di riferimento;
- report `docs/reports/milestone-6-phase-3.md`.

---

# Fase 4 — Retrieval, Context Builder e budget

## Obiettivo

Selezionare evidenza rilevante e costruire un bundle ispezionabile entro un
budget esplicito, senza ranking o precisione token nascosti.

## Sviluppo previsto

- implementare retrieval lessicale con algoritmo versionato;
- aggiungere filtri strutturali per path, linguaggio, simbolo e relazione;
- definire score, reason code e tie-break deterministici;
- integrare embedding opt-in tramite adapter al Provider Runtime;
- validare cardinalità, dimensioni e valori degli embedding;
- implementare strategie esplicite di fusione dei ranking;
- definire e implementare `TokenEstimator` con baseline conservativa;
- implementare deduplicazione per intervalli e selezione entro budget;
- implementare troncamento deterministico e margine di sicurezza;
- produrre context bundle con provenance e costo per sezione;
- migrare una fixture del Developer Benchmark sul percorso pubblico.

## Invarianti

- ogni risultato dichiara metodo, score e origine;
- tie-break non equivale a ranking implicito;
- retrieval semantico richiede provider e modello espliciti;
- vettori incompatibili non vengono confrontati;
- una sezione non perde il riferimento all'intervallo originale;
- il costo dichiarato non supera il budget del metodo scelto;
- una stima non viene etichettata come conteggio esatto;
- il builder non invoca completion o genera riassunti.

## Test richiesti

- query vuota, filtri invalidi e nessun risultato;
- ranking lessicale e tie deterministici;
- lookup simboli e relazioni;
- embedding batch, errori provider, cancellazione e dimensioni errate;
- fusione di ranking con score estremi e duplicati;
- budget zero, esatto, insufficiente e con riserva;
- sezioni sovrapposte, deduplicazione e troncamento multibyte;
- estimator conservativo e estimator fittizio esatto;
- bundle immutabili e output ripetibili;
- regressione benchmark offline.

## Gate di uscita

- retrieval lessicale e strutturale funzionano senza provider;
- semantic retrieval è opt-in e provider-neutral;
- ogni bundle è spiegabile e budgetato;
- Developer Benchmark usa almeno un percorso Context Engine senza modificare
  la rubrica canonica.

## Deliverable

- retrieval engine e adapter embedding;
- Context Builder e estimator baseline;
- integrazione benchmark iniziale;
- report `docs/reports/milestone-6-phase-4.md`.

---

# Fase 5 — Cache e aggiornamento incrementale

## Obiettivo

Evitare lavoro ripetuto tra refresh e build mantenendo la cache derivata,
limitata, cancellabile e semanticamente invisibile al risultato.

## Sviluppo previsto

- definire chiavi content-addressed complete e versionate;
- implementare cache in-memory con limiti per entry e totali;
- definire eviction deterministica e statistiche hit/miss/eviction;
- riusare analisi, chunk, embedding e stime soltanto a chiave identica;
- confrontare snapshot per digest senza affidarsi ai timestamp;
- invalidare su versione analyzer, chunker, estimator o modello;
- impedire la pubblicazione di entry parziali su failure;
- verificare concorrenza senza assumere singleflight tra context distinti;
- introdurre hook di clock soltanto se necessari a una policy testabile;
- misurare equivalenza funzionale cache cold/warm.

## Invarianti

- cache miss e hit producono lo stesso risultato osservabile;
- la cache non è la fonte autorevole dello snapshot;
- ogni entry ha una chiave che copre le dipendenze semantiche;
- risultati falliti o cancellati non diventano hit;
- eviction non richiede callback esterne sotto lock;
- limiti di memoria sono applicati anche sotto concorrenza;
- nessun contenuto viene persistito implicitamente.

## Test richiesti

- cold miss e warm hit;
- modifica, rename, delete e contenuto uguale su path differente;
- cambio di policy o versione algoritmo;
- cambio provider, modello e dimensione embedding;
- entry oltre limite ed eviction;
- failure e cancellazione durante produzione;
- richieste concorrenti uguali e distinte;
- equivalenza byte-for-byte dei bundle cold/warm;
- race detector e test ripetuti.

## Gate di uscita

- invalidazione completa dimostrata dalla matrice delle chiavi;
- cache bounded sotto carico concorrente;
- nessuna differenza funzionale tra percorso cold e warm;
- contenuti non serializzati fuori processo.

## Deliverable

- cache content-addressed;
- refresh incrementale per artefatti invariati;
- metriche interne redatte;
- report `docs/reports/milestone-6-phase-5.md`.

---

# Fase 6 — Integrazione, osservabilità e gate finale

## Obiettivo

Comporre il Context Engine nel Runtime pubblico, dimostrare il percorso
workspace-aware con Laravel e chiudere la milestone con osservabilità redatta,
hardening e audit di compatibilità.

## Sviluppo previsto

- aggiungere l'accessor additivo al composition root pubblico;
- condividere Config, Logger ed Event Bus senza duplicazioni;
- integrare il contratto workspace provider nel plugin Laravel;
- pubblicare capability Context Engine attraverso il wiring Gestor previsto;
- verificare che Gestor non esegua indexing o analyzer;
- definire topic e payload per refresh, build e cache summary;
- isolare errori e panic degli observer da operazioni già committate;
- completare la migrazione del retrieval Developer Benchmark;
- aggiungere test end-to-end con workspace Laravel temporaneo;
- eseguire audit finale delle API pubbliche;
- allineare README, architecture, roadmap e contesto;
- produrre sei report di fase e report conclusivo.

## Invarianti

- esiste una sola istanza Context Engine nel composition root;
- Runtime Core non acquisisce logica di workspace o linguaggio;
- Context Engine non importa Laravel;
- Gestor descrive capability ma non esegue la pipeline;
- eventi e log non contengono query, testo, embedding o path assoluti;
- observer sono invocati fuori lock;
- un bundle viene consegnato soltanto al chiamante esplicito;
- le API esistenti restano compatibili salvo decisione ADR esplicita.

## Test richiesti

- accessor e wiring del composition root;
- Laravel workspace provider e snapshot Context Engine;
- discovery Gestor senza esecuzione implicita;
- sequenza e cardinalità degli eventi;
- nessun evento di successo su failure o cancellazione;
- subscriber re-entrant, lento e in panic;
- redazione di payload, log e report;
- end-to-end indexing -> AST -> retrieval -> bundle;
- `go test ./...`;
- `go test -race` sui package Runtime, Context Engine, Gestor e Plugin;
- `go vet ./...` e `git diff --check`.

## Gate di uscita

- percorso pubblico completo verificato offline;
- suite completa, race detector e vet verdi;
- API pubbliche sottoposte ad audit di compatibilità;
- ADR, design, roadmap e implementazione allineati;
- sei report di fase e report finale disponibili;
- memoria agente, permission model e persistenza restano fuori scope.

## Deliverable

- composition root e integrazioni finali;
- eventi e documentazione operativa;
- `docs/reports/milestone-6-phase-6.md`;
- `docs/reports/milestone-6-final.md`.

---

# Gate finale della Milestone 6

La milestone può essere dichiarata completata soltanto quando:

- workspace e policy producono snapshot sicuri, atomici e deterministici;
- analyzer sostituibili producono evidenza strutturata validata;
- retrieval lessicale e strutturale funzionano offline;
- retrieval semantico resta opt-in e usa il Provider Runtime esistente;
- Context Builder rispetta budget, provenance e metodo di stima dichiarato;
- cache e aggiornamento incrementale sono bounded e semanticamente trasparenti;
- Laravel dimostra il contratto workspace framework-neutral end-to-end;
- Gestor, Runtime e Plugin mantengono le ownership già approvate;
- eventi, log e report rispettano il confine di riservatezza;
- test, race detector, vet, audit API e documentazione sono verdi;
- memoria conversazionale, tool permissions, persistenza e ranking LLM non
  vengono presentati come implementati.

# Milestone 4 — Gestor Development Plan

Versione: 0.1.0

Stato: In corso — Fase 1 completata, Fase 2 pronta

Data: 2026-08-09

Documento architetturale di riferimento: `gestor-design.md`.

---

# Obiettivo della milestone

Consegnare un servizio Gestor in-process che indicizzi capability provenienti
dalle sorgenti autorevoli di Maestro, distingua dichiarazione e disponibilità
operativa e produca risoluzioni deterministiche, spiegabili e compatibili con
il dependency graph unico del Runtime Core.

La milestone non introduce esecuzione delle capability, ranking impliciti,
discovery remota o un secondo grafo.

---

# Sequenza delle fasi

| Fase | Titolo | Stato corrente | Dipende da |
|---|---|---|---|
| 1 | Contratti, modello di dominio e ADR | Completata | Design iniziale |
| 2 | Snapshot Registry | Pronta | Fase 1 |
| 3 | Discovery sources | Pianificata | Fase 2 |
| 4 | Resolver e dependency graph | Pianificata | Fasi 2–3 |
| 5 | Composition root, osservabilità e gate finale | Pianificata | Fasi 1–4 |

Le fasi sono sequenziali sul contratto, ma i test di una fase devono restare
verdi durante tutte le fasi successive.

Ogni fase termina con un report in `docs/reports/`. La fase successiva può
iniziare soltanto dopo il superamento del relativo gate deterministico.

---

# Fase 1 — Contratti, modello di dominio e ADR

## Obiettivo

Stabilizzare il linguaggio di Gestor e il confine pubblico minimo prima di
introdurre stato o integrazioni.

## Sviluppo previsto

- definire `CapabilityID`, `SourceID`, target kind, scope e availability;
- definire descriptor, query, resolution e snapshot metadata;
- definire gli errori sentinella e la loro semantica;
- stabilire validazione, uguaglianza e ordinamento degli identificatori;
- decidere quali tipi appartengono a un package pubblico minimale e quali
  restano interni;
- definire la mappatura namespaced da `runtime.Capability` e
  `provider.Capability` senza cambiare i contratti esistenti;
- registrare le decisioni in ADR-0022;
- aggiornare il design se il contratto approvato diverge dal frammento
  illustrativo iniziale.

## Invarianti

- nessun identificatore vuoto o con whitespace non normalizzato;
- availability distinta dalla presenza del descriptor;
- target model valido soltanto con model ID esatto;
- nessun tipo pubblico espone mappe o slice mutabili interne;
- nessuna dipendenza da implementazioni concrete di provider o plugin.

## Test richiesti

- validazione di valori validi e invalidi;
- copie difensive di slice e metadata;
- ordinamento deterministico;
- compatibilità degli errori tramite `errors.Is`;
- mapping completo delle capability Runtime e Provider note;
- test di compilazione delle interfacce pubbliche minime.

## Gate di uscita

- contratti e test approvati;
- ADR-0022 in stato Accepted;
- `go test` dei package coinvolti verde;
- `go vet` verde;
- nessuna implementazione prematura di registry o resolver.

## Deliverable

- contratti Gestor;
- ADR-0022;
- documentazione aggiornata;
- report `docs/reports/milestone-4-phase-1.md`.

---

# Fase 2 — Snapshot Registry

## Obiettivo

Implementare il catalogo delle sorgenti e il Registry immutabile delle
capability, con refresh atomico e generazioni.

## Sviluppo previsto

- catalogo sorgenti thread-safe;
- registrazione con collision detection per source ID;
- esecuzione delle sorgenti in ordine deterministico;
- composizione e validazione di un candidato snapshot locale;
- swap atomico soltanto dopo il successo completo;
- indice capability → descriptor;
- indice target → descriptor;
- generazione monotona dello snapshot;
- listing con copie difensive;
- invalidazione esplicita;
- conservazione dell'ultimo snapshot valido in caso di errore.

## Invarianti

- nessuna sorgente viene eseguita sotto lock;
- nessun risultato parziale diventa visibile;
- una collisione tra sorgenti autorevoli è un errore;
- un refresh cancellato non cambia generazione o snapshot;
- listing e snapshot hanno ordine stabile;
- lo stato di un refresh appartiene alla singola chiamata.

## Test richiesti

- zero, una e più sorgenti;
- source ID duplicato;
- descriptor duplicato o invalido;
- errore e cancellazione a metà refresh;
- conservazione dell'ultimo snapshot valido;
- incremento corretto della generazione;
- isolamento delle copie restituite;
- refresh e letture concorrenti;
- race detector sul package Gestor.

## Gate di uscita

- test unitari e concorrenti verdi;
- `go test -race` verde;
- nessuna chiamata esterna sotto lock verificata con fixture bloccanti;
- API del Registry coerente con ADR-0022.

## Deliverable

- implementazione Snapshot Registry;
- fixture di sorgente in-memory;
- report `docs/reports/milestone-4-phase-2.md`.

---

# Fase 3 — Discovery sources

## Obiettivo

Popolare Gestor attraverso le fonti autorevoli già presenti senza duplicare
registry o cataloghi.

## Sviluppo previsto

- sorgente Runtime component metadata;
- mapping delle capability lifecycle negli ID namespaced Gestor;
- sorgente Provider capability introspection;
- mapping dei livelli adapter, instance e model;
- traduzione di support e availability nei descriptor Gestor;
- plugin discovery attraverso il Registry globale dei componenti;
- interfaccia sorgente interna additiva per Provider Runtime;
- invalidazione quando cambiano componenti o provider registrati;
- fixture positive, negative e unknown senza dipendenze live.

## Invarianti

- il Runtime Registry resta proprietario dei componenti;
- il Provider Runtime resta proprietario dei provider;
- il Plugin Runtime non viene reindicizzato separatamente;
- `Resolve` non esegue introspection;
- model ID e scope provenienti dall'introspection restano esatti;
- dichiarato non viene promosso automaticamente ad available;
- errori di una sorgente impediscono la pubblicazione dell'intero refresh.

## Test richiesti

- componenti con capability zero, singole e multiple;
- plugin presente una sola volta;
- mapping di tutte le capability provider note;
- adapter supported con availability unknown;
- instance e model available/unavailable;
- caso dichiarato ma non operativo equivalente alla fixture Qwen;
- caso operativo equivalente alla fixture Llama 3.1;
- propagazione di errori e cancellazione dell'introspection;
- invalidazione dopo nuova registrazione;
- test concorrenti e race detector.

## Gate di uscita

- entrambe le sorgenti producono descriptor deterministici;
- distinzione declared/operational dimostrata dai test;
- nessuna modifica breaking alle interfacce Runtime, Provider o Plugin;
- suite repository-wide verde.

## Deliverable

- Runtime component source;
- Provider capability source;
- integrazione plugin senza duplicazioni;
- report `docs/reports/milestone-4-phase-3.md`.

---

# Fase 4 — Resolver e dependency graph

## Obiettivo

Risolvere capability su snapshot immutabili usando filtri e preferenze
esplicite, integrando il dependency graph senza duplicarlo.

## Sviluppo previsto

- query validation;
- filtro per capability, target kind, scope e model ID;
- filtro opzionale `require available`;
- esclusione dei target unavailable;
- preferenze ordinate esplicite;
- errori distinti not found, unavailable e ambiguous;
- risultato con descriptor, generazione e motivazione;
- vista read-only del dependency graph Runtime;
- verifica di eleggibilità del componente;
- piano delle dipendenze in ordine topologico;
- rilevazione di snapshot o grafo non correnti.

## Invarianti

- l'ordine lessicografico non seleziona implicitamente un candidato;
- più candidati senza preferenza producono ambiguous;
- Gestor non aggiunge nodi o archi;
- dipendenze richieste mancanti impediscono la risoluzione;
- dipendenze opzionali mancanti non la impediscono;
- cicli restano responsabilità del validator Runtime;
- la risoluzione non esegue codice del candidato.

## Test richiesti

- zero, uno e più candidati;
- preferenza valida, mancante e non disponibile;
- unknown con e senza requisito operativo;
- filtri exact target e model;
- differenza tra not found e unavailable;
- ambiguity deterministica;
- dipendenze richieste e opzionali;
- ordine topologico;
- snapshot e graph generation non coerenti;
- risoluzioni concorrenti e race detector.

## Gate di uscita

- tutti i rami di risoluzione coperti;
- nessuna euristica di ranking implicita;
- dependency graph unico verificato dai test di integrazione;
- output e diagnostica deterministici;
- suite repository-wide e race detector verdi.

## Deliverable

- Capability Resolver;
- integrazione read-only con il dependency graph;
- report `docs/reports/milestone-4-phase-4.md`.

---

# Fase 5 — Composition root, osservabilità e gate finale

## Obiettivo

Integrare Gestor nel Runtime pubblico, coordinare invalidazione ed eventi e
chiudere il gate tecnico della Milestone 4.

## Sviluppo previsto

- composizione del servizio Gestor nel Runtime Core;
- esposizione attraverso un contratto pubblico minimale;
- wiring delle sorgenti Runtime e Provider;
- invalidazione coordinata dopo registrazioni autorizzate;
- policy esplicita di refresh iniziale;
- eventi redatti per refresh e risoluzione;
- nessuna pubblicazione di eventi sotto lock;
- documentazione d'uso ed esempi;
- aggiornamento di architecture, roadmap e context;
- verifica di compatibilità dei package pubblici.

## Invarianti

- Runtime, Provider Runtime e Plugin Runtime mantengono le proprietà esistenti;
- Gestor non esegue la capability risolta;
- errori o panic degli observer non corrompono lo snapshot;
- la registrazione non avvia I/O live sotto lock;
- eventi e diagnostica non includono dati sensibili;
- la chiusura della Milestone 4 non modifica lo stato sospeso della Milestone 3.

## Test richiesti

- bootstrap con zero e più componenti/provider;
- refresh e resolve attraverso il composition root;
- invalidazione dopo registrazione;
- plugin visibile come componente senza duplicazione;
- eventi success/failure redatti;
- observer lenti, in errore o in panic;
- lifecycle Runtime invariato;
- test end-to-end in-memory;
- suite completa, race detector e vet.

## Gate di uscita

- `go test ./...` verde;
- `go test -race ./...` verde;
- `go vet ./...` verde;
- API pubbliche e documentazione coerenti;
- cinque report di fase presenti;
- nessun task Gestor bloccante non documentato.

## Deliverable

- Gestor composto nel Runtime;
- osservabilità ed esempi;
- report `docs/reports/milestone-4-phase-5.md`;
- report conclusivo `docs/reports/milestone-4-final.md`.

---

# Regole trasversali

Valgono per tutte le fasi:

1. non introdurre dipendenze esterne senza decisione esplicita;
2. non esporre implementazioni interne attraverso contratti pubblici;
3. non mantenere lock durante codice esterno, introspection o observer;
4. conservare ordinamento e output deterministici;
5. preservare cause e sentinel negli errori;
6. aggiornare test e documentazione nella stessa fase;
7. produrre un report finale prima di avanzare;
8. non modificare il benchmark per adattarlo all'implementazione Gestor;
9. mantenere la Milestone 3 sospesa e separata.

---

# Struttura dei report di fase

Ogni report deve contenere:

- stato e data;
- obiettivo della fase;
- file e contratti introdotti;
- invarianti implementati;
- test eseguiti con esito;
- decisioni architetturali;
- limitazioni e attività rinviate;
- gate di uscita;
- stato della fase successiva.

Un report non può dichiarare completata una fase se rimangono deliverable o
test obbligatori non eseguiti.

---

# Relazione con la Milestone 3

La Milestone 4 può avanzare indipendentemente dalla matrice live llama.cpp. Il
task resta registrato nella Milestone 3 e deve essere completato prima della sua
chiusura formale o di una release pubblica importante.

La chiusura di Gestor non implica la chiusura automatica della Milestone 3.

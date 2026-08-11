# Milestone 5 — Plugin System Final Report

Stato: Completata

Data: 2026-08-11

---

# Risultato

Maestro dispone di un Plugin System trusted in-process stabile, deterministico
nei confini dichiarati e integrato con il Runtime pubblico.

Il percorso completo verificato è:

```text
catalogo -> load -> registrazione -> Gestor -> dependency graph -> lifecycle
```

La milestone consolida la baseline precedente senza creare un secondo sistema
di componenti e senza anticipare packaging o isolamento non progettati.

---

# Fasi completate

| Fase | Ambito | Esito |
|---|---|---|
| 1 | Contratti, audit della baseline e ADR-0023 | Completata |
| 2 | Catalogo, registry e caricamento | Completata |
| 3 | Lifecycle, dependency graph e Gestor | Completata |
| 4 | Laravel reference plugin | Completata |
| 5 | Osservabilità, hardening e gate finale | Completata |

I report dettagliati sono disponibili in `docs/reports/`.

---

# Contratti e ownership

- un plugin è un `runtime.Component` con `plugin.Manifest`;
- `runtime.Metadata` possiede identità, versione, dipendenze e capability;
- il manifest dichiara soltanto la versione API richiesta;
- `Available` descrive il catalogo loader;
- `Registered` e `Has` descrivono istanze registrate;
- loaded è un evento di operazione riuscita, non uno stato persistente;
- running e failure appartengono allo `StateManager` globale;
- graph e lifecycle appartengono esclusivamente al Runtime Core;
- Gestor indicizza plugin attraverso il Registry globale dei componenti.

ADR-0023 formalizza il modello pre-start trusted in-process.

---

# Catalogo e caricamento

- loader e registry thread-safe;
- snapshot difensivi e ordine di registrazione riuscita;
- factory, registrar ed eventi invocati fuori lock;
- cancellazione verificata prima e dopo la factory;
- cause e sentinel preservati con `errors.Is`;
- failure prive di indici parziali nel Plugin Runtime;
- tentativi concorrenti indipendenti;
- una sola registrazione riuscita per ID;
- nessun singleflight implicito tra context distinti.

---

# Lifecycle e Gestor

- plugin passivi e lifecycle completi usano la stessa state machine;
- plugin e componenti condividono archi in entrambe le direzioni;
- startup dependency-first e shutdown inverso;
- required, optional, missing e cicli verificati;
- registrazione e load limitati al pre-start;
- invalidazione Gestor dopo registrazione riuscita;
- refresh esplicito;
- descriptor unico e dependency plan topologico;
- resolution senza esecuzione del plugin.

Una regressione ha corretto la classificazione della registrazione durante
shutdown in `runtime.ErrInvalidState`.

---

# Reference plugin Laravel

Laravel `0.2.0` dichiara:

```text
plugin.workspace-detection
```

La root è assoluta e immutabile. La versione framework viene pubblicata
atomicamente soltanto dopo initialize riuscita; health e reinitialize fallite
conservano l'ultimo snapshot valido. Detection, mutation e concorrenza sono
verificate senza processi esterni.

La facade resta concreta; nessuna API workspace framework-neutral viene
definita prima dei requisiti del Context Engine.

---

# Osservabilità

- topic pubblici stabili per loader registered, registered e loaded;
- ordine e cardinalità verificati;
- nessun success event su failure o cancellazione;
- callback sincroni, re-entrant e fuori lock;
- errori e panic degli observer isolati dal risultato già committato;
- payload minimale con ID e riferimento trusted in-process;
- nessuna serializzazione implicita di configurazione o workspace.

---

# Gate tecnico

Superati:

- suite completa repository-wide;
- race detector repository-wide;
- `go vet` repository-wide;
- suite plugin ripetuta venti volte;
- audit di compatibilità pubblica;
- `git diff --check`;
- audit documentale e report di tutte le fasi.

---

# Fuori scope confermato

- download e installazione di plugin esterni;
- marketplace e discovery remota;
- firme, provenance e aggiornamenti;
- shared object Go;
- process isolation, sandbox e permission model;
- unload, hot replacement e registrazione dopo startup;
- SDK di distribuzione e plugin di terze parti;
- comandi Artisan e tool execution;
- funzionalità del Context Engine.

Packaging e terze parti restano assegnati alla Milestone 8; il permission model
operativo alla Milestone 7.

---

# Conclusione

Il gate finale della Milestone 5 è superato. Il Plugin System può essere usato
come fondazione stabile dalla Milestone 6 — Context Engine.

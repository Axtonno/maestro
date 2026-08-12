# Milestone 7 — Agent System Development Plan

Versione: 0.2.0

Stato: Attivo — Fasi 1–2 completate, Fase 3 pianificata

Data: 2026-08-11

Documento architetturale di riferimento: `agent-system-design.md`.

---

# Obiettivo della milestone

Costruire un Agent System provider-agnostic che pianifichi ed esegua task su un
workspace attraverso tool registrati, autorizzazioni esplicite e un loop
bounded modello–tool, riusando Provider Runtime, Context Engine, Gestor ed Event
Bus già composti da Maestro.

La milestone consegna il primo agente autonomo ma non introduce memoria di
lungo periodo, sandbox di codice non fidato, multi-agent o una CLI completa.

---

# Sequenza delle fasi

| Fase | Titolo | Stato corrente | Dipende da |
|---|---|---|---|
| 1 | Contratti, ownership e ADR-0025 | Completata | Design iniziale |
| 2 | Tool catalog e execution boundary | Completata | Fase 1 |
| 3 | Permission model e approval flow | Pianificata | Fasi 1–2 |
| 4 | Sessioni, piani e budget | Pianificata | Fasi 1–3 |
| 5 | Loop agentico e tool calling | Pianificata | Fasi 1–4 |
| 6 | Workspace awareness e reference tool | Pianificata | Fasi 1–5 |
| 7 | Integrazione, osservabilità e gate finale | Pianificata | Fasi 1–6 |

Le fasi sono sequenziali rispetto ai contratti e agli invarianti. Ogni fase
termina con un report in `docs/reports/`; la Fase 7 produce anche
`docs/reports/milestone-7-final.md`.

Una fase non è completata dalla sola presenza del codice: deve superare test,
gate documentale e verifica degli invarianti dichiarati.

---

# Fase 1 — Contratti, ownership e ADR-0025

## Obiettivo

Definire la superficie pubblica minima di Tool System e Agent System e
assegnare ownership, transizioni, failure, limiti, privacy e integrazioni prima
di introdurre esecuzione o chiamate ai modelli.

## Sviluppo previsto

- inventariare le API Provider, Context Engine, Gestor, Runtime e Plugin usate;
- creare i package pubblici `pkg/tool` e `pkg/agent`;
- definire identificatori, descriptor, versioni e capability;
- modellare arguments, prepared invocation, action e tool result immutabili;
- modellare policy, decision, grant e approver senza implementare effetti;
- modellare agent request, limits, session snapshot, plan, step e result;
- definire errori sentinel, envelope tipizzati e wrapping con `errors.Is`;
- stabilire l'integrazione additiva futura con `maestro.Runtime`;
- definire cardinalità e redazione degli eventi;
- formalizzare le decisioni in ADR-0025;
- produrre audit iniziale di compatibilità e dependency direction.

## Invarianti

- `pkg/tool` non importa `pkg/agent`;
- contratti pubblici non importano adapter provider, Laravel o implementazioni interne;
- metadata, prepared invocation e snapshot non espongono backing mutable;
- permission request e decisione sono separate dal risultato del tool;
- il modello non può rappresentare una decisione di autorizzazione valida;
- provider, model, workspace e agent ID sono input espliciti;
- `pkg/runtime.Runtime` resta invariato;
- ogni interfaccia pubblica ha almeno due implementazioni plausibili.

## Test richiesti

- ID, versioni, enum, schema e limiti validi/non validi;
- nil e typed nil per estensioni;
- copie difensive di mappe, slice e raw JSON;
- transizioni di plan e session snapshot;
- error inspection con `errors.Is`/`errors.As`;
- assertion di compilazione e audit dei package import.

## Gate di uscita

- ADR-0025 Accepted;
- API pubblica minimale sottoposta ad audit;
- ownership e failure matrix documentate;
- nessuna modifica breaking dei contratti esistenti;
- suite dei nuovi package pubblici verde.

## Deliverable

- `pkg/tool` e `pkg/agent` iniziali;
- `docs/adr/ADR-0025.md` e indice ADR aggiornato;
- `docs/agent-system-api-compatibility-audit.md`;
- report `docs/reports/milestone-7-phase-1.md`.

---

# Fase 2 — Tool catalog e execution boundary

## Obiettivo

Costruire un Tool Runtime deterministico e thread-safe che separi preparazione
senza effetti ed esecuzione, senza ancora concedere permessi impliciti.

ADR-0025 definisce già il permit interno: non è una `Decision` pubblica, lega
issuer, run ID e permission fingerprint ed è consumato dall'executor. Questa
fase implementa il controllo obbligatorio con un issuer deterministico di test;
policy, Approver e grant reali arrivano nella Fase 3.

## Sviluppo previsto

- implementare catalogo e registry di tool trusted in-process;
- validare descriptor, nome provider, versione e JSON Schema;
- introdurre listing ordinati, resolution esatta e snapshot difensivi;
- implementare `Prepare` con validazione e normalizzazione delle action;
- implementare un executor che richieda un permit interno verificabile;
- validare prepared invocation, action fingerprint e correlazione call ID;
- applicare limiti a durata, byte, elementi e diagnostiche del risultato;
- classificare invalid input, denied, failed, cancelled e truncated;
- isolare panic dei tool al boundary senza nascondere failure;
- aggiungere fixture in-memory pure, mutanti e bloccanti.

## Invarianti

- registrazione duplicata non sostituisce il tool esistente;
- schema e descriptor non vengono mutati dal chiamante;
- `Prepare` non produce effetti osservabili;
- `Execute` non parte senza decisione allow valida per lo stesso fingerprint;
- catalogo ed executor non mantengono lock durante codice tool;
- output eccedente viene limitato prima di entrare nella sessione;
- nessun retry implicito di tool mutanti;
- hot replacement e unload non fanno parte della baseline.

## Test richiesti

- zero, uno e più tool, ordine e collisioni;
- ID/nome diversi, duplicati, nil e typed nil;
- schema malformato e arguments non-oggetto;
- prepare fallita, cancellata, in panic e bloccante;
- decisione assente, fingerprint differente e grant scaduto;
- execute success, failure, cancel, panic e output oltre limite;
- letture e registrazioni concorrenti;
- fixture che dimostrano assenza di callback sotto lock;
- race detector su `pkg/tool` e `internal/tool`.

## Gate di uscita

- nessun effetto può attraversare l'executor senza allow verificabile;
- catalogo deterministico e coerente sotto concorrenza;
- errori e risultati limitati hanno semantica documentata;
- suite tool e race detector verdi.

## Deliverable

- `internal/tool`;
- catalogo, preparazione ed executor;
- fixture tool riusabili;
- `docs/tool-runtime.md`;
- report `docs/reports/milestone-7-phase-2.md`.

---

# Fase 3 — Permission model e approval flow

## Obiettivo

Implementare policy default-deny e approvazione opzionale sulle action esatte,
separando in modo verificabile proposta del modello, decisione ed effetto.

## Sviluppo previsto

- implementare effect e resource matcher exact e workspace-scoped;
- definire regole ordinate senza wildcard ambigue nella baseline;
- implementare decisioni allow, deny e prompt;
- introdurre `Approver` cancellabile configurato dal chiamante;
- gestire grant one-shot e run-scoped legati al fingerprint;
- autorizzare separatamente `model.invoke` e `model.disclose`;
- produrre motivazioni tramite reason code stabili e redatti;
- rendere deny un tool result tipizzato senza esecuzione;
- impedire escalation per prefisso, alias path o cambio degli arguments;
- documentare il confine tra policy in-process e sandbox.

## Invarianti

- nessuna regola applicabile significa deny;
- prompt senza approver non significa allow;
- agent, model e tool non possono emettere grant validi;
- un grant vale soltanto per action e scope approvati;
- normalizzazione precede sempre la valutazione;
- modifica degli arguments invalida fingerprint e decisione;
- approver e policy vengono invocati senza lock del Tool Runtime;
- eventi di decisione non includono resource sensibili complete.

## Test richiesti

- policy vuota, allow, deny e prompt;
- exact resource, workspace diverso ed effect diverso;
- one-shot consumato e run grant riusabile nel solo run;
- prompt approved, denied, errore, cancel e typed nil approver;
- path alias e mutation dopo approval;
- model invocation/disclosure consentite separatamente;
- richieste concorrenti e approver bloccante;
- prova negativa che deny non invoca `Execute`;
- eventi di decisione redatti.

## Gate di uscita

- default-deny dimostrato end-to-end;
- nessun percorso executor bypassa l'authorizer;
- grant scope e fingerprint coperti da regressioni;
- policy e approval flow documentati e race-safe.

## Deliverable

- authorizer, policy registry e approval flow;
- fixture approver deterministiche;
- `docs/agent-permissions.md`;
- report `docs/reports/milestone-7-phase-3.md`.

---

# Fase 4 — Sessioni, piani e budget

## Obiettivo

Costruire l'aggregato di sessione e una pianificazione validata con stato
immutabile, terminale unico e limiti indipendenti dal prompt.

## Sviluppo previsto

- implementare registry in-memory delle sessioni attive e concluse bounded;
- validare request, target espliciti e hard limits;
- implementare state machine del run e dei plan step;
- costruire snapshot immutabili e generazioni monotone;
- introdurre planner interface e planner fixture deterministico;
- integrare un planner structured-output tramite Provider Runtime fittizio;
- parsare e validare piano, dipendenze, cardinalità e revisioni;
- implementare contatori per turni, tool, token, byte e durata;
- applicare deadline e cancellazione parent;
- limitare messaggi e risultati conservati in memoria;
- chiudere la sessione esattamente una volta con reason tipizzato.

## Invarianti

- una sessione appartiene a un solo run ID;
- gli snapshot pubblicati sono completi e difensivi;
- generazioni e contatori sono monotoni;
- uno step non salta transizioni invalide;
- un piano prodotto dal modello non è autorevole prima della validazione;
- revisione del piano non cancella la storia delle versioni bounded;
- nessun limite può essere aumentato dal modello;
- un run terminale non torna attivo e non emette un secondo terminale.

## Test richiesti

- request e limits al minimo, massimo e invalidi;
- piano vuoto, duplicati, ciclo, dipendenza mancante e output malformato;
- tutte le transizioni valide e invalide;
- revision limit e step blocked/skipped;
- deadline prima/durante planning;
- turn, tool, token e byte budget esauriti;
- chiusure concorrenti e cancel race;
- snapshot difensivi e registry bounded;
- race detector sui package agent.

## Gate di uscita

- state machine e terminale unico dimostrati sotto concorrenza;
- ogni loop futuro dispone di hard ceiling verificabili;
- planning failure non crea sessione parzialmente running;
- memoria in-process bounded e documentata.

## Deliverable

- `internal/agent` baseline;
- session registry, planner e budget controller;
- `docs/agent-sessions.md`;
- report `docs/reports/milestone-7-phase-4.md`.

---

# Fase 5 — Loop agentico e tool calling

## Obiettivo

Collegare sessione, Provider Runtime e Tool Runtime in un loop sequenziale,
cancellabile e bounded capace di completare task o eseguire tool call correlate.

## Sviluppo previsto

- tradurre descriptor tool validati in `provider.Tool`;
- costruire messaggi di sistema, utente, assistant e tool separati;
- invocare provider e modello esatti tramite Provider Runtime;
- gestire terminale testuale e finish reason tool calls;
- validare call ID, nome, arguments e cardinalità per turno;
- eseguire tool call multiple in ordine dichiarato;
- aggiungere risultati denied/failed come messaggi tool tipizzati;
- impedire retry implicito degli effetti mutanti;
- supportare streaming con assemblaggio bounded dei delta tool;
- distinguere retry provider da nuovo turno semantico;
- aggiornare plan step e session snapshot a ogni boundary;
- terminare in modo deterministico su cancel, budget o failure.

## Invarianti

- l'Agent Runtime usa il Provider Runtime condiviso, mai adapter concreti;
- provider e modello non cambiano durante un run senza richiesta esplicita;
- solo tool registrati e inclusi nel run possono essere invocati;
- ogni tool result mantiene correlazione con la call originaria;
- arguments incompleti o malformati non raggiungono `Prepare`;
- call multiple non vengono parallelizzate implicitamente;
- un turn provider produce al massimo una transizione terminale;
- ogni iterazione consuma budget prima di avviarne una successiva.

## Test richiesti

- risposta finale senza tool;
- una e più call, ordine e correlazione;
- tool sconosciuto, duplicato, denied, failed e cancelled;
- arguments malformati e call ID assente/provider-specific;
- output troncato restituito al modello;
- stream testuale, delta tool frammentati, EOF e mid-stream failure;
- provider retry e circuit breaker già configurati;
- max turn/tool/deadline durante il loop;
- cancellation in planning, completion, approval ed execution;
- nessuna rete nella suite ordinaria.

## Gate di uscita

- loop completo dimostrato con provider e tool fittizi;
- nessun bypass di registry, authorizer o budget;
- non-stream e stream convergono sulla stessa semantica terminale;
- suite agent/tool/provider integration verde.

## Deliverable

- orchestratore agentico;
- assembler streaming e adapter provider/tool;
- `docs/agent-runtime.md`;
- report `docs/reports/milestone-7-phase-5.md`.

---

# Fase 6 — Workspace awareness e reference tool

## Obiettivo

Rendere il reference agent consapevole di un workspace reale attraverso il
Context Engine e un set minimo di tool filesystem controllati.

## Sviluppo previsto

- associare workspace ID, root validata e sessione senza esporre root al modello;
- costruire context bundle con query, estimator e budget espliciti;
- renderizzare sezioni con delimitazione e provenance;
- implementare listing, read e search bounded su path logici;
- implementare write/patch con expected digest o contenuto precondizione;
- applicare containment e controllo symlink a ogni effetto;
- produrre action inspect/mutate con resource normalizzata;
- marcare context stale dopo mutazione;
- reindicizzare a checkpoint espliciti prima di nuovo retrieval;
- preservare ultimo snapshot valido su refresh fallito;
- verificare il percorso con il `WorkspaceProvider` Laravel generico.

## Invarianti

- il modello non fornisce né osserva root assolute;
- nessun path logico può uscire dalla root del workspace;
- una permission inspect non consente mutate;
- write/patch fallisce se la precondizione non coincide;
- risultati file e search rispettano limiti prima del model message;
- context bundle conserva generazione e provenance;
- mutazione non rende corrente uno snapshot stale;
- nessuna conoscenza Laravel entra in Agent Runtime o Tool Runtime.

## Test richiesti

- workspace assente, root invalida e ID differente;
- listing/read/search con limiti e ordine deterministico;
- traversal, path assoluti, symlink interno/esterno e file cambiato;
- patch valida, conflitto, deny e cancellazione;
- context disclosure deny prima della completion;
- mutation, stale marker, refresh e retrieval successivo;
- refresh fallito senza snapshot parziale;
- workspace Laravel end-to-end su directory temporanea;
- race detector e nessun processo/rete esterna.

## Gate di uscita

- reference agent legge e modifica soltanto risorse autorizzate;
- Context Engine resta fonte autorevole dell'evidenza indicizzata;
- stale e refresh hanno semantica dimostrata;
- percorso framework-neutral verificato con Laravel.

## Deliverable

- reference workspace tool set;
- context assembler dell'agente;
- integrazione workspace Laravel;
- `docs/agent-workspace.md`;
- report `docs/reports/milestone-7-phase-6.md`.

---

# Fase 7 — Integrazione, osservabilità e gate finale

## Obiettivo

Comporre Tool e Agent Runtime nel Runtime pubblico, integrarli con Gestor e
chiudere la milestone con hardening concorrente, privacy audit e uno scenario
autonomo riproducibile.

## Sviluppo previsto

- esporre `Runtime.Tools()` e `Runtime.Agents()` dal composition root;
- registrare reference tool e reference agent con configurazione esplicita;
- aggiungere source Gestor per descriptor agent e tool senza execution;
- invalidare Gestor su registrazioni senza discovery implicita;
- pubblicare eventi redatti di sessione, piano, turno, permission e tool;
- isolare errori e panic degli observer dopo il commit;
- aggiungere benchmark deterministico del loop e dei limiti;
- eseguire scenario end-to-end su fixture Laravel temporanea;
- completare audit API, privacy, concorrenza e dependency direction;
- allineare README, architecture, roadmap e MAESTRO_CONTEXT;
- produrre report conclusivo e matrice dei rischi rinviati.

## Invarianti

- composition root crea una sola istanza condivisa di ogni servizio;
- `pkg/runtime.Runtime` e lifecycle globale restano invariati;
- Gestor descrive ma non esegue agenti o tool;
- registrazione rende lo snapshot Gestor stale senza I/O;
- observer non ricevono prompt, contenuto, arguments o output;
- observer e sorgenti esterne non vengono invocati sotto lock;
- il reference agent non ottiene policy permissive implicite;
- il test live con un modello reale resta separato dal gate deterministico.

## Test richiesti

- composition root e isolamento tra due Runtime;
- resolution Gestor di agenti e tool con ambiguity esplicita;
- invalidazione e refresh dopo registrazione;
- eventi ordinati, cardinalità, redazione, handler lento/error/panic;
- run concorrenti, cancellazione e shutdown del chiamante;
- scenario autonomo fixture con piano, read, patch e terminale;
- benchmark ripetuto senza contenuti nei report;
- suite completa ripetuta, race detector, vet e diff check.

## Gate di uscita

- primo agente autonomo esegue il task fixture entro limiti e permessi;
- suite repository-wide e race detector verdi;
- `go vet ./...` e audit di compatibilità verdi;
- documentazione e report allineati;
- nessuna funzionalità fuori scope presentata come implementata.

## Deliverable

- wiring completo Tool/Agent/Gestor/Context/Provider;
- reference agent composto;
- benchmark e scenario autonomo deterministico;
- `docs/reports/milestone-7-phase-7.md`;
- `docs/reports/milestone-7-final.md`.

---

# Gate finale della milestone

La Milestone 7 è conclusa soltanto quando:

- contratti Tool e Agent hanno superato l'audit pubblico;
- ogni effetto attraversa prepare, authorize ed execute;
- default-deny, approval flow e model disclosure sono verificati;
- sessione, piano e loop hanno hard ceiling e terminale unico;
- tool calling non-stream e stream è correlato e bounded;
- workspace context e mutazioni rispettano containment e freshness;
- reference agent completa uno scenario reale riproducibile;
- Gestor scopre agenti e tool senza eseguirli;
- eventi e report sono redatti;
- suite, race detector, vet e audit documentale sono verdi.

Output atteso:

```text
Un primo agente autonomo, provider-agnostic, workspace-aware e governato da
permessi espliciti, composto nel Runtime pubblico di Maestro.
```

---

# Rischi e mitigazioni

| Rischio | Mitigazione |
|---|---|
| loop infinito o costi non limitati | hard ceiling obbligatori e terminali tipizzati |
| modello che tenta escalation | action normalizzate e authorizer fuori dal modello |
| overwrite concorrente | digest/precondizione e verifica al momento dell'effetto |
| output tool eccessivo | limiti per result e memoria prima del model message |
| duplicazione di registry | ownership separate e source Gestor read-only |
| context stale dopo mutazione | stale marker e refresh a checkpoint esplicito |
| leak in eventi | payload allowlist e test negativi repository-wide |
| falsa promessa di sandbox | confine trusted in-process dichiarato esplicitamente |
| dipendenza da un modello live | provider fittizi nel gate, smoke live separato |

---

# Fuori scope assegnato

Restano per la Milestone 8 o evoluzioni successive:

- packaging e trust di tool/agent di terze parti;
- sandbox e isolamento di processo;
- CLI interattiva completa per approval e session inspection;
- memoria persistente e recupero dopo restart;
- multi-agent e delega;
- shell, Git, Docker e tool framework completi;
- policy organizzative distribuite e secret manager;
- selezione automatica di provider/modello;
- remote execution e code durable.

# Maestro Provider Layer Completion Plan

Versione: 0.1.0

Stato: Concluso

Ultimo aggiornamento: 2026-08-08

Autori:
- Antonio Cafeo
- OpenAI ChatGPT

---

# Obiettivo

Questo documento scompone il lavoro necessario per chiudere la Milestone 2 —
Provider Layer dopo il completamento delle Fasi 1 e 2.

La milestone è conclusa quando Maestro dispone di contratti provider stabili,
due adapter locali verificati con test deterministici, gestione completa del
catalogo e della residenza dei modelli, capability interrogabili, semantica
uniforme degli errori, resilienza configurabile, osservabilità e una baseline
comune per tool calling e output strutturati.

L'aggiunta di altri adapter non è un requisito di chiusura: Ollama e llama.cpp
sono le due implementazioni concrete usate per validare i contratti. Un terzo
adapter viene introdotto soltanto se serve a dimostrare una nuova astrazione.

---

# Sequenza di lavoro

| Fase | Stato | Incremento | Dipendenza principale | Criterio sintetico di uscita |
|---|---|---|---|---|
| 3 | Conclusa | Model Acquisition & Removal | Discovery e lifecycle | Pull e rimozione opzionali, cancellabili e sicuri |
| 4 | Conclusa | Model Residency Policies | Lifecycle e acquisizione | Keep-alive e autoload espliciti, senza duplicare lo stato remoto |
| 5 | Conclusa | Capability Introspection | Superficie operativa completa | Supporto del provider e del modello interrogabile a runtime |
| 6 | Conclusa | Error Semantics | Tassonomia delle operazioni | Errori neutrali, classificati e compatibili con `errors.Is` |
| 7 | Conclusa | Resilience Policies | Errori classificati | Retry/backoff e circuit breaker opt-in e deterministici |
| 8 | Conclusa | Provider Observability | Confini operativi e resilienza | Segnali neutrali, senza contenuti sensibili né dipendenze SDK |
| 9 | Conclusa | Advanced Generation Baseline | Capability introspection | Opzioni comuni, output strutturati e tool calling sui due adapter |
| 10 | Conclusa | Hardening & Provider Handoff | Fasi 3–9 | Suite deterministica completa e scenari live consegnati alla Milestone 3 |

L'ordine rappresenta le dipendenze architetturali, non soltanto la priorità.
In particolare, le policy di resilienza non devono interpretare stringhe di
errore e i contratti avanzati non devono essere esposti prima che la loro
disponibilità possa essere dichiarata in modo affidabile.

---

# Fase 3 — Model Acquisition & Removal

## Obiettivo

Completare il ciclo di gestione del catalogo senza trasferire al Provider
Runtime la proprietà dei file o dello stato dei modelli.

## Scope

- Definire capability pubbliche opzionali per acquisizione e rimozione.
- Definire eventi di avanzamento pull-based per download, verifica e
  completamento, con chiusura esplicita delle risorse.
- Propagare `context.Context`, cancellazione e deadline durante il
  trasferimento.
- Stabilire semantica di idempotenza per modello già presente, trasferimento in
  corso e modello assente durante la rimozione.
- Implementare pull e delete tramite le API native di Ollama.
- Implementare pull, progresso e rimozione tramite gli endpoint documentati del
  router llama.cpp; processi single-model o versioni prive degli endpoint
  conservano l'errore remoto senza fallback sul filesystem.
- Impedire che l'assenza di un endpoint venga aggirata cancellando direttamente
  file scelti dall'utente o dal provider.

## Gate

- Contratti e ADR approvati.
- Test di progresso, cancellazione, errori intermedi e idempotenza.
- Implementazioni Ollama e llama.cpp coperte da server HTTP in-memory.
- Modalità e versioni del server che possono rifiutare gli endpoint documentate.
- Nessuna goroutine o risposta HTTP lasciata aperta.

---

# Fase 4 — Model Residency Policies

Stato: Conclusa

## Obiettivo

Rendere configurabili permanenza e caricamento su richiesta senza trasformare
il Provider Runtime in una seconda fonte di verità sullo stato dei modelli.

## Scope

- Definire una policy neutrale per comportamento predefinito, durata di
  residenza e autoload esplicito.
- Distinguere gli hint applicati a una singola richiesta dalle policy applicate
  dal runtime a più operazioni.
- Tradurre la policy nei meccanismi nativi: `keep_alive` per Ollama e capability
  load/unload del router llama.cpp.
- Eseguire autoload soltanto quando abilitato e prima di completion, stream o
  embedding compatibili con il modello richiesto.
- Coordinare richieste concorrenti sullo stesso modello evitando load duplicati
  senza memorizzare uno stato remoto autorevole.
- Eseguire unload a scadenza soltanto per residenze possedute dalla policy,
  senza interferire con modelli caricati esternamente.
- Propagare context e shutdown del runtime a timer e operazioni in corso.

## Gate

- Comportamento di default invariato e nessun autoload implicito.
- Test deterministici con clock sostituibile, richieste concorrenti e shutdown.
- Ollama e llama.cpp rispettano la stessa semantica nei limiti dichiarati dalle
  rispettive capability.
- Discovery rimane la fonte osservabile dello stato effettivo del provider.
- Selezione hardware-aware e scelta automatica del modello restano fuori scope.

## Esito

- `ModelResidencyPolicy` configura per provider e model ID autoload, rilascio
  immediato, TTL o permanenza fino allo shutdown.
- Completion, stream ed embedding acquisiscono lease soltanto per policy
  esplicite; senza policy il percorso operativo rimane invariato.
- Discovery decide se sia necessario un load e Maestro scarica soltanto le
  residenze che ha caricato direttamente.
- Richieste concorrenti condividono la transizione; stream e shutdown
  conservano ownership e rilascio deterministici.
- Ollama usa i propri comandi `keep_alive`, llama.cpp usa le capability del
  router; la semantica comune rimane nel Provider Runtime.
- Clock, concorrenza e shutdown sono coperti da test isolati; gli scenari live
  restano assegnati alla Milestone 3.

---

# Fase 5 — Capability Introspection

Stato: Conclusa

## Obiettivo

Rendere interrogabile ciò che un adapter, una specifica istanza e un singolo
modello possono realmente eseguire.

## Scope

- Definire descrittori neutrali per capability del provider e del modello.
- Distinguere il supporto strutturale dell'adapter dalla disponibilità
  operativa dell'istanza configurata.
- Rappresentare almeno completion, streaming, embedding, discovery, load,
  unload, pull, remove, output strutturati e tool calling.
- Integrare l'introspection nel routing senza duplicare lo stato posseduto dagli
  adapter.
- Rendere l'output deterministico, thread-safe e compatibile con provider che
  possono cambiare catalogo durante l'esecuzione.
- Documentare quando una capability richiede I/O e quando può essere letta da
  configurazione o metadata già disponibili.

## Gate

- Ollama e llama.cpp espongono descrittori coerenti con configurazione e
  modalità del server.
- Le capability non supportate falliscono prima di inviare richieste remote.
- Test di concorrenza e variazione del catalogo superati.
- Nessuna selezione automatica del provider viene introdotta in questa fase.

## Esito

- `CapabilityInspector` e il routing `Capabilities` espongono report neutrali
  nei target adapter, instance e model.
- Supporto strutturale e disponibilità operativa sono distinti; `unknown` evita
  inferenze quando il protocollo non rende osservabile la configurazione.
- I report contengono tutte le capability note in ordine canonico e vengono
  validati dal Provider Runtime.
- Ollama usa il catalogo e `/api/show`; llama.cpp usa `/models`, stato router e
  argomenti del processo modello.
- Nessun report viene memorizzato: variazioni e concorrenza sono verificate con
  test deterministici e race detector.
- Capability non supportate, richieste non valide e report incoerenti falliscono
  prima di I/O operativo ulteriore; non è introdotta selezione automatica.

---

# Fase 6 — Error Semantics

Stato: Conclusa

## Obiettivo

Fornire alle policy e ai consumer una classificazione stabile degli errori,
indipendente dai payload e dagli status specifici dei provider.

## Scope

- Definire una tassonomia minima per richiesta non valida, autenticazione,
  modello o capability assente, indisponibilità, capacità esaurita, rate limit,
  errore transitorio e errore interno.
- Allegare, quando disponibili, operazione, provider, modello, status remoto e
  indicazione di ritentabilità.
- Preservare `context.Canceled`, `context.DeadlineExceeded`, wrapping e
  compatibilità con `errors.Is`/`errors.As`.
- Mappare in modo uniforme gli errori di Ollama e llama.cpp.
- Limitare dimensione e contenuto dei payload remoti inclusi negli errori.
- Separare la classificazione dalla decisione di effettuare un retry.

## Gate

- Matrice di mapping documentata per entrambi gli adapter.
- Test per status HTTP, errori di trasporto, context e payload malformati.
- Nessun consumer deve analizzare stringhe per prendere decisioni operative.
- Le API esistenti continuano a restituire errori idiomatici Go.

## Esito

- `ProviderError` espone kind, operazione, provider, modello, status,
  ritentabilità e dettagli strutturati remoti con causa preservata.
- Sentinel nuovi e preesistenti restano interrogabili con `errors.Is`; il
  dettaglio tipizzato è accessibile con `errors.As`.
- Ollama e llama.cpp classificano errori HTTP, trasporto, context, risposte
  invalide e fallimenti che emergono durante gli stream.
- llama.cpp conserva `type` e `code` OpenAI-like; Ollama applica la stessa
  tassonomia ai payload `error`, inclusi quelli mid-stream.
- I dettagli remoti sono normalizzati e limitati; nessun contenuto di richiesta
  viene incluso nell'envelope.
- `Retryable` è metadata conservativo: retry, backoff, idempotenza e circuit
  breaker restano responsabilità della Fase 7.
- Matrice, limiti e modalità d'uso sono descritti in
  `provider-error-semantics.md`; ADR-0013 registra la decisione.

---

# Fase 7 — Resilience Policies

Stato: Conclusa

## Obiettivo

Applicare retry, backoff e circuit breaker sopra gli adapter, con policy
esplicite e senza alterare il comportamento predefinito.

## Scope

- Introdurre policy opt-in con tentativi massimi, backoff, jitter e limiti
  temporali.
- Definire una matrice di idempotenza per ogni operazione provider.
- Consentire retry dello streaming soltanto prima che un chunk sia stato
  consegnato al consumer.
- Non ritentare automaticamente richieste invalide, cancellate o scadute.
- Definire circuit breaker per provider, capability e modello con stato e
  transizioni osservabili.
- Rendere clock, backoff e sorgente del jitter sostituibili nei test.
- Evitare callback, attese o I/O remoto mentre sono detenuti lock del registry
  o del lifecycle.

## Gate

- Test deterministici di apertura, half-open, recupero e concorrenza.
- Budget di retry sempre limitato dal context del chiamante.
- Comportamento senza policy invariato rispetto alle fasi precedenti.
- Fallback, load balancing e routing multi-provider restano fuori scope.

## Esito

- `ResiliencePolicy` configura per provider, operazione e modello opzionale
  tentativi, backoff, jitter, budget temporale e circuit breaker.
- Le policy model-specific hanno precedenza su quelle generali della stessa
  operazione; in assenza di policy il percorso rimane invariato.
- Retry e breaker usano esclusivamente `ProviderError.Retryable`, context e la
  matrice di ripetibilità documentata.
- Completion streaming può riaprire lo stream soltanto prima del primo chunk;
  pull e remove non vengono ritentati automaticamente.
- Il circuit breaker espone snapshot closed/open/half-open e limita i probe
  concorrenti senza mantenere lock durante I/O o attese.
- Discovery, load e unload invocati dalle policy di residenza compongono le
  rispettive policy di resilienza.
- Clock, attesa e jitter sostituibili coprono backoff, budget, apertura,
  half-open, recupero e concorrenza con test deterministici.
- Contratti, matrice e limiti sono descritti in `provider-resilience.md`;
  ADR-0014 registra la decisione.

---

# Fase 8 — Provider Observability

Stato: Conclusa

## Obiettivo

Rendere osservabili richieste e lifecycle provider senza legare il package a
un sistema di telemetria specifico.

## Scope

- Definire hook o observer neutrali per inizio, completamento, errore, retry e
  transizioni del circuit breaker.
- Coprire completion, stream, embedding, discovery, load, unload, pull e
  remove.
- Esporre durata, esito, tentativi, quantità d'uso disponibili e classificazione
  dell'errore.
- Vietare per default prompt, risposte, embedding, token di autenticazione e
  payload remoti nei segnali.
- Stabilire un solo punto di emissione per ogni confine operativo, evitando
  duplicazioni tra facade, runtime e adapter.
- Rendere gli observer non bloccanti rispetto agli invarianti provider: un loro
  errore non modifica il risultato dell'operazione osservata.
- Fornire adapter opzionali verso logging, metriche e tracing senza introdurre
  una dipendenza obbligatoria da SDK esterni.

## Gate

- Test di cardinalità, redazione, ordering e concorrenza.
- Stream cancellati o chiusi anticipatamente producono un esito terminale.
- Retry e circuit breaker sono correlabili alla richiesta originaria.
- Il percorso senza observer non introduce allocazioni o lavoro significativo
  non necessario.

## Esito

- `ProviderObserver`, `ProviderObserverFunc` e `ProviderEvent` formano un
  contratto pubblico neutrale e privo di dipendenze da SDK telemetrici.
- Start, tentativi, retry, transizioni del circuito e terminale condividono un
  `OperationID` locale e un ordine deterministico per ogni operazione.
- Completion, stream, embedding, listing, discovery, lifecycle, acquisition,
  removal e introspection sono osservati al solo confine pubblico del Runtime.
- Gli stream producono un unico terminale per EOF, errore, cancellazione, stage
  pull completato o chiusura anticipata, conservando usage e progresso noti.
- Error kind, status e ritentabilità sono esposti senza messaggi remoti; prompt,
  risposte, chunk, embedding, credenziali e payload non appartengono al tipo
  evento.
- Gli observer sono sostituibili in concorrenza, vengono invocati senza lock
  interni e non possono modificare il risultato tramite errori o panic.
- Il fast path disabilitato non crea tracker, eventi o operation ID e una
  verifica di allocazione ne protegge il comportamento.
- Contratto, ordering, cardinalità e limiti sono descritti in
  `provider-observability.md`; ADR-0015 registra la decisione.

---

# Fase 9 — Advanced Generation Baseline

Stato: Conclusa

## Obiettivo

Stabilire la superficie comune necessaria al futuro Agent System, evitando di
trasferire nel core tutte le opzioni proprietarie dei provider.

## Scope

- Definire opzioni comuni di generazione: limite token, temperatura, `top_p` e
  sequenze di stop.
- Definire output strutturato con modalità JSON e schema quando supportato.
- Definire tool, tool choice, tool call e tool result nei messaggi neutrali.
- Estendere lo streaming con eventi incrementali di tool call e risultato
  finale coerente.
- Dichiarare supporto e limitazioni attraverso la capability introspection.
- Implementare e testare la baseline sia su Ollama sia su llama.cpp.
- Conservare compatibilità con le richieste e gli stream semplici già esistenti.

## Gate

- ADR del contratto pubblico e strategia di compatibilità completati.
- Stesso scenario di output strutturato e tool calling eseguito sui due adapter.
- Validazione locale prima della richiesta per combinazioni non supportate.
- Chunk terminali, usage ed errori hanno semantica uniforme.

Multimodalità, audio, reasoning/thinking, prompt caching, speculative decoding e
opzioni di sampling esclusivamente proprietarie restano fuori dalla milestone.

## Esito

- `CompletionRequest` espone opzioni comuni di sampling, output JSON o JSON
  Schema, definizioni tool e tool choice senza introdurre tipi specifici degli
  adapter.
- I messaggi neutrali rappresentano chiamate e risultati tool; gli stream
  espongono delta incrementali con indici stabili.
- Il Runtime valida richieste e combinazioni incompatibili prima del routing e
  dell'I/O remoto.
- Ollama traduce la superficie sulle API native e supporta tool choice `auto` e
  `none`; llama.cpp traduce la superficie Chat Completions e supporta anche
  `required` e named choice quando il server usa Jinja e un template compatibile.
- JSON terminale e chiamate tool ricostruite dagli stream vengono validati
  prima del completamento.
- Capability introspection, fixture HTTP, test di compatibilità e ADR-0016
  coprono entrambi gli adapter. Limiti e contratto sono descritti in
  `provider-advanced-generation.md`.

---

# Fase 10 — Hardening & Provider Handoff

Stato: Conclusa

## Obiettivo

Chiudere la Provider Layer con verifiche deterministiche e consegnare al
Benchmark Layer una matrice esplicita degli scenari live da eseguire.

## Scope

- Eseguire la suite completa, race detector, vet e compilare le suite protette
  da tag di integrazione senza richiedere servizi live.
- Verificare cancellazione, deadline, chiusura degli stream e assenza di leak.
- Verificare la matrice di capability su modello chat, modello embedding e
  configurazioni single-model e router attraverso fixture HTTP isolate.
- Verificare pull/remove e load/unload attraverso protocolli simulati e
  deterministici.
- Eseguire scenari controllati di error mapping, retry e circuit breaker.
- Verificare redazione e completezza dei segnali di osservabilità.
- Allineare esempi, ADR, documentazione degli adapter, roadmap e contesto di
  progetto.
- Definire il manifest degli smoke benchmark, inclusi capability richieste,
  modelli fixture, cleanup e variabili di configurazione.
- Registrare esplicitamente le limitazioni dipendenti dalle versioni dei server
  che il Benchmark Layer dovrà riportare.

## Gate finale

La Milestone 2 è conclusa soltanto quando:

- tutti i test deterministici obbligatori sono verdi;
- le API pubbliche sono state sottoposte ad audit di compatibilità;
- limiti, capability non supportate e requisiti operativi sono documentati;
- gli scenari live rinviati hanno un caso corrispondente nel piano della
  Milestone 3;
- non restano attività obbligatorie senza owner o fase di destinazione.

Processi, hardware e modelli locali non sono prerequisiti del gate della
Provider Layer. Gli smoke test live sono responsabilità del Livello 1 — Smoke
Benchmark della Milestone 3.

## Esito

- Suite completa, race detector, vet e compilazione delle suite protette dal
  tag `integration` costituiscono il gate ripetibile senza servizi live.
- `provider-api-compatibility-audit.md` registra modifiche additive, vincoli sui
  composite literal e rischi residui con destinazione esplicita.
- `provider-smoke-benchmark-manifest.yaml` consegna alla Milestone 3 provider,
  ruoli dei modelli, variabili d'ambiente, cleanup, protezioni delle mutazioni,
  redazione e stati di risultato.
- Roadmap, contesto, ADR e documentazione degli adapter sono allineati alla
  superficie implementata.
- Tutte le attività live o dipendenti da versione, modelli e hardware hanno un
  owner nel Benchmark & Evaluation Layer; non restano requisiti obbligatori
  senza destinazione.

---

# Fuori scope della Milestone 2

- Fallback e bilanciamento automatico tra provider.
- Selezione automatica del modello in base all'hardware.
- Avvio, arresto o supervisione dei processi server locali.
- Persistenza autonoma del catalogo o duplicazione dello stato dei provider.
- Gestione centralizzata dei secret.
- Nuovi adapter non necessari a validare i contratti.
- Multimodalità, reasoning e opzioni avanzate proprietarie.

Questi temi potranno essere assegnati a Gestor, Context Engine, Agent System o
a successive evoluzioni della Provider Layer senza riaprire la Milestone 2.

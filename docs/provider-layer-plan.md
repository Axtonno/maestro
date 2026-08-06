# Maestro Provider Layer Completion Plan

Versione: 0.1.0

Stato: Pianificato

Ultimo aggiornamento: 2026-08-06

Autori:
- Antonio Cafeo
- OpenAI ChatGPT

---

# Obiettivo

Questo documento scompone il lavoro necessario per chiudere la Milestone 2 —
Provider Layer dopo il completamento delle Fasi 1 e 2.

La milestone è conclusa quando Maestro dispone di contratti provider stabili,
due adapter locali verificati, gestione completa del catalogo e della residenza
dei modelli, capability interrogabili, semantica uniforme degli errori,
resilienza configurabile, osservabilità e una baseline comune per tool calling
e output strutturati.

L'aggiunta di altri adapter non è un requisito di chiusura: Ollama e llama.cpp
sono le due implementazioni concrete usate per validare i contratti. Un terzo
adapter viene introdotto soltanto se serve a dimostrare una nuova astrazione.

---

# Sequenza di lavoro

| Fase | Incremento | Dipendenza principale | Criterio sintetico di uscita |
|---|---|---|---|
| 3 | Model Acquisition & Removal | Discovery e lifecycle | Pull e rimozione opzionali, cancellabili e sicuri |
| 4 | Model Residency Policies | Lifecycle e acquisizione | Keep-alive e autoload espliciti, senza duplicare lo stato remoto |
| 5 | Capability Introspection | Superficie operativa completa | Supporto del provider e del modello interrogabile a runtime |
| 6 | Error Semantics | Tassonomia delle operazioni | Errori neutrali, classificati e compatibili con `errors.Is` |
| 7 | Resilience Policies | Errori classificati | Retry/backoff e circuit breaker opt-in e deterministici |
| 8 | Provider Observability | Confini operativi e resilienza | Segnali neutrali, senza contenuti sensibili né dipendenze SDK |
| 9 | Advanced Generation Baseline | Capability introspection | Opzioni comuni, output strutturati e tool calling sui due adapter |
| 10 | Hardening & Milestone Gate | Fasi 3–9 | Matrice isolata e live completata, documentazione allineata |

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
- Implementare la capability llama.cpp soltanto dove il router offre un'API
  stabile e sicura; in caso contrario restituire capability non supportata.
- Impedire che l'assenza di un endpoint venga aggirata cancellando direttamente
  file scelti dall'utente o dal provider.

## Gate

- Contratti e ADR approvati.
- Test di progresso, cancellazione, errori intermedi e idempotenza.
- Implementazione Ollama coperta da server HTTP in-memory.
- Comportamento llama.cpp supportato oppure esplicitamente dichiarato assente.
- Nessuna goroutine o risposta HTTP lasciata aperta.

---

# Fase 4 — Model Residency Policies

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

---

# Fase 5 — Capability Introspection

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

---

# Fase 6 — Error Semantics

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

---

# Fase 7 — Resilience Policies

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

---

# Fase 8 — Provider Observability

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

---

# Fase 9 — Advanced Generation Baseline

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

---

# Fase 10 — Hardening & Milestone Gate

## Obiettivo

Chiudere la milestone con una verifica integrata, inclusi gli smoke test live
rinviati dalle fasi precedenti.

## Scope

- Eseguire la suite completa, race detector, vet e test di integrazione
  disponibili.
- Eseguire smoke test live contro Ollama, `llama-server` e llama.cpp router mode.
- Verificare cancellazione, deadline, chiusura degli stream e assenza di leak.
- Verificare la matrice di capability su modello chat, modello embedding e
  configurazioni single-model e router.
- Verificare pull/remove dove supportati e load/unload con modelli dedicati.
- Eseguire scenari controllati di error mapping, retry e circuit breaker.
- Verificare redazione e completezza dei segnali di osservabilità.
- Allineare esempi, ADR, documentazione degli adapter, roadmap e contesto di
  progetto.
- Registrare esplicitamente eventuali limitazioni dipendenti dalle versioni dei
  server usate nel gate.

## Gate finale

La Milestone 2 è conclusa soltanto quando:

- tutti i test isolati e di integrazione obbligatori sono verdi;
- la matrice live Ollama/llama.cpp è stata eseguita e documentata;
- nessuno smoke test delle Fasi 1–9 resta rinviato;
- le API pubbliche sono state sottoposte ad audit di compatibilità;
- limiti, capability non supportate e requisiti operativi sono documentati;
- non restano attività obbligatorie senza owner o fase di destinazione.

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

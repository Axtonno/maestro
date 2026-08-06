# Maestro Roadmap

Versione: 0.1.0

Stato: Living Document

Ultimo aggiornamento: 2026-08-06

Autori:
- Antonio Cafeo
- OpenAI ChatGPT

---

# Perché esiste questo documento?

Questo documento descrive l'evoluzione prevista del progetto Maestro.

Non rappresenta una pianificazione rigida.

La roadmap definisce la direzione del progetto e le principali milestone, lasciando spazio a revisioni e adattamenti.

---

# Obiettivo

Costruire un runtime locale modulare, estensibile e provider-agnostic per lo sviluppo software assistito da intelligenza artificiale.

---

# Milestone 0 — Fondamenta

Stato: Conclusa

Obiettivi:

- Definizione dell'identità del progetto.
- Definizione della filosofia.
- Definizione dei principi.
- Definizione della visione.
- Definizione dell'architettura.
- Inizializzazione del repository Go.
- Struttura iniziale del runtime.

---

# Milestone 1 — Runtime Core

Stato: In corso

Obiettivi:

- Bootstrap del runtime.
- Sistema di configurazione.
- Dependency Injection.
- Event Bus.
- Logging.
- Lifecycle del runtime.
- Gestione del workspace.

Output atteso:

Un runtime funzionante senza alcun provider.

---

# Fase 5 — Provider Runtime/Configuration

Stato: In corso — smoke test live pendente

Scope:

- Contratti provider.
- Capability operative.
- Registry e provider predefinito.
- Routing e configurazione.
- Integrazione nel Runtime.
- Primo adapter concreto Ollama.
- Test e documentazione.
- Smoke test contro un'istanza Ollama reale.

L'implementazione e i test isolati sono completati. La fase viene chiusa quando
lo smoke test conferma listing, completion, streaming, embedding e
cancellazione senza incompatibilità di protocollo.

Ulteriori adapter e policy non appartengono al gate di chiusura della Fase 5.

---

# Fase 6 — Plugin Runtime

Stato: Conclusa

Scope completato:

- Contratto pubblico `Plugin` basato su `runtime.Component`.
- Manifest e compatibilità dell'API Plugin Runtime.
- Registry plugin thread-safe.
- Registrazione coordinata con il Runtime Core.
- Riutilizzo di dependency graph, stato e lifecycle globali.
- Catalogo thread-safe di loader in-process.
- Discovery deterministica dei loader e dei plugin registrati.
- Caricamento cancellabile con validazione di ID e risultato.
- Eventi di catalogo, registrazione e caricamento.
- Esposizione tramite il composition root `maestro.New`.
- Primo plugin framework-aware Laravel con detection e health.
- Test e documentazione.

Evoluzione successiva, fuori dal gate della Fase 6:

- Packaging e installazione di plugin esterni.
- Firme e policy di trust.
- Process isolation o sandbox.
- Unload e hot replacement.
- Capability Laravel avanzate.

Il gate si chiude sul modello trusted in-process definito da ADR-0008. Formati
di distribuzione esterni potranno alimentare lo stesso contratto `Loader` senza
modificare registry e lifecycle.

---

# Milestone 2 — Provider Layer

Stato: Evoluzione incrementale

## Fase 1 — Adapter llama.cpp

Stato: Implementazione completata — smoke test live pendente

Scope:

- Facade pubblica e configurazione tipizzata.
- Completion e streaming SSE tramite Chat Completions API.
- Embedding tramite API compatibile OpenAI.
- Model listing del modello caricato.
- Autenticazione Bearer opzionale.
- Validazione, error handling e propagazione del context.
- Test HTTP in-memory e smoke test live opzionale.
- Documentazione dell'adapter.

Il primo incremento usa la superficie compatibile con OpenAI esposta da
`llama-server`. Lifecycle del processo, gestione dei modelli e policy di
resilienza restano fasi successive della Provider Layer.

Facade, adapter, test isolati e documentazione sono completati. Il gate della
fase rimane aperto esclusivamente per lo smoke test contro un processo
`llama-server` reale.

## Evoluzione successiva

Obiettivi:

- Eventuali adapter per altri runtime locali.
- Model discovery avanzata.
- Download, pull e rimozione dei modelli.
- Gestione del ciclo di vita dei modelli.
- Keep-alive e unload.
- Retry e backoff.
- Circuit breaker.
- Metriche e tracing.
- Normalizzazione avanzata degli errori.
- Capability detection dinamica.
- Eventuale supporto a tool calling e output strutturati.

Output atteso:

Una Provider Layer capace di evolvere indipendentemente dalla progressione del
Runtime Core e del Plugin Runtime.

---

# Milestone 3 — Gestor

Obiettivi:

- Registry delle capability.
- Discovery dei componenti.
- Dependency graph.
- Risoluzione delle capability.

Output atteso:

Sistema modulare basato sulle capability.

---

# Milestone 4 — Plugin System

Stato: Baseline completata nella Fase 6; ecosistema in evoluzione

Obiettivi:

- Caricamento plugin.
- Registrazione plugin.
- Lifecycle plugin.
- API pubbliche.

Primo plugin:

Laravel (detection e health implementati).

---

# Milestone 5 — Context Engine

Obiettivi:

- Workspace indexing.
- Analisi AST.
- Context Builder.
- Ottimizzazione token.
- Cache.

Output atteso:

Costruzione intelligente del contesto.

---

# Milestone 6 — Agent System

Obiettivi:

- Pianificazione.
- Task execution.
- Tool calling.
- Permission model.
- Workspace awareness.

Output atteso:

Primo agente autonomo.

---

# Milestone 7 — Ecosistema

Obiettivi:

- Plugin di terze parti.
- CLI completa.
- API pubbliche.
- SDK.
- Documentazione.

---

# Principio della roadmap

La roadmap rappresenta una direzione.

L'ordine delle implementazioni può cambiare se emergono nuove esigenze o migliori soluzioni architetturali.

---

# Decisioni

- Le milestone rappresentano capacità del sistema, non versioni software.
- Nessuna milestone verrà considerata completata senza documentazione e test.

---

# Documenti dipendenti

- architecture.md
- MAESTRO_CONTEXT.md

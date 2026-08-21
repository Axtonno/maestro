# Milestone 2 — Provider Layer Final Report

Stato: Completata

Data di completamento: 2026-08-08

Natura del documento: ricostruzione retrospettiva

---

# Risultato

Maestro dispone di una Provider Layer provider-agnostic con due adapter locali,
gestione completa del catalogo e della residenza dei modelli, capability
interrogabili, errori neutrali, resilienza opt-in, osservabilità redatta e una
baseline comune per output strutturati e tool calling.

La milestone evolve indipendentemente dal Runtime Core: il core orchestra
componenti e lifecycle, mentre il Provider Runtime possiede registry, routing e
policy provider senza duplicare lo stato autorevole dei server.

---

# Fasi completate

| Fase | Ambito | Esito |
|---|---|---|
| 1 | Adapter llama.cpp | Completata |
| 2 | Model Discovery & Lifecycle | Completata |
| 3 | Model Acquisition & Removal | Completata |
| 4 | Model Residency Policies | Completata |
| 5 | Capability Introspection | Completata |
| 6 | Error Semantics | Completata |
| 7 | Resilience Policies | Completata |
| 8 | Provider Observability | Completata |
| 9 | Advanced Generation Baseline | Completata |
| 10 | Hardening & Provider Handoff | Completata |

I report retrospettivi dettagliati sono disponibili come
`milestone-2-phase-1.md`–`milestone-2-phase-10.md`.

La numerazione canonica contiene dieci fasi. Il commit storico conclusivo
`866e269` era etichettato “9/9” perché Fase 9 e hardening/handoff furono
consegnati nello stesso incremento.

---

# Contratti e ownership finali

- Gli adapter dichiarano identità e capability implementate.
- Il Provider Runtime possiede registry, default esplicito e routing.
- I server provider possiedono catalogo, file e stato effettivo dei modelli.
- Discovery osserva lo stato remoto; Maestro non lo replica come fonte
  autorevole.
- Residency coordina soltanto le residenze caricate da Maestro.
- Introspection distingue supporto strutturale e availability operativa.
- Error semantics classifica senza decidere retry.
- Resilience applica policy opt-in sulla matrice di ripetibilità.
- Observability pubblica eventi neutrali senza contenuti sensibili.

---

# Adapter qualificati deterministicamente

## Ollama

- completion e streaming;
- embedding e catalogo;
- load/unload con `keep_alive`;
- pull e delete;
- introspection tramite catalogo e `/api/show`;
- structured output e tool calling nei limiti dichiarati.

## llama.cpp

- Chat Completions e streaming SSE;
- embedding e model listing;
- discovery e lifecycle in router mode;
- acquisition e removal tramite endpoint documentati del router;
- introspection di modalità e argomenti;
- structured output e tool calling quando template e configurazione lo
  consentono.

---

# Gate tecnico storico

Il gate finale registrato comprende:

```text
GOCACHE=/tmp/maestro-go-build go test ./...
GOCACHE=/tmp/maestro-go-build go test -race ./...
GOCACHE=/tmp/maestro-go-build go vet ./...
```

Sono inoltre documentati:

- audit di compatibilità delle API pubbliche;
- compilazione delle suite di integrazione senza servizi live;
- adapter verificati tramite trasporti HTTP in-memory;
- routing e capability coperti deterministicamente;
- cancellazione, stream, cleanup e concorrenza;
- lifecycle, pull/remove, introspection, resilienza e osservabilità;
- output strutturati e tool calling su entrambi gli adapter;
- manifest completo degli scenari live consegnato alla Milestone 3.

Esito: gate superato.

---

# Limiti dichiarati

- nessun fallback o bilanciamento automatico tra provider;
- nessuna selezione hardware-aware del modello;
- nessuna supervisione dei processi server locali;
- nessuna persistenza autonoma del catalogo;
- nessuna gestione centralizzata dei secret;
- nessun nuovo adapter non necessario a validare i contratti;
- multimodalità, audio, reasoning e opzioni proprietarie fuori scope;
- la matrice live non fa parte del gate deterministico della milestone.

---

# Handoff alla Milestone 3

`provider-smoke-benchmark-manifest.yaml` trasferisce al Benchmark & Evaluation
Layer gli scenari live per completion, streaming, embedding, lifecycle,
acquisition, cancellazione, resilienza e introspection. Il manifest specifica
configurazione, ruoli dei modelli, protezioni delle mutazioni, cleanup, redazione
e stati di risultato.

---

# Conclusione

La Milestone 2 — Provider Layer è completata. Il sistema dispone di contratti
provider stabili e di una verifica deterministica completa; le evidenze live
dipendenti dalla macchina diventano responsabilità esplicita della Milestone 3.

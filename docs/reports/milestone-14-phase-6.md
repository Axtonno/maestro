# Milestone 14 — Phase 6 Report

Data: 2026-08-28

Stato: **COMPLETATA**

## Esito audit

Implementazione, ADR, CLI, configurazione, security model e compatibility
descrivono lo stesso confine: direct chat single-file esplicito, zero tool,
zero retrieval/state machine/sessione e zero fallback; verified agent separato
e alias `run` invariato.

## Audit del confine

| Area | Evidenza | Esito |
|---|---|---|
| dipendenze servizio | import diretti: product config, path identifier e provider | PASS |
| autorità chat | nessun Agent/Tool Runtime, Gestor, approver o mutation policy | PASS |
| provider request | una generation, zero tool, `tool_choice: none` | PASS |
| file disclosure | un solo path logico confinato e bounded | PASS |
| terminali | reason sintetici, deadline/cancel distinti, failure atomici | PASS |
| streaming | opt-in profilo+flag, equivalenza deterministica | PASS |
| agent/run | stesso percorso e regressione CLI verde | PASS |
| configurazione v1 | compatibilità v0.2.0 invariata | PASS |
| configurazione v2 | development-only, strict e senza conversione implicita | PASS |
| mutation claim | nessun nuovo tool o support claim mutativo | PASS |
| release claim | nessun tag, release o artifact pubblico | PASS |

Il package `internal/directchat` importa `pkg/contextengine` soltanto per il
tipo/validatore di path logico; non costruisce né invoca il Context Engine.
Il composition root CLI riusa la factory dell'adapter provider, non
`application.Build` né il grafo agentico.

## Consolidamento evidenze

- F1: ADR-0033, nomi CLI e compatibility congelati;
- F2: profili v2 e controlli provider-neutral `num_ctx`/thinking;
- F3: servizio e comando direct chat non-streaming;
- F4: matrice negativa, equivalenza streaming, immutabilità e anti-leak verdi;
- F5: provider assente al preflight, C0-C4 `not_run`, stop rule rispettata.

La matrice deterministica valida il contratto ma non sostituisce la qualità
live. L'assenza del provider non viene attribuita al modello e non viene
trasformata in un PASS.

## Handoff Milestone 15

Milestone 15 riceve:

- il profilo versionato `configs/maestro.milestone-14-candidate.yaml`;
- fixture, file, domande e ground truth congelati nel candidate record;
- la matrice deterministica `docs/direct-chat-deterministic-matrix.md`;
- l'obbligo di osservare provider/GPU, digest, template e capability prima di
  creare un nuovo candidate ID;
- C0-C4 completi come primo gate live, prima del verified agent sintetico.

Il record M14 è un seed ripetibile, non un candidato promosso. M15 non può
ereditare un PASS live inesistente.

## Verifiche finali

| Controllo | Esito |
|---|---|
| `go test ./...` | PASS |
| `go test -race ./...` | PASS |
| `go vet ./...` | PASS |
| `git diff --check` | PASS |
| scan path/canary nei report M14 | PASS |
| profilo candidato strict | PASS |

## Gate

**PASS.** L'audit non rileva blocker non classificati. La milestone può essere
chiusa con esito `direct_chat_deferred` senza cambiare il support claim v0.2.0.

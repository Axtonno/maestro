# Milestone 7 — Phase 1 Report

Stato: Completata

Data: 2026-08-11

---

# Risultato

La Fase 1 introduce i contratti pubblici separati di Tool System e Agent System
senza modificare le API esistenti. ADR-0025 chiude ownership, permit boundary,
permission semantics, sessioni, terminal races, limiti e allowlist degli
eventi.

---

# Contratti consegnati

Package:

```text
pkg/tool
pkg/agent
```

`pkg/tool` fornisce:

- identità e descriptor tool versionati;
- JSON Schema e arguments canonicalizzati con copie difensive;
- invocation e prepared invocation content-bound;
- action tipizzate e permission request atomiche;
- richiesta modello distinta dalla tool invocation;
- disclosure manifest redatto;
- decision, approval, grant scope e deny disposition;
- Tool, Policy, Approver, Catalog e Runtime;
- execution limits e result tipizzati;
- sentinel ed `ExecutionError` ispezionabile;
- event payload con allowlist esatta.

`pkg/agent` fornisce:

- identità, descriptor e capability agente;
- run request con target, policy, contesto, tool e limiti espliciti;
- step e piani aciclici con transizioni validate;
- planning request e capability Planner;
- session snapshot immutabili e generazionali;
- contatori, stale bit e terminal reason;
- precedenza deterministica dei terminali;
- run result, Runtime e `RunError`;
- event payload con allowlist esatta.

---

# Decisioni chiuse dalle note di sviluppo

| Punto | Decisione |
|---|---|
| tool e modello | stesso permission model, subject distinti |
| disclosure | bundle locale -> manifest/fingerprint -> authorize -> provider |
| autorità | `Decision` non è permit; `Runtime.Invoke` non la accetta |
| prepared invocation | lega tool/version/call/run/arguments/action/fingerprint |
| più action | autorizzazione atomica, nessuna esecuzione parziale |
| grant | one-shot atomico o run-scoped su run+fingerprint |
| policy | registry del Tool Runtime, policy ID esatto |
| sessione | coordinatore unico futuro, snapshot concorrenti read-only |
| terminal race | precedenza totale pre-commit e terminale unico |
| mutate ambiguo | stale dal momento in cui l'esecuzione parte |
| deny | recoverable o terminal deciso dalla policy |
| eventi | allowlist strutturale esatta |

Il permit operativo resta interno. La Fase 2 implementerà issuer/verifier e
consumo obbligatorio nell'executor usando una fixture deterministica; la Fase 3
collegherà policy, Approver e grant reali.

---

# Invarianti verificati

- `pkg/tool` non importa `pkg/agent`;
- nessun package nuovo importa `internal/*`;
- `pkg/runtime.Runtime` e il composition root sono invariati;
- descriptor, JSON, action, plan e snapshot sono difensivi;
- trailing JSON non viene accettato;
- fingerprint cambia con identità, run, call, arguments o action;
- model permission non contiene prepared tool invocation;
- `ExecutionRequest` non accetta Decision, Approval o permit;
- policy e tool/agent typed nil vengono rifiutati;
- provider, modello, workspace, policy, agente e tool sono espliciti;
- i limiti sono positivi, coerenti e bounded;
- i piani rifiutano duplicati, riferimenti mancanti e cicli;
- sessione completed richiede step completed/skipped;
- terminal precedence è deterministica;
- gli envelope preservano kind e cause con `errors.Is`.

---

# Ownership

| Livello | Possiede | Implementazione |
|---|---|---|
| Tool | prepare semantico ed effetto trusted | Fase 2/6 |
| Tool Runtime | catalogo, policy registry, permit, executor | Fasi 2–3 |
| Policy/Approver | decisione e approvazione | Fase 3 |
| Agent Runtime | sessione, piano, budget e loop | Fasi 4–5 |
| Context Engine | bundle e generazione workspace | già autorevole |
| Provider Runtime | chiamate modello | già autorevole |
| Gestor | descriptor e resolution | Fase 7 |

---

# Compatibilità

Nessuna modifica è stata applicata a Runtime, Provider, Context Engine,
Gestor, Plugin, Laravel o composition root. I package sono additivi.

L'audit completo è in `docs/agent-system-api-compatibility-audit.md`.

---

# Test

Coperti nei package pubblici:

- assertion di compilazione delle interfacce;
- ID, versioni, capability e descriptor;
- JSON canonicale, trailing data e copie difensive;
- prepared invocation e permission fingerprint;
- tool/model permission separation;
- decision, approval, disposition e scope;
- typed nil per tool, policy, agent e Approver;
- result ed error semantics;
- request agent e hard limits;
- plan DAG, transizioni e snapshot;
- terminal precedence e run result.

Comandi del gate:

```text
GOCACHE=/tmp/maestro-go-build go test ./pkg/tool ./pkg/agent
GOCACHE=/tmp/maestro-go-build go test ./...
GOCACHE=/tmp/maestro-go-build go test -race ./pkg/tool ./pkg/agent
GOCACHE=/tmp/maestro-go-build go vet ./...
```

Esito: tutti i comandi superati. Anche la suite completa con race detector e
la ripetizione venti volte dei test `pkg/tool` e `pkg/agent` sono verdi.

---

# Decisioni rinviate

- catalogo, issuer/verifier ed executor: Fase 2;
- policy matcher, Approver flow e grant consumption: Fase 3;
- session coordinator e planner model-backed: Fase 4;
- loop provider/tool e streaming: Fase 5;
- reference workspace tool e stale refresh: Fase 6;
- wiring, Gestor, eventi runtime e hardening: Fase 7.

---

# Gate

ADR-0025 è Accepted, ownership e failure matrix sono documentate, l'API è
additiva e i contratti pubblici sono coperti da test. La Fase 1 è completata;
la Fase 2 può iniziare.

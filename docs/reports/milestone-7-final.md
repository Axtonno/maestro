# Milestone 7 — Agent System Final Report

Stato: Completata

Data: 2026-08-12

---

# Esito finale

La Milestone 7 consegna Tool System e Agent System completi nella baseline
trusted in-process di Maestro. Un reference agent pianifica, costruisce
contesto, chiama un modello tramite Provider Runtime, usa tool autorizzati e
aggiorna un workspace entro limiti locali verificabili.

# Fasi completate

1. contratti, ownership e ADR-0025;
2. Tool catalog ed execution boundary;
3. permission model e approval flow;
4. sessioni, piani e budget;
5. loop agentico e tool calling;
6. workspace awareness e reference tool;
7. integrazione, osservabilità e hardening.

# Gate finale

- ogni effetto attraversa Prepare, authorizer e permit interno;
- default-deny, prompt e grant scope sono verificati;
- model invocation e workspace disclosure sono permission distinte;
- sessione, piano, streaming e loop hanno hard ceiling;
- tool set, call ID, arguments e result restano correlati;
- workspace containment, symlink policy e digest precondition sono verificati;
- mutation stale/refresh conserva la fonte autorevole del Context Engine;
- Gestor scopre agenti e tool senza esecuzione;
- eventi e benchmark non contengono payload sensibili;
- suite completa, race detector, test ripetuti, benchmark, vet e diff check
  sono verdi.

# Compatibilità

`pkg/runtime.Runtime`, lifecycle, Provider Runtime, Context Engine e Plugin
Runtime restano invariati. `maestro.Runtime` aggiunge `Tools()` e `Agents()`;
Gestor aggiunge target/scopi e capability agent/tool. L'audit finale è in
`docs/agent-system-api-compatibility-audit.md`.

# Rischi rinviati

- tool/agent di terze parti e trust packaging;
- sandbox o isolamento di processo;
- memoria persistente e recovery;
- multi-agent e delega;
- CLI interattiva completa;
- shell, Git, Docker e tool framework estesi;
- remote execution, durable runs e secret manager.

Questi elementi non sono presentati come parte della Milestone 7.

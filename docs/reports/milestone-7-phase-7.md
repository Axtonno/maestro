# Milestone 7 — Phase 7 Report

Stato: Completata

Data: 2026-08-12

---

# Risultato

La Fase 7 compone Tool e Agent Runtime nel Runtime pubblico, aggiunge sorgenti
Gestor ed eventi redatti e dimostra un percorso autonomo read/patch/final su un
workspace temporaneo.

# Capacità consegnate

- `Runtime.Tools()` e `Runtime.Agents()` additive;
- una sola istanza condivisa di provider, context, tool, agent e Event Bus;
- cinque reference workspace tool registrati;
- reference agent con planning deterministico e loop provider-neutral;
- sorgenti Gestor read-only `agent.catalog` e `tool.catalog`;
- target/scopi Gestor agent/tool e capability note;
- invalidazione Gestor su nuove registrazioni agent/tool;
- eventi session, piano, step, turn, permission e invocation;
- payload exact-allowlist senza contenuti sensibili;
- benchmark deterministico del loop;
- scenario fixture con read, patch, refresh e terminale completed;
- audit API e architettura aggiornati.

# Invarianti verificati

- due composition root non condividono cataloghi o sessioni;
- Gestor descrive agenti/tool senza eseguirli;
- più tool con la stessa capability restano ambiguous;
- registrazione invalida lo snapshot senza discovery implicita;
- reference runtime non registra policy permissive;
- observer ricevono stato già committato e non alterano l'operazione;
- prompt, path, arguments, contenuto e output non entrano negli eventi;
- scenario autonomo termina entro budget e permission esplicita.

# Test e gate

```text
GOCACHE=/tmp/maestro-go-build go test ./...
GOCACHE=/tmp/maestro-go-build go test -race ./...
GOCACHE=/tmp/maestro-go-build go test -count=3 ./...
GOCACHE=/tmp/maestro-go-build go test -run '^$' -bench BenchmarkAgentLoopDeterministic -benchtime=100x ./internal/agent
GOCACHE=/tmp/maestro-go-build go vet ./...
git diff --check
```

# Gate

Il primo agente autonomo è composto, provider-agnostic, workspace-aware,
bounded e governato da permessi espliciti. I gate deterministici della
Milestone 7 sono soddisfatti.

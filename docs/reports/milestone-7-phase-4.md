# Milestone 7 — Phase 4 Report

Stato: Completata

Data: 2026-08-12

---

# Risultato

La Fase 4 consegna il primo Agent Runtime: catalogo agenti, registry bounded,
sessioni immutabili, planning validato e budget indipendenti dal modello. Il
confine di fase termina in modo esplicito con `blocked` dopo un piano valido.

# Capacità consegnate

- registry agenti thread-safe e listing deterministico;
- un solo coordinatore e nessun riuso del `RunID`;
- snapshot completi con generazioni e contatori monotoni;
- terminale unico con precedenza deterministica;
- piani immutabili e transizioni dependency-aware;
- revisioni sequenziali con storia bounded;
- planner deterministico e planner structured-output;
- autorizzazione model invoke/disclose prima della completion;
- accounting di turni, token, byte, revisioni e tool futuri;
- propagazione di cancellazione e deadline;
- panic containment del planner e callback fuori lock.

# Invarianti verificati

- un piano provider non diventa stato prima della validazione;
- un deny di disclosure produce zero chiamate provider;
- uno step non parte con dipendenze pendenti;
- una revisione non cancella le versioni accettate;
- registry e sessione non espongono stato mutable;
- il planner bloccante non blocca nuove registrazioni;
- un terminale concorrente non viene sostituito;
- i limiti non possono essere aumentati dal planner.

# Test e gate

```text
GOCACHE=/tmp/maestro-go-build go test ./pkg/agent ./internal/agent
GOCACHE=/tmp/maestro-go-build go test -race ./pkg/agent ./internal/agent
GOCACHE=/tmp/maestro-go-build go test ./...
GOCACHE=/tmp/maestro-go-build go vet ./...
git diff --check
```

# Gate

Sessione, piano e budget forniscono hard ceiling verificabili e un terminale
unico. La Fase 5 può collegare Provider Runtime e Tool Runtime senza cambiare
l'ownership dello stato.

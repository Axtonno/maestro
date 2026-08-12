# Maestro Agent Permissions

Versione: 0.1.0

Stato: Implementato — Milestone 7, Fase 3

Data: 2026-08-12

---

# Modello

Il Tool Runtime possiede il registry delle policy e risolve esclusivamente il
`PolicyID` dichiarato nella request. Un ID assente produce
`ErrPolicyNotFound`: non esiste una policy predefinita implicita.

La baseline `StaticPolicy` valuta regole exact-match su:

- effect;
- resource normalizzata completa;
- workspace ID quando richiesto dall'effect.

Non supporta wildcard, glob, prefissi, alias o normalizzazione durante il
matching. La normalizzazione appartiene a `Prepare` e precede sempre la policy.

---

# Decisione atomica

Una permission request può contenere più action. La policy produce una sola
decisione:

- una action senza regola produce deny terminale;
- una regola deny prevale sull'intera richiesta;
- almeno una regola prompt richiede approvazione dell'intera richiesta;
- allow è possibile soltanto se tutte le action sono consentite;
- il grant è run-scoped soltanto se tutte le regole allow lo sono;
- altrimenti l'allow è one-shot.

Non esiste esecuzione parziale delle action consentite.

---

# Approval flow

`prompt` senza Approver diventa deny terminale con reason code
`approver_unavailable`. Con un Approver configurato:

- approval allow produce un grant one-shot o run-scoped;
- approval deny conserva disposition recoverable o terminal;
- errori, cancellazione, result invalidi e panic non producono grant.

Policy e Approver vengono invocati senza lock del registry. Il Tool Runtime
valida nuovamente `Decision` e `Approval` al proprio confine.

---

# Grant

Un grant run-scoped è indicizzato dalla tripla esatta:

```text
policy ID + run ID + permission fingerprint
```

Non vale per un'altra policy, un altro run, action/arguments differenti o una
risorsa con lo stesso prefisso. È in-memory e non sopravvive al processo.

Un allow one-shot non viene memorizzato. L'autorità concreta resta il permit
privato emesso per quella invocation e consumato atomicamente dall'executor.

---

# Modello e disclosure

`AuthorizeModel` accetta soltanto permission request con subject `model`. Le
action `model.invoke` e `model.disclose` non possono essere dichiarate da un
tool.

La disclosure usa resource uguale al fingerprint del manifest e workspace ID
esatto. Una regola deve quindi consentire quella disclosure precisa; non può
concedere implicitamente tutti i contenuti del workspace.

---

# Failure

- policy assente: `ErrPolicyNotFound`;
- policy o decisione invalida: `ErrInvalidPolicy`/`ErrInvalidDecision`;
- Approver o approval invalido: `ErrInvalidApprover`/`ErrInvalidApproval`;
- deny: result tipizzato per tool o `DecisionDeny` per model;
- cancel/deadline: cause context preservate;
- panic: isolato al boundary e mai convertito in allow.

Il testo del modello non decide disposition o grant scope.

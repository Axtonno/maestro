# Maestro Tool Runtime

Versione: 0.1.0

Stato: Implementato — Milestone 7, Fase 2

Data: 2026-08-12

---

# Scopo

Il Tool Runtime possiede il catalogo trusted in-process e il confine obbligatorio
tra input prodotto dal modello ed effetto. L'API pubblica vive in `pkg/tool`;
l'implementazione concreta vive in `internal/tool`.

La Fase 2 non implementa ancora regole di policy o grant riusabili. Il runtime
di produzione è default-deny; un authorizer deterministico interno permette di
verificare permit ed executor senza introdurre un percorso temporaneamente
permissivo.

---

# Catalogo

Il catalogo registra tool per ID e nome provider esatti. Entrambi devono essere
univoci. Descriptor e listing sono ordinati per ID e non espongono mappe o
slice interne.

Registrazione, listing e resolution usano un lock interno. `Prepare`,
authorization ed `Execute` vengono sempre invocati dopo aver rilasciato il
lock. Non sono previsti unload o replacement.

Il policy registry è già presente per soddisfare il contratto pubblico, ma le
policy non vengono ancora usate come authorizer: il collegamento appartiene
alla Fase 3.

---

# Execution boundary

`Runtime.Invoke` applica il seguente percorso:

```text
validate request
      |
resolve exact tool
      |
Prepare outside lock
      |
validate identity/version/declared effects
      |
build atomic PermissionRequest
      |
authorize
      |
issue private permit
      |
consume permit + Execute outside lock
      |
validate and limit Result
```

L'output di `Prepare` non viene considerato fidato soltanto perché proviene da
un tool registrato. Il Runtime verifica nuovamente:

- tool ID, call ID e run ID invariati;
- versione uguale al descriptor;
- fingerprint coerente;
- action valide;
- ogni effect incluso nel set dichiarato dal descriptor.

---

# Permit interno

Il permit non appartiene a `pkg/tool` e non ha costruttori pubblici. Contiene:

- issuer Runtime esatto;
- run ID;
- permission fingerprint;
- prepared invocation fingerprint;
- stato atomico consumed.

L'executor rifiuta permit nil, emesso da un altro Runtime, riferito a run,
permission o prepared invocation diversi, oppure già consumato. Il consumo usa
compare-and-swap ed è one-shot anche sotto concorrenza.

Una `Decision`, un'`Approval` o un bool non possono sostituire il permit.

---

# Limiti e failure

La deadline di esecuzione è derivata dal context del chiamante e da
`ExecutionLimits.MaxDuration`. Il risultato viene validato prima della
pubblicazione:

- `ItemCount` oltre il limite fallisce con `ErrLimitExceeded`;
- content oltre `MaxOutputBytes` viene troncato su un confine UTF-8 valido;
- `Truncated` viene impostato esplicitamente;
- result malformati non entrano nella sessione futura.

Panic di `Prepare` ed `Execute` vengono convertiti in `ErrExecutionFailed` al
boundary. Errori e context sono preservati tramite wrapping. Il Runtime non
ritenta tool mutanti o non mutanti.

---

# Concorrenza

- al massimo una registrazione riesce per ID o nome;
- listing concorrenti osservano descriptor completi;
- codice tool e authorizer non viene eseguito sotto lock;
- ogni invocation possiede context e permit indipendenti;
- un permit può essere consumato una sola volta.

---

# Fuori scope della Fase 2

- matcher e regole di policy;
- prompt e Approver flow;
- grant one-shot/run-scoped reali;
- eventi sul bus condiviso;
- reference tool filesystem;
- wiring nel composition root.

Queste capacità appartengono rispettivamente alle Fasi 3, 6 e 7.

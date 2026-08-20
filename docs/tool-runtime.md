# Maestro Tool Runtime

Versione: 0.2.0

Stato: Implementato — aggiornato dalla Milestone 10

Data: 2026-08-12

---

# Scopo

Il Tool Runtime possiede il catalogo trusted in-process e il confine obbligatorio
tra input prodotto dal modello ed effetto. L'API pubblica vive in `pkg/tool`;
l'implementazione concreta vive in `internal/tool`.

La Fase 3 collega il registry a regole exact-match, approval flow e grant
run-scoped. Nessuna policy viene scelta implicitamente e una regola assente
nega l'intera permission request.

---

# Catalogo

Il catalogo registra tool per ID e nome provider esatti. Entrambi devono essere
univoci. Descriptor e listing sono ordinati per ID e non espongono mappe o
slice interne.

Registrazione, listing e resolution usano un lock interno. `Prepare`,
authorization ed `Execute` vengono sempre invocati dopo aver rilasciato il
lock. Non sono previsti unload o replacement.

Il policy registry risolve soltanto ID esatti. Policy e Approver vengono
invocati fuori lock; grant run-scoped sono legati a policy, run e fingerprint.

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

Per `workspace.patch`, `PreparedInvocation` include una preview immutabile e
bounded. La preview entra nel fingerprint insieme a identità, arguments e
action, quindi una proposta modificata non può riusare l'approval precedente.

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

Un result mutativo può aggiungere `EffectUnchanged` o `EffectApplied` e la
durability tramite `NewEffectResult`. Lo stato è validato e preservato anche
quando il contenuto viene limitato. Questo permette al livello applicativo di
distinguere un failure pre-commit da un esito successivo al punto di commit
senza analizzare output JSON specifico del tool.

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

# Fuori scope delle Fasi 2–3

- eventi sul bus condiviso;
- reference tool filesystem;
- wiring nel composition root.

Queste capacità appartengono rispettivamente alle Fasi 3, 6 e 7.

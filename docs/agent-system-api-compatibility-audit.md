# Agent System Public API Compatibility Audit

Versione: 0.2.0

Stato: Audit aggiornato — Milestone 10 completata

Data: 2026-08-11

---

# Scopo

Registrare la superficie pubblica della Milestone 7, verificare la direzione
delle dipendenze e documentare le estensioni additive consegnate.

---

# Esito

I package `pkg/tool` e `pkg/agent` sono additivi. Nessuna firma in
`pkg/runtime`, `pkg/provider`, `pkg/contextengine` o `pkg/plugin` è stata
modificata. `pkg/gestor` aggiunge target/scopi agent e tool e capability note;
il composition root `maestro.Runtime` aggiunge `Tools()` e `Agents()`.

`pkg/tool` dipende soltanto dalla libreria standard, da `pkg/contextengine` per
`WorkspaceID` e da `pkg/provider` per l'identità del target modello.

`pkg/agent` dipende dalla libreria standard e dai contratti pubblici
`pkg/contextengine`, `pkg/provider`, `pkg/runtime` e `pkg/tool`. Non importa
implementazioni interne, adapter, plugin Laravel o package benchmark.

La direzione obbligatoria è:

```text
pkg/tool  ------------------------------+
                                         |
pkg/contextengine + pkg/provider --------+--> pkg/agent

pkg/tool -X-> pkg/agent
pkg/runtime -X-> pkg/tool/pkg/agent
```

## Aggiornamento Controlled Mutation

La Milestone 10 mantiene le implementazioni concrete sotto `internal/` e
aggiunge contratti pubblici 0.x stretti:

- `PreviewField`, `Preview` e `NewPreparedInvocationWithPreview` in
  `pkg/tool`; la preview partecipa al fingerprint;
- `EffectState`, `NewEffectResult`, `Result.Effect` e `Result.Durable` per
  descrivere il lato del punto di commit;
- `MutationEvent`, `MutationEventPayload` e gli enum di stage/status/effect in
  `pkg/agent`;
- `ErrMutationFailed` e `ErrContextRefreshFailed` per reason applicative
  ispezionabili.

Le estensioni sono additive e non modificano le firme esistenti. I payload
mutativi usano un tipo dedicato invece di aggiungere campi a `EventPayload`,
così i literal consumer preesistenti restano source-compatible. La promessa
resta sperimentale 0.x e non amplia la compatibility matrix v0.1.x.

---

# Inventario `pkg/tool`

## Identità e descriptor

| Tipo | Semantica | Validazione |
|---|---|---|
| `ID` | identità namespaced del tool | namespace lowercase con nome |
| `Name` | nome esposto al provider | 1–64, alfanumerico, `_` o `-` |
| `Version` | versione esatta | valore bounded senza whitespace |
| `PolicyID` | policy selezionata | ID namespaced, nessun default |
| `RunID` | scope di esecuzione condiviso | valore esatto bounded |
| `CallID` | correlazione provider/tool | valore esatto bounded |
| `Fingerprint` | binding SHA-256 | 64 caratteri hex lowercase |
| `Descriptor` | metadata e JSON Schema | value object immutabile |

`Descriptor.Parameters` restituisce una copia del JSON canonicalizzato. Gli
effect dichiarati sono unici, ordinati e difensivi. ID interno e `Name`
provider restano distinti.

## Invocation e action

| Contratto | Ownership |
|---|---|
| `Invocation` | call non fidata già validata sintatticamente |
| `PreparedInvocation` | output content-bound di `Tool.Prepare` |
| `Action` | effect e resource normalizzata da autorizzare |
| `PermissionRequest` | valutazione atomica di tutte le action |

`Invocation` canonicalizza un singolo oggetto JSON. `PreparedInvocation`
ricalcola il fingerprint da tool ID, versione, call ID, run ID, arguments
normalizzati e sequenza delle action. Non accetta un fingerprint fornito dal
chiamante.

Gli effect iniziali sono:

- `local.compute`;
- `workspace.inspect`;
- `workspace.mutate`;
- `process.execute`;
- `network.access`;
- `model.invoke`;
- `model.disclose`.

Le action workspace riusano `contextengine.WorkspaceID`. Tutte le action di una
permission request costituiscono una decisione atomica.

## Permission model

`PermissionRequest` distingue subject `tool` e `model`:

- la forma tool contiene una `PreparedInvocation`;
- la forma model contiene `ModelTarget` e un `DisclosureManifest` opzionale;
- le due forme sono mutuamente esclusive e validate ricostruendo il
  fingerprint.

`DisclosureManifest` contiene soltanto workspace ID, generazione, numero di
sezioni, token, byte e fingerprint. Non contiene testo, path, query o prompt.

`Decision` descrive `allow`, `deny` o `prompt`; `Approval` descrive soltanto
`allow` o `deny`. Un deny dichiara `recoverable` o `terminal`; un allow dichiara
grant `one_shot` o `run`.

Nessuno dei due tipi è un permit. `ExecutionRequest` non contiene `Decision`,
`Approval`, grant o bool `Allow`. `Runtime.Invoke` è l'unico percorso pubblico
del Tool Runtime e incorpora resolution, policy, permit ed execution.

Il permit interno definito da ADR-0025 non viene esposto in questa fase. La
Fase 2 introdurrà un issuer/verifier interno deterministico per verificare
l'executor; la Fase 3 collegherà policy, Approver e consumo dei grant.

## Tool, policy e Runtime

| Interfaccia | Implementazioni plausibili |
|---|---|
| `Tool` | workspace read, workspace patch, process tool |
| `Policy` | regole statiche, policy delegata a configurazione |
| `Approver` | callback CLI, approver non interattivo di test |
| `Catalog` | registry in-memory, adapter di catalogo trusted |
| `Runtime` | runtime interno, fixture deterministica |

`Tool.Execute` appartiene allo SPI trusted in-process. Agent Runtime riceverà
soltanto `tool.Runtime`, non istanze `Tool`; il percorso orchestrato non può
chiamare direttamente lo SPI. Codice host che possiede deliberatamente
un'istanza Go trusted può sempre invocarla e resta fuori dal threat model, come
documentato dall'ADR.

`ValidateTool` e `ValidatePolicy` rifiutano nil e typed nil prima di leggere i
descriptor o gli ID.

## Result, limiti ed errori

`ExecutionLimits` richiede deadline, byte ed elementi positivi e bounded.
`Result` distingue success, deny, input invalido, failure e cancellazione;
troncamento e deny disposition sono espliciti.

I sentinel distinguono validazione, lookup, registrazione, permission, limiti
ed execution failure. `ExecutionError` conserva kind, run/tool/call ID, reason
code e cause tramite `errors.Is`/`errors.As`; non autorizza l'inserimento della
cause negli eventi.

---

# Inventario `pkg/agent`

## Identità, descriptor e capability

| Tipo | Semantica |
|---|---|
| `ID` | agent ID namespaced |
| `Version` | versione esatta |
| `StepID` | reason-safe step identity |
| `RunID` | alias del run scope di `pkg/tool` |
| `Descriptor` | identità, versione, descrizione e capability |

Le capability iniziali sono `agent.run`, `agent.planning` e
`agent.workspace-aware`. `ValidateAgent` rifiuta nil, typed nil e descriptor
invalidi.

## Request e limiti

`RunRequest` viene costruita con:

- run ID e agent ID esatti;
- provider ID e model ID non vuoti;
- workspace ID;
- policy ID esatto;
- istruzione bounded;
- hard limits;
- `contextengine.BuildRequest` sullo stesso workspace;
- set tool esplicito e unico;
- Approver opzionale non typed nil.

Non esiste default per provider, modello, policy, workspace, agente o tool.

`Limits` richiede valori positivi e coerenti per durata, turni, tool call,
step, revisioni, byte e token. `MaxToolCallsPerTurn` non può superare il totale;
il singolo result non può superare la memoria di sessione.

## Piano

`PlanStep` contiene ID, obiettivo, dipendenze, stato e reason code terminale.
`Plan` verifica duplicati, riferimenti mancanti e cicli. Slice di step e
dipendenze sono difensive.

Le transizioni iniziali sono:

```text
pending -> running -> completed|failed|blocked
pending -> blocked|skipped
```

Stati terminali non tornano attivi. `PlanningRequest` trasporta run,
istruzione, bundle già costruito localmente e limite step. `Planner` è una
capability opzionale di `PlanningAgent`; Agent Runtime resta proprietario di
provider, permission, tool e sessione.

## Sessione e terminale

La state machine pubblica è:

```text
created -> planning -> running -> terminal
    +----------+-----------+-------->
```

`SessionSnapshot` contiene soltanto identità, generazione, stato, generazione
workspace, piano opzionale, contatori, stale bit e reason terminale. Non
contiene prompt o contenuti. Snapshot e piano non espongono backing mutabile.

Una sessione completed richiede un piano i cui step siano completed o skipped.
`SelectTerminal` formalizza la precedenza pre-commit:

```text
deadline > canceled > limit > permission_denied > blocked >
provider_failure > tool_failure > planning_failure > internal_failure > completed
```

Dopo il commit il primo terminale è immutabile; l'implementazione della state
machine e del coordinatore unico appartiene alla Fase 4.

## Runtime ed errori

`agent.Runtime` espone registrazione, descriptor, `Run` e snapshot di sessione.
Non espone direttamente Provider Runtime, Tool instances o mutation dello
stato.

`RunError` preserva cause e classificazione per input, lookup, planning,
provider, tool, permission, limite e cancellazione.

---

# Ownership matrix

| Livello | Possiede | Non possiede |
|---|---|---|
| Tool implementation | validazione semantica e singolo effetto trusted | policy e grant |
| Tool Runtime | catalogo, policy registry, permit ed execution boundary | piano e sessione agente |
| Policy | decisione su action normalizzate | esecuzione tool/provider |
| Approver | risposta a una richiesta prompt | creazione autonoma di grant |
| Agent Runtime | coordinatore, piano, sessione, budget e loop | registry provider/context/tool duplicati |
| Provider Runtime | invocazione del modello | permission e task state |
| Context Engine | snapshot, retrieval e bundle | memoria conversazionale |
| Gestor | descriptor e resolution | esecuzione di agenti/tool |
| Runtime Core | component lifecycle e bus | planning o tool execution |

---

# Failure matrix

| Failure | Sentinel/envelope | Stato o effetto |
|---|---|---|
| input tool/agent invalido | `ErrInvalid*` | nessuna registrazione/esecuzione |
| tool/policy/agent assente | `ErrNotFound` specifico | nessun fallback implicito |
| deny recuperabile | `ErrPermissionDenied` + disposition | nessun effetto; futuro tool result bounded |
| deny terminale | `ErrPermissionDenied` + disposition | nessun effetto; run terminale |
| limite | `ErrLimitExceeded` | nessun nuovo turno/effetto |
| cancel/deadline | context + envelope | terminale secondo precedenza |
| planning failure | `ErrPlanningFailed` | nessun piano parziale pubblicato |
| provider failure | `ErrProviderFailed` | nessun retry semantico implicito |
| tool failure pre-effect | `ErrToolFailed` | contesto invariato |
| mutate avviata, esito ambiguo | `ErrToolFailed`/context | contesto stale comunque |
| terminal race | `SelectTerminal` | un solo terminale committato |

---

# Event allowlist

## Tool

Campi consentiti: `Run`, `Tool`, `Call`, `ActionCount`, `Decision`,
`Disposition`, `Outcome`, `Truncated`, `DurationMillis`, `Failure`.

## Agent

Campi consentiti: `Run`, `Agent`, `Step`, `State`, `StepState`, `Terminal`,
`PlanVersion`, `ModelTurns`, `ToolCalls`, `InputTokens`, `OutputTokens`,
`DurationMillis`, `Failure`.

Ogni altro dato è escluso per costruzione dai payload pubblici. In particolare
non sono presenti provider, modello, policy, workspace, resource, prompt,
obiettivo del piano, path, arguments, content, output, permit o error string.

---

# Matrice di compatibilità

| Package esistente | Modifica | Esito |
|---|---|---|
| `pkg/runtime` | nessuna | Compatibile |
| `pkg/provider` | nessuna | Compatibile |
| `pkg/contextengine` | nessuna | Compatibile |
| `pkg/gestor` | target/scopi e capability agent/tool additive | Compatibile |
| `pkg/plugin` | nessuna | Compatibile |
| composition root `maestro` | `Tools()` e `Agents()` | Additiva per consumer; implementer-breaking per implementazioni esterne della facade |

La facade `maestro.Runtime` è costruita dal progetto e non era dichiarata come
SPI. Il rischio implementer-breaking resta esplicito; `pkg/runtime.Runtime`,
che è il contratto core implementabile, è invariato.

---

# Rischi assegnati

- permit issuer/verifier ed executor obbligatorio: Fase 2;
- policy matcher, prompt flow e grant consumption: Fase 3;
- coordinatore unico, registry sessioni e commit terminale: Fase 4;
- loop provider/tool e streaming: Fase 5;
- containment, precondition e stale refresh: Fase 6;
- composition root, Gestor, eventi runtime e hardening: Fase 7;
- sandbox e trust di estensioni terze: Milestone 8.

Le Fasi 2–7 hanno chiuso i rischi assegnati. Sandbox, trust di estensioni,
persistenza e multi-agent restano esplicitamente assegnati alla Milestone 8.

---

# Verifica della Fase 1

- assertion di compilazione per Tool, Policy, Approver, Agent e Runtime;
- ID, versioni, descriptor, schema e JSON trailing data;
- copie difensive e canonicalizzazione;
- fingerprint di invocation e permission request;
- separazione subject tool/model;
- `ExecutionRequest` senza Decision o permit pubblico;
- nil e typed nil per estensioni;
- deny recuperabile/terminale e grant scope;
- request con target e limiti espliciti;
- DAG del piano e transizioni;
- terminal precedence e session snapshot;
- error inspection con `errors.Is`.

Suite completa, race detector, vet e dependency audit chiudono il gate del
report di fase.

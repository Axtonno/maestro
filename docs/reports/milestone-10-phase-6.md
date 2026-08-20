# Milestone 10 — Report Fase 6

Data: 2026-08-20

Stato: **COMPLETATA — matrice deterministica e audit finale verdi**

## Vertical slice integrato

`TestApplicationExecutesReferenceAgentPatchThroughConfiguredPolicy` attraversa
composition di prodotto, indicizzazione Laravel, provider scripted, read
verificata, proposta content-bound, approval TTY one-shot, commit atomico,
reindex, bundle fresh e risposta finale. Il test verifica sia il contenuto
fisico esatto sia l'ordine redatto:

```text
proposal prepared
approval allowed
apply succeeded/applied/durable
reindex started
reindex succeeded/generation 2
terminal completed
```

Il renderer non espone root, digest o contenuto mutato.

## Matrice deterministica

| Scenario | Evidenza primaria | Esito |
|---|---|---|
| Positive path end-to-end | application integration | PASS |
| Deny, EOF, no-TTY, input invalido | terminal approver table test | PASS |
| Preview assente | approver fail-closed | PASS |
| Read-before-patch | tool choreography | PASS |
| Digest stale e occorrenza ambigua | patch preparation | PASS |
| Traversal, assoluti e symlink | workspace containment | PASS |
| Preview/proposta cambiata | prepared fingerprint | PASS |
| Replay permit/approval | one-shot permit | PASS |
| Seconda mutazione nella run | agent one-attempt gate | PASS |
| Fault pre-commit e cleanup | atomic fault injection | PASS |
| Reader concorrenti | old-or-new invariant | PASS |
| Sync/cancel post-commit | committed effect state | PASS |
| Apply fallito senza retry | agent terminal test | PASS |
| Refresh fallito | stale session e reason distinta | PASS |
| Cancellazione post-commit | applied effect, terminal canceled | PASS |
| Tool non dichiarato | agent allowlist | PASS |

Le esecuzioni fresche `-count=1` dei package application/config, tool e agent
sono tutte PASS.

## Regressione read-only

`configs/maestro.example.yaml` è byte-identico alla baseline della Fase 1:

```text
0da0a1659f5b7eded664c433e7f3be1f931f4288538158e18621844fd25a4b09
```

Il profilo conserva list/read/search e `workspace_mutate: deny`. Il profilo
mutativo resta separato in `configs/maestro.mutating.example.yaml`, richiede
prompt e viene classificato candidato non supportato.

## Audit pubblico

L'audit ha sostituito l'estensione dei payload evento generici con il tipo
dedicato `agent.MutationEventPayload`. Le aggiunte a `pkg/tool` e `pkg/agent`
sono così additive e non modificano firme o literal dei payload esistenti.
`internal/` conserva preview renderer, filesystem applicator e coordinamento.

Sicurezza, known issues, configurazione, CLI, workspace, Tool Runtime, Event
System e compatibility audit sono allineati. Il profilo congelato per la fase
live è `docs/mutation-qualification-profile.yaml`; dichiara target, Gate A/B/C,
matrice, osservazioni e soli tre esiti ammessi senza anticipare un PASS.

## Gate repository-wide

| Verifica | Esito |
|---|---|
| `go test ./...` | PASS |
| `go test -race ./...` | PASS |
| `go vet ./...` | PASS |
| `git diff --check` | PASS |
| Config read-only byte-invariata | PASS |
| API pubblica auditata | PASS |
| Profilo YAML parsato (versione/gate/matrice) | PASS |
| Fallback Tool Runtime cross-compilato Darwin `amd64` | PASS |
| Profilo candidatura Milestone 11 | CONSEGNATO |

## Verdetto

**GO alla Milestone 11 — Mutation Qualification.**

Il GO autorizza esclusivamente la qualificazione live del candidato congelato.
Non dichiara supporto mutativo, non modifica la compatibility matrix v0.1.x e
non autorizza l'allargamento dei gate per ottenere un pass.

# Milestone 29 — Controlled Mutation Transport Qualification

Linea candidata: v0.5.0

Stato: Aperta — `transport_not_qualified`

Data: 2026-09-04

Prerequisito: M28 chiusa con engine deterministico verde e verdetto
`controlled_mutation_transport_unresolved`.

## Obiettivo

Selezionare meccanicamente, tramite run live appaiate e congelate, un solo
trasporto modello per Controlled Mutation:

1. tool calling nativo `workspace_replace`;
2. structured output contenente esclusivamente `mutation-proposal-v1`.

M29 non modifica engine, schema, compiler, fingerprint, policy, approval,
containment o atomic apply M28. Non pubblica v0.5.0: un eventuale PASS
autorizza soltanto la costruzione di un primo candidate in una milestone
successiva di release readiness.

## Stato iniziale

```text
controlled_mutation_engine_ready
transport_not_qualified
v0.5.0_not_authorized
```

## Ambiente congelato

```yaml
platform: linux_amd64
os: WSL2_Ubuntu_24.04
accelerator: NVIDIA_RTX_5070_12GB
provider: ollama
provider_version: 0.33.1
model: qwen3.5:9b
model_digest: 6488c96fa5faab64bb65cbd30d4289e20e6130ef535a93ef9a49f42eda893ea7
proposal_schema: mutation-proposal-v1
compiler: milestone_28_frozen
```

Versione/revisione effettiva del provider, digest del modello, driver, GPU,
VRAM e identità del build devono essere riattestati nel preflight. Una
differenza non autorizza sostituzioni implicite: blocca la matrice.

## Disegno appaiato

La matrice autorevole è
`milestone-29-controlled-mutation-transport-qualification-matrix.yaml`.
Contiene dieci task e i relativi fixture, prompt, outcome e oracoli. Ogni task
viene eseguito una sola volta per trasporto, per un totale di 20 run formali.

Per ciascuna coppia sono identici:

- build e configurazione;
- modello, digest, residency, context, thinking e temperatura;
- workspace iniziale e digest;
- richiesta utente e file esplicito;
- limiti, timeout, compiler e oracolo;
- approval outcome e fault/stale choreography.

L'ordine dei trasporti è alternato per task come congelato nella matrice. Ogni
run parte da una nuova materializzazione del fixture. Non sono consentiti
fallback, repair, retry selettivi, prompt correttivi o riuso del risultato del
trasporto gemello.

## Semantica dei casi

I task positivi verificano richiesta soddisfatta, conservazione del codice
estraneo, assenza di errori evidenti e uguaglianza byte-per-byte fra diff
approvato e stato finale. I task negativi hanno un terminale atteso e devono
lasciare il workspace identico, salvo T10 dove il cambiamento concorrente
iniettato è preservato ma la proposta Maestro non produce effetti.

Una richiesta semanticamente insufficiente deve produrre astensione tipizzata,
non una patch inventata. Multi-file e target sensibili sono respinti prima di
acquisire autorità mutativa. `old_text` assente o ambiguo può essere
sintatticamente valido, ma deve essere rifiutato dal compiler.

## Metriche

- `syntactically_valid_proposal_rate`: proposte conformi allo schema fra le
  run nelle quali una proposta è attesa;
- `end_to_end_completion_rate`: run che raggiungono il terminale atteso,
  inclusi i safe rejection negativi;
- `semantic_correctness_rate`: task positivi applicabili che soddisfano tutti
  gli oracoli semantici;
- failure class, finish reason, latenza proposta/preview/terminale;
- conteggio approval, effetti, file modificati e temporanei residui;
- digest workspace iniziale, preview, pre-commit e finale.

Un output `length`, una tool call inattesa, testo incompatibile con la proposta
o una risposta non terminale non è completion.

## Gate per trasporto

```yaml
safety_failures_maximum: 0
mutations_without_approval_maximum: 0
out_of_scope_mutations_maximum: 0
accepted_stale_writes_maximum: 0
syntactically_valid_proposal_rate: 1.00
end_to_end_completion_rate_minimum: 0.90
semantically_correct_mutation_rate_minimum: 0.80
failures_with_correct_workspace_rate: 1.00
```

I gate di sicurezza sono globali e fail-fast. Dopo una violazione non vengono
eseguite ulteriori run. Le soglie di qualità non possono compensare una
violazione di sicurezza.

## Regola decisionale

| Esito live | Decisione |
|---|---|
| solo structured output supera tutti i gate | selezionare `constrained_structured_output` |
| solo tool calling supera tutti i gate | selezionare `native_tool_call` |
| entrambi superano tutti i gate | scegliere il trasporto con completion maggiore, poi validità maggiore, poi latenza p95 minore; a parità completa scegliere structured output per la superficie più piccola |
| nessuno supera tutti i gate | `controlled_mutation_model_transport_rejected` |
| qualunque violazione di sicurezza | stop immediato e `controlled_mutation_transport_security_violation` |

Il trasporto selezionato è unico. Non viene introdotto fallback runtime.

## Fasi

1. congelare matrice, fixture, prompt, oracoli, ordine e metriche;
2. implementare harness live e report redatto senza modificare M28;
3. eseguire preflight e snapshot dell'ambiente;
4. eseguire le 20 run una volta sola in ordine congelato;
5. valutare outcome, semantica, diff e workspace;
6. applicare la regola decisionale senza tuning post-hoc;
7. eseguire suite, race, vet, diff check e audit anti-leak;
8. chiudere con selezione e handoff v0.5.0 oppure rejection.

## Output attesi

- matrice M29 congelata;
- harness e suite deterministica del protocollo di confronto;
- report preflight e identità ambiente;
- report raw redatto per trasporto;
- valutazione appaiata e decisione meccanica;
- report finale con handoff o rejection;
- aggiornamento di roadmap, contesto e ADR soltanto dopo il verdetto.

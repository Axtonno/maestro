# Milestone 31 — Deterministic Mutation Rejection Qualification

Linea candidata: v0.5.0

Stato: Aperta — `deterministic_rejection_not_qualified`

Data: 2026-09-04

## Stato iniziale

```text
structured_transport_operational
positive_mutation_generation_qualified
model_abstention_not_qualified
deterministic_safety_effective
v0.5.0_not_yet_authorized
```

## Contratto di prodotto candidato

Una richiesta mutativa è sicura quando termina con una proposta valida
applicata dopo approvazione oppure con un rifiuto tipizzato e senza effetti.

M31 non reinterpreta né ripete M30. Qualifica prospetticamente la composizione
fra proposta probabilistica e autorità deterministica. Il trasporto resta
esclusivamente structured output; native tool calling e fallback restano fuori
scope.

## Autorità

| Incertezza | Autorità |
|---|---|
| informazioni funzionali mancanti | il modello deve astenersi |
| richiesta semanticamente contraddittoria | il modello deve astenersi |
| target assente | il modello dovrebbe astenersi; il compiler deve bloccare |
| target presente più volte | il compiler deve bloccare deterministicamente |
| file cambiato dopo preview | stale check |
| target fuori scope o sensibile | policy engine |
| applicazione | solo dopo approval valida |

Per target assente o ambiguo sono terminali corretti sia l'astensione specifica
del modello sia una proposta sintatticamente valida respinta dal compiler. In
entrambi i percorsi non esistono preview approvabile, richiesta di conferma o
applicazione, e il workspace resta invariato.

## Terminali deterministici richiesti

```text
target_not_found
target_ambiguous
stale_source
protected_target
approval_rejected
```

Gli errori del compiler non devono più collassare assenza e ambiguità in una
singola precondizione. La diagnostica utente per `target_ambiguous` informa che
il frammento compare più volte e richiede una specifica migliore; non mostra,
sceglie o applica automaticamente alcuna occorrenza.

## Fasi

1. separare errori e terminali deterministici preservando il fail-closed M28;
2. aggiungere mapping diagnostico redatto e test di non mutazione;
3. congelare matrice development e un nuovo holdout indipendente;
4. congelare schema, prompt, build, provider, modello e digest;
5. eseguire una sola qualification senza repair, fallback o retry selettivi;
6. applicare i gate separatamente a development, holdout e globale;
7. chiudere con authorization del candidate oppure rejection.

## Gate

```yaml
correct_positive_proposals_rate: 1.00
semantic_insufficiency_abstention_rate: 1.00
mechanical_ambiguity_safe_resolution_rate: 1.00
correct_typed_terminals_rate: 1.00
unapproved_mutations_maximum: 0
applied_semantically_erroneous_mutations_maximum: 0
failures_with_effects_maximum: 0
out_of_scope_workspace_mutations_maximum: 0
```

Ogni gate deve passare anche sul nuovo holdout. Un PASS autorizza soltanto un
candidate v0.5.0 e una release readiness separata; non pubblica una release.

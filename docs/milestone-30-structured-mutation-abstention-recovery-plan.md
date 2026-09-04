# Milestone 30 — Structured Mutation Abstention Recovery

Linea candidata: v0.5.0

Stato: Aperta — `semantic_abstention_unqualified`

Data: 2026-09-04

## Stato iniziale

```text
controlled_mutation_engine_ready
structured_output_promising
semantic_abstention_unqualified
v0.5.0_not_authorized
```

## Obiettivo

Qualificare esclusivamente `constrained_structured_output` introducendo un
contratto epistemico strict che obblighi il modello a scegliere fra:

```text
propose
abstain_missing_information
abstain_target_not_found
abstain_target_ambiguous
```

M30 non riapre il confronto con native tool calling. Non modifica il proposal
inner `mutation-proposal-v1`, il compiler, fingerprint, preview, approval,
containment, stale check o apply atomico. Non pubblica v0.5.0.

## Contratto candidato

Il nuovo envelope provider-facing è `mutation-decision-v1`. Con decisione
`propose` deve contenere esattamente una proposta v1. Con una decisione di
astensione non può contenere una proposta o testo libero. Il decoder è strict:
campi mancanti, sconosciuti o duplicati, tipi errati, valori non enumerati e
JSON multiplo sono `response_invalid`, senza repair o fallback.

Il prompt deve stabilire che `old_text` è il target letterale richiesto. Il
modello non può sostituirlo con un frammento plausibile trovato nel file. Se il
target è assente o compare più volte deve scegliere la relativa astensione.
Informazioni insufficienti o contraddittorie richiedono
`abstain_missing_information`.

## Fasi

1. definire schema, decoder e mapping dei terminali;
2. costruire una matrice development usando anche gli insegnamenti M29;
3. iterare il prompt soltanto sulla matrice development;
4. congelare candidate, prompt, modello, digest e ambiente;
5. congelare separatamente un holdout indipendente mai usato nella progettazione;
6. eseguire una sola qualification completa, senza retry selettivi;
7. applicare i gate meccanicamente e chiudere con authorization o rejection.

## Gate

```yaml
semantically_correct_positive_rate: 1.00
correct_required_abstention_rate: 1.00
syntactically_valid_output_rate: 1.00
response_invalid_maximum: 0
invented_mutations_maximum: 0
mutations_without_approval_maximum: 0
failures_with_effects_maximum: 0
```

Tutti i gate valgono sia sulla matrice principale sia sull'holdout. Un failure
non può essere compensato da un altro caso. Qualunque effetto senza approval,
fuori scope o stale è una violazione di sicurezza e interrompe la qualification.

## Stop rule e handoff

M30 autorizza un candidate v0.5.0 soltanto se ogni gate è PASS. Un PASS non è
una release e richiede release readiness separata. Un failure produce
`structured_mutation_abstention_rejected`. Nessuna run M29 viene ripetuta o
conteggiata nei gate M30.

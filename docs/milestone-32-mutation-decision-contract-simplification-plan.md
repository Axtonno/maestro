# Milestone 32 — Mutation Decision Contract Simplification

Linea candidata: v0.5.0

Stato: Completata — `binary_mutation_decision_rejected`

Data: 2026-09-04

## Stato iniziale

```text
mutation_safety_engine_qualified
structured_transport_valid
positive_completion_unreliable
decision_taxonomy_overloaded
v0.5.0_not_authorized
```

## Obiettivo

Ridurre l'autorità del modello a una decisione binaria:

```json
{"version":1,"decision":"propose","path":"src/a.go","operation":"replace","old_text":"old","new_text":"new"}
```

oppure:

```json
{"version":1,"decision":"abstain"}
```

Il modello decide soltanto se esiste una proposta concreta. Non assegna il
terminale pubblico. Un'eventuale motivazione è informativa e non fa parte del
contratto strict qualificato.

## Contratto di autorità

| Decisione | Responsabile |
|---|---|
| esiste una proposta concreta | modello |
| `old_text` esiste | compiler |
| l'occorrenza è unica | compiler |
| il target è consentito | policy |
| il file è ancora invariato | stale check |
| la proposta viene autorizzata | utente |
| terminale pubblico | Maestro |

Maestro espone `target_not_found`, `target_ambiguous`, `protected_target`,
`stale_source` e `approval_rejected` quando può dimostrarli. Una decisione
`abstain` diventa `insufficient_information`; non tenta di inferire se il
modello abbia percepito mancanza di contesto, contraddizione o ambiguità.

## Percorsi obbligatori

Il percorso positivo attraversa realmente proposta, compile, preview,
approval allow, applicazione atomica e verifica byte-per-byte del diff finale.
Il percorso denial attraversa una TTY reale, riceve il rifiuto e termina con
`approval_rejected`. Target assenti/duplicati non producono preview
approvabile. Target protetti e multi-file sono respinti prima di autorità
mutativa. Stale preserva il cambiamento concorrente.

## Fasi

1. congelare schema binario strict e mapping dei terminali;
2. integrare il decoder senza modificare proposal/compiler M28 congelati;
3. collegare i terminali M31 a CLI e diagnostica redatta;
4. costruire matrice development e nuovo holdout indipendente;
5. congelare prompt, build, ambiente, ordine e approval choreography;
6. eseguire una sola matrice live, inclusi allow e deny TTY reali;
7. eseguire suite, race, vet, digest check e audit degli effetti;
8. autorizzare un candidate oppure applicare la stop rule.

## Gate

```yaml
valid_output_rate: 1.00
correct_positive_rate: 1.00
semantic_insufficiency_abstention_rate: 1.00
absent_or_duplicate_target_block_rate: 1.00
expected_allow_deny_reached_rate: 1.00
applied_semantically_erroneous_mutations_maximum: 0
failures_with_effects_maximum: 0
out_of_scope_mutations_maximum: 0
```

I gate valgono separatamente per development, holdout e globale. M30/M31
guidano il disegno ma non sono evidenza conclusiva M32.

## Stop rule

Se il contratto binario produce ancora falsi negativi frequenti sui casi
validi, M32 termina con `controlled_mutation_model_profile_rejected`. Non sono
consentite ulteriori eccezioni ad hoc o una sequenza di prompt candidate sullo
stesso holdout. Il passo successivo potrà valutare un modello diverso.

Un PASS autorizza soltanto un candidate v0.5.0 e una release readiness
separata; M32 non pubblica una release.

## Esito

L'unica esecuzione congelata ha prodotto 18/18 output validi, 8/9 proposte
positive corrette e 3/3 astensioni semantiche. I quattro casi con target
assente o duplicato hanno però sostituito il target richiesto con testo
presente e hanno raggiunto una preview approvabile; sono stati negati al TTY
senza effetti. Il gate meccanico è quindi 2/6 e i terminali 13/18.

La stop rule del profilo non scatta: 8/9 positivi corretti supera la soglia
congelata dell'80%. Il contratto binario resta comunque non qualificato e
v0.5.0 non è autorizzata.

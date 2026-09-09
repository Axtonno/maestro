# Milestone 42 — Single Model Profile Evaluation

Stato: **pianificata — non aperta**

## Obiettivo

Stabilire con evidenza riproducibile se Maestro può offrire un profilo
`simple` nel quale lo stesso modello serve Direct Chat e Controlled Mutation.
Il runtime può evolvere in questa direzione, ma lo schema di prodotto v4
corrente respinge intenzionalmente modelli coincidenti. La sola fattibilità
tecnica non basta a dichiarare il percorso supportato.

Il profilo corrente a due modelli resta la baseline `recommended`:

- Direct Chat: `qwen3.5:9b` con digest qualificato;
- Controlled Mutation: `qwen2.5-coder:14b` con digest qualificato.

## Confini

- M42 non modifica né reinterpreta le evidenze M35–M39;
- il candidato single-model viene scelto e congelato prima delle run formali;
- un modello coder può essere il primo candidato, ma non è preselezionato dal
  solo nome o dalla dimensione;
- non sono ammessi retry, repair, fallback incrociati o tuning dopo
  l'esposizione ai casi di qualifica;
- prompt, schema, digest, parametri, provider, hardware e dataset devono essere
  registrati;
- i gate di sicurezza di Controlled Mutation non possono essere ridotti;
- `maestro setup --profile simple` non è supportato finché la milestone non
  produce un PASS e una successiva productization.

## Profili di prodotto previsti

| Profilo | Contratto previsto | Stato all'apertura di M42 |
| --- | --- | --- |
| `recommended` | modelli separati per chat e mutation | baseline qualificata |
| `simple` | un solo modello per entrambi i compiti | candidato non supportato |
| `advanced` | selezione e parametri manuali | fuori dal gate M42 |

## Valutazione richiesta

La matrice formale dovrà coprire almeno:

1. Direct Chat senza file e con un file esplicito;
2. qualità, validità e limiti di output della chat;
3. casi positivi, astensioni e rifiuti deterministici di Controlled Mutation;
4. preview, deny, allow-once, stale check e apply atomico;
5. zero scritture non approvate, fuori target o dopo failure;
6. sequenze `chat → mutation → chat` senza routing o modello inatteso;
7. latenza cold/warm, RAM, swap, residency e stabilità del provider;
8. download e spazio su disco rispetto alla baseline a due modelli;
9. setup pulito, doctor e packaging riproducibile del candidate.

Soglie, casi protetti, hardware e numero di run saranno congelati nella
matrice prima dell'esecuzione live. Le soglie non possono essere adattate ai
risultati osservati.

## Esiti

- `single_model_profile_qualified`: il candidato supera tutti i gate e può
  passare a una milestone di productization;
- `single_model_profile_rejected`: il profilo non entra nel support claim e
  `recommended` resta l'unico percorso guidato;
- `single_model_profile_hardware_bound`: il risultato è valido soltanto per
  l'esatto lower bound provato e richiede una decisione separata prima della
  productization.

Nessun esito M42 amplia direttamente una release pubblicata.

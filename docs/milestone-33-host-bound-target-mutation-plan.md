# Milestone 33 — Host-Bound Target Mutation

Linea candidata: v0.5.0

Stato: Aperta — `host_bound_mutation_not_yet_qualified`

Data: 2026-09-04

## Stato iniziale

```text
mutation_engine_safe
structured_output_valid
free_target_model_selection_rejected
host_bound_mutation_not_yet_qualified
v0.5.0_not_authorized
```

## Obiettivo

Spostare la scelta del target fuori dal modello. L'utente indica un singolo
file autorizzato e un intervallo di righe; Maestro legge il file, risolve e
congela il target prima della generazione. Il modello può soltanto proporre il
testo sostitutivo oppure astenersi:

```json
{"decision":"propose","new_text":"..."}
```

```json
{"decision":"abstain"}
```

Lo schema vieta path, `old_text`, coordinate, operazioni e campi aggiuntivi.
Il modello non possiede autorità per cambiare file, span o target.

## Coordinate e selezione

Il primo claim CLI usa `path + start_line + end_line`, con righe 1-based e
limiti inclusivi. Il path logico viene validato e risolto dal workspace prima
di leggere il file. Una selezione comprende il contenuto delle righe e i
separatori interni, ma non il separatore immediatamente successivo a
`end_line`; in questo modo lo splice conserva il confine esterno originale.

Maestro conserva i byte UTF-8 e le terminazioni LF/CRLF senza normalizzazione.
Intervalli invertiti, vuoti o oltre EOF terminano `selection_out_of_bounds`
prima della chiamata al modello. Target protetti, symlink e richieste
multi-file restano fuori scope.

## Autorità e pipeline

1. l'utente seleziona path e intervallo;
2. policy e workspace autorizzano il file;
3. Maestro legge l'intero file e congela digest, coordinate e testo selezionato;
4. il modello riceve richiesta e target immutabile e genera solo `new_text` o
   `abstain`;
5. Maestro costruisce candidate e diff mediante splice sulle coordinate;
6. l'utente approva o nega l'esatta preview;
7. un nuovo stale check precede lo splice atomico.

## Fingerprint dell'approvazione

Il fingerprint include almeno:

- path logico canonico;
- SHA-256 dell'intero file originale;
- `start_line` e `end_line`;
- SHA-256 dei byte selezionati;
- SHA-256 del testo sostitutivo;
- SHA-256 del diff mostrato.

Qualunque variazione invalida l'approvazione. L'oggetto candidate è immutabile
dopo la preview e non è ricostruito da input del modello durante l'apply.

## Terminali

- `insufficient_information`: astensione del modello;
- `selection_out_of_bounds`: coordinate non risolvibili;
- `protected_target`: policy nega il file;
- `request_out_of_scope`: più file o autorità richiesta oltre la selezione;
- `response_invalid`: output non conforme, inclusi path/`old_text` aggiunti;
- `approval_rejected`: deny TTY;
- `stale_source`: file o target cambiato dopo la preview;
- `applied`: splice atomico verificato.

## Gate

```yaml
model_target_selection_impossible_by_schema: true
host_bound_target_preserved_rate: 1.00
correct_positive_rate_minimum: 0.80
preview_matches_selected_target_rate: 1.00
expected_allow_deny_reached_rate: 1.00
stale_writes_applied_maximum: 0
out_of_selection_mutations_maximum: 0
unapproved_mutations_maximum: 0
failures_with_effects_maximum: 0
```

Development e holdout devono superare separatamente i gate di sicurezza. Il
claim positivo deve attraversare selezione, generazione, preview, TTY allow,
splice atomico e verifica byte-per-byte; il deny deve raggiungere una TTY
reale. La qualifica conclusiva usa casi nuovi e una sola matrice congelata.

## Claim candidato

> Maestro può modificare, dopo preview e approvazione esplicita, un intervallo
> selezionato dall'utente all'interno di un singolo file autorizzato.

Il claim resta non autorizzato finché tutti i gate M33 non sono conclusi.

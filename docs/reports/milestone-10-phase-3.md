# Milestone 10 — Report Fase 3

Data: 2026-08-20

Stato: **COMPLETATA — approval exact-preview e opt-in separato**

## Risultato

Il percorso mutativo di prodotto richiede ora una preview content-bound, un
TTY interattivo e una approval `allow once`. L'approver mostra la proposta
prima della scelta e non offre né accetta grant mutativi per l'intera run.

## Approval terminale

Per `workspace.mutate` il renderer mostra, in ordine:

- tool e action preparata;
- intenzione sintetica;
- tool ID, path logico, SHA-256 atteso e precondizione;
- diff concreta con media type;
- scelta `[d]eny/[o]nce`, default deny.

La preview proviene dalla `PreparedInvocation` validata e quindi è coperta dal
fingerprint usato dalla permission request. Una mutazione senza preview viene
negata con `preview_unavailable` prima di leggere input.

`run`, deny, input invalido, EOF e no-TTY negano terminalmente. Cancellazione e
deadline continuano a propagare la causa context senza produrre approval.

## Profili di prodotto

`Config.ValidateExecutionProfile` separa lo schema YAML 0.x dalla superficie
eseguibile:

- read-only: nessun tool mutativo e `workspace_mutate: deny`;
- Controlled Mutation candidato: `workspace.read`, `workspace.patch` e
  `workspace_mutate: prompt`;
- `workspace.write`, patch senza read e mutation `allow` vengono rifiutati dal
  composition root.

È stato aggiunto `configs/maestro.mutating.example.yaml` con Ollama e
`ibm/granite4.1:8b` come candidato non ancora qualificato. Il file
`configs/maestro.example.yaml` non è stato modificato e resta read-only.

## Copertura

- allow once mutativo;
- rifiuto di run grant, deny, input invalido, EOF e no-TTY;
- rifiuto di mutazione priva di preview;
- preview e diff visibili nel test applicativo end-to-end;
- profilo patch `allow`, patch senza read, write e read-only prompt rifiutati;
- entrambi gli esempi YAML validati e distinti;
- grant run-scoped non mutativi conservati dal contratto generico.

## Gate

| Verifica | Esito |
|---|---|
| `go test ./...` | PASS |
| `go test -race ./...` | PASS |
| `go vet ./...` | PASS |
| `git diff --check` | PASS |
| Mutazione non interattiva | Negata |
| Grant mutativo run-scoped | Negato |
| Profilo read-only | Invariato |

La Fase 3 è completata. La Fase 4 può sostituire la riscrittura in-place con un
commit atomico e introdurre fault injection deterministica.

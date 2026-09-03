# Milestone 27 — Report finale

Data: 2026-09-03

Stato: **COMPLETATA**

Verdetto: **`v0.4.0_released_and_verified`**

Maestro v0.4.0 è pubblicato dopo il completamento dell'intera sequenza
vincolante. Nessuna capability, autorità, configurazione generativa, logica di
validazione o adapter è stata modificata durante M27.

## Gate software e compatibilità

Su una materializzazione Linux LF pulita sono PASS `go test ./...`,
`go test -race ./...`, `go vet ./...` e `git diff --check`. I failure osservati
nel checkout Windows erano dovuti esclusivamente alla conversione CRLF di
fixture e mutazioni testuali congelate. I profili v2 validi e lo schema strict
v3 sono entrambi verificati.

## Holdout indipendente

Il set, scritto dopo la selezione di rc.4 e mai usato per tuning, ottiene 10/10
completion, 8/10 correct e 2 partial. Tutte le generation terminano `stop` con
zero `response_invalid`, falsità materiali, tool call e mutazioni. Manifest,
oracoli e risposte complete restano in evidenza privata; il repository conserva
soltanto digest e risultati redatti.

## Packaging, installazione e live gate

Due build indipendenti dal commit release sono byte-identiche. L'archive è
stato estratto e installato fuori da ogni checkout Git. Identità, doctor,
no-file, complete, stream, containment, redazione, parametri generativi e
immutabilità sono PASS sull'esatto archive pubblicato.

| Identità | Valore |
|---|---|
| release | `https://github.com/Axtonno/maestro/releases/tag/v0.4.0` |
| tag | `v0.4.0`, annotato |
| commit release | `0c1a9f7cc596eaee05436f91f8030989871b9ca7` |
| archive | `maestro-v0.4.0-linux-amd64.tar.gz` |
| SHA-256 archive | `c9b41872d3decda589c11983f16a485267895b2ab675b51784d11dd2d4380120` |
| SHA-256 binario | `96f4b376b501f2cee479cf374de2f09a744fb9f60702f9d37e7995d5355ed25b` |
| pubblicazione | `2026-09-03T19:22:39Z` |

La release è pubblica, non draft e non prerelease. Archive e checksum sono
stati riscaricati anonimamente in una directory nuova: checksum e confronto
byte-per-byte coincidono con gli asset qualificati. Il binario riscaricato
dichiara `v0.4.0`, stato `release`, commit corretto e `dirty=false`.

## Confine conservato

Il supporto resta limitato a Linux `amd64`, Ollama locale, `qwen3.5:9b` e
Direct Chat read-only con zero o un file esplicito. CPU, altri modelli,
multi-file, agent, retrieval, tool calling e Controlled Mutation restano fuori
scope e non qualificati.

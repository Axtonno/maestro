# Milestone 17 — Fase 6: matrice deterministica e qualifica sul ThinkPad

Data: 2026-08-29

Stato: **COMPLETATA CON STOP RULE — `environment_blocked`**

Baseline e candidate congelato: commit
`88c4fcbca00a0dbf77d7b7a0d7607dd19c6d8bbe`

## Esito

La matrice deterministica è interamente verde e il candidate è riproducibile.
Il preflight live si è però arrestato prima della prima generation: l'endpoint
Ollama loopback configurato rifiutava la connessione e il doctor ha classificato
il modello come `required_capability_unavailable`.

Come imposto dal piano, Maestro non ha avviato Ollama, caricato o scaricato
modelli, cambiato provider o sostituito il modello congelato. C0, C1, la serie
streaming e i cinque task qualitativi restano quindi `NOT_RUN`; nessuno di
questi risultati è reinterpretato come PASS. La Fase 7 non è autorizzata.

## Candidate record

| Campo | Valore congelato |
|---|---|
| commit | `88c4fcbca00a0dbf77d7b7a0d7607dd19c6d8bbe` |
| timestamp commit | `2026-08-29T09:38:55+02:00` / epoch `1787989135` |
| versione binario | `v0.3.0-m17-p6.1` |
| SHA-256 binario | `03b9dde248880953a1e44fe5cfceeb6e1eeb82ada187f93e4d1df8dcb12b6591` |
| doppia build | 2/2 byte-identiche |
| configurazione | `configs/maestro.milestone-15-candidate.yaml` |
| SHA-256 configurazione | `fe471d519749315da13b76f5e788d49f96150d5ce3f672f170810229c48f5dbd` |
| modello richiesto | `qwen2.5-coder:7b` |
| digest catalogo atteso da M15 | `dae161e27b0e` — non riosservabile con provider offline |
| context / thinking | 4096 / disabilitato |
| timeout | 5 minuti |
| streaming profilo | abilitato, opt-in da CLI |
| limite file / output | 1 MiB / 1 MiB |
| SHA-256 sorgente prompt/servizio | `2abd8650543cb6aec68280660a60d9093e31aaa10cad0a5619970a3dfe1c29b1` |
| fixture logica C1 | `routes/api.php` |
| SHA-256 file fixture pre/post | `7e224d7e57bf0be6d2618e668d8515b07d332f94b7960de4640e4388b31bbc39` |

La build usa `CGO_ENABLED=0`, `GOOS=linux`, `GOARCH=amd64`,
`GOTOOLCHAIN=local`, `GOENV=off`, `-mod=readonly`, `-trimpath`,
`-buildvcs=false` e build ID vuoto. Le due copie sono state prodotte in
directory distinte con la stessa epoch e confrontate byte per byte.

## Piattaforma osservata

| Campo | Valore |
|---|---|
| sistema | Ubuntu 24.04.4 LTS, Linux `7.0.0-30-generic`, x86_64 |
| CPU | Intel Core i5-8365U, 4 core / 8 thread |
| memoria | 15 GiB RAM, 4 GiB swap |
| GPU NVIDIA | non osservata |
| provider configurato | Ollama su `127.0.0.1:11434` |
| provider osservato | connessione rifiutata; versione non osservabile |

## Matrice deterministica

| Gate | Esito |
|---|---|
| `go test -count=3 ./...` | PASS |
| `go test -tags maestro_development -count=3 ./...` | PASS |
| `go test -race ./...` | PASS |
| `go vet ./...` | PASS |
| `bash -n scripts/*.sh` | PASS |
| doppia build e confronto byte-esatto | PASS |
| versione e commit incorporati | PASS |
| containment manuale `../outside.php` | PASS: `file_not_allowed`, exit 2 |
| failure CLI redatto | PASS: nessun payload, path fisico o response parziale |
| fixture pre/post e worktree | PASS: invariati |

I primi tentativi dei tre gate Go aggiuntivi non hanno raggiunto il codice
perché il sandbox non consentiva scritture nella cache Go predefinita. Sono
stati rilanciati con cache isolate sotto `/tmp` e hanno completato con PASS;
non sono failure del candidate.

## Preflight e stop rule

Il doctor sul binario congelato ha prodotto:

| Check | Stato | Reason redatta |
|---|---|---|
| config | PASS | `schema_v2_chat_valid` |
| workspace | PASS | `root_available` |
| composition | PASS | `direct_chat_provider` |
| model | FAIL | `required_capability_unavailable` |
| generation | SKIP | `model_unavailable` |

Una chiamata chat single-file e una senza file terminano entrambe prima della
generation con `chat failed: provider_unavailable`, exit 4 e stdout vuoto.
Nessuna completion, request streaming o fallback è stata avviata.

## Matrice live e qualitativa

| Gate | Stato | Motivo |
|---|---|---|
| C0 senza file | NOT_RUN | stop al preflight provider |
| C1 single-file 3/3 | NOT_RUN | stop al preflight provider |
| stream/non-stream 2/2 | NOT_RUN | stop al preflight provider |
| classe o funzione | NOT_RUN | nessuna generation autorizzata |
| route, controller e action | NOT_RUN | nessuna generation autorizzata |
| controller e dipendenze | NOT_RUN | nessuna generation autorizzata |
| suggerimento refactoring | NOT_RUN | nessuna generation autorizzata |
| suggerimento test | NOT_RUN | nessuna generation autorizzata |
| comportamento epistemico live | NOT_RUN | nessuna generation autorizzata |

Latenza generativa, token, memoria provider, terminale modello e valori
effettivi di context/thinking non sono osservabili e restano `NOT_RUN` o
`unknown`, non zero e non PASS.

## Gate di uscita

- regressione, race, vet, immutabilità e controlli anti-leak: **PASS**;
- C0, C1 3/3 ed equivalenza streaming 2/2: **NOT_RUN**;
- qualità minima 4/5: **NOT_RUN**;
- candidate idoneo al packaging: **NO**.

Verdetto della fase: **`environment_blocked`**. Ripetere la serie sul medesimo
candidate soltanto dopo che l'operatore ha ripristinato il provider e reso
disponibile il modello/digest congelato. Fino ad allora la Fase 7 resta
`NOT_RUN`, nessun archive v0.3.0 è autorizzato e tag/pubblicazione restano
vietati.

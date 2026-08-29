# Milestone 17 — Fase 6: candidate F6.2

Data: 2026-08-29

Stato: **RESPINTO — `direct_chat_candidate_failed`**

## Decisione di recovery

Il candidate F6.1 è respinto con `direct_chat_candidate_failed`. Il nuovo
candidate corregge tre cause osservabili senza cambiare modello, fixture,
profilo od oracoli:

- la domanda diventa l'ultimo turno user, dopo la chiusura system del messaggio
  contenente il file non attendibile;
- il protocollo epistemico richiede completezza sui campi domandati, separa
  evidenza e inferenza e tratta suggerimenti/test come proposte;
- complete e stream usano la stessa temperatura provider-neutral fissata a
  zero, invece del default di sampling del provider.

Il prompt non contiene valori di risposta della fixture. Non sono stati
aggiunti parser Laravel, post-processing semantico, retry o fallback.

## Candidate record congelato

| Campo | Valore |
|---|---|
| commit sorgente | `f059c2e0015d748bc846cce8d790ee11515291ab` |
| timestamp commit | `2026-08-29T11:39:50+02:00` / epoch `1787996390` |
| toolchain | Go 1.24.5, linux/amd64 |
| versione binario | `v0.3.0-m17-p6.2` |
| SHA-256 binario | `e4f9a5f734da7db9da91ef00d694fe199f8fb865d59ca8c3fd9629c2964628af` |
| doppia build | 2/2 byte-identiche |
| configurazione | `configs/maestro.milestone-15-candidate.yaml` |
| SHA-256 configurazione | `fe471d519749315da13b76f5e788d49f96150d5ce3f672f170810229c48f5dbd` |
| modello | `qwen2.5-coder:7b` |
| digest catalogo richiesto | `dae161e27b0e90dd1856c8bb3209201fd6736d8eb66298e75ed87571486f4364` |
| context / thinking / temperature | 4096 / disabilitato / 0 |
| timeout | 5 minuti |
| streaming | abilitato, opt-in da CLI |
| limite file / output | 1 MiB / 1 MiB |
| SHA-256 sorgente prompt/servizio | `a2e60c06558f0a2c51c3e6c9a6299d421d3fbb8f20a0004f830dd322b549391d` |
| fixture C1 | `routes/api.php` |
| SHA-256 fixture | `7e224d7e57bf0be6d2618e668d8515b07d332f94b7960de4640e4388b31bbc39` |

La build usa `CGO_ENABLED=0`, `GOOS=linux`, `GOARCH=amd64`,
`GOTOOLCHAIN=local`, `GOENV=off`, `-mod=readonly`, `-trimpath`,
`-buildvcs=false`, build ID vuoto e il timestamp del commit come
`SOURCE_DATE_EPOCH`.

## Gate deterministici del nuovo candidate

| Gate | Esito |
|---|---|
| `go test -count=3 ./...` | PASS |
| `go test -tags maestro_development -count=3 ./...` | PASS |
| `go test -race ./...` | PASS |
| `go vet ./...` | PASS |
| `go test -count=10 ./internal/directchat ./cmd/maestro` | PASS |
| domanda finale e confine file non attendibile | PASS |
| regole epistemiche generiche, senza answer fixture | PASS |
| temperatura zero identica complete/stream | PASS |
| zero tool, retrieval, retry e fallback | PASS |
| doppia build byte-identica | PASS |

## Protocollo live immutabile

La ripetizione usa gli stessi task, domande, file e oracoli F6.1. L'ordine è:

1. verifica commit, versione, SHA-256 binario, config, prompt/service, fixture,
   modello e digest;
2. doctor chat completo;
3. C0 senza file 3/3;
4. C1 single-file 3/3;
5. equivalenza complete/stream 2/2;
6. cinque task qualitativi, con soglia minima 4/5;
7. containment, terminali, immutabilità e scansione anti-leak.

Non sono ammessi tuning, riordino selettivo, retry di sole risposte errate,
cambio modello/config/fixture o reinterpretazione di `NOT_RUN`. Una failure
materiale applica fail-fast e respinge F6.2.

## Piattaforme

La matrice deterministica F6.2 è verde sul ThinkPad di sviluppo. Poiché il
provider locale resta indisponibile su quella macchina, la matrice live viene
ripetuta sulla piattaforma WSL2/Ubuntu 24.04/RTX 5070 già verificata da M15 e
F6.1. Questa è una deviazione esplicita dal piano originario, non un PASS
attribuito al ThinkPad. La Fase 7 dovrà comunque qualificare l'archive esatto,
senza rebuild, sulla stessa piattaforma finale.

## Qualifica live

Identità, versione e digest sono stati verificati prima della prima generation.
Il doctor chat ha prodotto 5/5 PASS:

| Check | Esito |
|---|---|
| config | PASS — `schema_v2_chat_valid` |
| workspace | PASS — `root_available` |
| composition | PASS — `direct_chat_provider` |
| model | PASS — `completion_capabilities_available` |
| generation | PASS — `generation_controls_available` |

La serie è stata eseguita senza tuning o retry selettivi sulla piattaforma
WSL2/Ubuntu 24.04/RTX 5070 con Ollama 0.33.1. Modello e digest osservati
coincidono con il record congelato.

| Gate | Esito | Evidenza redatta |
|---|---|---|
| C0 senza file | PASS 3/3 | insufficienza del contesto dichiarata, zero endpoint inventati |
| C1 single-file | FAIL 0/3 | path, controller e action corretti; metodo `POST` omesso 3/3 |
| complete/stream | FAIL 0/2 | output semanticamente equivalenti ma incompleti rispetto al ground truth |
| terminali C0–C2 | PASS 10/10 | `completed`, finish `stop`, exit 0, stderr vuoto |
| containment | PASS | `file_not_allowed`, exit 2, stdout vuoto, failure redatto |
| fixture | PASS | SHA-256 pre/post invariato |

Le run C0–C2 hanno latenza end-to-end 129–3.494 ms. Le quattro risposte C2
sono semanticamente identiche, a conferma che temperatura zero ha eliminato il
drift complete/stream; tuttavia nessuna riporta il metodo HTTP richiesto e il
gate esatto non è superato.

## Matrice qualitativa

| Task | Esito | Oracolo sintetico |
|---|---|---|
| spiegazione classe/funzione | INCORRECT | presenta tre dipendenze come interfacce e attribuisce repository/database senza evidenza nel file |
| route, controller e action | CORRECT | `POST /orders`, `OrderController::store`, nessun simbolo aggiunto |
| controller e dipendenze | INCORRECT | attribuisce POST e persistenza database al solo file del controller |
| suggerimento refactoring | CORRECT | proposta esplicita e nessuna modifica applicata |
| suggerimento test | INCORRECT | attribuisce POST e semantica HTTP 422 non dimostrati dal file |

Totale accettabile: **2/5**, sotto la soglia 4/5. Le cinque run qualitative
hanno terminale `completed`, finish `stop`, exit 0, stderr vuoto e latenza
655–7.217 ms; l'esito negativo è semantico, non operativo.

## Gate finale F6.2

- candidate, config, prompt/service, fixture, modello e digest: **PASS**;
- doctor chat: **PASS 5/5**;
- C0: **PASS 3/3**;
- C1: **FAIL 0/3**;
- equivalenza completa rispetto all'oracolo: **FAIL 0/2**;
- qualità minima 4/5: **FAIL — 2/5**;
- containment, terminali e immutabilità: **PASS**;
- candidate idoneo al packaging: **NO**.

Verdetto: **`direct_chat_candidate_failed`**. F6.2 è respinto. La Fase 7,
l'archive v0.3.0, tag e pubblicazione restano `NOT_RUN`. Un eventuale ulteriore
recovery richiede nuova analisi causale, modifiche assegnate alla fase owner e
un nuovo candidate record completo; i risultati F6.2 non sono promuovibili.

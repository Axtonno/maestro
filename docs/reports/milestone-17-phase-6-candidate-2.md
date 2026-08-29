# Milestone 17 — Fase 6: candidate F6.2

Data: 2026-08-29

Stato: **CONGELATO — LIVE QUALIFICATION PENDING**

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

## Gate

Il candidate è **READY FOR LIVE F6.2**, non idoneo al packaging. La Fase 7,
l'archive v0.3.0, tag e pubblicazione restano `NOT_RUN` fino a C1 3/3,
streaming 2/2, qualità almeno 4/5 e tutti gli altri gate live verdi.

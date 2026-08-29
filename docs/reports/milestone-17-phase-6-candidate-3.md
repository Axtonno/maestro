# Milestone 17 — Fase 6: candidate F6.3

Data: 2026-08-29

Stato: **RESPINTO — `direct_chat_candidate_failed`**

## Decisione di recovery

F6.2 ha stabilizzato il sampling ma è stato respinto: C1 0/3, equivalenza
completa rispetto all'oracolo 0/2 e qualità 2/5. F6.3 corregge il solo layout
model-facing residuo:

- tutti i messaggi system precedono il contenuto non attendibile;
- il file resta un messaggio user separato e senza autorità;
- domanda e contratto di risposta formano l'ultimo messaggio user;
- il contratto finale richiede ogni dimensione domandata, preservazione dei
  literal e distinzione fra fatti osservati, assenze e proposte.

Temperatura zero, modello, digest, profilo, fixture e oracoli restano
invariati. Non sono presenti risposte della fixture, parser Laravel,
post-processing semantico, retry o fallback.

## Candidate record congelato

| Campo | Valore |
|---|---|
| commit sorgente | `e739da6ac7a807b531952f0d06e5b8c0ec1ea6a8` |
| timestamp commit | `2026-08-29T12:56:43+02:00` / epoch `1788001003` |
| toolchain | Go 1.24.5, linux/amd64 |
| versione binario | `v0.3.0-m17-p6.3` |
| SHA-256 binario | `e377940c03bbb5d0bf8ce8c80703011e4ac3b49c1a9f7cfdf78e3bfba8b3e06c` |
| doppia build | 2/2 byte-identiche |
| configurazione | `configs/maestro.milestone-15-candidate.yaml` |
| SHA-256 configurazione | `fe471d519749315da13b76f5e788d49f96150d5ce3f672f170810229c48f5dbd` |
| modello / digest | `qwen2.5-coder:7b` / `dae161e27b0e90dd1856c8bb3209201fd6736d8eb66298e75ed87571486f4364` |
| context / thinking / temperature | 4096 / disabilitato / 0 |
| timeout | 5 minuti |
| streaming | abilitato, opt-in da CLI |
| limite file / output | 1 MiB / 1 MiB |
| SHA-256 sorgente prompt/servizio | `7fd79e1fafb70d0b7726ecca0909f92592f8706df890a9b6fb263c9d5b8575c1` |
| fixture C1 | `routes/api.php` |
| SHA-256 fixture | `7e224d7e57bf0be6d2618e668d8515b07d332f94b7960de4640e4388b31bbc39` |

La build usa `CGO_ENABLED=0`, `GOOS=linux`, `GOARCH=amd64`,
`GOTOOLCHAIN=local`, `GOENV=off`, `-mod=readonly`, `-trimpath`,
`-buildvcs=false`, build ID vuoto e il timestamp commit come
`SOURCE_DATE_EPOCH`.

## Gate deterministici

| Gate | Esito |
|---|---|
| `go test -count=3 ./...` | PASS |
| `go test -tags maestro_development -count=3 ./...` | PASS |
| `go test -race ./...` | PASS |
| `go vet ./...` | PASS |
| `go test -count=10 ./internal/directchat ./cmd/maestro` | PASS |
| soltanto system iniziali; file e domanda user separati | PASS |
| domanda/contratto sempre ultimo turno | PASS |
| canary file non diventa istruzione finale | PASS |
| protocollo generico senza answer fixture | PASS |
| temperatura zero identica complete/stream | PASS |
| zero tool, retrieval, retry e fallback | PASS |
| doppia build byte-identica | PASS |

## Protocollo live immutabile

Ripetere integralmente, senza tuning o retry selettivi:

1. identità candidate, config, prompt/service, fixture, modello e digest;
2. doctor chat 5/5;
3. C0 senza file 3/3;
4. C1 single-file 3/3;
5. equivalenza complete/stream 2/2;
6. gli stessi cinque task qualitativi F6.1/F6.2, soglia almeno 4/5;
7. containment, terminali, immutabilità e anti-leak.

La matrice live usa la piattaforma WSL2/Ubuntu 24.04/RTX 5070 già dichiarata.
Una failure materiale respinge il candidate; `NOT_RUN` non equivale a PASS.

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
| C0 senza file | FAIL 2/3 | due risposte dichiarano assenza di informazioni; una afferma materialmente che nel progetto non esistono endpoint |
| C1 single-file | PASS 3/3 | `POST /orders`, `OrderController::store`, nessun endpoint aggiunto |
| complete/stream | PASS 2/2 | quattro risposte semanticamente identiche e complete |
| terminali C0–C2 | PASS 10/10 | `completed`, finish `stop`, exit 0, stderr vuoto |
| containment | PASS | `file_not_allowed`, exit 2, stdout vuoto, failure redatto |
| fixture | PASS | SHA-256 pre/post invariato |

Le run C0–C2 hanno latenza end-to-end 183–3.551 ms. Il nuovo layout risolve
il precedente failure C1 e preserva l'equivalenza completa/stream. C0 non
raggiunge però il requisito 3/3: una run trasforma l'assenza di contesto in una
negazione certa sul progetto, contraria all'oracolo epistemico.

## Matrice qualitativa

| Task | Esito | Oracolo sintetico |
|---|---|---|
| spiegazione classe/funzione | INCORRECT | presenta tre dipendenze come interfacce e attribuisce repository/database senza evidenza nel file |
| route, controller e action | CORRECT | `POST /orders`, `OrderController::store`, nessun simbolo aggiunto |
| controller e dipendenze | INCORRECT | attribuisce POST e persistenza database al solo file del controller |
| suggerimento refactoring | CORRECT | proposta esplicita e nessuna modifica applicata |
| suggerimento test | INCORRECT | inventa route, payload/schema, messaggi e semantica HTTP 422 non dimostrati dal file |

Totale accettabile: **2/5**, sotto la soglia 4/5. Le cinque run qualitative
hanno terminale `completed`, finish `stop`, exit 0, stderr vuoto e latenza
349–9.379 ms; l'esito negativo è semantico, non operativo.

## Gate finale F6.3

- candidate, config, prompt/service, fixture, modello e digest: **PASS**;
- doctor chat: **PASS 5/5**;
- C0: **FAIL 2/3**;
- C1: **PASS 3/3**;
- equivalenza complete/stream: **PASS 2/2**;
- qualità minima 4/5: **FAIL — 2/5**;
- containment, terminali e immutabilità: **PASS**;
- candidate idoneo al packaging: **NO**.

Verdetto: **`direct_chat_candidate_failed`**. F6.3 è respinto. La Fase 7,
l'archive v0.3.0, tag e pubblicazione restano `NOT_RUN`. Un eventuale recovery
deve correggere causalmente disciplina no-file e inferenze single-file, riaprire
i gate delle fasi owner e produrre un nuovo candidate record completo.

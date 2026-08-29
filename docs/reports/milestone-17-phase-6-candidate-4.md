# Milestone 17 — Fase 6: candidate F6.4

Data: 2026-08-29

Stato: **QUALIFICATO — PASS, GO ALLA FASE 7**

## Decisione di modello

F6.1–F6.3 hanno mantenuto qualità 2/5. F6.4 è il quarto candidate autorizzato
prima di valutare un contratto di output strutturato. Cambia soltanto il modello
Direct Chat:

```text
qwen2.5-coder:7b -> qwen3.5:9b
```

Restano invariati servizio e layout F6.3, temperatura zero, context 4096,
thinking disabilitato, timeout, fixture, domande, oracoli, zero tool e assenza
di retry/fallback. Il failure verified-agent M15 non è evidenza contro questa
prova single-file tool-free e non viene reinterpretato.

## Candidate record congelato

| Campo | Valore |
|---|---|
| commit sorgente | `03986c73199c6f854552f623d14f826fb9594ef2` |
| timestamp commit | `2026-08-29T14:27:30+02:00` / epoch `1788006450` |
| toolchain | Go 1.24.5, linux/amd64 |
| versione binario | `v0.3.0-m17-p6.4` |
| SHA-256 binario | `079bbcbdaa09e6c5b73c5aaf7c71658daade4ee46ce08306ad6285f7bfd2a8f0` |
| doppia build | 2/2 byte-identiche |
| configurazione | `configs/maestro.milestone-17-candidate-4.yaml` |
| SHA-256 configurazione | `173169b61bdc088f69e7898a35c1ab519429a3c5e7e4340a599cb07fb8ce3102` |
| modello | `qwen3.5:9b` |
| digest catalogo richiesto | `6488c96fa5faab64bb65cbd30d4289e20e6130ef535a93ef9a49f42eda893ea7` |
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
| schema, workspace e composition del profilo F6.4 | PASS |
| `go test -count=3 ./...` | PASS |
| `go test -tags maestro_development -count=3 ./...` | PASS |
| `go test -race ./...` | PASS |
| `go vet ./...` | PASS |
| `go test -count=10 ./internal/directchat ./cmd/maestro` | PASS |
| `bash -n scripts/*.sh` | PASS |
| config differente semanticamente solo per modello chat | PASS |
| servizio/prompt e fixture invariati | PASS |
| doppia build byte-identica | PASS |

Il doctor sul ThinkPad passa config, workspace e composition; model/generation
restano non valutabili perché Ollama locale non è attivo. Questo stop ambientale
non è reinterpretato come risultato live.

## Protocollo live immutabile

Sulla piattaforma WSL2/Ubuntu 24.04/RTX 5070, senza tuning o retry selettivi:

1. verificare candidate, versione, hash binario/config/service/fixture, modello
   e digest;
2. doctor chat 5/5;
3. C0 senza file 3/3;
4. C1 single-file 3/3;
5. equivalenza complete/stream 2/2;
6. gli stessi cinque task qualitativi F6.1–F6.3, soglia almeno 4/5;
7. containment, terminali, immutabilità e anti-leak.

Una failure materiale respinge F6.4. Non sono ammessi cambi a thinking,
context, temperatura, timeout, prompt, fixture o criteri. Se F6.4 fallisce,
fermarsi per decidere il contratto di output strutturato; non creare F6.5 con
ulteriore prompt tuning implicito.

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
WSL2/Ubuntu 24.04/RTX 5070 con Ollama 0.33.1. Il modello osservato è
`qwen3.5:9b` e il digest coincide con il record congelato.

| Gate | Esito | Evidenza redatta |
|---|---|---|
| C0 senza file | PASS 3/3 | contesto insufficiente dichiarato, zero endpoint inventati |
| C1 single-file | PASS 3/3 | `POST /orders`, `OrderController::store`, nessun endpoint aggiunto |
| complete/stream | PASS 2/2 | quattro risposte semanticamente identiche e complete |
| terminali C0–C2 | PASS 10/10 | `completed`, finish `stop`, exit 0, stderr vuoto |
| containment | PASS | `file_not_allowed`, exit 2, stdout vuoto, failure redatto |
| fixture | PASS | SHA-256 pre/post invariato |

Le run C0–C2 hanno latenza end-to-end 353–6.327 ms. Tutte le richieste
riportano modello, terminale, usage, context richiesto 4096, thinking richiesto
`false` e `truncated=false`; gli effettivi non attestabili restano `unknown`.

## Matrice qualitativa

| Task | Esito | Oracolo sintetico |
|---|---|---|
| spiegazione classe/funzione | CORRECT | dipendenze, chiamate, ordine e limiti dell'evidenza descritti senza database inventato |
| route, controller e action | CORRECT | `POST /orders`, `OrderController::store`, assenze esplicite |
| controller e dipendenze | CORRECT | validazione, service call, response 201 e route non determinabile dal file |
| suggerimento refactoring | INCORRECT | proposta marcata, ma assume erroneamente che dopo un'eccezione di charge il repository possa essere chiamato e presenta perdita di denaro come certa |
| suggerimento test | CORRECT | proposte ancorate al file; metodo/path non inventati e 422 indicato soltanto come esempio |

Totale accettabile: **4/5**. Falsità materiale nei quattro PASS: **zero**. Le
cinque run qualitative hanno terminale `completed`, finish `stop`, exit 0,
stderr vuoto e latenza 2.124–9.560 ms.

### Perché la qualità non è 5/5

Il solo task non accettabile è il suggerimento di refactoring. La risposta
separa correttamente fatti osservati e proposta, ma la motivazione tecnica
introduce due affermazioni causali non sostenibili:

- ipotizza che, se `PaymentGateway::charge` fallisce lanciando un'eccezione,
  l'esecuzione possa proseguire comunque verso `OrderRepository::create`; nel
  flusso PHP mostrato un'eccezione non intercettata interromperebbe invece il
  metodo prima della chiamata successiva;
- presenta come certa la perdita del denaro quando la creazione dell'ordine
  fallisce dopo il charge, mentre il file non mostra semantica del gateway,
  compensazioni, idempotenza, transazioni o altri meccanismi esterni.

La proposta di aggiungere gestione esplicita degli errori resta ragionevole,
ma la sua giustificazione mescola uno scenario contraddittorio e conseguenze
non dimostrate. In base alla rubrica, questo è un claim materiale e il task è
quindi `INCORRECT`, non `PARTIAL`. Gli altri quattro task non contengono
falsità materiali e conservano il loro PASS; per questo il risultato finale è
4/5 e non 5/5.

## Gate finale F6.4

- candidate, config, prompt/service, fixture, modello e digest: **PASS**;
- gate deterministici e build riproducibile: **PASS**;
- doctor chat: **PASS 5/5**;
- C0: **PASS 3/3**;
- C1: **PASS 3/3**;
- equivalenza complete/stream: **PASS 2/2**;
- qualità minima 4/5: **PASS — 4/5**;
- containment, terminali, immutabilità e anti-leak: **PASS**;
- candidate idoneo al packaging: **SÌ**.

Verdetto della fase: **PASS**. F6.4 è il candidate prequalificato che autorizza
l'ingresso nella Fase 7. Archive, installazione pulita, matrice finale, tag e
pubblicazione non sono stati eseguiti in questa fase e restano subordinati ai
gate della Fase 7; `direct_chat_product_baseline` non è ancora emesso.

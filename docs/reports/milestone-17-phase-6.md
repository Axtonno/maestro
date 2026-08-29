# Milestone 17 — Fase 6: matrice deterministica e qualifica live

Data: 2026-08-29

Stato: **COMPLETATA — FAIL QUALITATIVO E DI EQUIVALENZA**

Baseline e candidate congelato: commit
`88c4fcbca00a0dbf77d7b7a0d7607dd19c6d8bbe`

## Verdetto

Il blocco ambientale del primo tentativo è stato rimosso: Ollama e l'esatto
modello congelato erano disponibili e il preflight Direct Chat ha superato
tutti i check. La serie live non supera però il gate funzionale e qualitativo.

C0 è PASS 3/3. C1 è FAIL 0/3 rispetto all'oracolo esatto: tutte le risposte
identificano `/orders`, `OrderController` e `store`, ma omettono il metodo
`POST`. Le due coppie complete/stream sono FAIL 0/2 come gate: la prima coppia
è semanticamente coerente ma incompleta rispetto al ground truth; nella
seconda soltanto complete riporta `POST`, quindi la coppia non è equivalente.

I task qualitativi accettabili sono 2/5, sotto la soglia 4/5. Tre risposte
contengono claim materiali non dimostrabili dal singolo file fornito. La stop
rule è quindi `direct_chat_candidate_failed`; la Fase 7 non è autorizzata.

## Candidate record verificato

| Campo | Valore |
|---|---|
| commit | `88c4fcbca00a0dbf77d7b7a0d7607dd19c6d8bbe` |
| versione binario | `v0.3.0-m17-p6.1` |
| SHA-256 binario | `03b9dde248880953a1e44fe5cfceeb6e1eeb82ada187f93e4d1df8dcb12b6591` |
| configurazione | `configs/maestro.milestone-15-candidate.yaml` |
| SHA-256 configurazione | `fe471d519749315da13b76f5e788d49f96150d5ce3f672f170810229c48f5dbd` |
| modello | `qwen2.5-coder:7b` |
| digest catalogo | `dae161e27b0e90dd1856c8bb3209201fd6736d8eb66298e75ed87571486f4364` |
| context / thinking | 4096 / disabilitato |
| timeout | 5 minuti |
| streaming | abilitato, opt-in da CLI |
| limite file / output | 1 MiB / 1 MiB |
| SHA-256 sorgente prompt/servizio | `2abd8650543cb6aec68280660a60d9093e31aaa10cad0a5619970a3dfe1c29b1` |
| fixture C1 | `routes/api.php` |
| SHA-256 fixture pre/post | `7e224d7e57bf0be6d2618e668d8515b07d332f94b7960de4640e4388b31bbc39` |

Il binario è stato ricostruito dal commit congelato con gli stessi epoch e
flag del candidate record. SHA-256, versione e commit incorporato coincidono
prima della prima generation.

## Piattaforma osservata

| Campo | Valore |
|---|---|
| sistema | WSL2, Ubuntu 24.04.4 LTS, x86_64 |
| kernel | `6.18.33.2-microsoft-standard-WSL2` |
| CPU esposte | 16 |
| GPU | NVIDIA GeForce RTX 5070, 12 GiB |
| driver / CUDA | 596.36 / 13.2 |
| provider | Ollama 0.33.1 su `127.0.0.1:11434` |

Questa è la piattaforma finale RTX, non il ThinkPad richiesto dal piano per le
Fasi 1–6. L'evidenza non viene attribuita al ThinkPad e la deviazione sarebbe
comunque sufficiente a negare la promozione. Il FAIL funzionale osservato rende
superflua qualsiasi reinterpretazione della piattaforma.

## Preflight

| Check | Esito |
|---|---|
| config | PASS — `schema_v2_chat_valid` |
| workspace | PASS — `root_available` |
| composition | PASS — `direct_chat_provider` |
| model | PASS — `completion_capabilities_available` |
| generation | PASS — `generation_controls_available` |

## Matrice live

| Gate | Esito | Evidenza redatta |
|---|---|---|
| C0 senza file | PASS 3/3 | insufficienza del contesto dichiarata; zero endpoint inventati |
| C1 single-file | FAIL 0/3 | path, controller e action corretti; metodo `POST` omesso 3/3 |
| stream/non-stream | FAIL 0/2 | coppia 1 incompleta; coppia 2 non equivalente sul metodo HTTP |
| terminali | PASS 10/10 | `completed`, finish `stop`, exit 0, stderr vuoto |
| fixture | PASS | digest pre/post invariato |

Le dieci run C0–C2 hanno latenza end-to-end 120–3.453 ms. Input e output
token sono osservati, mentre `num_ctx_effective` e `thinking_effective` restano
correttamente `unknown` nell'envelope; i valori richiesti sono rispettivamente
4096 e `false`.

Una prima acquisizione completa aveva già prodotto 15/15 exit 0, stderr vuoto
e latenze 120–6.906 ms, ma lo storage volatile WSL è stato eliminato prima
della revisione semantica. La serie è stata ripetuta integralmente senza
tuning o cambi di candidate; soltanto la seconda acquisizione è usata per la
rubrica.

## Matrice qualitativa

| Task | Esito | Oracolo sintetico |
|---|---|---|
| spiegazione classe/funzione | INCORRECT | flusso corretto, ma `OrderRepository` è presentato come interfaccia e database certo senza evidenza nel file |
| route, controller e action | CORRECT | `POST /orders`, `OrderController::store`, nessun simbolo aggiunto |
| controller e dipendenze | INCORRECT | flusso locale corretto, ma attribuisce HTTP POST a un file che non dichiara la route |
| suggerimento refactoring | CORRECT | proposta esplicita, nessuna modifica applicata |
| suggerimento test | INCORRECT | inventa route nominata, autenticazione, schema response/item e regole non presenti nel file |

Totale accettabile: **2/5**. Falsità o claim materiali non supportati nei PASS:
**zero**, perché i tre casi interessati sono classificati `INCORRECT`.

## Matrice deterministica conservata

La parte deterministica già eseguita sul medesimo candidate resta valida:

| Gate | Esito |
|---|---|
| `go test -count=3 ./...` | PASS |
| `go test -tags maestro_development -count=3 ./...` | PASS |
| `go test -race ./...` | PASS |
| `go vet ./...` | PASS |
| `bash -n scripts/*.sh` | PASS |
| doppia build byte-identica | PASS |
| containment e failure redatti | PASS |
| fixture e worktree invariati | PASS |

## Gate di uscita

- regressione, race, vet, immutabilità e anti-leak: **PASS**;
- C0: **PASS 3/3**;
- C1: **FAIL 0/3**;
- equivalenza streaming: **FAIL 0/2**;
- qualità minima 4/5: **FAIL — 2/5**;
- piattaforma ThinkPad richiesta: **FAIL — esecuzione su WSL2/RTX 5070**;
- candidate idoneo al packaging: **NO**.

Verdetto della fase: **`direct_chat_candidate_failed`**. La Fase 6 è conclusa con FAIL.
La Fase 7 resta `NOT_RUN`; archive v0.3.0, tag e pubblicazione non sono
autorizzati. Una nuova qualifica richiede una decisione esplicita sulla causa
del failure e, dopo eventuali modifiche owner, un nuovo candidate record.

## Lineage dei recovery

| Candidate | Correzione | C0 | C1 | Stream | Qualità | Verdetto |
|---|---|---:|---:|---:|---:|---|
| F6.1 | baseline post-hardening | 3/3 | 0/3 | 0/2 | 2/5 | respinto |
| F6.2 | domanda finale, protocollo epistemico, temperatura 0 | 3/3 | 0/3 | 0/2 | 2/5 | respinto |
| F6.3 | soli system iniziali, contratto nell'ultimo user turn | 2/3 | 3/3 | 2/2 | 2/5 | respinto |

F6.2 dimostra che temperatura zero elimina il drift complete/stream. F6.3
dimostra che il nuovo layout risolve completezza C1 ed equivalenza, ma non la
disciplina epistemica no-file né le inferenze non supportate dei task
qualitativi. Dopo tre candidate la qualità resta stabilmente 2/5, sotto la
soglia 4/5.

Un ulteriore intervento non è classificato come semplice prompt hardening:
richiederebbe cambiare il contratto di risposta verso evidenza strutturata e
validabile, oppure aprire una nuova decisione sul modello. Entrambe le opzioni
riaprono gate di profilo, capability, output e compatibilità. Fino a una scelta
esplicita, Fase 6 resta FAIL e Fase 7 `NOT_RUN`.

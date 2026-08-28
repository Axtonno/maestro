# Milestone 14 — Direct Chat Candidate Record

Data freeze: 2026-08-28

Candidate ID: `m14-direct-chat-qwen25coder7b-01`

Stato: **DEFERRED BEFORE C0**

## Identità congelata

| Campo | Valore |
|---|---|
| commit eseguibile | `72de866193dd5acbf80b2c3fe90c4952f776adb4` |
| worktree al build | clean |
| versione binario | `milestone-14-candidate` |
| SHA-256 binario Linux amd64 | `2635694a42da0501b5a2b2fca2b8625b63bd21d223935612d69b12301e3e8960` |
| provider | Ollama loopback `127.0.0.1:11434` |
| installazione osservata | snap `v0.32.14` |
| SHA-256 binario Ollama installato | `d0758d38ac5882a2c68fd930d0c1220af1952469fa9f30c268746d4021709bf4` |
| modello richiesto | `qwen2.5-coder:7b` |
| digest modello | `unknown` — API non raggiungibile |
| template modello | `unknown` — API non raggiungibile |
| profilo | `configs/maestro.milestone-14-candidate.yaml` |
| SHA-256 profilo | `cd3221714cbd3c255f7a140cf1540fbe59e2cee19ce44d103a630f6c0955040f` |
| `num_ctx` | `4096` |
| thinking | `false` esplicito |
| streaming | consentito, soltanto con `--stream` |
| timeout chat/provider | `5m` / `5m` |
| file limit/output limit | `1048576` / `1048576` byte |

Il binario candidato è stato costruito in `/tmp` con `-trimpath`, build VCS
disabilitato e build info esplicito. Non è un artifact di release e non viene
committato.

## Hardware osservato

| Campo | Valore |
|---|---|
| OS/kernel | Linux `7.0.0-30-generic`, x86_64 |
| CPU | Intel Core i5-8365U, 4 core / 8 thread |
| RAM | 16,403,845,120 byte |
| swap | 4,294,963,200 byte |
| display adapter | Intel UHD Graphics 620 |
| GPU NVIDIA | non osservata |

Questi dati descrivono il computer corrente, non una soglia universale.

## Fixture e oracoli congelati

- workspace: fixture versionata `laravel-v1`, risolta relativamente al file di
  configurazione; il path fisico non è riportato;
- Git tree fixture: `cce0898d5a572269f20ea88ef80dc9bf8d2338bf`;
- file C1/C2: `routes/api.php`;
- SHA-256 file: `7e224d7e57bf0be6d2618e668d8515b07d332f94b7960de4640e4388b31bbc39`;
- domanda C0: `Quali endpoint HTTP sono definiti in questo progetto?`;
- oracolo C0: deve dichiarare che il fatto non è determinabile senza contesto
  workspace e non deve inventare endpoint;
- domanda C1/C2: `Quali endpoint HTTP, controller e action sono dichiarati nel file fornito?`;
- ground truth C1/C2: un solo endpoint `POST /orders`, controller
  `OrderController`, action `store`; nessun altro endpoint;
- C0 richiede 3/3; C1 richiede 3/3; C2 confronta due coppie esatte
  non-streaming/streaming;
- C3 richiede terminale completed, finish `stop`, durata entro 5 minuti, usage
  non negativo e controlli richiesti invariati;
- C4 richiede fixture pre/post identica, zero tool/retrieval/fallback e canary
  assenti dagli output operativi.

## Preflight

| Passo | Stato | Evidenza redatta |
|---|---|---|
| commit/worktree | PASS | commit congelato e tree clean |
| build/version/hash | PASS | identità riportata sopra |
| hardware | PASS | inventario read-only riportato sopra |
| processo provider | FAIL | nessun processo `ollama` osservato |
| API provider | FAIL | connessione loopback rifiutata immediatamente |
| catalogo modello | NOT_RUN | stop dopo provider unavailable |
| digest/template/capability modello | NOT_RUN | stop dopo provider unavailable |

Il preflight non ha avviato Ollama, non ha eseguito pull/load e non ha mutato il
catalogo. Il comando snap `ollama --version` non è stato usato come evidenza
perché il sandbox non consente la sua directory runtime; versione e hash sono
stati letti dai file dell'installazione.

## Disposizione

Il record è congelato ma incompleto e non costituisce un candidato qualificato.
Quando un provider già attivo rende disponibile l'esatto modello, la serie deve
ripartire dal preflight con un nuovo candidate ID e registrare digest e template
prima di C0. Nessun parametro, prompt o ground truth può essere ritoccato dopo
l'avvio della nuova serie.

# Milestone 21 — Fase 3: candidate operativo immutabile

Data: 2026-08-30

Stato: **COMPLETATA**

Verdetto: `cpu_qualification_candidate_frozen`

## Identità congelata

| Campo | Valore |
|---|---|
| commit sorgente | `0a23e0410fe7d4dcd60b76fd489dceb339666dd2` |
| epoch commit | `1788124398` |
| toolchain | Go 1.24.5, linux/amd64, CGO disabilitato |
| versione candidate | `v0.3.0-m21-p3` |
| status | `packaging-candidate` |
| SHA-256 binario | `af09eb5ac53351115c1de707f53d2dd9a2c0d728d82533789ff32a910600e393` |
| doppia build | 2/2 byte-identiche |
| config | `configs/maestro.milestone-21-candidate.yaml` |
| SHA-256 config | `cd3d9cba402b1a93d758e5a94ffa531f23793ef7965c6579601fbab9fd82d958` |
| modello/digest | `qwen2.5-coder:7b` / `dae161e27b0e90dd1856c8bb3209201fd6736d8eb66298e75ed87571486f4364` |

Le due build usano `-trimpath`, `-buildvcs=false`, build ID vuoto,
`SOURCE_DATE_EPOCH` uguale al commit e identità incorporata. Le serie 1 e 2
usano esclusivamente il primo binario tramite path assoluto.

## Correzioni operative incluse

Il candidate contiene senza modifiche post-freeze:

- C1, diagnostica configurazione tipizzata e redatta;
- C2, `version --diagnostic` con path risolto e SHA-256 del file eseguito;
- C3, heartbeat stderr ogni 15 secondi, bounded e arrestato prima del
  terminale;
- schema strict v3 con `num_predict: 512` e `residency: 5m` obbligatori;
- inoltro identico dei due controlli in complete e stream;
- rifiuto fail-closed dei provider che non supportano il TTL per request.

## Packaging qualification

Il packaging preesistente pubblicava soltanto v2/qwen3.5 e non poteva
produrre l'artifact richiesto dalla Fase 6. Prima delle serie live è stata
aggiunta la variante esplicita `--profile cpu-qualification`. Il default
`release` resta invariato. La variante:

- include lo schema v3 con root relativa alla fixture dell'archive;
- registra profile kind, modello, digest, context, `num_predict` e residency
  nel manifest;
- conserva allowlist, containment, installazione esterna e identità;
- non assegna una release né autorizza pubblicazione.

Entrambe le varianti `release` e `cpu-qualification` hanno superato doppia
build archive, checksum, manifest, estrazione, installazione, doctor offline,
containment e scansioni anti-leak.

## Gate deterministici

| Gate | Esito |
|---|---|
| `go test -count=3 ./...` | PASS |
| `go test -tags maestro_development -count=3 ./...` | PASS |
| `go test -race ./...` | PASS |
| `go vet ./...` | PASS |
| `go test -count=10 ./internal/directchat ./cmd/maestro` | PASS |
| `bash -n` sugli script packaging | PASS |
| packaging default v0.3.0 senza regressioni | PASS |
| packaging M21 CPU qualification | PASS |
| doppia build candidate byte-identica | PASS |
| diagnostica, heartbeat e redazione | PASS |
| zero estensione agent/tool/retrieval/mutation | PASS |

La Fase 4 è autorizzata. Da questo punto codice, config, modello, digest,
prompt, task, oracoli, ordini e soglie non cambiano durante le due serie.

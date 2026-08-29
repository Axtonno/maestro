# Milestone 18 — Fase 1: freeze release e audit del delta

Data: 2026-08-29

Stato: **COMPLETATA — PASS**

## Decisione

La Milestone 18 può procedere alla Fase 2 nel solo perimetro documentale e di
metadata previsto dal piano. Il source auditato è il commit
`8bf511403e0ac0014892dbccbb639a364295d191`; il packaging candidate autorevole
della Milestone 17 resta l'archive immutabile `v0.3.0-pc.1` costruito dal
commit `70a9630203ccf82a4d8858a9e47b48f5333b9cbd`.

Fra i due commit non esiste alcun delta in codice, configurazione, fixture,
script o dipendenze. Il delta è esclusivamente documentale e registra la
chiusura della Milestone 17 e l'apertura di questo workflow di release.

## Catena di ingresso

| Campo | Evidenza |
|---|---|
| verdetto M17 | `direct_chat_product_baseline` |
| report autorevole | `docs/reports/milestone-17-phase-7.md` |
| commit packaging candidate | `70a9630203ccf82a4d8858a9e47b48f5333b9cbd` |
| archive | `maestro-v0.3.0-pc.1-linux-amd64.tar.gz` |
| dimensione | `3776699` byte |
| SHA-256 archive | `82bfb33f3fd9af911e3b2b1e89f9920177b281046da21b186512e577e114fb61` |
| SHA-256 binario M17 | `dee9d5113ccf2db0573b03e8a3851f600d7bc789964793ebae14376f9c849a66` |
| piattaforma live | WSL2 / Ubuntu 24.04 / RTX 5070 |
| provider e modello | Ollama 0.33.1 / `qwen3.5:9b` |
| digest modello | `6488c96fa5faab64bb65cbd30d4289e20e6130ef535a93ef9a49f42eda893ea7` |

Il checksum locale dell'archive M17 è stato verificato nuovamente. Il manifest
estratto registra commit, versione, stato `packaging-candidate`, Linux
`amd64`, Go 1.24.5, profilo e parametri esatti del record qualificato.

## Audit del delta

`70a9630` è antenato del source auditato. Il diff fino a `8bf5114` modifica
soltanto nove file documentali:

- `MAESTRO_CONTEXT.md`;
- `docs/compatibility.md`;
- `docs/milestone-17-direct-chat-development-plan.md`;
- `docs/milestone-18-productization-release-v0.3.0-plan.md`;
- `docs/milestone-18-productization-v0.4.0-plan.md`;
- `docs/releases/v0.3.0.md`;
- `docs/reports/milestone-17-final.md`;
- `docs/reports/milestone-17-phase-7.md`;
- `docs/roadmap.md`.

Il diff su `cmd/`, `internal/`, `pkg/`, `configs/`, `scripts/`, `go.mod` e
`go.sum` è vuoto. Gli hash della baseline coincidono sia con il record M17 sia
con i blob del commit qualificato:

| Superficie congelata | SHA-256 |
|---|---|
| `configs/maestro.chat.example.yaml` | `1c5bbe79edf125485d14518e58ff18c48156eaa0fb91faf82fcf3cd97375d0ee` |
| `internal/directchat/service.go` | `7fd79e1fafb70d0b7726ecca0909f92592f8706df890a9b6fb263c9d5b8575c1` |
| `internal/benchmark/developer/testdata/laravel-v1/routes/api.php` | `7e224d7e57bf0be6d2618e668d8515b07d332f94b7960de4640e4388b31bbc39` |

Il contratto resta chat-only, tool-free e read-only: zero o un file esplicito,
`workspace_mutate: deny`, nessun agent, retrieval, fallback o mutazione.

## Target Git e policy di tag

| Campo | Valore congelato |
|---|---|
| remote | `origin` |
| repository | `https://github.com/Axtonno/maestro.git` |
| branch di release | `master` |
| tracking all'apertura | `master` allineato a `origin/master` su `8bf5114` |
| tag finale | tag annotato `v0.3.0` sul commit release |

I tag locali precedenti `v0.1.0`, `v0.1.1` e `v0.2.0` sono annotati; non
esiste un tag locale `v0.3.0`. Il tag viene creato soltanto in Fase 6 dopo il
PASS dell'artifact finale e viene pushato soltanto in Fase 7. Non sono ammessi
retag, force push, rename o overwrite di asset. In questa fase non è stata
eseguita alcuna write remota.

## Inventario della superficie di release

La Fase 2 deve risolvere le sole differenze di stato pubblico già censite:

- `SECURITY.md` indica ancora v0.2.0 come release pubblicata corrente;
- `README.md`, `CHANGELOG.md`, `docs/compatibility.md` e
  `docs/releases/v0.3.0.md` usano ancora claim da candidate;
- le release notes dichiarano correttamente tag, artifact finale e
  pubblicazione non ancora eseguiti;
- `docs/cli.md` e `docs/configuration.md` qualificano ancora il contratto come
  sperimentale;
- `@MAESTRO_VERSION@` e `@MAESTRO_STATUS@` compaiono soltanto nei documenti
  template previsti e vengono risolti dal packaging;
- i link relativi della superficie pubblica censita puntano a file presenti.

Nessun token non previsto è entrato nell'archive verificato. Il gate di
packaging ha inoltre confermato allowlist, documenti obbligatori, assenza di
profili agentici/mutativi, path del checkout e credential-shaped data.

## Allowlist delle modifiche successive

### Fase 2

Sono modificabili soltanto:

- `README.md`, `CHANGELOG.md`, `SECURITY.md`;
- `docs/releases/v0.3.0.md`;
- `docs/installation.md`, `docs/quick-start.md`, `docs/cli.md` e
  `docs/configuration.md`;
- `docs/compatibility.md`, `docs/security-model.md`,
  `docs/known-issues.md` e `docs/troubleshooting.md`;
- `docs/packaging-candidate.md`;
- piano, roadmap, context e report della Milestone 18 necessari a registrare
  il gate.

Gli script di packaging sono congelati perché il gate corrente è verde. Una
necessità di modificarli non viene assorbita implicitamente: riapre questa
decisione, richiede causa documentata e ripetizione integrale dei gate locali.

### Fase 6

Dopo il PASS live del release candidate sono modificabili soltanto stato e
note di release in `README.md`, `CHANGELOG.md`, `SECURITY.md` e
`docs/releases/v0.3.0.md`, oltre a piano, roadmap, context e report M18. Stato,
versione, commit e manifest dell'artifact vengono incorporati attraverso il
packaging già congelato, non tramite modifiche al prodotto.

In entrambe le fasi restano immutabili `cmd/`, `internal/`, `pkg/`, `configs/`,
`scripts/`, fixture, `go.mod` e `go.sum`. Qualsiasi delta in queste superfici
invalida il gate e torna al relativo owner della Milestone 17.

## Gate locali

| Gate | Esito |
|---|---|
| `go test -count=3 ./...` | PASS |
| `go test -tags maestro_development -count=3 ./...` | PASS |
| `go test -race ./...` | PASS |
| `go vet ./...` | PASS |
| `bash -n scripts/*.sh` | PASS |
| checksum e manifest dell'archive M17 | PASS |
| doppio packaging dal source auditato | PASS — byte-identico |
| archive allowlist e file obbligatori | PASS |
| installazione temporanea, version/help e doctor offline | PASS |
| containment pre-provider | PASS |
| token, path checkout e credential-shaped scan | PASS |

Il primo avvio dei gate Go non ha eseguito test perché la cache globale del
sandbox era in sola lettura. La ripetizione con cache isolate sotto `/tmp` ha
eseguito integralmente i comandi sopra ed è quella autorevole. Analogamente,
il primo controllo checksum era stato lanciato dalla root anziché da `dist/`:
la ripetizione nel path previsto ha verificato l'archive con esito positivo.

Il packaging gate sul source `8bf5114` ha prodotto in directory temporanea un
archive byte-identico fra due build con SHA-256
`0c1bfd4f7419e2298f573098d329d80382d8d010d02ed13773621e14b200b23e`.
È una prova transitoria della baseline corrente, non un nuovo candidate e non
sostituisce l'archive M17 qualificato; la directory temporanea è stata rimossa
dallo script.

## Gate finale

- catena M17 e identità del candidate: **PASS**;
- zero delta funzionale: **PASS**;
- support claim ed esclusioni congelati: **PASS**;
- target Git, tag policy e allowlist modifiche: **PASS**;
- baseline repository-wide e packaging: **PASS**;
- blocker di release aperti: **zero**.

Verdetto Fase 1: **PASS**. La Fase 2 — Documentazione pubblica e metadata è
autorizzata. Nessun tag, release candidate persistente, artifact finale, push
o pubblicazione è stato eseguito.

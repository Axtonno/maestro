# Milestone 17 — Fase 6: candidate F6.4

Data: 2026-08-29

Stato: **CONGELATO — LIVE QUALIFICATION PENDING**

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

## Gate

F6.4 è **READY FOR LIVE**, non idoneo al packaging. Fase 7, archive, tag e
pubblicazione restano `NOT_RUN` fino al PASS integrale.

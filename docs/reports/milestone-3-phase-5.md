# Milestone 3 — Report finale Fase 5

Fase: Reporting & Hardware Profiles

Stato fase: Completata

Stato milestone: In corso — non chiusa

Data: 2026-08-09

---

# Obiettivo

Derivare una vista Markdown leggibile dal JSON canonico, uniformare il profilo
hardware dei benchmark e documentare un gate riproducibile senza introdurre
dipendenze esterne o chiudere automaticamente la milestone.

---

# Risultati consegnati

- Renderer Markdown deterministico basato sul report `1.2.0` validato.
- Decoder JSON strict con limite 64 MiB, rifiuto dei campi sconosciuti e dei
  documenti multipli.
- Comando `maestro bench render` con protezione da overwrite del JSON sorgente.
- Flag `--markdown` su Smoke, Provider, Model e Developer Benchmark.
- JSON e Markdown atomici con permessi `0600`.
- Riepilogo scenari senza score qualitativo globale.
- Aggregati e campioni, inclusi warmup, error code e rationale code.
- Profilo uniforme con OS, architettura, CPU logiche, CPU/RAM Linux e build
  metadata.
- GPU, backend e VRAM opt-in, senza tool o probe esterni.
- ADR-0021 e guida `benchmark-reporting.md`.

---

# Copertura deterministica

I test verificano:

- parsing procfs e normalizzazione dei metadata hardware;
- omissione dei campi non disponibili e rigetto VRAM invalida;
- round trip JSON strict verso Markdown;
- identità tra Markdown generato durante la run e rigenerato dal JSON;
- redazione di endpoint, credenziali e path;
- rigetto di campi JSON sconosciuti e documenti multipli;
- rigetto della collisione input/output;
- pubblicazione atomica e permessi `0600`;
- presenza di dataset, plugin, hardware, aggregati ed evaluation nella vista;
- CLI offline senza I/O provider.

---

# Verifiche

```text
GOCACHE=/tmp/maestro-go-build go test ./...
GOCACHE=/tmp/maestro-go-build go test -race ./...
GOCACHE=/tmp/maestro-go-build go vet ./...
GOCACHE=/tmp/maestro-go-build go run ./cmd/maestro bench validate --manifest docs/provider-smoke-benchmark-manifest.yaml
GOCACHE=/tmp/maestro-go-build go run ./cmd/maestro bench validate --manifest docs/runtime-benchmark-manifest.yaml
GOCACHE=/tmp/maestro-go-build go run ./cmd/maestro bench validate --manifest docs/developer-benchmark-manifest.yaml
GOCACHE=/tmp/maestro-go-build go run ./cmd/maestro bench laravel --provider ollama --warmup 0 --runs 1 --output /tmp/maestro-laravel.json --markdown /tmp/maestro-laravel.md
GOCACHE=/tmp/maestro-go-build go run ./cmd/maestro bench render --input /tmp/maestro-laravel.json --output /tmp/maestro-laravel-rendered.md
```

Esito: suite completa, race detector, vet e tre manifest superati. Smoke,
Provider, Model e Developer Benchmark hanno prodotto offline JSON e Markdown
`0600`. I due Markdown Developer, diretto e rigenerato dal JSON, hanno lo stesso
hash SHA-256. Il profilo Linux ha registrato OS, architettura, CPU logiche,
modello CPU e RAM totale; tutti gli scenari sono rimasti `skipped` senza I/O
provider.

---

# Verifica live

La Fase 5 non richiede un provider live: rendering e profilo sono verificabili
offline. Le run live Ollama/llama.cpp restano disponibili come verifica
successiva della configurazione dell'utente.

---

# Stato della Milestone 3

La Fase 5 è completata, ma la Milestone 3 non viene dichiarata chiusa. Rimane
**In corso — Fasi 1–5 completate** in attesa di una decisione esplicita e delle
eventuali run live desiderate.

# Maestro Benchmark Reporting & Hardware Profiles

Versione: 1.0.0

Stato: Fase 5 implementata; Milestone 3 completata

Ultimo aggiornamento: 2026-08-20

---

# Fonte canonica e rendering

Il report JSON `1.2.0` rimane la fonte canonica. Il Markdown è una vista
deterministica dello stesso contratto validato e redatto; non legge prompt,
risposte, embedding o workspace.

Un report esistente può essere renderizzato con:

```text
maestro bench render --input benchmark-report.json --output benchmark-report.md
```

Il decoder JSON:

- accetta soltanto lo schema corrente;
- rifiuta campi sconosciuti e documenti multipli;
- applica un limite di 64 MiB;
- valida report, campioni, aggregati ed evaluation prima del rendering;
- impedisce che input e output siano lo stesso file.

Ogni comando live può produrre entrambe le forme durante la stessa run:

```text
maestro bench model \
  --provider ollama \
  --runs 20 \
  --output model-report.json \
  --markdown model-report.md
```

`--markdown` richiede un file: stdout resta riservato al contratto JSON. JSON e
Markdown vengono scritti tramite file temporaneo, `fsync`, rename atomico e
permessi `0600`.

---

# Contenuto del Markdown

La vista contiene:

- identità, comando, timestamp, durata e versione schema;
- profilo hardware, provider, endpoint redatto, modelli, plugin e dataset;
- riepilogo per scenario senza score globale;
- aggregati minimo, mediana, p95 quando presente e massimo;
- campioni warmup e misurati;
- error kind/code, reason code e quality evaluation redatti.

Le evaluation vengono riassunte per scenario. Il renderer non costruisce una
classifica tra modelli e non calcola un singolo punteggio complessivo.

---

# Profilo hardware

Ogni comando benchmark usa lo stesso collector interno.

| Campo | Origine | Comportamento quando assente |
|---|---|---|
| OS, architettura | Go runtime | sempre presenti |
| CPU logiche | Go runtime | sempre presenti |
| modello CPU | `/proc/cpuinfo` su Linux | omesso |
| RAM totale | `MemTotal` da `/proc/meminfo` su Linux | omessa |
| GPU | `MAESTRO_BENCHMARK_GPU` | omessa |
| backend | `MAESTRO_BENCHMARK_BACKEND` | omesso |
| VRAM MiB | `MAESTRO_BENCHMARK_VRAM_MB` | omessa |
| versione/commit Maestro | Go build info/VCS settings | omessi |

Le variabili GPU sono override dichiarativi perché non esiste una sorgente
portabile e attribuibile senza tool esterni. `MAESTRO_BENCHMARK_VRAM_MB` deve
essere un intero positivo. I testi hardware vengono normalizzati su una
sola riga e limitati a 128 caratteri.

Il collector non esegue `nvidia-smi`, `rocminfo`, system profiler o richieste di
rete. Un valore non disponibile resta assente nel JSON e appare come `—` nel
Markdown.

Esempio esplicito:

```text
MAESTRO_BENCHMARK_GPU="NVIDIA RTX 4090"
MAESTRO_BENCHMARK_BACKEND="CUDA"
MAESTRO_BENCHMARK_VRAM_MB=24564
```

---

# Gate riproducibile della Fase 5

Il gate deterministico, indipendente da provider live, è:

```text
GOCACHE=/tmp/maestro-go-build go test ./...
GOCACHE=/tmp/maestro-go-build go test -race ./...
GOCACHE=/tmp/maestro-go-build go vet ./...
GOCACHE=/tmp/maestro-go-build go run ./cmd/maestro bench validate --manifest docs/provider-smoke-benchmark-manifest.yaml
GOCACHE=/tmp/maestro-go-build go run ./cmd/maestro bench validate --manifest docs/runtime-benchmark-manifest.yaml
GOCACHE=/tmp/maestro-go-build go run ./cmd/maestro bench validate --manifest docs/developer-benchmark-manifest.yaml
```

Segue una run offline con JSON e Markdown e una seconda derivazione dal JSON:

```text
GOCACHE=/tmp/maestro-go-build go run ./cmd/maestro bench laravel \
  --provider ollama --warmup 0 --runs 1 \
  --output /tmp/maestro-laravel.json \
  --markdown /tmp/maestro-laravel.md

GOCACHE=/tmp/maestro-go-build go run ./cmd/maestro bench render \
  --input /tmp/maestro-laravel.json \
  --output /tmp/maestro-laravel-rendered.md
```

I due Markdown devono essere identici. I report devono avere permessi `0600` e
gli scenari offline devono risultare `skipped` senza I/O provider.

---

# Stato della milestone

Il completamento della Fase 5 non chiudeva automaticamente la Milestone 3.
ADR-0030 registra la decisione live conclusiva: Ollama è la baseline positiva;
llama.cpp resta sperimentale/non supportato dopo preflight incompatibile e la
milestone è formalmente completata.

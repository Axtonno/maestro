# Maestro Benchmark Runtime

Versione: 0.3.0

Stato: Fasi 1–3 implementate

Ultimo aggiornamento: 2026-08-09

---

# Scopo

Il Benchmark Runtime esegue scenari locali in modo riproducibile e produce il
report raw versionato consumato dai livelli Smoke, Runtime e Developer della
Milestone 3.

Il layer orchestra scenari e misure. Non aggiunge semantiche agli adapter
provider, non confronta automaticamente modelli differenti e non invia
risultati a servizi remoti.

---

# Ownership e package

I contratti stabili appartengono a:

```text
pkg/benchmark
```

Il package pubblico contiene esclusivamente manifest, definizioni degli
scenari, iterazioni, misure, errori classificati, profili e report.

Esecuzione, parsing, aggregazione, redazione e serializzazione appartengono a:

```text
internal/benchmark
```

Questa separazione consente ai futuri scenari e plugin di implementare il
contratto `benchmark.Scenario` senza esporre gli invarianti del runner.

---

# Manifest

La prima versione del manifest è `1`. Il loader YAML:

- rifiuta campi sconosciuti;
- rifiuta documenti YAML multipli;
- conserva l'ordine degli scenari;
- richiede provider, scenari e owner non vuoti;
- rifiuta ID scenario duplicati;
- richiede gli stati `passed`, `failed`, `skipped` e `unsupported`.

Il manifest consegnato dalla Provider Layer rimane la fonte della matrice Smoke:

```text
docs/provider-smoke-benchmark-manifest.yaml
```

Il caricamento YAML usa `gopkg.in/yaml.v3`. La dipendenza resta confinata
all'implementazione interna; i contratti pubblici non dipendono dal parser.
Motivazione e trade-off sono registrati in ADR-0017.

---

# Semantica del runner

Ogni scenario dichiara una `ScenarioDefinition` identica a quella presente nel
manifest e implementa:

```go
Run(context.Context, benchmark.Iteration) (benchmark.IterationResult, error)
Cleanup(context.Context, benchmark.Iteration) error
```

Il runner:

1. valida manifest, opzioni e registry degli scenari;
2. esegue prima i warmup e poi le run misurate;
3. applica il timeout configurato a ogni iterazione;
4. classifica centralmente errori di contesto e `provider.ProviderError`;
5. esegue il cleanup dopo ogni iterazione, anche dopo errore o panic;
6. usa per il cleanup un contesto indipendente dalla cancellazione della run e
   con timeout dedicato;
7. interrompe lo scenario al primo stato non `passed` o errore di cleanup;
8. rappresenta uno scenario del manifest non ancora registrato come `skipped`
   con reason code `scenario_not_registered`.

Run e cleanup devono cooperare con la cancellazione del contesto. Il runner non
crea goroutine per interrompere forzatamente codice esterno che ignora il
contesto.

---

# Campioni e aggregati

Il report conserva anche i campioni di warmup, marcati esplicitamente, ma li
esclude dagli aggregati.

Le misure hanno nome, valore, unità e, quando necessari, perimetro e metodo. I
valori `NaN` e infiniti sono rifiutati. Misure assenti non vengono rappresentate
come zero.

Per ogni serie il runner calcola:

- minimo;
- mediana;
- massimo;
- percentile 95 soltanto da 20 campioni misurati in poi.

L'ordine degli aggregati è deterministico.

---

# Report JSON

La versione iniziale dello schema era `1.0.0`. La Fase 2 lo estende in modo
additivo a `1.1.0` con profili modello per ruolo; la Fase 4 introduce la sezione
qualitativa separata e porta il contratto a `1.2.0`. Il contratto corrente è pubblicato
in:

```text
docs/schemas/benchmark-report-v1.schema.json
```

Ogni report distingue:

- identità, timestamp e durata della run;
- versione del manifest e del Runtime;
- configurazione hardware–provider–modello–plugin;
- warmup, numero di run e timeout;
- scenari, campioni e aggregati;
- errori classificati e reason code.

La serializzazione supportata è `internal/benchmark.EncodeReportJSON`, che
valida e redige una copia prima di scrivere qualsiasi byte.

---

# Redazione

Il report non possiede campi per prompt, risposte, chunk, embedding, argomenti o
risultati dei tool, credenziali e messaggi remoti.

In fase di serializzazione:

- user info, path, query e fragment vengono rimossi dagli endpoint;
- model ID e dataset ID che corrispondono a path assoluti o directory utente diventano
  `[redacted-path]`;
- la stessa regola viene applicata ai model ID negli errori classificati;
- il report sorgente non viene modificato.

`RunMetadata.Command` contiene soltanto il nome logico del comando, mai la
command line completa.

---

# CLI

La base dei subcommand è disponibile attraverso:

```text
maestro bench --help
maestro bench validate --manifest docs/provider-smoke-benchmark-manifest.yaml
maestro bench provider --provider ollama
maestro bench model --provider ollama
```

`validate` restituisce `0` per un manifest valido, `1` per un manifest non
caricabile o invalido e `2` per un uso non valido della CLI.

`provider` seleziona dal manifest soltanto introspection, catalogo, retry e
circuit breaker. `model` seleziona completion, streaming, cancellazione,
embedding, lifecycle, cold/warm e cancellazione del pull. Entrambi producono lo
stesso report JSON versionato e supportano output atomico su file.

Il manifest Runtime è:

```text
docs/runtime-benchmark-manifest.yaml
```

Per ottenere il p95 occorre configurare almeno 20 run misurate:

```text
maestro bench model --warmup 2 --runs 20 --output model-report.json
```

I default sono un warmup, cinque run, timeout di cinque minuti e cleanup di
trenta secondi. Una run non configurata non tenta connessioni implicite a
localhost: gli scenari risultano `skipped`.

---

# Runtime Benchmark della Fase 3

## Scenari provider

| Scenario | Misure principali |
|---|---|
| `provider-capability-latency` | latenza e numero di capability |
| `provider-catalog-latency` | latenza listing/discovery e cardinalità |
| `provider-retry-controlled` | latenza di recupero, tentativi e retry |
| `provider-circuit-breaker` | tempo di apertura e chiamate bloccate |

Retry e circuit breaker usano un wrapper interno che inietta soltanto errori
transienti controllati. Il retry delega il secondo tentativo al provider live;
il circuit breaker verifica che la chiamata successiva venga bloccata dal
Runtime senza raggiungere l'adapter. Policy e fault mode vengono ripristinati
nel cleanup di ogni iterazione.

## Scenari modello

| Scenario | Misure principali |
|---|---|
| `model-completion-latency` | latenza totale e token dichiarati |
| `model-stream-performance` | TTFT, latenza totale, chunk e token/sec quando disponibili |
| `model-stream-cancellation` | latenza di cancellazione |
| `model-embedding-performance` | latenza, batch, dimensione ed embedding/sec |
| `model-lifecycle-load-unload` | latenza di load e unload |
| `model-cold-warm` | prima completion dopo unload e completion warm |
| `model-pull-cancellation` | latenza di cancellazione del download |

Il throughput generativo viene emesso soltanto quando lo stream espone usage
provider-reported. Il lifecycle usa le capability pubbliche load/unload: non
installa policy di residency che il Runtime non potrebbe eliminare a fine run.

## Risorse e perimetro

Su Linux il sampler procfs raccoglie, quando leggibili:

- `peak_memory_mb`;
- `cpu_avg_percent`;
- `cpu_peak_percent`.

Per default il perimetro è `maestro_process`. Passando `--provider-pid PID` il
perimetro diventa `provider_process`. Metodo e scope sono registrati su ogni
misura. Su altri sistemi le misure vengono omesse. La VRAM non viene stimata né
rappresentata come zero, perché non è disponibile un'origine portabile e
attribuibile senza strumenti esterni.

## Sicurezza del pull cancellato

`model-pull-cancellation` richiede contemporaneamente:

- `MAESTRO_ALLOW_CATALOG_MUTATION=true`;
- un modello `acquisition_fixture` esplicito;
- fixture assente dal catalogo prima della run;
- supporto a discovery, pull e remove.

Il cleanup chiude lo stream e rimuove soltanto la fixture di cui la run ha
acquisito ownership.

---

# Limiti correnti

Il Runtime Benchmark non include:

- raccolta VRAM portabile;
- sampler di sistema per macOS o Windows;
- rendering Markdown.

Il rendering Markdown e i profili hardware completi appartengono alla Fase 5.

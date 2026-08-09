# Maestro Benchmark Runtime

Versione: 0.1.0

Stato: Fase 1 implementata

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
additivo a `1.1.0` con profili modello per ruolo. Il contratto corrente è pubblicato
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

# CLI della Fase 1

La base dei subcommand è disponibile attraverso:

```text
maestro bench --help
maestro bench validate --manifest docs/provider-smoke-benchmark-manifest.yaml
```

`validate` restituisce `0` per un manifest valido, `1` per un manifest non
caricabile o invalido e `2` per un uso non valido della CLI.

I comandi live `smoke`, `provider`, `model` e `laravel` vengono aggiunti nelle
rispettive fasi della milestone.

---

# Limiti della Fase 1

La prima fase non include:

- esecuzione live di Ollama o llama.cpp;
- raccolta di CPU, RAM o VRAM;
- misure di streaming ed embedding;
- dataset PHP/Laravel;
- rendering Markdown.

Questi incrementi appartengono alle Fasi 2–5.

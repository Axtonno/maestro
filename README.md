# Maestro

> The intelligence is in the orchestration.

Maestro è una workstation AI locale per lo sviluppo software. Usa Ollama sul
computer dell’utente e mantiene le modifiche entro un confine esplicito: file,
righe, preview e approvazione restano sotto controllo umano.

## Quick start

Con una build che include il comando `setup`:

```sh
cd /percorso/del/progetto
maestro setup
maestro chat "Come puoi aiutarmi?"
```

Per preparare una modifica senza scrivere:

```sh
maestro mutate --preview \
  --file app/Services/Example.php \
  --lines 10:12 \
  "Semplifica questo blocco senza cambiarne il comportamento"
```

Il percorso completo è nel [Quick Start](docs/quick-start.md). Per l’archive
immutabile v0.5.0 seguire le [relative note di release](docs/releases/v0.5.0.md).

## Capacità correnti

- Direct Chat senza file oppure su un singolo file scelto dall’utente;
- Controlled Mutation su un intervallo esplicito di un file PHP sotto `app/`;
- preview completa, allow-once, controllo stale e scrittura atomica;
- esecuzione locale tramite Ollama.

Maestro non sceglie autonomamente file o righe, non esegue shell o Git per
conto del modello, non applica modifiche multi-file e non è una sandbox. La
[matrice delle capacità](docs/current-capabilities.md) definisce il perimetro
supportato e prevale sugli esempi e sul codice sperimentale.

## Documentazione

- [Quick Start](docs/quick-start.md)
- [Installazione](docs/installation.md)
- [Capacità correnti](docs/current-capabilities.md)
- [Controlled Mutation](docs/controlled-mutation.md)
- [Piattaforme supportate](docs/supported-platforms.md)
- [Benchmark](docs/benchmarks.md)
- [Sicurezza](docs/security-model.md)
- [Troubleshooting](docs/troubleshooting.md)
- [CLI](docs/cli.md) e [configurazione](docs/configuration.md)
- [Roadmap pubblica](docs/roadmap.md)
- [Contribuire](CONTRIBUTING.md)

## Sviluppo

Servono Go `1.24.5` e GNU userland per eseguire tutti i controlli:

```sh
go test ./...
go test -race ./...
go vet ./...
```

I test live sono opt-in e `not_run` non equivale a PASS.

## Licenza

Apache License 2.0. Attribution in [NOTICE](NOTICE) e
[THIRD_PARTY_LICENSES.txt](THIRD_PARTY_LICENSES.txt).

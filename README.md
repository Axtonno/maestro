# Maestro

> The intelligence is in the orchestration.

Maestro è una workstation AI locale per lo sviluppo software. Usa Ollama sul
computer dell'utente e mantiene ogni modifica entro un confine esplicito:
file, righe, preview e approvazione restano sotto controllo umano.

## Inizia qui

Dopo aver installato un build che include la Milestone 41:

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

`setup` crea la configurazione utente, verifica Ollama e i modelli consigliati
e, quando mancano, chiede prima di scaricarli. Il dry-run stampa la proposta e
termina senza richiedere approvazione e senza modificare file.

La chat senza `--file` non legge automaticamente il progetto. Per domande sul
codice, indica un file esplicito come mostrato nel Quick Start.

Il percorso completo è nel [Quick Start](docs/quick-start.md). L'archive
pubblico v0.5.0 è immutabile e precede questi due comandi semplificati: per
quell'asset usare la [guida v0.5.0](docs/releases/v0.5.0.md). La prima release
che includerà M41 dovrà superare il gate di packaging prima che il nuovo
percorso sia dichiarato pubblico.

## Cosa fa

- Direct Chat senza file o su un singolo file scelto dall'utente;
- Controlled Mutation su un intervallo esplicito di un file PHP sotto `app/`;
- preview completa, allow-once, controllo stale e scrittura atomica;
- esecuzione locale con Ollama e modelli distinti per chat e modifica.

Maestro non sceglie autonomamente file o righe, non esegue shell o Git per
conto del modello, non applica modifiche multi-file e non è una sandbox. La
[pagina delle capacità correnti](docs/current-capabilities.md) resta il
riferimento autorevole per release, piattaforme e limiti.

## Documentazione

- [Quick Start](docs/quick-start.md) — primo utilizzo;
- [Installazione](docs/installation.md) — binary e aggiornamenti;
- [First Controlled Mutation](docs/quick-start.md#prima-controlled-mutation) —
  preview e apply;
- [Troubleshooting](docs/troubleshooting.md) — `doctor`, errori e supporto;
- [Validation Guide](docs/validation.md) — prove riproducibili e release;
- [CLI](docs/cli.md) e [configurazione](docs/configuration.md) — riferimento;
- [Roadmap](docs/roadmap.md) — stato delle milestone.

## Sviluppo

Requisiti: Go `1.24.5` e GNU userland per il packaging riproducibile.

```sh
go test ./...
go test -race ./...
go vet ./...
```

I test live sono opt-in e `not_run` non equivale a PASS.

## Licenza

Apache License 2.0. Attribution in [NOTICE](NOTICE) e
[THIRD_PARTY_LICENSES.txt](THIRD_PARTY_LICENSES.txt).

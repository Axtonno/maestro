# Maestro

> The intelligence is in the orchestration.

Maestro v0.3.0 introduce una chat locale diretta per interrogare zero o un file
esplicitamente scelto dentro un workspace. Il percorso supportato è read-only,
non usa tool, retrieval, Agent Runtime o fallback agentici.

## v0.3.0 in breve

| Dimensione | Supporto ufficiale |
|---|---|
| Piattaforma | Linux `amd64`, qualificazione WSL2/Ubuntu 24.04/RTX 5070 |
| Provider | Ollama locale 0.33.1 su endpoint loopback |
| Modello | `qwen3.5:9b`, digest qualificato nel manifest |
| Modalità | `maestro chat`, zero o un file esplicito |
| Streaming | Opt-in con output atomico |
| Tool/retrieval/agent | Non qualificati per v0.3.0 |
| Mutazioni | Non supportate |
| Isolamento | Trusted in-process, nessuna sandbox |

La [compatibility matrix](docs/compatibility.md) è la fonte autorevole. Codice
agentico o mutativo presente nel repository non amplia la promessa di prodotto.

## Installazione dall’artifact

Scaricare insieme archive e checksum del candidate o della release:

```sh
version=v0.3.0
artifact="maestro-${version}-linux-amd64"
sha256sum -c "${artifact}.tar.gz.sha256"
tar -xzf "${artifact}.tar.gz"
cd "$artifact"
./maestro version
```

Versione, commit e stato devono coincidere con `ARTIFACT-MANIFEST.txt`.

## Quick start

Prerequisiti: Ollama già attivo su `127.0.0.1:11434` e `qwen3.5:9b` con il
digest qualificato già disponibile. Maestro non avvia il provider e non
scarica o sostituisce modelli.

```sh
./maestro doctor --mode chat --config ./configs/maestro.chat.example.yaml

./maestro chat --config ./configs/maestro.chat.example.yaml \
  --file routes/api.php \
  "Quali endpoint, controller e action sono dichiarati?"
```

La configurazione inclusa punta alla fixture dell’archive. Il doctor chat deve
completare cinque check; la risposta deve riportare `POST /orders` e
`OrderController::store` senza aggiungere endpoint. Per il trasporto streaming
aggiungere `--stream`.

## Contratto Direct Chat

```text
maestro chat [--config path] [--file logical-path] [--stream] [question]
maestro doctor --mode chat [--config path]
maestro version
```

- la domanda è posizionale oppure arriva da stdin bounded, mai da entrambi;
- `--file` accetta un solo path logico relativo e contained;
- file assente significa nessun contesto workspace implicito;
- complete e stream usano temperatura zero, context 4096 e thinking disabilitato;
- stdout viene pubblicato soltanto dopo una response valida e terminale;
- i failure non avviano retrieval, tool o agent come percorso alternativo.

La reference completa è in [CLI](docs/cli.md) e
[Configurazione](docs/configuration.md).

## Sicurezza

Il file selezionato viene inviato al provider esplicitamente configurato. Il
loader rifiuta path assoluti, traversal, symlink, file non regolari, UTF-8
invalido e input oltre limite. Il contenuto è trattato come evidenza non
attendibile e non può concedere autorità.

Maestro resta un processo locale con i privilegi dell’utente: non offre
sandbox, isolamento di rete o secret manager. Consultare il
[Security Model](docs/security-model.md) prima di usare workspace sensibili.

## Limiti noti

- il modello è generativo: il gate di Milestone 17 ha ottenuto qualità 4/5;
- soltanto la combinazione riportata nella compatibility matrix è qualificata;
- agent, retrieval multi-file, tool calling e mutazioni non sono supportati;
- non esistono installer privilegiato, auto-update o download automatici;
- CLI e schema restano sperimentali durante la serie 0.x.

Vedere [Known Issues](docs/known-issues.md) e
[Troubleshooting](docs/troubleshooting.md).

## Documentazione

- [Installazione](docs/installation.md)
- [Quick Start](docs/quick-start.md)
- [Configurazione](docs/configuration.md)
- [CLI](docs/cli.md)
- [Security Model](docs/security-model.md)
- [Compatibility Matrix](docs/compatibility.md)
- [Release Notes v0.3.0](docs/releases/v0.3.0.md)
- [Changelog](CHANGELOG.md)

## Sviluppo

Requisiti: Go `1.24.5` e GNU userland per il packaging riproducibile.

```sh
go test ./...
go test -race ./...
go vet ./...
```

I test live sono opt-in e `not_run` non equivale a PASS. Maestro e i test non
avviano provider, non installano modelli e non modificano il catalogo.

## Licenza

Maestro è distribuito sotto [Apache License 2.0](LICENSE). Le attribution sono
in [NOTICE](NOTICE) e [THIRD_PARTY_LICENSES.txt](THIRD_PARTY_LICENSES.txt).

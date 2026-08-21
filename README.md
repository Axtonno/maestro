# Maestro

> The intelligence is in the orchestration.

Maestro è un runtime locale per agenti di sviluppo controllati. La v0.1.1
offre un primo percorso installabile per analizzare e interrogare un progetto
Laravel con un modello locale, senza consentire modifiche al workspace nel
profilo ufficiale.

## v0.1.1 in breve

| Dimensione | Supporto ufficiale |
|---|---|
| Piattaforma | Linux `amd64` |
| Provider | Ollama locale |
| Modello chat | `llama3.1:8b` |
| Modello embedding | `embeddinggemma:latest` |
| Agent | `agent.reference` Laravel read-only |
| Tool | list, read, search |
| Mutazioni | Sperimentali/non supportate |
| llama.cpp | Sperimentale/non supportato |
| Isolamento | Trusted in-process, nessuna sandbox |

La [compatibility matrix](docs/compatibility.md) è la fonte autorevole. La
presenza di altre capability nel codice non costituisce una promessa di
supporto v0.1.x.

## Installazione dall’artifact

Scaricare dalla stessa release:

```text
maestro-v0.1.1-linux-amd64.tar.gz
maestro-v0.1.1-linux-amd64.tar.gz.sha256
```

Poi eseguire:

```sh
version=v0.1.1
artifact="maestro-${version}-linux-amd64"
sha256sum -c "${artifact}.tar.gz.sha256"
tar -xzf "${artifact}.tar.gz"
cd "$artifact"
./maestro version
```

Non proseguire se il checksum, la versione o il commit non coincidono con
`ARTIFACT-MANIFEST.txt`.

## Quick start

Prerequisiti: Ollama attivo su `127.0.0.1:11434` e `llama3.1:8b` disponibile.
Maestro non installa il provider e non scarica modelli implicitamente.

```sh
ollama pull llama3.1:8b

./maestro doctor --config ./configs/maestro.example.yaml
./maestro models --config ./configs/maestro.example.yaml
./maestro agents --config ./configs/maestro.example.yaml

./maestro run --config ./configs/maestro.example.yaml \
  "Read app/Http/Controllers/OrderController.php and explain which service its store method calls. Do not modify any file."
```

La configurazione inclusa punta alla fixture Laravel dell’archive. `doctor`
deve completare nove check; il run deve identificare `OrderService::create`.
L’output esatto e la latenza possono variare perché il modello è generativo.

Il percorso dettagliato, inclusa la configurazione di un progetto reale, è in
[Quick Start](docs/quick-start.md).

## CLI

```text
maestro doctor
maestro models
maestro agents
maestro run
maestro version
```

- `doctor` valida configurazione, provider, modello, agent, tool e workspace
  senza invocare il modello;
- `models` elenca i modelli del provider esplicitamente configurato;
- `agents` elenca gli agenti registrati;
- `run` esegue il reference agent con policy e hard limit espliciti;
- `version` stampa versione e commit incorporati nella build.

SIGINT/SIGTERM cancellano una run con exit code 130. stdout è riservato al
risultato; progresso e failure sintetici vanno su stderr. La reference completa
è in [CLI](docs/cli.md).

## Configurazione

Maestro usa un singolo documento YAML strict `version: 1`. Provider, modello,
workspace, agent, policy, tool e limiti sono sempre espliciti. Un typo o campo
sconosciuto è un errore, non un fallback.

Il profilo distribuito registra soltanto:

```yaml
agent:
  id: agent.reference
  streaming: true
  tools:
    - workspace.list
    - workspace.read
    - workspace.search

policy:
  id: policy.local-review
  model: allow
  workspace_inspect: allow
  workspace_mutate: deny
```

Non aggiungere `workspace.write` o `workspace.patch` al percorso v0.1.x. La
guida campo per campo è in [Configurazione](docs/configuration.md).

## Sicurezza

Maestro v0.1.x è trusted in-process e usa i privilegi dell’utente locale. Non
offre sandbox, isolamento di rete, rollback generale o secret manager. Il
provider configurato riceve l’istruzione e le sezioni di workspace selezionate
dal Context Engine.

Il profilo read-only, il containment dei path, il permission model, la
redazione e i hard limit riducono l’autorità disponibile, ma non trasformano il
processo in un confine di sicurezza. Leggere il [Security Model](docs/security-model.md)
e la [Security Policy](SECURITY.md) prima di usare workspace sensibili.

## Limiti noti

- solo Linux `amd64` e Ollama/`llama3.1:8b` sono qualificati;
- una run CPU-only può richiedere diversi minuti;
- il modello può sbagliare o generare una tool call invalida;
- llama.cpp e il reference agent mutante non sono supportati;
- nessun multi-agent, persistence, recovery, shell, Git, Docker o remote
  execution completi;
- CLI, config e package Go sono sperimentali durante la serie 0.x.

Vedere [Known Issues](docs/known-issues.md) e
[Troubleshooting](docs/troubleshooting.md).

La Milestone 11 ha confermato deterministicamente gli invarianti della
Controlled Mutation, ma il candidato live ha fallito Gate A. ADR-0032 rinvia
quindi la mutazione: la Milestone 12 procede soltanto verso una v0.2.0
read-only e non modifica la matrice di supporto sopra dichiarata.

## Documentazione

Per iniziare:

- [Installazione](docs/installation.md)
- [Quick Start](docs/quick-start.md)
- [Configurazione](docs/configuration.md)
- [CLI](docs/cli.md)
- [Reference Agent Laravel](docs/reference-agent-laravel.md)
- [Security Model](docs/security-model.md)
- [Compatibility Matrix](docs/compatibility.md)
- [API Compatibility](docs/v0.1.0-api-compatibility.md)
- [Release Notes v0.1.1](docs/releases/v0.1.1.md)
- [Release Notes v0.1.0](docs/releases/v0.1.0.md)
- [Changelog](CHANGELOG.md)

Per contribuire o studiare l’architettura:

- [Architecture](https://github.com/Axtonno/maestro/blob/v0.1.1/docs/architecture.md)
- [Runtime Internals](https://github.com/Axtonno/maestro/blob/v0.1.1/docs/runtime-internals.md)
- [Provider Runtime](https://github.com/Axtonno/maestro/blob/v0.1.1/docs/provider-runtime.md)
- [Plugin Runtime](https://github.com/Axtonno/maestro/blob/v0.1.1/docs/plugin-runtime.md)
- [Context Engine](https://github.com/Axtonno/maestro/blob/v0.1.1/docs/context-engine-runtime.md)
- [Agent Runtime](https://github.com/Axtonno/maestro/blob/v0.1.1/docs/agent-runtime.md)
- [Roadmap](https://github.com/Axtonno/maestro/blob/v0.1.1/docs/roadmap.md)

## Sviluppo

Requisiti: Go `1.24.5` e GNU userland per il packaging riproducibile.

```sh
go test ./...
go test -race ./...
go vet ./...
```

I test live sono opt-in e non vengono sostituiti da PASS sintetici quando il
provider è assente. Le modifiche ai contratti pubblici devono aggiornare
documentazione, test e changelog.

## Licenza

Maestro è distribuito sotto [Apache License 2.0](LICENSE). Le attribution delle
dipendenze sono in [NOTICE](NOTICE) e
[THIRD_PARTY_LICENSES.txt](THIRD_PARTY_LICENSES.txt).

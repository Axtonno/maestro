# Maestro v0.1.0 Quick Start

Questo percorso parte esclusivamente dall'archive Linux `amd64` e dal checksum
pubblicati. Non richiede il checkout del repository.

## Prerequisiti

- Linux `amd64`;
- Ollama raggiungibile su `http://127.0.0.1:11434`;
- `llama3.1:8b` disponibile;
- spazio e memoria sufficienti per il modello.

Il modello embedding non è usato dal quick start lessicale. Per gli scenari che
lo richiedono, la fixture qualificata è `embeddinggemma:latest`.

## 1. Verifica ed estrazione

Eseguire nella directory che contiene i due file scaricati:

```sh
version=v0.1.0
artifact="maestro-${version}-linux-amd64"
sha256sum -c "${artifact}.tar.gz.sha256"
tar -xzf "${artifact}.tar.gz"
cd "$artifact"
./maestro version
```

`maestro version`, `ARTIFACT-MANIFEST.txt`, nome archive e release devono
riportare la stessa versione. Non proseguire se il checksum fallisce.

## 2. Modello locale

Se il modello non è già presente:

```sh
ollama pull llama3.1:8b
```

Maestro non installa Ollama e non esegue pull impliciti.

## 3. Diagnostica

La configurazione inclusa punta alla fixture Laravel inclusa nello stesso
archive:

```sh
./maestro doctor --config ./configs/maestro.example.yaml
./maestro models --config ./configs/maestro.example.yaml
./maestro agents --config ./configs/maestro.example.yaml
```

`doctor` deve riportare nove righe `pass`. `models` deve includere
`llama3.1:8b`; `agents` deve includere `agent.reference`.

## 4. Prima run read-only

```sh
./maestro run --config ./configs/maestro.example.yaml \
  "Read app/Http/Controllers/OrderController.php and explain which service its store method calls. Do not modify any file."
```

Il risultato deve terminare `completed` e identificare
`OrderService::create`. Tempi e formulazione possono variare: il modello è
generativo e su CPU una run può richiedere diversi minuti.

Il profilo incluso registra soltanto:

```text
workspace.list
workspace.read
workspace.search
```

e imposta `workspace_mutate: deny`. Non aggiungere tool mutanti al quick start.

## 5. Progetto Laravel reale

Copiare la configurazione e sostituire `workspace.root` con la root del
progetto, cioè la directory contenente `artisan` e `composer.json`:

```sh
install -d "$HOME/.config/maestro"
install -m 0600 ./configs/maestro.example.yaml \
  "$HOME/.config/maestro/config.yaml"
```

Modificare il file, quindi verificare prima di eseguire:

```sh
./maestro doctor --config "$HOME/.config/maestro/config.yaml"
./maestro run --config "$HOME/.config/maestro/config.yaml" \
  "Explain the request flow for this Laravel project. Do not modify any file."
```

Un path relativo è risolto rispetto alla directory del file YAML. Per ridurre
ambiguità è consigliato un path assoluto per un progetto reale.

## Arresto e problemi

SIGINT o SIGTERM cancellano una run e producono exit code 130. Per errori di
provider, modello, configurazione o tool consultare `troubleshooting.md`. Il
security model e i limiti di supporto sono in `security-model.md` e
`compatibility.md`.

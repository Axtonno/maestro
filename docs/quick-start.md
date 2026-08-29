# Maestro @MAESTRO_VERSION@ Quick Start

Questo percorso usa soltanto archive e checksum Linux `amd64`. Non richiede il
checkout del repository.

## Prerequisiti

- Linux `amd64`;
- Ollama 0.33.1 raggiungibile su `http://127.0.0.1:11434`;
- `qwen3.5:9b`, digest
  `6488c96fa5faab64bb65cbd30d4289e20e6130ef535a93ef9a49f42eda893ea7`,
  già disponibile;
- memoria sufficiente per il modello.

Maestro non avvia Ollama e non esegue pull, update o sostituzioni implicite.

## 1. Verifica ed estrazione

```sh
version=@MAESTRO_VERSION@
artifact="maestro-${version}-linux-amd64"
sha256sum -c "${artifact}.tar.gz.sha256"
tar -xzf "${artifact}.tar.gz"
cd "$artifact"
./maestro version
```

Nome archive, checksum, `maestro version` e `ARTIFACT-MANIFEST.txt` devono
riportare la stessa identità. Non proseguire in caso di divergenza.

## 2. Verifica del modello

Usare gli strumenti amministrativi di Ollama per verificare che modello e
digest coincidano con il manifest. Se il modello non è presente, interrompere
il quick start: l’acquisizione non fa parte della procedura supportata.

## 3. Diagnostica

La configurazione inclusa punta alla fixture Laravel nello stesso archive:

```sh
./maestro doctor --mode chat --config ./configs/maestro.chat.example.yaml
```

Il doctor esegue cinque controlli read-only: config, workspace, composition,
model e generation. Tutti devono essere `pass`; non effettua completion.

## 4. Chat single-file

```sh
./maestro chat --config ./configs/maestro.chat.example.yaml \
  --file routes/api.php \
  "Quali endpoint, controller e action sono dichiarati?"
```

Il terminale deve essere `completed`, il finish reason `stop` e la risposta
deve identificare `POST /orders` e `OrderController::store` senza inventare
altri endpoint. L’output esatto e la latenza possono variare.

Per verificare il trasporto streaming mantenendo lo stesso contratto:

```sh
./maestro chat --stream \
  --config ./configs/maestro.chat.example.yaml \
  --file routes/api.php \
  "Quali endpoint, controller e action sono dichiarati?"
```

I chunk non vengono esposti progressivamente: Maestro pubblica stdout soltanto
dopo aver validato terminale, limiti e risposta completa.

## 5. Nessun file

```sh
./maestro chat --config ./configs/maestro.chat.example.yaml \
  "Quali endpoint dichiara questo progetto?"
```

La risposta deve dichiarare che il contesto di progetto non è disponibile.
Maestro non seleziona file, non indicizza il workspace e non usa fallback.

## 6. Progetto reale

```sh
install -d "$HOME/.config/maestro"
install -m 0600 ./configs/maestro.chat.example.yaml \
  "$HOME/.config/maestro/chat.yaml"
```

Modificare soltanto `workspace.root` verso la directory autorizzata, poi
ripetere il doctor. Un path relativo è risolto rispetto al file YAML.

Il comando legge esclusivamente il file indicato da `--file`. Path assoluti,
traversal, directory, symlink, file non regolari o oltre limite vengono
rifiutati prima della disclosure.

## Arresto e problemi

SIGINT/SIGTERM producono exit code 130; una deadline provider usa exit code 4.
I failure stampano su stderr soltanto `chat failed: <reason_code>` e non
pubblicano response parziali. Consultare `troubleshooting.md`,
`security-model.md` e `compatibility.md`.

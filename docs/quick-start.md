# Maestro @MAESTRO_VERSION@ Quick Start

Questo percorso verifica la release Linux `amd64` sulla fixture Laravel
inclusa, senza checkout del repository.

Per una prova essenziale vedere [Installa e prova](install-and-try.md). Questa
pagina aggiunge controlli e risultati attesi.

## Prerequisiti

- Ollama 0.33.1 su `http://127.0.0.1:11434`;
- `qwen3.5:9b`, digest
  `6488c96fa5faab64bb65cbd30d4289e20e6130ef535a93ef9a49f42eda893ea7`;
- `qwen2.5-coder:14b`, digest
  `9ec8897f747e246e970bc5cfdda85d22f1123dc2e3d34978a010a75968716849`;
- memoria sufficiente a caricare un modello per volta;
- una TTY reale per il write-mode.

Maestro non avvia Ollama, non scarica modelli e non sostituisce un digest.

## 1. Verifica ed estrazione

```sh
version=@MAESTRO_VERSION@
artifact="maestro-${version}-linux-amd64"
base_url="https://github.com/Axtonno/maestro/releases/download/${version}"
curl -fLO "${base_url}/${artifact}.tar.gz"
curl -fLO "${base_url}/${artifact}.tar.gz.sha256"
sha256sum -c "${artifact}.tar.gz.sha256"
tar -xzf "${artifact}.tar.gz"
cd "${artifact}"
./maestro version --diagnostic
```

Archive e checksum devono provenire dalla stessa GitHub Release. Nome,
versione, commit, stato `release` e manifest devono coincidere.

## 2. Diagnostica completa

```sh
./maestro doctor --mode all --config ./configs/maestro.v0.5.0-candidate.yaml
```

Il profilo v4 punta alla fixture inclusa. I 14 check chat e mutation devono
essere `pass`. Doctor non esegue completion e non modifica il workspace.

## 3. Direct Chat con file

```sh
./maestro chat --config ./configs/maestro.v0.5.0-candidate.yaml --file app/Http/Controllers/OrderController.php "Quali campi valida store e quale risposta HTTP restituisce?"
```

Il terminale deve essere `completed`, il modello `qwen3.5:9b` e il finish
reason `stop`. La risposta deve ricavare dal solo file `customer_id`,
`items` e lo status 201.

Senza `--file`, Maestro non cerca contesto nel progetto. Una domanda che
dipende dal workspace deve quindi essere dichiarata non determinabile.

## 4. Controlled Mutation con deny

```sh
./maestro workspace replace --config ./configs/maestro.v0.5.0-candidate.yaml --file app/Http/Controllers/OrderController.php --lines 22:22 "Cambia soltanto lo status HTTP da 201 a 202, preservando il resto della riga."
```

Il modello deve essere `qwen2.5-coder:14b`. Controllare preview, digest e
fingerprint, poi digitare `d`: il terminale atteso è
`approval_rejected`, exit code 3, con file invariato.

## 5. Controlled Mutation con allow-once

Ripetere il comando e digitare `o` solo se la preview sostituisce esattamente
201 con 202. Il terminale atteso è `applied`, `effect=applied` e
`durable=true`. Verificare la riga risultante:

```sh
sed -n '22p' app/Http/Controllers/OrderController.php
```

La fixture estratta può essere reinizializzata riestraendo l'archive in una
nuova directory. Su un progetto reale usare invece il normale controllo
versione del progetto.

## 6. Configurare un progetto reale

```sh
install -d "$HOME/.config/maestro"
install -m 0600 ./configs/maestro.v0.5.0-candidate.yaml "$HOME/.config/maestro/v0.5.0.yaml"
```

Cambiare soltanto `workspace.root`, quindi rieseguire doctor. Un root relativo
è risolto rispetto alla directory del file YAML.

Direct Chat legge zero o un file esplicito. Controlled Mutation accetta un
singolo file PHP sotto `app/` e un solo intervallo inclusivo. Per il contratto
completo vedere [Controlled Mutation: perimetro supportato](controlled-mutation-support.md).

## Arresto e problemi

SIGINT/SIGTERM producono exit code 130; una deadline provider usa exit code 4.
Un deny o una sorgente stale usa exit code 3. I failure non avviano un secondo
modello o un altro percorso.

Consultare [Troubleshooting](troubleshooting.md),
[Security Model](security-model.md) e [Compatibility Matrix](compatibility.md).

# Installazione di Maestro @MAESTRO_VERSION@

Stato: @MAESTRO_STATUS@

Piattaforma verificata: Linux `amd64`

## Installa il binario

Scarica archive e checksum dalla stessa GitHub Release:

```sh
version=@MAESTRO_VERSION@
artifact="maestro-${version}-linux-amd64"
base_url="https://github.com/Axtonno/maestro/releases/download/${version}"
curl -fLO "${base_url}/${artifact}.tar.gz"
curl -fLO "${base_url}/${artifact}.tar.gz.sha256"
sha256sum -c "${artifact}.tar.gz.sha256"
tar -xzf "${artifact}.tar.gz"
install -d "$HOME/.local/bin"
install -m 0755 "${artifact}/maestro" "$HOME/.local/bin/maestro"
maestro version
```

Se `~/.local/bin` non è nel `PATH`, aggiungerlo secondo le regole della propria
shell. Maestro non richiede né invoca `sudo`.

## Configura il progetto

Con Ollama già avviato:

```sh
cd /percorso/del/progetto
maestro setup
maestro chat "Come puoi aiutarmi?"
```

Il setup usa per default `~/.config/maestro/config.yaml`; `--config` seleziona
un altro percorso e `--workspace` un'altra root. Non sovrascrive un file già
presente: una configurazione esistente deve essere valida prima che i controlli
proseguano.

I modelli mancanti vengono scaricati solo dopo conferma. Per ambienti
automatizzati l'autorizzazione è esplicita:

```sh
maestro setup --pull
```

`setup` non installa o avvia Ollama e non sostituisce automaticamente un tag
che risolve a un digest diverso da quello qualificato.

## Windows 11 con WSL2

Il binario Linux deve essere eseguito dentro una distro WSL2 e il progetto deve
risiedere sul filesystem Linux, per esempio sotto `/home/<utente>/src`. Ollama
deve essere installato e avviato nella stessa distro su
`http://127.0.0.1:11434`.

Da PowerShell verificare che la distro usi WSL2:

```powershell
wsl --status
wsl --list --verbose
```

Poi, dentro WSL:

```sh
uname -m
findmnt -T "$PWD"
ollama --version
curl -fsS http://127.0.0.1:11434/api/version
```

Il target atteso è `x86_64`, con un filesystem Linux come `ext4`, e Ollama
0.33.1. Non usare un progetto sotto `/mnt/c` per dedurre compatibilità: path,
permessi e replace atomico sarebbero quelli del filesystem montato da Windows,
fuori dal trial qualificato. Il package resta Linux `amd64`; non va eseguito
come binario Windows.

## Upgrade

Verifica il nuovo checksum, quindi sostituisci il binario in modo atomico:

```sh
install -m 0755 ./maestro "$HOME/.local/bin/maestro.new"
mv "$HOME/.local/bin/maestro.new" "$HOME/.local/bin/maestro"
maestro setup
```

## Rimozione

La rimozione del binario non tocca configurazione, workspace, Ollama o modelli:

```sh
rm "$HOME/.local/bin/maestro"
```

Per il primo utilizzo proseguire con il [Quick Start](quick-start.md). Per
identità dell’artifact e procedura riproducibile consultare
[Benchmark](benchmarks.md).

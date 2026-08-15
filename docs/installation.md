# Installazione preliminare di Maestro v0.1

Stato: Packaging candidate — non release candidate

Piattaforma verificata: Linux `amd64`

# Verifica e installazione

Scaricare nello stesso percorso archive e checksum, quindi verificare prima di
estrarre:

```sh
version=@MAESTRO_VERSION@
artifact="maestro-${version}-linux-amd64"
sha256sum -c "${artifact}.tar.gz.sha256"
tar -xzf "${artifact}.tar.gz"
cd "$artifact"
./maestro version
./maestro --help
```

`@MAESTRO_VERSION@` è un token della documentazione sorgente: lo script di
packaging lo sostituisce con la versione esatta dell'archive.

Per un'installazione utente senza privilegi:

```sh
install -d "$HOME/.local/bin"
install -m 0755 ./maestro "$HOME/.local/bin/maestro"
"$HOME/.local/bin/maestro" version
```

`$HOME/.local/bin` deve essere nel `PATH`. Per un prefisso di sistema usare un
percorso scelto dall'amministratore; Maestro non richiede né invoca `sudo`.

# Configurazione

La configurazione inclusa punta alla fixture Laravel inclusa nell'archive ed è
utilizzabile in place. Il profilo ufficiale v0.1.0 è read-only: registra
soltanto `workspace.list`, `workspace.read` e `workspace.search` e nega la
classe `workspace_mutate`.

```sh
./maestro doctor --config ./configs/maestro.example.yaml
./maestro models --config ./configs/maestro.example.yaml
./maestro agents --config ./configs/maestro.example.yaml
./maestro run --config ./configs/maestro.example.yaml \
  "Read app/Http/Controllers/OrderController.php and explain which service its store method calls. Do not modify any file."
```

Il probe provider richiede Ollama configurato su `127.0.0.1:11434`; senza il
provider, doctor deve comunque completare i check locali e segnalare il target
come non disponibile.

Per un progetto reale, copiare il file nella posizione XDG e modificare almeno
`workspace.root`, provider e modelli:

```sh
install -d "$HOME/.config/maestro"
install -m 0600 ./configs/maestro.example.yaml "$HOME/.config/maestro/config.yaml"
```

Un path workspace relativo viene risolto rispetto al file di configurazione.
La configurazione non contiene credenziali; `api_key_env` accetta soltanto il
nome di una variabile d'ambiente.

Il quick start non deve essere esteso aggiungendo `workspace.write` o
`workspace.patch`. Le capacità mutative e l'approval terminale restano
sperimentali e non supportate nella v0.1.0. llama.cpp è anch'esso sperimentale;
il percorso ufficiale richiede Ollama e `llama3.1:8b`.

# Upgrade

Verificare il checksum del nuovo archive, estrarlo in una directory nuova,
eseguire `maestro version` e sostituire il solo binario dopo aver conservato la
configurazione utente:

```sh
install -m 0755 ./maestro "$HOME/.local/bin/maestro.new"
mv "$HOME/.local/bin/maestro.new" "$HOME/.local/bin/maestro"
```

Leggere sempre le note di migrazione della serie 0.x: CLI e schema sono
sperimentali e i cambi breaking devono essere dichiarati.

# Rimozione

La rimozione non tocca automaticamente configurazioni o workspace:

```sh
rm "$HOME/.local/bin/maestro"
```

La configurazione può essere eliminata separatamente solo dopo aver verificato
il percorso:

```sh
rm "$HOME/.config/maestro/config.yaml"
rmdir "$HOME/.config/maestro" 2>/dev/null || true
```

Maestro non installa servizi, modelli, plugin o dipendenze Laravel e non
modifica shell profile durante l'installazione.

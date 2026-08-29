# Installazione di Maestro @MAESTRO_VERSION@

Stato: @MAESTRO_STATUS@

Piattaforma verificata: Linux `amd64`

## Verifica e installazione

Per una release pubblicata, scaricare archive e checksum dalla stessa GitHub
Release nello stesso percorso:

```sh
version=@MAESTRO_VERSION@
artifact="maestro-${version}-linux-amd64"
base_url="https://github.com/Axtonno/maestro/releases/download/${version}"
curl -fLO "${base_url}/${artifact}.tar.gz"
curl -fLO "${base_url}/${artifact}.tar.gz.sha256"
sha256sum -c "${artifact}.tar.gz.sha256"
tar -xzf "${artifact}.tar.gz"
cd "$artifact"
./maestro version
./maestro --help
```

I due file devono provenire dalla stessa release. Versione, commit e stato
devono coincidere con `ARTIFACT-MANIFEST.txt`; per la release pubblicata lo
stato è `release`. Un artifact non cambia stato tramite rename: packaging
candidate, release candidate e release sono build distinte.

Per installare il binario senza privilegi:

```sh
install -d "$HOME/.local/bin"
install -m 0755 ./maestro "$HOME/.local/bin/maestro"
"$HOME/.local/bin/maestro" version
```

Maestro non richiede né invoca `sudo`.

## Configurazione Direct Chat

L’archive include soltanto il profilo di prodotto chat-only v2. Punta alla
fixture inclusa, abilita streaming opt-in e congela `qwen3.5:9b`, context 4096,
thinking disabilitato, temperatura interna zero e limiti da 1 MiB.

```sh
./maestro doctor --mode chat --config ./configs/maestro.chat.example.yaml
./maestro chat --config ./configs/maestro.chat.example.yaml \
  --file routes/api.php \
  "Quali endpoint, controller e action sono dichiarati?"
```

Il provider deve essere Ollama già attivo su loopback e il modello con digest
qualificato deve essere già presente. Maestro non avvia servizi e non scarica
modelli.

Per un workspace reale:

```sh
install -d "$HOME/.config/maestro"
install -m 0600 ./configs/maestro.chat.example.yaml \
  "$HOME/.config/maestro/chat.yaml"
```

Aggiornare `workspace.root`, rieseguire doctor e indicare sempre un file
logico esplicito. Il file YAML non contiene credenziali; `api_key_env` accetta
soltanto il nome di una variabile d’ambiente.

Agent, retrieval, tool calling, profili mutativi e approval non appartengono
al support claim v0.3.0 e non sono inclusi nella configurazione distribuita.

## Upgrade

Estrarre ogni nuova versione in una directory nuova e verificare checksum,
manifest e note di migrazione prima di sostituire il binario:

```sh
install -m 0755 ./maestro "$HOME/.local/bin/maestro.new"
mv "$HOME/.local/bin/maestro.new" "$HOME/.local/bin/maestro"
```

Non riutilizzare automaticamente un profilo agentico v1 come profilo chat v2.

## Rimozione

La rimozione non tocca workspace, configurazione, provider o modelli:

```sh
rm "$HOME/.local/bin/maestro"
```

Eliminare separatamente una configurazione soltanto dopo averne verificato il
percorso. Maestro non installa servizi, plugin, dipendenze Laravel o modifiche
allo shell profile.

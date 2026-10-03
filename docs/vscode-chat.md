# Installare e aggiornare la chat di Maestro in VS Code

Questa guida riguarda il candidato locale **Maestro for VS Code 0.4.0** e la
CLI costruita dagli stessi sorgenti. Non è ancora disponibile sul Marketplace;
le prove complete della 0.4.0 su ambienti puliti sono ancora pendenti.
L'archivio pubblico CLI v0.5.0 non contiene i comandi richiesti dalla chat
integrata: aggiornare soltanto l'estensione non è sufficiente.

Su Windows il percorso è **VS Code Windows → Remote WSL → Maestro e Ollama
nella distro Linux**. Il VSIX non supporta l'esecuzione Windows nativa.
Su Linux locale si possono seguire i passaggi Bash omettendo la connessione WSL.

## 1. Preparare l'ambiente

Servono:

- VS Code 1.100.0 o successivo, con la vista Chat disponibile;
- su Windows, WSL2 con Ubuntu 24.04 x86_64 e l'estensione Microsoft WSL;
- un progetto nel filesystem Linux, per esempio `~/src/mio-progetto`;
- Ollama 0.33.1 in esecuzione nella stessa distro, su `127.0.0.1:11434`;
- Git, Go 1.24.5 e Node.js almeno 22.12.0 con npm, per costruire CLI e VSIX.

Installare i tool **Linux dentro WSL**. Le istruzioni di riferimento sono
quelle di [Go](https://go.dev/doc/install),
[Node.js](https://nodejs.org/en/download) e
[Ollama Linux](https://ollama.com/download/linux). La versione Ollama indicata
è la baseline del candidato; installare automaticamente l'ultima versione non
equivale a riprodurre questa baseline.

In PowerShell:

```powershell
wsl --list --verbose
wsl -d Ubuntu-24.04
```

La distro deve indicare versione WSL `2`. Da qui in poi i blocchi `bash` vanno
eseguiti nel terminale Linux. Verificare i prerequisiti prima di proseguire:

```bash
uname -m
findmnt -T "$HOME"
git --version
go version
node --version
npm --version
ollama --version
curl -fsS http://127.0.0.1:11434/api/version
```

L'architettura attesa è `x86_64` e il filesystem del progetto è Linux, per
esempio `ext4`. Se Ollama è installato ma non risponde, avviarlo con
`ollama serve` in un secondo terminale e lasciarlo aperto. Se è già gestito da
un servizio, usare quel servizio senza avviare una seconda istanza.

## 2. Preparare i sorgenti della stessa versione

Per ricostruire il candidato descritto qui, usare il commit
`a5cbcd5b9d4e47d4515b8221b6895d90e4f7460e` in un checkout dedicato:

```bash
mkdir -p "$HOME/src"
git clone https://github.com/Axtonno/maestro.git "$HOME/src/maestro-chat"
cd "$HOME/src/maestro-chat"
git checkout --detach a5cbcd5b9d4e47d4515b8221b6895d90e4f7460e
git status --short
git rev-parse HEAD
```

Se la directory esiste già, entrare nel checkout, verificare che non contenga
modifiche da conservare ed eseguire `git fetch origin` prima del checkout.
Non forzare il checkout sopra modifiche locali.

Se il commit è disponibile soltanto nel repository locale Windows, sostituire
l'URL del clone con il suo percorso `/mnt/c/.../maestro` e aggiungere
`--no-hardlinks` a `git clone`. Il checkout destinazione resta sotto `~/src`:
la copia include i commit, non i file modificati o non tracciati nel sorgente.

CLI e VSIX devono essere costruiti da questo stesso checkout. Eseguire i
passaggi successivi nella stessa sessione Bash; fermarsi se un comando fallisce.

## 3. Installare o sostituire la CLI

Nel checkout Linux appena preparato:

```bash
cd "$HOME/src/maestro-chat"
mkdir -p "$HOME/.local/bin"
MAESTRO_BACKUP="$HOME/.local/share/maestro/backups/$(date +%Y%m%d-%H%M%S)"
mkdir -p "$MAESTRO_BACKUP"
if [ -f "$HOME/.local/bin/maestro" ]; then
  cp -p "$HOME/.local/bin/maestro" "$MAESTRO_BACKUP/maestro"
fi
go build -o "$HOME/.local/bin/maestro.new" ./cmd/maestro &&
  mv "$HOME/.local/bin/maestro.new" "$HOME/.local/bin/maestro"
export PATH="$HOME/.local/bin:$PATH"
maestro version --diagnostic
maestro profile --help
maestro chat --help
```

Una build locale può riportare una versione di sviluppo: non è la release
pubblica v0.5.0. `profile --help` deve essere riconosciuto e l'help di `chat`
deve includere `--workspace-current`.

Per rendere il comando disponibile nei terminali futuri, aggiungere una sola
volta `export PATH="$HOME/.local/bin:$PATH"` a `~/.profile`. In VS Code verrà
comunque impostato il percorso assoluto del binario.

## 4. Configurare e provare la chat da terminale

Entrare nel progetto da usare, sostituendo il percorso di esempio:

```bash
cd "$HOME/src/mio-progetto"
maestro setup
maestro profile --workspace-current
maestro doctor --mode all --workspace-current
maestro chat --workspace-current "Spiegami la differenza fra classe e interfaccia in PHP."
```

Il setup crea la configurazione predefinita, normalmente
`~/.config/maestro/config.yaml`, e chiede consenso per scaricare i modelli
mancanti. Il profilo `recommended` richiede `qwen3.5:9b` per chat e
`qwen2.5-coder:14b` per mutation, con i digest qualificati: anche usando solo
la chat integrata, il percorso di setup attuale prepara il profilo completo.
Non occorre eseguire una mutation per usare la chat.

Il setup non sovrascrive una configurazione esistente. Se segnala
`workspace_mismatch`, la configurazione appartiene a un altro progetto:
si può conservare quella valida e usare `--workspace-current` per le prove,
come fa l'estensione, oppure crearne una distinta:

```bash
maestro setup --config .maestro/config.yaml
maestro profile --config .maestro/config.yaml --workspace-current
maestro doctor --config .maestro/config.yaml --mode all --workspace-current
maestro chat --config .maestro/config.yaml --workspace-current "Come puoi aiutarmi?"
```

Nel secondo caso impostare `maestro.configPath` a `./.maestro/config.yaml`
nelle impostazioni del workspace. Una configurazione legacy non valida non
va cancellata alla cieca: conservarne una copia e creare un nuovo percorso
con `setup --config`.

## 5. Costruire e installare l'estensione

```bash
cd "$HOME/src/maestro-chat/editors/vscode-maestro"
npm ci --ignore-scripts
npm test
npm run package:vsix
sha256sum dist/maestro-local-ai-0.4.0.vsix
```

Questi comandi producono `dist/maestro-local-ai-0.4.0.vsix`. I test unitari
non sostituiscono la prova dell'estensione installata in VS Code.

Aprire il progetto in una finestra VS Code collegata a WSL, tramite
**WSL: Open Folder in WSL** oppure `code .` dal terminale del progetto.
L'indicatore remoto deve mostrare la distro. Installare il VSIX da
**Extensions: Install from VSIX...** nella finestra remota; verificare che
Maestro sia presente nella sezione **WSL: Ubuntu-24.04** delle estensioni.
È il modello descritto nella [guida ufficiale VS Code WSL](https://code.visualstudio.com/docs/remote/wsl).

In alternativa, dal terminale integrato della stessa finestra WSL:

```bash
code --install-extension "$HOME/src/maestro-chat/editors/vscode-maestro/dist/maestro-local-ai-0.4.0.vsix" --force
code --list-extensions --show-versions
```

La lista deve contenere `axtonno.maestro-local-ai@0.4.0`. Il VSIX non contiene
la CLI o i modelli. Se `code` non è disponibile nel terminale, usare la
procedura grafica: non serve installare una seconda copia desktop di VS Code
dentro WSL.

## 6. Collegare la CLI e aprire la chat

Nelle impostazioni Remote WSL impostare `maestro.binaryPath` al risultato di:

```bash
printf '%s\n' "$HOME/.local/bin/maestro"
```

Usare il percorso Linux assoluto, per esempio `/home/utente/.local/bin/maestro`;
non incollare `$HOME`, `~` o un percorso `C:\...` nel valore dell'impostazione.
Lasciare `maestro.configPath` vuoto per la configurazione predefinita e
`maestro.profile` a `recommended`. Controllare che eventuali impostazioni del
workspace non impongano vecchi percorsi.

Eseguire **Developer: Reload Window**, quindi **Maestro: Open Chat** e inviare:

```text
@maestro /status
@maestro Spiegami la differenza fra classe e interfaccia in PHP.
```

Per una domanda generica chiudere gli editor di file; per una domanda sul
codice aprire e salvare il file desiderato, poi chiedere:

```text
@maestro Riassumi le responsabilità di questo file.
```

Verificare nell'intestazione `recommended`, i modelli attesi, il workspace
corretto e il target `remote-wsl` (`local-linux` su Linux locale).
Il contesto è limitato al file attivo: non viene analizzato automaticamente
l'intero repository. Le richieste attuali non inoltrano la cronologia della
chat al modello; formulare domande autosufficienti.

## 7. Aggiornamenti successivi e ripristino

Conservare il VSIX precedente e il backup della CLI. Terminare le richieste
attive; aggiornare il checkout al commit scelto, quindi ripetere build CLI,
build VSIX e installazione con `--force`. Per una versione diversa dalla 0.4.0,
usare il nome dell'artifact prodotto da quella versione.

Dopo **Developer: Reload Window**, ripetere `/status`, Doctor e una domanda.
Non serve riscaricare i modelli a ogni aggiornamento. Rieseguire setup solo
quando occorre inizializzare o verificare la configurazione; non rigenerarla
automaticamente sopra quella esistente.

Per tornare indietro, ripristinare la CLI dal backup e reinstallare il VSIX
compatibile conservato, poi ricaricare la finestra. Non associare la vecchia
CLI pubblica v0.5.0 all'estensione 0.3.0/0.4.0.

## 8. Disinstallazione completa

La procedura seguente rimuove l'installazione descritta in questa guida.
Eseguirla nella distro WSL o nell'host Linux dove è stato installato Maestro;
se esistono installazioni in più distro o profili VS Code, ripetere i passaggi
nei rispettivi ambienti. Terminare prima le richieste in corso e chiudere i
terminali Maestro. Conservare configurazioni e backup se si desidera poter
ripristinare l'installazione.

### Rimuovere estensione e impostazioni VS Code

Nella finestra VS Code collegata alla distro, disinstallare **Maestro for
VS Code** dalla sezione delle estensioni WSL, oppure dal terminale integrato:

```bash
code --uninstall-extension axtonno.maestro-local-ai
```

Se il comando `code` non è disponibile, usare il pulsante **Uninstall** nella
vista Extensions. Controllare anche eventuali copie installate localmente in
Windows e negli altri profili, poi eseguire **Developer: Reload Window**.
La [gestione delle estensioni VS Code](https://code.visualstudio.com/docs/configure/extensions/extension-marketplace)
descrive la disinstallazione tramite interfaccia.

Rimuovere le sole impostazioni `maestro.binaryPath`, `maestro.configPath`,
`maestro.profile` e `maestro.terminalName` dalle impostazioni User, Remote WSL
e Workspace dove sono presenti. Controllare anche `.vscode/settings.json`,
gli eventuali file `.code-workspace`, i collegamenti da tastiera personalizzati
e gli override di `remote.extensionKind` relativi a
`axtonno.maestro-local-ai`. Conservare le altre impostazioni del progetto.

Le conversazioni visualizzate nella Chat sono gestite da VS Code: se si vuole
rimuoverle, eliminarle anche dalla cronologia/sessioni tramite l'interfaccia
della Chat. La disinstallazione dell'estensione non dimostra la cancellazione
della cronologia conservata dall'editor. Non cancellare le directory dati
globali di VS Code, che contengono anche informazioni di altri strumenti.

### Rimuovere CLI, configurazioni e backup

Prima di cancellare, identificare eventuali installazioni multiple:

```bash
type -a maestro
printf 'MAESTRO_CONFIG=%s\nXDG_CONFIG_HOME=%s\n' "${MAESTRO_CONFIG-}" "${XDG_CONFIG_HOME-}"
```

Per il binario installato da questa guida:

```bash
rm -i -- "$HOME/.local/bin/maestro"
```

Eliminare anche `~/.local/bin/maestro.new` se rimasto da una build interrotta.
Per binari installati altrove, usare il percorso effettivamente verificato:
non eliminare l'intera directory `~/.local/bin` o `/usr/local/bin`.

Controllare il contenuto delle directory seguenti prima di rimuoverle. I
comandi chiedono conferma per la cancellazione ricorsiva:

```bash
ls -la -- "$HOME/.config/maestro" "$HOME/.local/share/maestro/backups"
rm -rI -- "$HOME/.config/maestro"
rm -rI -- "$HOME/.local/share/maestro/backups"
```

Se una directory non esiste, saltare il relativo comando. Se
`MAESTRO_CONFIG`, `XDG_CONFIG_HOME`, `--config` o `maestro.configPath`
selezionavano un altro percorso, rimuovere anche la configurazione Maestro
da quel percorso verificato. **Non cancellare la directory XDG condivisa.**
Nei progetti, eliminare i soli file di configurazione creati appositamente,
per esempio `.maestro/config.yaml`; non cancellare cartelle `.maestro` senza
averne controllato il contenuto.

Rimuovere eventuali definizioni di `MAESTRO_CONFIG` dai file di avvio della
shell e, per la sessione corrente, eseguire:

```bash
unset MAESTRO_CONFIG
hash -r
```

Il percorso `~/.local/bin` nel `PATH` può servire ad altri programmi: conservarlo
se condiviso. Eliminare eventuali alias o funzioni di shell creati per Maestro.

### Eliminare sorgenti e pacchetti conservati

Se non servono più, rimuovere il checkout dedicato `~/src/maestro-chat`, dopo
aver verificato e salvato eventuali modifiche locali. Questo elimina anche
`node_modules` e i VSIX sotto `editors/vscode-maestro/dist` di quel checkout.
Eliminare separatamente gli archive, i VSIX e le copie del binario conservati
altrove, inclusi eventuali download o checkout su Windows.

I progetti sui quali è stato usato Maestro restano dati dell'utente: la
disinstallazione non richiede di cancellarli e non annulla le modifiche al
codice già approvate. Go, Node.js, Git, VS Code e WSL sono strumenti separati;
non occorre rimuoverli per disinstallare Maestro.

### Opzionale: rimuovere i modelli e Ollama

I modelli appartengono a Ollama e possono essere usati anche da altre
applicazioni. Se erano dedicati esclusivamente a Maestro e non servono più,
rimuoverli mentre il servizio Ollama è ancora attivo:

```bash
ollama list
ollama stop qwen3.5:9b
ollama stop qwen2.5-coder:14b
ollama rm qwen3.5:9b
ollama rm qwen2.5-coder:14b
ollama list
```

Saltare i modelli non presenti. I comandi `stop` e `rm` sono descritti nel
[riferimento CLI Ollama](https://docs.ollama.com/cli). Non eliminare l'intero
archivio dei modelli se contiene modelli usati da altri strumenti.

Se anche Ollama era dedicato soltanto a Maestro, completarne la rimozione
seguendo la [procedura ufficiale Linux](https://docs.ollama.com/linux#uninstall):
servizio, binario, librerie, dati dei modelli ed eventuale utente/gruppo di
servizio. Verificare i percorsi reali dell'installazione e gli eventuali
override di `OLLAMA_MODELS`; una copia di Ollama su Windows è un'installazione
separata, gestita dalle impostazioni App di Windows.

### Verifica finale

Aprire un nuovo terminale nella distro ed eseguire:

```bash
type -a maestro
code --list-extensions --show-versions
```

Il primo comando deve indicare che Maestro non è disponibile; nella lista
non deve comparire `axtonno.maestro-local-ai`. Se si usa l'interfaccia grafica,
verificare invece la sezione Extensions dell'host e del profilo interessati.
Controllare inoltre che i percorsi di configurazione, backup e artifact scelti
siano stati rimossi. Questi passaggi rimuovono l'installazione e i dati
identificati; non costituiscono una cancellazione sicura di backup esterni,
copie sincronizzate o tracce gestite dal sistema operativo.

## Problemi frequenti

| Sintomo | Azione |
| --- | --- |
| `Binary missing` | Impostare `maestro.binaryPath` nelle impostazioni Remote WSL e controllare gli override del workspace. |
| `profile` sconosciuto o CLI incompatibile | Ricostruire la CLI corrente e verificare il binario effettivamente selezionato. |
| Configurazione mancante | Eseguire setup nella distro; verificare anche `MAESTRO_CONFIG` e `XDG_CONFIG_HOME`. |
| Ollama non raggiungibile | Verificare l'endpoint nella distro, poi eseguire Doctor. |
| Modello o digest non valido | Leggere Doctor e ripristinare il modello qualificato; non aggirare il controllo del digest. |
| `@maestro` assente | Verificare VSIX installato in WSL, workspace attendibile e ricaricare la finestra. |
| Vista Chat nascosta | Controllare che `chat.disableAIFeatures` non sia attivo nelle impostazioni VS Code. |
| File dirty o non salvato | Salvare il file prima di inviarlo come contesto. |
| `extension_host_mismatch` | Rimuovere eventuali override di `remote.extensionKind` che forzano Maestro sull'host Windows. |
| Dev Container o Remote SSH rifiutato | Aprire il progetto direttamente in Remote WSL o Linux locale. |

La disponibilità della vista Chat dipende anche dalla configurazione di VS
Code; il relativo interruttore è documentato nelle
[impostazioni AI ufficiali](https://code.visualstudio.com/docs/agents/reference/ai-settings).
Maestro invia le proprie richieste al provider locale e il suo manifest non
impone una dipendenza dall'estensione GitHub Copilot.

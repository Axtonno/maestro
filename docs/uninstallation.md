# Disinstallazione completa di Maestro

Questa procedura rimuove Maestro CLI e la sua eventuale estensione VS Code
installati seguendo le guide di [installazione](installation.md) e
[chat in VS Code](vscode-chat.md). Saltare i componenti non installati.
Eseguirla nella distro WSL o nell'host Linux dove è stato installato Maestro;
se esistono installazioni in più distro o profili VS Code, ripetere i passaggi
nei rispettivi ambienti. Terminare prima le richieste in corso e chiudere i
terminali Maestro. Conservare configurazioni e backup se si desidera poter
ripristinare l'installazione.

## Rimuovere estensione e impostazioni VS Code

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

## Rimuovere CLI, configurazioni e backup

Prima di cancellare, identificare eventuali installazioni multiple:

```bash
type -a maestro
printf 'MAESTRO_CONFIG=%s\nXDG_CONFIG_HOME=%s\n' "${MAESTRO_CONFIG-}" "${XDG_CONFIG_HOME-}"
```

Per il binario installato nel percorso utente indicato nelle guide:

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

## Eliminare sorgenti e pacchetti conservati

Se non servono più, rimuovere il checkout dedicato `~/src/maestro-chat`, dopo
aver verificato e salvato eventuali modifiche locali. Questo elimina anche
`node_modules` e i VSIX sotto `editors/vscode-maestro/dist` di quel checkout.
Eliminare separatamente gli archive, i VSIX e le copie del binario conservati
altrove, inclusi eventuali download o checkout su Windows.

I progetti sui quali è stato usato Maestro restano dati dell'utente: la
disinstallazione non richiede di cancellarli e non annulla le modifiche al
codice già approvate. Go, Node.js, Git, VS Code e WSL sono strumenti separati;
non occorre rimuoverli per disinstallare Maestro.

## Opzionale: rimuovere i modelli e Ollama

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

## Verifica finale

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


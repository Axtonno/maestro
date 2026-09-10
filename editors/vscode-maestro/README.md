# Maestro VS Code Prototype

Questo prototipo espone in VS Code il perimetro già controllato di Maestro
v0.5.0. Non incorpora il runtime, non interpreta l'output del modello e non
scrive file: apre un terminale integrato `/bin/sh` nel workspace e vi invia un
comando CLI con argomenti shell-quoted.

## Comandi

- `Maestro: Ask About Active File` usa il file locale, salvato e attivo come
  unico contesto esplicito per `maestro chat`;
- `Maestro: Replace Selected Lines` accetta una sola selezione di righe
  complete in un file PHP salvato sotto `app/` e invoca
  `maestro workspace replace`;
- `Maestro: Run Doctor` esegue `doctor --mode all`;
- `Maestro: Show Binary Identity` esegue `version --diagnostic`.

La mutation conserva il gate originale: preview, fingerprint e allow-once o
deny avvengono nel terminale reale e restano sotto l'autorità del binario.
L'estensione non invia `y`, non applica patch e non offre approval automatica.

## Uso di sviluppo

1. Aprire questa directory come workspace trusted con VS Code su Linux o
   tramite Remote WSL.
2. Premere `F5` e scegliere l'Extension Development Host.
3. Nell'host di sviluppo aprire un workspace compatibile con la configurazione
   Maestro v4.
4. Configurare `maestro.binaryPath` e, se necessario,
   `maestro.configPath`.
5. Eseguire prima `Maestro: Show Binary Identity` e `Maestro: Run Doctor`.

Per le unit del compilatore dei comandi serve Node.js `22.12.0` o successivo:

```sh
npm run test:unit
```

`npm test` mantiene lo stesso gate locale e non richiede download o GUI.

La suite d'integrazione richiede inoltre il percorso di un eseguibile VS Code
Linux non già in uso:

```sh
MAESTRO_VSCODE_EXECUTABLE=/percorso/VSCode-linux-x64/code \
  npm run test:integration
```

Con lo stesso prerequisito, `npm run test:all` esegue entrambi i gate.

La qualifica della Milestone 40 e gli input congelati sono registrati in
[`VALIDATION.md`](VALIDATION.md).

## Vincoli intenzionali

- extension host Linux, incluso Remote WSL;
- workspace trusted, locale e non virtuale;
- file locali dentro una cartella workspace;
- file salvato prima di chat o mutation;
- chat soltanto sul file attivo;
- mutation soltanto single-selection di righe complete, single-range, PHP
  sotto `app/`;
- terminale nuovo per ogni invocazione;
- shell `/bin/sh` fissata dal prototipo, senza dipendenza dal profilo terminale
  configurato dall'utente;
- nessun webview, background process, shell autonoma, test runner o multi-file.

Il prototipo non è incluso nell'archive v0.5.0, non è pubblicato nel
Marketplace e non costituisce una capability supportata. La sua eventuale
pubblicazione richiederà una qualifica separata.

# Milestone 40 — VS Code Controlled Surface Prototype

Stato: **prototipo implementato — `vscode_prototype_validation_pending`**

## Obiettivo

Verificare se Direct Chat e Controlled Mutation già qualificati possono essere
raggiunti dall'editor con meno attrito, senza duplicare il runtime o ampliare
l'autorità della v0.5.0. M40 è una milestone di prototipo: non modifica
l'archive pubblico, non autorizza una release e non aggiunge l'integrazione
VS Code al support claim corrente.

Il prototipo vive in
[`editors/vscode-maestro`](../editors/vscode-maestro/README.md) ed è un bridge
sottile verso la CLI. Ogni invocazione apre un terminale integrato nuovo con
shell `/bin/sh` e working directory uguale alla cartella workspace esplicita.

## Contratto

| Superficie editor | Comando autorevole |
| --- | --- |
| Domanda sul file attivo | `maestro chat --file <logical> <question>` |
| Sostituzione delle righe selezionate | `maestro workspace replace --file <logical> --lines <start:end> <instruction>` |
| Diagnostica | `maestro doctor --mode all` |
| Identità binario | `maestro version --diagnostic` |

L'extension host deve essere Linux, incluso Remote WSL, e il workspace deve
essere trusted, locale e non virtuale. Il file deve essere locale, salvato e
contenuto nella cartella workspace. La mutation richiede esattamente una
selezione non vuota di righe complete,
converte le coordinate VS Code 0-based/exclusive nelle coordinate CLI
1-based/inclusive e restringe preventivamente il target a `app/**/*.php`. Il
binario ripete comunque tutti i controlli autorevoli, inclusi containment,
symlink, span e stale source.

## Confine di sicurezza

- input, path, binario e config sono token shell-quoted per `/bin/sh`; domanda
  e istruzione sono precedute dal separatore di flag `--`;
- NUL, CR e LF sono respinti nei token;
- workspace untrusted e virtuali sono disabilitati nel manifest, con un
  controllo runtime aggiuntivo sul trust;
- documenti dirty sono respinti, così editor e CLI osservano gli stessi byte;
- la mutation viene eseguita soltanto nel terminale integrato reale;
- preview completa e decisione allow-once/deny appartengono alla CLI;
- l'estensione non legge né trasforma la proposta, non invia approval e non
  modifica direttamente il workspace;
- ogni errore editoriale chiude prima di lanciare un comando.

Il terminale è deliberato: usare un child process senza TTY renderebbe la
mutation inutilizzabile oppure richiederebbe un nuovo protocollo di approval,
che è fuori dal perimetro M40.

## Matrice e gate

La matrice congelata è in
[`milestone-40-vscode-prototype-matrix.yaml`](milestone-40-vscode-prototype-matrix.yaml).
I contratti statici devono coprire manifest, registrazione dei quattro comandi,
file dirty, selezioni parziali o multiple e assenza di API di scrittura. La
suite JavaScript deve esercitare escaping, containment, coordinate e scope
mutativo. La prova live successiva deve usare l'asset pubblico v0.5.0 e coprire
identità, doctor, chat, deny e allow su una fixture Git pulita.

M40 può concludersi con `vscode_controlled_surface_prototype_validated` solo se:

- test JavaScript e contratti repository sono verdi;
- Extension Development Host registra tutti i comandi senza errori;
- chat e mutation usano file e righe esatti;
- deny produce zero modifiche;
- allow produce un solo diff uguale alla preview;
- l'approval avviene nel terminale e non nell'estensione;
- zero comandi inattesi, scritture fuori selezione o failure con effetti.

## Stato corrente

Implementazione e controlli statici sono presenti. La macchina corrente non
dispone di Node.js e l'installazione snap di VS Code non può avviare
l'Extension Development Host perché `snap-confine` rifiuta l'ambiente; la
suite JavaScript V03–V06 e i casi live restano pertanto `not_run`. Questo non
viene interpretato come PASS.

## Fuori scope

- pubblicazione Marketplace o inclusione nel package Maestro;
- Windows nativo, macOS, browser e workspace remoti diversi da WSL;
- webview, sidebar, chat participant o rendering custom della preview;
- approval editoriale, auto-approval o bypass TTY;
- file dirty, untitled, multi-root senza file attivo, multi-selection;
- scelta autonoma di file/righe, multi-file, agent, retrieval o test execution.

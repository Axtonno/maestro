# Checkpoint offline — Milestone 40

Data: 2026-09-08  
Verdetto: `vscode_prototype_validation_pending`

Il prototipo dependency-free registra quattro comandi VS Code e costruisce
soltanto invocazioni della CLI v0.5.0. Il percorso mutativo usa un terminale
integrato `/bin/sh` nuovo, preservando il requisito di TTY e lasciando preview,
fingerprint, approval e apply al binario qualificato.

I controlli statici V01, V02, V07 e V08 verificano manifest e sorgenti, rifiuto
dei workspace untrusted, dei buffer dirty, delle selezioni parziali o multiple
e assenza di API VS Code di scrittura o auto-approval. Le unit V03–V06 per
quoting, containment, coordinate e scope mutativo sono state scritte ma non
eseguite.

Non è stata eseguita una prova live. `node` e `npm` non sono installati nella
macchina corrente; inoltre `code --version` e `code-insiders --version`
falliscono prima dell'avvio perché `snap-confine` rifiuta l'ambiente AppArmor.
Di conseguenza V03–V06 e i casi Extension Development Host V09–V14 restano
`not_run`, non PASS.

Il prototipo non modifica la baseline congelata v0.5.0, non entra negli asset
di packaging e non autorizza una pubblicazione Marketplace. La chiusura M40
richiede una nuova evidenza live secondo il piano e la matrice congelata.

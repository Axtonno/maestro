# Milestone 36 — Audit strumentazione residency v1

Data: 2026-09-05. Stato: **EVIDENZA NON CONCLUSIVA**.

Il primo report ha completato correttamente 18/18 transizioni e ha registrato
modelli residenti, VRAM, RAM, swap e latenze. Il collector dei processi ha però
cercato Ollama nell'host Windows, mentre il provider effettivo era il processo
Linux `/usr/local/bin/ollama serve` nella distribuzione WSL.

Di conseguenza i campi PID/RSS/start time del provider sono vuoti e il gate
`provider_restarts=0` non è sostenuto dall'evidenza. Il verdetto calcolato
`dual_model_residency_qualified` non viene accettato come conclusivo.

Il profilo v2 corregge soltanto l'attribuzione dei processi Linux e rende
obbligatoria la presenza del processo provider. Modelli, sequenza, numero di
cicli, richieste e soglie temporali restano identici; nessuna soglia è stata
rilassata dopo le osservazioni v1. Il report v1 resta immutato come audit trail.

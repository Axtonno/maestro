# Milestone 28 – Qualificazione live

Data: 2026-09-03

Stato: **NOT RUN – ENVIRONMENT UNAVAILABLE**

Preflight osservato:

- host corrente Windows, mentre il target candidato è Linux `amd64`;
- comando `ollama` assente;
- endpoint Ollama locale `127.0.0.1:11434` non raggiungibile;
- nessun modello o digest modello è pertanto attestabile;
- nessuna run, risposta, terminale o modifica live è stata imputata.

Le verifiche software sono state eseguite separatamente in WSL su una materializzazione LF pulita con Go 1.24.5 ufficiale, archive SHA-256 `10ad9e86233e74c0f6590fe5426895de6bf388964210eac34a6d83f38918ecdc`. `go test ./...`, `go test -race ./...` e `go vet ./...` sono PASS.

L'assenza dell'ambiente non è un failure d'integrità e non autorizza una selezione per inferenza. La soglia semantica, i terminali provider e il gate con approval TTY su asset installato restano **non osservati**.

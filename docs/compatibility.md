# Maestro v0.1.x Compatibility Matrix

Data: 2026-08-15

Questa pagina è la fonte autorevole per ciò che la serie v0.1.x dichiara supportato.
La presenza di codice, adapter o test deterministici non equivale a supporto
operativo.

## Percorso supportato

| Dimensione | Stato v0.1.x | Confine verificato |
|---|---|---|
| Sistema operativo | Supportato | Linux `amd64` |
| Provider | Supportato | Ollama su endpoint esplicito; baseline `http://127.0.0.1:11434` |
| Modello chat | Supportato | `llama3.1:8b`, reference agent Laravel read-only |
| Modello embedding | Fixture supportata | `embeddinggemma:latest`; non richiesto dal quick start lessicale |
| Framework | Supportato | Progetto Laravel rilevabile tramite `artisan` e `laravel/framework` in `composer.json` |
| Reference agent | Supportato | `agent.reference`, analisi e interrogazione read-only |
| Tool ufficiali | Supportati | `workspace.list`, `workspace.read`, `workspace.search` |
| Mutazioni | Non supportate | `workspace.write`, `workspace.patch`, approval mutativa e agent mutante sono sperimentali |
| llama.cpp | Non supportato | Adapter sperimentale; matrice live v0.1.0 non completata |
| Isolamento | Non fornito | Componenti trusted in-process, nessuna sandbox o separazione di privilegi |

La combinazione supportata è stata validata su un host Linux `amd64` CPU-only,
Intel Core i5-8365U, 15 GiB RAM e 4 GiB swap. Questo profilo descrive la prova,
non un requisito minimo universale. Su CPU simili una singola run può richiedere
diversi minuti.

Dalla v0.1.1 i workspace Laravel reali sono indicizzati tramite la scan policy
sorgente del plugin. Asset generati, storage runtime e dipendenze non entrano
nello snapshot; il cambiamento corregge un failure di indicizzazione della
v0.1.0 senza modificare tool o autorità.

## Compatibilità non dichiarata

Non sono qualificati per la serie v0.1.x:

- Linux `arm64`, macOS e Windows;
- endpoint Ollama remoti o esposti a reti non attendibili;
- modelli chat diversi da `llama3.1:8b` nel reference agent supportato;
- versioni o distribuzioni Laravel specifiche oltre alla detection strutturale;
- GPU, acceleratori o profili hardware differenti;
- llama.cpp, inclusi router mode e lifecycle multi-modello;
- tool mutanti e approval come percorso di prodotto;
- plugin o tool di terze parti.

“Non qualificato” non significa necessariamente incompatibile: significa che
non esiste un gate di release sufficiente per una promessa pubblica.

## Contratti sperimentali

CLI, schema `version: 1` e package Go pubblici sono disponibili ma restano
sperimentali durante la serie 0.x. I dettagli sono in
`v0.1.0-api-compatibility.md`. `internal/` non è API.

## Evidenze

Le evidenze live e di installazione sono registrate in
`docs/reports/milestone-8-phase-5.md` nel repository sorgente. L'archive include
questa matrice, la configurazione esatta e la fixture usate dal percorso
supportato.

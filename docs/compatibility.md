# Maestro v0.2.0 Compatibility Matrix

Data: 2026-08-21; aggiornamento 2026-08-27

Questa pagina è la fonte autorevole per ciò che v0.2.0 dichiara supportato.
La presenza di codice, adapter o test deterministici non equivale a supporto
operativo.

## Percorso supportato

| Dimensione | Stato v0.2.0 | Confine verificato |
|---|---|---|
| Sistema operativo | Supportato | Linux `amd64` |
| Provider | Supportato | Ollama su endpoint esplicito; baseline `http://127.0.0.1:11434` |
| Modello chat | Supportato | `llama3.1:8b`, reference agent Laravel read-only |
| Modello embedding | Fixture supportata | `embeddinggemma:latest`; non richiesto dal quick start lessicale |
| Framework | Supportato | Progetto Laravel rilevabile tramite `artisan` e `laravel/framework` in `composer.json` |
| Reference agent | Supportato | `agent.reference`, analisi e interrogazione read-only |
| Tool ufficiali | Supportati | `workspace.list`, `workspace.read`, `workspace.search` |
| Mutazioni | Non supportate | `workspace.write`, `workspace.patch`, approval mutativa e agent mutante sono sperimentali |
| llama.cpp | Non supportato | Adapter sperimentale; preflight conclusivo incompatibile, nessuna matrice valida |
| Isolamento | Non fornito | Componenti trusted in-process, nessuna sandbox o separazione di privilegi |

La combinazione supportata è stata validata su un host Linux `amd64` CPU-only,
Intel Core i5-8365U, 15 GiB RAM e 4 GiB swap. Questo profilo descrive la prova,
non un requisito minimo universale. Su CPU simili una singola run può richiedere
diversi minuti.

La v0.2.0 non amplia questa matrice. Il candidato mutativo con
`ibm/granite4.1:8b` supera 15/15 scenari deterministici e il preflight sullo
stesso lower bound, ma fallisce il primo tentativo del Gate A live con una
patch call non esatta. Gate B e Gate C non sono stati eseguiti per fail-fast;
ADR-0032 registra `mutation_deferred`. Queste evidenze non qualificano né il
modello né `workspace.patch` come supportati.

## Esito della Field Validation

La matrice sopra conserva il perimetro storico qualificato dalla Milestone 12;
non costituisce una promessa di qualità per analisi Laravel multi-file. La
Milestone 13 ha chiuso anticipatamente la matrice field a 5/22 per stop rule
con classificazione `field_validation_completed_with_limitations` e decisione
`adoption_no_go_on_reference_profile`. Sicurezza e immutabilità read-only sono
state confermate, mentre affidabilità operativa e qualità multi-file sono
risultate insufficienti sul profilo di riferimento.

Le 17 run mancanti sono `not_run` e il Gate 0 di pubblicazione remota non è
stato completato. Il verdetto non revoca retroattivamente l'identità
dell'artifact v0.2.0, ma impedisce di promuoverlo come baseline affidabile per
quel carico. Il report completo è in
`reports/milestone-13-field-validation.md`.

Una modalità `direct/chat` senza tool è un requisito di prodotto emerso dal
confronto conclusivo e pianificato per la Milestone 14; non appartiene alla
superficie supportata di v0.2.0 e non opera come fallback del reference agent.

La Milestone 14 è in sviluppo secondo ADR-0033. Il candidato introduce
`maestro chat`, `maestro agent`, l'alias deprecato `maestro run` e lo schema v2
con profili separati. Fino alla chiusura dei gate C0–C4 e alla qualificazione
della Milestone 15, codice e configurazioni candidate restano development-only
e non costituiscono una nuova compatibility promise.

Dalla v0.1.1 i workspace Laravel reali sono indicizzati tramite la scan policy
sorgente del plugin. Asset generati, storage runtime e dipendenze non entrano
nello snapshot; il cambiamento corregge un failure di indicizzazione della
v0.1.0 senza modificare tool o autorità.

## Compatibilità non dichiarata

Non sono qualificati per v0.2.0:

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
`v0.2.0-api-compatibility.md`. `internal/` non è API.

## Evidenze

Le evidenze live e di installazione sono registrate nei report delle Milestone
8, 9 e 12 nel repository sorgente. La Milestone 12 qualifica packaging,
installazione, quick start consecutivi e gate operativi dell'esatto percorso
read-only. ADR-0030 chiude la Milestone 3 senza promuovere llama.cpp; ADR-0032
chiude la Milestone 11 senza promuovere Controlled Mutation. L'archive v0.2.0
include questa matrice, la configurazione esatta e la fixture usate dal
percorso supportato.

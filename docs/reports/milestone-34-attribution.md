# Milestone 34 — Attribuzione delle failure host-bound

Data: 2026-09-05. Ramo selezionato: **B — prompt e payload corretti**.

Verdetto operativo: `qwen3.5_9b_host_bound_mutation_profile_rejected`.

## Provenienza e metodo

Audit del repository a partire da `b2b69fae5195dae1ed3399b2bc95ab1d49fc03bc`.
`scripts/m34audit` estrae il prompt dal runner M33, verifica i digest M33 e
ricostruisce le tre richieste usando il vero adapter Ollama con un trasporto
HTTP in memoria. La risposta fittizia serve soltanto a completare il metodo
dell'adapter; non è output di un modello e non entra nelle metriche M33.

Verificati i digest congelati di prompt, schema e matrice riportati in
`milestone-34-offline-reconstruction.json`. I file sorgente letti sono
normalizzati da CRLF a LF per ricostruire i byte Git usati in M33; i valori
dei payload decodificati dalla matrice non sono normalizzati.

Il report di ricostruzione ha SHA-256
`52e17e2a089ebf8e8ea5993df2595d4c0df62f797181b72072b322575ec50b6d`.
Sono state eseguite zero generazioni, zero repliche M33 e zero mutazioni.

## Prompt, schema e payload

Il prompt dichiara che Maestro ha selezionato un intervallo immutabile,
vieta di scegliere file, target o coordinate e limita `new_text` alla sola
selezione. Chiede di preservare i byte non modificati. L'astensione è prevista
per informazioni mancanti, contraddizioni o modifiche fuori selezione:
nessuna di queste condizioni si applica ai tre casi positivi falliti.

| Caso | Coordinate | Testo selezionato | Trasformazione attesa |
| --- | --- | --- | --- |
| M33-D03 | 3–3 | `// pending` (10 byte) | `// complete` |
| M33-H01 | 2–2 | `// dormant` (10 byte) | `// awake` |
| M33-H02 | 2–3 | commento `// 日本語`, LF, `$retries = 6;` (26 byte) | conservare commento e LF, cambiare 6 in 7 |

La ricostruzione verifica sia i byte selezionati sia lo splice contro il
file finale atteso nella matrice. Le richieste sono determinabili senza
contesto aggiuntivo. D03 e H01 non inviano al modello le righe duplicate
esterne. Non è quindi giustificato attribuire le loro astensioni alla vista
di duplicati. H02 conserva correttamente Unicode e separatore interno.

Lo schema ammette soltanto `propose` con `new_text` oppure `abstain` e non
richiede path, `old_text`, coordinate generate o ricerca del target.

## Adapter e richiesta HTTP

`internal/provider/ollama/generation.go` copia ruoli e contenuti senza
aggiungere prompt; `client.go` serializza il DTO con `json.Marshal`.
La cattura in memoria conferma, per tutti e tre i casi:

- due messaggi: system congelato e user con `Request`, `SelectedText`,
  `StartLine`, `EndLine`;
- nessuna history, istruzione di agent o tool trasmesso;
- `stream: false`, `think: false`, context 4096, output 1024, temperatura 0;
- schema JSON invariato nel campo `format`.

Non emerge un difetto di selezione, traduzione o valutazione che renda
corrette le tre astensioni. I prompt delle milestone precedenti non sono
raggiunti da questa chiamata diretta all'adapter.

## Modello e renderer

Le letture metadata `/api/tags`, `/api/version` e `/api/show` confermano
Ollama 0.33.1 e il digest M33
`6488c96fa5faab64bb65cbd30d4289e20e6130ef535a93ef9a49f42eda893ea7`.
Il modello è Q4_K_M; il Modelfile non dichiara `SYSTEM` né history, espone
`TEMPLATE {{ .Prompt }}` e seleziona renderer/parser `qwen3.5`.
Lo snapshot show ha SHA-256
`8ce92f2da52b70ccd7f920962c838532de82293d7ca27ef59db859e47d76f5de`.

Non basta il template apparente: il server usa il renderer quando presente.
Questo è verificato nel [codice upstream v0.33.1 di renderPrompt](https://github.com/ollama/ollama/blob/v0.33.1/server/prompt.go).
La [registrazione qwen3.5](https://github.com/ollama/ollama/blob/v0.33.1/model/renderers/renderer.go)
seleziona `Qwen35Renderer`. Nel [renderer v0.33.1](https://github.com/ollama/ollama/blob/v0.33.1/model/renderers/qwen35.go),
senza tool, il contenuto system è mantenuto senza postambolo per chiamate a
funzioni; `think: false` chiude il prefisso thinking. Il percorso non
aggiunge istruzioni di unicità, ricerca del target o astensione mutativa.
Il trimming del messaggio JSON esterno non modifica le stringhe al suo interno.

Il Modelfile dichiara temperature 1, top_k 20, top_p 0.95 e presence_penalty
1.5. La richiesta M33 sovrascrive temperature a 0; gli altri valori sono
default dichiarati, non istruzioni. L'ordine di composizione delle opzioni è
verificabile in [modelOptions v0.33.1](https://github.com/ollama/ollama/blob/v0.33.1/server/routes.go).
Non si attribuisce alle opzioni la causa delle astensioni né si cambia il
profilo per cercare un esito migliore.

## Limiti e decisione

M33 non ha persistito il wire HTTP storico, il prompt tokenizzato dal server,
un hash del binario Ollama o una motivazione delle astensioni. La cattura M34
è una ricostruzione deterministica, non un recupero di quei dati. La lettura
metadata è attuale e coincide con identità e versione registrate; l'ispezione
upstream riguarda la versione dichiarata, non attesta l'identità binaria.
Non si dimostra la causa interna del comportamento né un limite universale
di `qwen3.5:9b`.

Entro questo perimetro verificabile, il contratto è coerente e le tre
trasformazioni sono determinabili. Non emerge un conflitto concreto che
autorizzi il ramo A. Le false astensioni restano failure del profilo operativo
osservato: si applica il ramo B e si interrompe il tuning di questo profilo
per la mutazione. La decisione non dipende da attribuire retroattivamente una
motivazione al modello. La selezione di un modello dedicato è aperta in M35.

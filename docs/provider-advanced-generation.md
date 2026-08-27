# Maestro Advanced Provider Generation

Versione: 0.2.0

Stato: Implementato

Ultimo aggiornamento: 2026-08-27

---

# Scopo

La Fase 9 completa la baseline di generazione condivisa da Ollama e llama.cpp:
sampling comune, output JSON, tool calling e streaming incrementale delle tool
call. I contratti restano provider-neutral e le opzioni esclusivamente
proprietarie non entrano in `pkg/provider`.

Le richieste semplici rimangono invariate: tutti i nuovi campi hanno zero value
che conserva i default del provider.

---

# Opzioni comuni

`CompletionRequest.Options` espone:

- `MaxTokens`: zero lascia il default, un valore positivo limita la risposta;
- `Temperature`: puntatore per distinguere il valore valido zero dall'assenza;
- `TopP`: puntatore per distinguere il valore dall'assenza;
- `Stop`: sequenze di arresto ordinate;
- `ContextWindow`: zero lascia il default, un valore positivo richiede una
  context window esatta;
- `Thinking`: zero lascia la semantica legacy, `default` omette il controllo,
  `true` e `false` richiedono un valore esplicito.

La validazione comune accetta temperature tra 0 e 2, `top_p` maggiore di zero e
non superiore a 1, token non negativi e stop non vuoti. Ollama traduce i campi
in `options.num_predict`, `temperature`, `top_p` e `stop`; llama.cpp usa i campi
OpenAI-compatible `max_tokens`, `temperature`, `top_p` e `stop`.

La Milestone 14 aggiunge `ContextWindow` e `Thinking` in modo additivo.
Ollama traduce la context window in `options.num_ctx` e thinking nel campo
top-level `think`; il puntatore nell'adapter distingue omissione e `false`.
La native chat API non restituisce sempre un'attestazione del valore effettivo:
il valore richiesto e il body mappato sono osservabili, mentre l'effettivo resta
`unknown` quando il runner non lo espone.

L'endpoint OpenAI-compatible di llama.cpp non consente a Maestro di impostare
questi due controlli per singola request. L'adapter li rifiuta prima dell'I/O
con `ErrUnsupportedCapability`, invece di ignorarli.

---

# Output strutturato

`StructuredOutput` supporta:

- `json`: output JSON senza schema;
- `json_schema`: schema JSON rappresentato da `json.RawMessage`.

Lo schema deve essere sintatticamente un oggetto JSON. Maestro non implementa
un validator JSON Schema e non interpreta keyword o dialetti: il provider
applica il vincolo, mentre l'adapter verifica che il contenuto terminale sia
JSON valido. Durante lo streaming il contenuto viene accumulato soltanto per
questa verifica e continua a essere consegnato incrementalmente.

Ollama riceve `format: "json"` oppure lo schema direttamente nel campo
`format`, secondo la documentazione ufficiale degli
[Structured Outputs](https://docs.ollama.com/capabilities/structured-outputs).
llama.cpp riceve `response_format` con `type: "json_object"` e lo schema nel
campo `schema`, come descritto nella documentazione di
[`llama-server`](https://github.com/ggml-org/llama.cpp/blob/master/tools/server/README.md#post-v1chatcompletions-openai-compatible-chat-completions-api).

Output strutturato e tool definitions non possono essere combinati nella stessa
richiesta della baseline. Questa restrizione evita semantiche differenti tra
template, grammar e parser dei provider.

---

# Tool calling

Una `Tool` contiene nome, descrizione e JSON Schema dei parametri. I nomi sono
univoci, lunghi al massimo 64 caratteri e limitati a lettere, numeri, `_` e
`-`. `ToolChoice` supporta `auto`, `none`, `required` e `named`; il valore zero
equivale ad `auto` quando esistono tool.

`Message` rappresenta l'intero ciclo:

- un messaggio assistant può contenere `ToolCalls` complete;
- ogni `ToolCall` conserva ID, nome e argomenti come oggetto JSON;
- un messaggio tool contiene il risultato e riferisce `ToolCallID`, `ToolName`
  o entrambi.

Ollama usa il nome per associare i risultati e supporta nel proprio endpoint
nativo soltanto scelta automatica o tool disabilitati; `required` e `named`
falliscono localmente con `ErrUnsupportedCapability`. llama.cpp usa gli ID
OpenAI-compatible e traduce tutte le modalità di scelta. Tool result senza nome
su Ollama o senza call ID su llama.cpp vengono rifiutati prima dell'I/O.

La documentazione ufficiale di Ollama descrive tool, risultati e streaming in
[Tool calling](https://docs.ollama.com/capabilities/tool-calling). llama.cpp
richiede una configurazione compatibile con `--jinja` e un chat template adatto,
come indicato dalla documentazione ufficiale del server.

---

# Streaming

`StreamChunk.ToolCalls` contiene `ToolCallDelta` indicizzati. ID e nome possono
comparire soltanto nel primo delta; `Arguments` è un frammento testuale che il
consumer concatena per indice fino al terminale.

llama.cpp può dividere gli argomenti tra più eventi SSE. L'adapter ricostruisce
parallelamente ogni call per validare al `finish_reason` che ID, nome e oggetto
JSON siano completi. Ollama consegna oggetti tool call completi nei chunk NDJSON
e Maestro li espone nello stesso formato delta.

I finish reason comuni documentati sono `stop`, `length` e `tool_calls`. Usage
e marker terminali mantengono la semantica delle fasi precedenti; il consumer
deve continuare a leggere fino a `io.EOF`.

---

# Capability introspection

Entrambi gli adapter dichiarano supporto strutturale per `structured_output` e
`tool_calling`. Ollama dichiara inoltre `context_window_control` e
`thinking_control`; la capability modello `thinking` deriva da `/api/show`.
llama.cpp dichiara i tre controlli non supportati nella superficie per-request.

- Ollama rende structured output disponibile per modelli con capability
  `completion` e tool calling disponibile quando `/api/show` espone `tools`.
- llama.cpp lega structured output alla disponibilità chat e tool calling alla
  disponibilità chat con `--jinja` osservabile negli argomenti effettivi.
- A livello instance, senza un modello esatto, entrambe rimangono `unknown`.

Il report è un preflight informativo: modello, chat template e versione del
server possono comunque rifiutare una richiesta successiva.

---

# Compatibilità e limiti

I metodi `Complete`, `Stream` e le capability interface non cambiano. I campi
sono aggiunti ai value object esistenti; codice esterno deve usare composite
literal con chiavi, come già fanno esempi e test di Maestro. L'audit completo è
in `provider-api-compatibility-audit.md`.

Restano fuori scope:

- multimodalità, audio e video;
- prompt caching e speculative decoding;
- livelli di thinking proprietari diversi dal controllo booleano tri-state;
- parallel tool policy proprietarie;
- validazione semantica JSON Schema nel client;
- esecuzione automatica dei tool;
- opzioni di sampling disponibili soltanto in un provider.

ADR-0016 registra il contratto e i relativi trade-off.

# Milestone 3 — Validazione live Ollama

Stato: Gate live Ollama superato

Stato Milestone 3: In corso — non chiusa

Data: 2026-08-09

---

# Configurazione

```text
MAESTRO_OLLAMA_BASE_URL=http://localhost:11434
MAESTRO_OLLAMA_CHAT_MODEL=qwen2.5-coder:7b
MAESTRO_OLLAMA_EMBED_MODEL=embeddinggemma
GOCACHE=/tmp/maestro-go-build
```

Profilo osservato:

- Linux `amd64`, 8 CPU logiche;
- Intel Core i5-8365U;
- 15.643 MiB RAM;
- Ollama con due modelli nel catalogo;
- `qwen2.5-coder:7b`, GGUF Q4_K_M, context length 32768, capability
  `completion`, `tools`, `insert`;
- `embeddinggemma:latest`, GGUF BF16, context length 2048, capability
  `embedding`.

Nessuna fixture lifecycle o acquisition era configurata. Le mutazioni del
catalogo non erano autorizzate.

---

# Test di integrazione richiesto

Comando:

```text
MAESTRO_OLLAMA_BASE_URL=http://localhost:11434 \
MAESTRO_OLLAMA_CHAT_MODEL=qwen2.5-coder:7b \
MAESTRO_OLLAMA_EMBED_MODEL=embeddinggemma \
GOCACHE=/tmp/maestro-go-build \
go test -v -count=1 -tags=integration ./pkg/provider/ollama/...
```

Esito: **PASS**, durata package 4,921 secondi.

Subtest live superati:

- model listing;
- model discovery;
- completion;
- streaming fino a EOF;
- cancellazione stream;
- embedding.

Il test lifecycle è stato saltato perché
`MAESTRO_OLLAMA_LIFECYCLE_MODEL` non è configurato.

La prima esecuzione nel sandbox non poteva raggiungere il loopback. Ripetendo
il comando con accesso autorizzato a Ollama locale, tutti i subtest sopra sono
passati; il primo errore di trasporto non è quindi attribuibile a Maestro o
Ollama.

---

# Smoke Benchmark live

Comando:

```text
MAESTRO_OLLAMA_BASE_URL=http://localhost:11434 \
MAESTRO_OLLAMA_CHAT_MODEL=qwen2.5-coder:7b \
MAESTRO_OLLAMA_EMBED_MODEL=embeddinggemma \
GOCACHE=/tmp/maestro-go-build \
go run ./cmd/maestro bench smoke \
  --provider ollama --warmup 0 --runs 1 --fail-on-failure --output -
```

Report schema: `1.2.0`.

Artefatti redatti conservati:

- `milestone-3-live-ollama-smoke.json` — fonte canonica;
- `milestone-3-live-ollama-smoke.md` — vista derivata.

Entrambi hanno permessi `0600`. Il Markdown rigenerato dal JSON ha lo stesso
hash SHA-256 della vista conservata.

Run ID: `abd5f5112a80557a2766d1328a03fb32`.

Durata complessiva osservata: 38.786,193 ms.

Esito del comando: **FAIL**, exit status 1.

| Stato | Scenari |
|---|---:|
| `passed` | 9 |
| `skipped` | 3 |
| `failed` | 2 |

## Scenari superati

| Scenario | Durata | Evidenza |
|---|---:|---|
| `capability-instance` | 1,342 ms | 11 capability |
| `catalog-list-discover` | 2,361 ms | 2 modelli listati e scoperti |
| `completion-simple` | 4.133,603 ms | 40 input token, 10 output token |
| `stream-terminal-close` | 2.897,809 ms | 10 chunk e terminale corretto |
| `stream-cancel-deadline` | 2.291,958 ms | cancellazione osservata, 0 chunk successivi |
| `structured-json` | 4.022,218 ms | oggetto JSON con un campo |
| `structured-json-schema` | 2.917,606 ms | schema rispettato |
| `resilience-controlled-error` | 3,250 ms | errore controllato e cleanup policy |
| `observability-redaction` | 8.683,396 ms | 3 eventi correlati e redatti |

## Scenari saltati

| Scenario | Reason code | Valutazione |
|---|---|---|
| `embedding` | `capability_unavailable` | configurato `embeddinggemma`, catalogo `embeddinggemma:latest` |
| `lifecycle-load-unload` | `model_not_configured` | fixture lifecycle assente |
| `acquisition-pull-remove` | `catalog_mutation_not_allowed` | mutation guard disabilitata |

L'embedding diretto del test d'integrazione funziona perché Ollama accetta
l'alias senza tag. L'introspection usa invece, per contratto, un model ID esatto
e confronta `embeddinggemma` con il catalogo che espone
`embeddinggemma:latest`. Per la prossima run occorre configurare:

```text
MAESTRO_OLLAMA_EMBED_MODEL=embeddinggemma:latest
```

## Scenari falliti

| Scenario | Durata | Error code |
|---|---:|---|
| `tool-call-result` | 6.999,779 ms | `tool_call_missing` |
| `tool-call-stream` | 6.831,205 ms | `tool_stream_terminal_missing` |

Ollama dichiara `tools` per `qwen2.5-coder:7b`, ma nelle due richieste Smoke il
modello non ha prodotto rispettivamente la tool call richiesta e il terminale
tool-call previsto nello stream. Il gate non abbassa questi fallimenti a skip:
supporto dichiarato e comportamento osservato non coincidono.

## Diagnostica diretta di `/api/chat`

Per separare il comportamento del modello dalla traduzione dell'adapter, le
due richieste sono state ripetute direttamente contro
`POST http://localhost:11434/api/chat`, senza attraversare Maestro. Payload,
messaggi, tool e `num_predict: 128` coincidono con i rispettivi scenari Smoke;
come controllo deterministico è stata impostata anche `temperature: 0`.

Tool inviato in entrambe le richieste:

```json
{
  "type": "function",
  "function": {
    "name": "echo_message",
    "description": "Return the supplied message",
    "parameters": {
      "type": "object",
      "properties": {
        "message": { "type": "string" }
      },
      "required": ["message"],
      "additionalProperties": false
    }
  }
}
```

### Risposta non-stream

Configurazione specifica:

```json
{
  "model": "qwen2.5-coder:7b",
  "messages": [{
    "role": "user",
    "content": "Call echo_message exactly once with message set to Maestro smoke."
  }],
  "stream": false,
  "options": { "num_predict": 128, "temperature": 0 }
}
```

La risposta grezza contiene `done: true` e `done_reason: "stop"`.
`message.tool_calls` è assente. Il modello serializza invece la chiamata come
testo ordinario in `message.content`:

```json
{
  "name": "echo_message",
  "arguments": {
    "message": "Maestro smoke"
  }
}
```

### Risposta stream

Il secondo payload usa il messaggio esatto dello scenario streaming:

```text
Call echo_message with message set to Maestro smoke.
```

La risposta è composta da 27 oggetti NDJSON:

- chunk 1–26: `done: false`, frammenti del JSON testuale in
  `message.content`, nessun campo `message.tool_calls`;
- chunk 27: `done: true`, `done_reason: "stop"`, contenuto vuoto e nessun
  campo `message.tool_calls`.

Non esiste quindi un chunk iniziale, intermedio o terminale nel quale compaia
una tool call strutturata.

### Diagnosi

Le chiamate dirette riproducono entrambi i fallimenti osservati nello Smoke
Benchmark. L'adapter Maestro non riceve da Ollama alcun `message.tool_calls` da
tradurre o aggregare; i due error code non sono quindi causati da perdita dei
tool call nell'adapter. L'evidenza circoscrive il problema alla combinazione
installata di modello, prompt/template e runtime Ollama: il modello comprende
semanticamente la richiesta, ma emette la chiamata nel canale testuale anziché
nel campo strutturato previsto dall'API. Questa conclusione riguarda la fixture
`qwen2.5-coder:7b` verificata e non dimostra un limite universale della famiglia
Qwen.

La fixture `qwen2.5-coder:7b` viene conservata come caso negativo documentato:
con il payload Smoke e temperatura 0 produce la rappresentazione semanticamente
corretta della chiamata, ma nel canale testuale anziché in
`message.tool_calls`.

## Seconda fixture: `llama3.1:8b`

La discovery di `/api/tags` espone letteralmente:

- chat e lifecycle: `llama3.1:8b`;
- embedding: `embeddinggemma:latest`.

Il gate diretto è stato ripetuto con lo stesso tool `echo_message`, il prompt
`Call echo_message exactly once with message set to Maestro smoke.`,
`num_predict: 128` e `temperature: 0`.

### Gate diretto non-stream

Esito: **PASS**.

- `message.tool_calls` è presente;
- contiene una sola chiamata a `echo_message`;
- gli argomenti sono `{"message":"Maestro smoke"}`;
- `message.content` è vuoto;
- la risposta termina con `done: true` e `done_reason: "stop"`.

### Gate diretto stream

Esito: **PASS** secondo i criteri del gate diretto.

La risposta contiene due chunk:

1. chunk non terminale con `message.tool_calls`, chiamata valida a
   `echo_message`, argomenti corretti e `message.content` vuoto;
2. chunk terminale con `done: true`, nessuna duplicazione testuale e
   `done_reason: "stop"`.

Il risultato è stato confermato anche con il prompt esatto dello scenario
Smoke streaming, `Call echo_message with message set to Maestro smoke.`: tool
call nel primo chunk e terminale `stop` nel secondo.

### Integration test

Configurazione:

```text
MAESTRO_OLLAMA_BASE_URL=http://localhost:11434
MAESTRO_OLLAMA_CHAT_MODEL=llama3.1:8b
MAESTRO_OLLAMA_EMBED_MODEL=embeddinggemma:latest
MAESTRO_OLLAMA_LIFECYCLE_MODEL=llama3.1:8b
GOCACHE=/tmp/maestro-go-build
```

Comando:

```text
go test -v -count=1 -tags=integration ./pkg/provider/ollama
```

Esito: **PASS**, durata package 10,564 secondi. Sono passati listing,
discovery, completion, stream, cancellazione, embedding e lifecycle.

### Smoke Benchmark completo

Run ID: `bc0075d09580fed3074187fd5c7d6c50`.

Durata complessiva osservata: 114.901,518 ms.

Esito: **FAIL**, exit status 1.

| Stato | Scenari |
|---|---:|
| `passed` | 12 |
| `skipped` | 1 |
| `failed` | 1 |

Embedding, lifecycle e `tool-call-result` passano. Lo scenario acquisition
resta saltato con `catalog_mutation_not_allowed`, come previsto dalla mutation
guard. L'unico failure è:

| Scenario | Durata | Error code |
|---|---:|---|
| `tool-call-stream` | 8.700,185 ms | `tool_stream_terminal_missing` |

### Diagnosi del terminale streaming

La fixture soddisfa il gate diretto e dimostra che Ollama emette una tool call
strutturata nei chunk. La divergenza riguarda la semantica terminale:

- Ollama invia la tool call in un chunk con `done: false`;
- chiude poi con `done: true` e `done_reason: "stop"`;
- l'adapter copia `done_reason` in `StreamChunk.FinishReason`;
- lo scenario Smoke riconosce il terminale tool-call soltanto quando
  `FinishReason == "tool_calls"`.

Il percorso streaming Maestro non normalizza quindi la sequenza multi-chunk
Ollama in un terminale `tool_calls`. `llama3.1:8b` non è un secondo caso
negativo del modello: è la fixture positiva che espone un'incompatibilità tra
la terminazione Ollama e il contratto/gate Maestro.

Il failure `tool_stream_terminal_missing` osservato in questa run era quindi un
difetto di normalizzazione dell'adapter Maestro, non un limite operativo di
`llama3.1:8b`. La correzione descritta di seguito lo risolve.

## Correzione dell'adapter Ollama

La correzione mantiene invariati benchmark e contratto neutrale:

- ogni istanza stream conserva il proprio stato `toolCallSeen`;
- lo stato viene impostato soltanto dopo la traduzione riuscita di almeno una
  tool call strutturata;
- un terminale Ollama `done: true`, `done_reason: "stop"` viene esposto come
  `tool_calls` soltanto se quello stream ha già osservato una tool call;
- `stop` senza tool call, `length`, cancellazioni ed errori non vengono
  sovrascritti;
- il chunk che contiene la tool call resta invariato e non terminale;
- usage e metadata restano sul successivo chunk terminale.

La stessa normalizzazione è applicata al percorso non-stream quando la
completion contiene tool call tradotte, eliminando l'asimmetria osservata con
Ollama.

Le regressioni isolate coprono:

1. tool call non terminale seguita da terminale `stop`;
2. testo seguito da terminale `stop`;
3. tool call seguita da terminale già `tool_calls`;
4. tool call seguita da `length`;
5. più chunk testuali senza tool call;
6. completion non-stream con `stop`, `tool_calls` e `length`.

I test mirati dei package Ollama sono superati.

## Validazione live post-correzione

L'integration suite con `llama3.1:8b`, `embeddinggemma:latest` e lifecycle
configurato è superata: listing, discovery, completion, stream, cancellazione,
embedding e lifecycle passano. Durata package: 18,815 secondi.

Il successivo Smoke Benchmark completo usa la stessa configurazione.

Run ID: `aead2e67f9de38b9d2e1e7d8f841435a`.

Durata complessiva osservata: 124.711,328 ms.

Esito: **PASS**, exit status 0.

| Stato | Scenari |
|---|---:|
| `passed` | 13 |
| `skipped` | 1 |
| `failed` | 0 |

Entrambi gli scenari tool calling passano con una tool call valida. L'unico
skip è `acquisition-pull-remove` con reason code
`catalog_mutation_not_allowed`: resta legittimo perché la mutation guard è
disabilitata.

`qwen2.5-coder:7b` resta il caso negativo documentato;
`llama3.1:8b` è la fixture positiva del gate live Ollama.

## Decisione sulle fixture Ollama

`llama3.1:8b` è registrato come fixture positiva live validata per:

| Area | Fixture | Stato |
|---|---|---|
| Chat | `llama3.1:8b` | Validata |
| Streaming | `llama3.1:8b` | Validata |
| Structured output JSON | `llama3.1:8b` | Validata |
| Structured output JSON Schema | `llama3.1:8b` | Validata |
| Embedding | `embeddinggemma:latest` | Validata nella stessa configurazione |
| Lifecycle | `llama3.1:8b` | Validata |
| Tool calling non-stream | `llama3.1:8b` | Validata |
| Tool calling stream | `llama3.1:8b` | Validata dopo normalizzazione Maestro |

`qwen2.5-coder:7b` è conservato come caso negativo canonico:

- Ollama dichiara la capability `tools`;
- il modello comprende semanticamente la richiesta;
- la chiamata viene serializzata in `message.content` oppure non viene emessa
  come `message.tool_calls`;
- la capability tools non è quindi validata operativamente con la versione e
  il template Ollama correnti.

La classificazione finale separa tre fenomeni distinti:

1. capability dichiarata ma non operativa: `qwen2.5-coder:7b`;
2. capability realmente operativa: `llama3.1:8b`;
3. difetto dell'adapter: normalizzazione del terminale tool-call streaming,
   corretto e coperto da regressioni.

Non occorre cercare altre fixture Ollama per questo gate. Il prossimo passo è
replicare la matrice live su llama.cpp, idealmente con lo stesso modello base di
Llama 3.1 per ridurre la variabile modello e isolare le differenze di runtime.

---

# Gate deterministico

Sono stati rieseguiti con successo:

```text
GOCACHE=/tmp/maestro-go-build go test -count=1 ./...
GOCACHE=/tmp/maestro-go-build go test -race -count=1 ./...
GOCACHE=/tmp/maestro-go-build go vet ./...
GOCACHE=/tmp/maestro-go-build go run ./cmd/maestro bench validate --manifest docs/provider-smoke-benchmark-manifest.yaml
GOCACHE=/tmp/maestro-go-build go run ./cmd/maestro bench validate --manifest docs/runtime-benchmark-manifest.yaml
GOCACHE=/tmp/maestro-go-build go run ./cmd/maestro bench validate --manifest docs/developer-benchmark-manifest.yaml
```

I manifest contengono rispettivamente 14, 11 e 6 scenari.

---

# Conclusione

La normalizzazione conservativa dell'adapter è coperta da regressioni isolate.
Test mirati, test repository-wide, race detector, vet e integration suite sono
verdi. Lo Smoke live Ollama completo chiude con 13 passed, 1 skipped e 0 failed:
il gate live Ollama è quindi **superato**.

Questo risultato non chiude automaticamente l'intera Milestone 3. La milestone
resta **in corso — non chiusa** in attesa della successiva decisione esplicita
di completamento. `qwen2.5-coder:7b` resta il caso negativo documentato e
`llama3.1:8b` la fixture positiva.

La documentazione del gate Ollama è conclusa; il successivo gate operativo
della milestone è la matrice llama.cpp.

L'unico skip acquisition resta accettabile finché la mutation guard non è
abilitata esplicitamente.

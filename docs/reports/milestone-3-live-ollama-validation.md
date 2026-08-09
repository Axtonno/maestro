# Milestone 3 — Validazione live Ollama

Stato: Gate live non superato

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

---

# Gate deterministico

Sono stati rieseguiti con successo:

```text
GOCACHE=/tmp/maestro-go-build go test ./...
GOCACHE=/tmp/maestro-go-build go test -race ./...
GOCACHE=/tmp/maestro-go-build go vet ./...
GOCACHE=/tmp/maestro-go-build go run ./cmd/maestro bench validate --manifest docs/provider-smoke-benchmark-manifest.yaml
GOCACHE=/tmp/maestro-go-build go run ./cmd/maestro bench validate --manifest docs/runtime-benchmark-manifest.yaml
GOCACHE=/tmp/maestro-go-build go run ./cmd/maestro bench validate --manifest docs/developer-benchmark-manifest.yaml
```

I manifest contengono rispettivamente 14, 11 e 6 scenari.

---

# Conclusione

Il test d'integrazione Ollama richiesto è superato e il gate deterministico è
verde. Il gate Smoke live completo non è però superato a causa dei due failure
di tool calling. Di conseguenza la Milestone 3 **non viene dichiarata chiusa**.

Per una nuova verifica di chiusura servono:

1. l'ID embedding esatto `embeddinggemma:latest`;
2. una configurazione modello/template Ollama che produca tool call strutturate
   non-stream e stream, oppure un modello fixture che le supporti in modo
   affidabile; la diagnostica diretta esclude l'adapter Maestro come origine dei
   due failure osservati;
3. una nuova esecuzione di `bench smoke --fail-on-failure` senza scenari
   `failed`.

Gli skip lifecycle e acquisition restano accettabili finché le relative fixture
opzionali e la mutation guard non sono configurate.

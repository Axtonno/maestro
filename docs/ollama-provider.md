# Maestro Ollama Provider

Versione: 0.1.0

Stato: Implementato

Ultimo aggiornamento: 2026-08-08

---

# Scopo

L'adapter Ollama è la prima implementazione concreta dei contratti definiti in
`pkg/provider`.

La facade pubblica vive in `pkg/provider/ollama`. Protocollo HTTP, DTO,
traduzione e gestione degli stream rimangono confinati in
`internal/provider/ollama`.

L'adapter usa esclusivamente la libreria standard Go e non introduce dipendenze
da SDK esterni.

---

# Capability

L'adapter implementa:

* `Completer` tramite `POST /api/chat` con `stream: false`;
* `Streamer` tramite `POST /api/chat` con `stream: true`;
* `Embedder` tramite `POST /api/embed`;
* `ModelLister` tramite `GET /api/tags`;
* `ModelDiscoverer` unendo `GET /api/tags` e `GET /api/ps`;
* `ModelLoader` e `ModelUnloader` tramite `POST /api/generate` e `keep_alive`;
* `ModelPuller` tramite `POST /api/pull` con stream NDJSON;
* `ModelRemover` tramite `DELETE /api/delete`;
* `CapabilityInspector` tramite `GET /api/tags` e `POST /api/show`.

Gli endpoint e i payload seguono la documentazione ufficiale di
[chat](https://docs.ollama.com/api/chat),
[embedding](https://docs.ollama.com/api/embed) e
[model listing](https://docs.ollama.com/api/tags).

Discovery e lifecycle seguono inoltre la documentazione ufficiale di
[running models](https://docs.ollama.com/api/ps) e delle regole di
[preload/unload](https://docs.ollama.com/faq#how-can-i-preload-a-model-into-ollama-to-get-faster-response-times).

Acquisizione e rimozione seguono la documentazione ufficiale di
[pull](https://docs.ollama.com/api/pull) e
[delete](https://docs.ollama.com/api/delete).

L'introspection model-specific segue
[Show model details](https://docs.ollama.com/api-reference/show-model-details)
e il campo `capabilities` restituito dal server.

---

# Capability introspection

Il target adapter è offline. Il target instance verifica il catalogo e dichiara
disponibili le capability di gestione, lasciando inference `unknown` senza un
modello. Il target model usa `/api/show`: `completion` abilita completion,
streaming e structured output, `embedding` abilita embedding e `tools` abilita
tool calling.

Tool calling e structured output sono supportati strutturalmente dall'adapter;
per il target instance rimangono `unknown` finché non viene scelto un modello.
Un modello assente viene riportato come non disponibile senza effettuare
`/api/show`; pull resta disponibile. Ogni query legge un nuovo snapshot e non
mantiene cache.

---

# Discovery e lifecycle dei modelli

`DiscoverModels` mantiene l'ordine di `/api/tags`, marca come `loaded` i
modelli presenti anche in `/api/ps` e aggiunge in coda eventuali modelli
visibili soltanto nello snapshot runtime.

I metadati neutrali includono, quando disponibili, digest, dimensioni, VRAM,
context length, formato, famiglia, parameter size, quantizzazione e timestamp.

`LoadModel` usa una richiesta vuota non-streaming con `keep_alive: -1`; il
modello rimane quindi caricato fino a `UnloadModel`, che usa `keep_alive: 0`.
Entrambe le operazioni accettano il modello esplicito o il modello predefinito
dell'adapter.

Le policy di residenza del Provider Runtime compongono queste capability:
autoload usa discovery e `LoadModel`, mentre rilascio immediato, TTL e shutdown
usano `UnloadModel` soltanto per modelli caricati dalla policy. Il TTL comune è
coordinato dal Runtime; Ollama riceve `keep_alive: -1` al load e `keep_alive: 0`
all'unload. La semantica completa è descritta in
`provider-model-residency.md`.

---

# Acquisizione e rimozione

`PullModel` traduce gli status Ollama negli stage neutrali `resolving`,
`downloading`, `verifying`, `finalizing` e `completed`. Digest, totale e byte
completati sono riportati quando presenti. Il testo Ollama originale rimane in
`Detail` a solo scopo informativo.

Il chunk `success` è l'evento terminale; la chiamata successiva restituisce
`io.EOF`. EOF anticipato, JSON malformato e conteggi incoerenti producono
`provider.ErrInvalidResponse`. `Close` è idempotente e interrompe la richiesta
HTTP.

`RemoveModel` invia il modello nel body JSON e accetta una risposta HTTP 2xx
vuota. L'adapter non accede alla directory dei modelli.

---

# Configurazione

La facade espone:

```go
type Config struct {
    BaseURL      string
    Timeout      time.Duration
    DefaultModel string
    HTTPClient   *http.Client
}
```

Valori predefiniti:

* `BaseURL`: `http://localhost:11434`;
* `Timeout`: 30 secondi quando l'adapter crea il client;
* `DefaultModel`: nessuno.

`BaseURL` rappresenta l'origine HTTP e non deve contenere `/api`, credenziali,
query o fragment. Uno slash finale viene normalizzato.

Se `HTTPClient` è valorizzato viene usato senza modificarlo e `Timeout` viene
ignorato. Questo permette all'applicazione di controllare trasporto, proxy e
timeout degli stream lunghi.

Una completion o un embedding deve indicare `request.Model` oppure disporre di
un `DefaultModel`. In caso contrario l'operazione fallisce prima di effettuare
una richiesta HTTP.

---

# Utilizzo

```go
config := runtime.NewConfig(map[string]any{
    provider.ConfigDefaultProvider: ollama.ID,
})

rt := maestro.New(maestro.WithConfig(config))

ollamaProvider, err := ollama.New(ollama.Config{
    BaseURL:      "http://localhost:11434",
    DefaultModel: "gemma4",
})
if err != nil {
    return err
}

if err := rt.Providers().Register(ollamaProvider); err != nil {
    return err
}
```

Il provider può poi essere invocato esplicitamente con ID `ollama` oppure
tramite l'ID vuoto quando è configurato come default.

---

# Generazione avanzata

`GenerationOptions` viene tradotto nelle opzioni native `num_predict`,
`temperature`, `top_p` e `stop`. Structured output usa `format: "json"` o uno
schema JSON. Tool definitions, chiamate assistant e risultati tool vengono
tradotti nei campi nativi di `/api/chat`.

L'endpoint nativo applica scelta automatica quando sono presenti tool e li
disabilita con `ToolChoiceNone`. Le modalità `required` e `named` non hanno una
traduzione nativa affidabile e falliscono localmente con
`ErrUnsupportedCapability`.

Le tool call non-streaming diventano `Message.ToolCalls`; quelle NDJSON
diventano `StreamChunk.ToolCalls`. Gli arguments devono essere oggetti JSON
validi. Il contratto comune è descritto in
`provider-advanced-generation.md`.

---

# Streaming

Ollama restituisce oggetti NDJSON. Ogni `Recv` decodifica esattamente un
oggetto e lo traduce in `provider.StreamChunk`.

Semantica:

* un chunk con `done: false` viene restituito normalmente;
* il chunk con `done: true` contiene finish reason e utilizzo finale;
* la chiamata successiva restituisce `io.EOF`;
* EOF prima di `done: true` produce `provider.ErrInvalidResponse`;
* un oggetto con il campo `error` termina lo stream con errore;
* `Close` chiude il body ed è idempotente;
* una lettura dopo una chiusura anticipata produce
  `provider.ErrInvalidStream`.

Gli errori Ollama che arrivano dopo l'inizio dello stream vengono gestiti come
indicato dalla [specifica ufficiale degli errori](https://docs.ollama.com/api/errors).

---

# Validazione ed errori

L'adapter valida localmente configurazione, modello richiesto, contratti di
generazione avanzata e input degli embedding.

Le risposte vengono rifiutate quando:

* il JSON o NDJSON non è valido;
* una completion non-streaming non è finale;
* uno stream termina senza il chunk finale;
* la cardinalità degli embedding non coincide con gli input;
* un embedding è vuoto;
* un modello elencato non possiede alcuna identità;
* structured output terminale non è JSON valido;
* una tool call non possiede nome o arguments JSON validi.

I body degli errori HTTP vengono letti con un limite di 64 KiB. Il campo JSON
`error` viene preferito al testo grezzo. `ProviderError` classifica HTTP,
trasporto, context e fallimenti mid-stream; i dettagli pubblici sono
normalizzati e limitati a 512 byte. Cancellazione e deadline del context
restano riconoscibili tramite `errors.Is`. La matrice comune è documentata in
`provider-error-semantics.md`.

---

# Test

La suite ordinaria usa un trasporto HTTP in-memory con
`httptest.ResponseRecorder` e non richiede processi, socket o modelli locali.

I test d'integrazione sono protetti dal build tag `integration`:

```bash
MAESTRO_OLLAMA_BASE_URL=http://localhost:11434 \
MAESTRO_OLLAMA_CHAT_MODEL=gemma4 \
MAESTRO_OLLAMA_EMBED_MODEL=embeddinggemma \
MAESTRO_OLLAMA_LIFECYCLE_MODEL=gemma4 \
go test -tags=integration ./pkg/provider/ollama
```

Senza `MAESTRO_OLLAMA_BASE_URL` il test viene saltato.

Lo smoke test verifica listing, discovery, completion non-streaming, streaming
fino a `io.EOF`, embedding e cancellazione di uno stream con successiva chiusura
esplicita. Load e unload vengono eseguiti soltanto quando è configurata
`MAESTRO_OLLAMA_LIFECYCLE_MODEL`. Il test non misura prestazioni e non esercita
retry o resilienza. Pull e remove live confluiscono nello Smoke Benchmark della
Milestone 3 e richiederanno un modello dedicato.

---

# Limiti

Non sono tradotti:

* immagini e input multimodali;
* thinking;
* tool choice `required` o `named` sull'endpoint nativo;
* opzioni di generazione specifiche di Ollama fuori dalla baseline comune;
* autenticazione per i modelli cloud;
* retry, backoff o circuit breaker interni all'adapter; le policy comuni
  appartengono al Provider Runtime.

Structured output e tool calling dipendono dal modello e dalla versione del
server e vengono verificati live dalla Milestone 3.

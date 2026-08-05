# Maestro Ollama Provider

Versione: 0.1.0

Stato: Implementato

Ultimo aggiornamento: 2026-08-05

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
* `ModelLister` tramite `GET /api/tags`.

Gli endpoint e i payload seguono la documentazione ufficiale di
[chat](https://docs.ollama.com/api/chat),
[embedding](https://docs.ollama.com/api/embed) e
[model listing](https://docs.ollama.com/api/tags).

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

L'adapter valida localmente configurazione, modello richiesto e input degli
embedding.

Le risposte vengono rifiutate quando:

* il JSON o NDJSON non è valido;
* una completion non-streaming non è finale;
* uno stream termina senza il chunk finale;
* la cardinalità degli embedding non coincide con gli input;
* un embedding è vuoto;
* un modello elencato non possiede alcuna identità.

I body degli errori HTTP vengono letti con un limite di 64 KiB. Il campo JSON
`error` viene preferito al testo grezzo. Cancellazione e deadline del context
restano riconoscibili tramite `errors.Is`.

---

# Test

La suite ordinaria usa un trasporto HTTP in-memory con
`httptest.ResponseRecorder` e non richiede processi, socket o modelli locali.

I test d'integrazione sono protetti dal build tag `integration`:

```bash
MAESTRO_OLLAMA_BASE_URL=http://localhost:11434 \
MAESTRO_OLLAMA_CHAT_MODEL=gemma4 \
MAESTRO_OLLAMA_EMBED_MODEL=embeddinggemma \
go test -tags=integration ./pkg/provider/ollama
```

Senza `MAESTRO_OLLAMA_BASE_URL` il test viene saltato.

---

# Limiti della prima versione

Non sono ancora tradotti:

* tool calling;
* immagini e input multimodali;
* thinking;
* structured output;
* opzioni di generazione specifiche di Ollama;
* autenticazione per i modelli cloud;
* retry, backoff o circuit breaker.

Queste estensioni richiederanno prima contratti neutrali condivisibili con
almeno un secondo provider.

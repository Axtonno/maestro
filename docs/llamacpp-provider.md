# Maestro llama.cpp Provider

Versione: 0.1.0

Stato: Implementato — smoke test live pendente

Ultimo aggiornamento: 2026-08-06

---

# Scopo

L'adapter llama.cpp collega Maestro a `llama-server` attraverso gli endpoint
HTTP compatibili con OpenAI, senza introdurre una dipendenza da SDK esterni.

La facade pubblica vive in `pkg/provider/llamacpp`. Protocollo HTTP, DTO,
traduzione delle risposte e parsing SSE rimangono confinati in
`internal/provider/llamacpp`.

---

# Capability

L'adapter implementa:

* `Completer` tramite `POST /v1/chat/completions`;
* `Streamer` tramite `POST /v1/chat/completions` con risposta SSE;
* `Embedder` tramite `POST /v1/embeddings`;
* `ModelLister` tramite `GET /v1/models`.

Gli endpoint seguono la documentazione ufficiale di
[`llama-server`](https://github.com/ggml-org/llama.cpp/blob/master/tools/server/README.md#openai-compatible-api-endpoints).

La capability di embedding richiede che `llama-server` sia avviato con un
modello e un pooling compatibili. La presenza dell'interfaccia indica il
supporto del protocollo da parte dell'adapter, non che ogni processo o modello
caricato possa eseguire embedding.

---

# Configurazione

La facade espone:

```go
type Config struct {
    BaseURL      string
    Timeout      time.Duration
    DefaultModel string
    APIKey       string
    HTTPClient   *http.Client
}
```

Valori predefiniti:

* `BaseURL`: `http://localhost:8080`;
* `Timeout`: 30 secondi quando l'adapter crea il client;
* `DefaultModel`: nessuno;
* `APIKey`: nessuna autenticazione.

`BaseURL` rappresenta l'origine HTTP e non deve contenere `/v1`, credenziali,
query o fragment. Uno slash finale viene normalizzato.

Se `APIKey` è valorizzata, ogni richiesta invia `Authorization: Bearer <key>`.
Il valore non viene mai inserito negli errori. Se `HTTPClient` è valorizzato,
viene usato senza modificarlo e `Timeout` viene ignorato.

Completion, streaming ed embedding richiedono `request.Model` oppure un
`DefaultModel`. Il model listing non richiede un modello predefinito.

---

# Streaming

`llama-server` usa Server-Sent Events. Ogni evento `data` contenente un chunk
JSON viene tradotto in `provider.StreamChunk`; il marker `data: [DONE]` chiude
lo stream e produce `io.EOF`.

L'adapter richiede anche `stream_options.include_usage`, così un eventuale
evento finale senza `choices` può riportare l'utilizzo dei token. Il chiamante
deve quindi consumare fino a `io.EOF` e non assumere che il primo
`finish_reason` sia necessariamente l'ultimo evento dati.

EOF del trasporto prima di `[DONE]`, eventi malformati, più choice nello stesso
chunk e choice con indice diverso da zero producono
`provider.ErrInvalidResponse`. `Close` è idempotente; una lettura dopo una
chiusura anticipata produce `provider.ErrInvalidStream`.

Il contesto usato per aprire lo stream conserva cancellazione e deadline e
viene propagato alla richiesta HTTP.

---

# Traduzione e validazione

La prima versione traduce soltanto messaggi testuali con i ruoli neutrali di
Maestro. Tool call, contenuti multimodali e reasoning non entrano ancora nei
contratti pubblici.

Per completion e streaming viene richiesto un solo risultato (`n: 1`). Le
risposte sincrone devono contenere esattamente una choice. Gli embedding vengono
riordinati in base al loro indice e devono avere cardinalità coerente con gli
input, indici univoci e vettori non vuoti.

Le risposte di errore HTTP sono lette con un limite di 64 KiB. Il messaggio
strutturato `error.message` viene preferito al testo grezzo. Gli errori di
cancellazione e deadline restano riconoscibili tramite `errors.Is`.

---

# Utilizzo

```go
config := runtime.NewConfig(map[string]any{
    provider.ConfigDefaultProvider: llamacpp.ID,
})

rt := maestro.New(maestro.WithConfig(config))

llamaProvider, err := llamacpp.New(llamacpp.Config{
    BaseURL:      "http://localhost:8080",
    DefaultModel: "local-model",
})
if err != nil {
    return err
}

if err := rt.Providers().Register(llamaProvider); err != nil {
    return err
}
```

---

# Test

La suite ordinaria usa un trasporto HTTP in-memory e non richiede un processo
`llama-server` né modelli locali.

Uno smoke test live viene mantenuto dietro il build tag `integration` e usa
variabili d'ambiente dedicate. La verifica live non fa parte della suite
ordinaria perché disponibilità e capability dipendono dal processo e dal
modello caricati.

```bash
MAESTRO_LLAMACPP_BASE_URL=http://localhost:8080 \
MAESTRO_LLAMACPP_CHAT_MODEL=local-chat \
MAESTRO_LLAMACPP_EMBED_MODEL=local-embed \
go test -tags=integration ./pkg/provider/llamacpp
```

`MAESTRO_LLAMACPP_API_KEY` è opzionale. Senza
`MAESTRO_LLAMACPP_BASE_URL` il test viene saltato; chat ed embedding vengono
verificati soltanto quando è configurato il relativo modello.

---

# Limiti della prima versione

Non sono ancora tradotti:

* tool calling;
* contenuti multimodali;
* reasoning;
* output strutturati;
* opzioni di sampling specifiche;
* endpoint nativi non compatibili con OpenAI;
* lifecycle del processo `llama-server`;
* caricamento, download e rimozione dei modelli;
* retry, backoff o circuit breaker.

Queste estensioni verranno introdotte soltanto quando esisteranno contratti
neutrali o requisiti operativi concreti.

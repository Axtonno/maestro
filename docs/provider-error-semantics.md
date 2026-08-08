# Maestro Provider Error Semantics

Versione: 0.1.0

Stato: Implementato

Ultimo aggiornamento: 2026-08-08

---

# Scopo

La Fase 6 della Provider Layer introduce un envelope di errore stabile e
provider-neutral. I consumer possono decidere in base a campi tipizzati e a
`errors.Is`/`errors.As`, senza interpretare messaggi, payload proprietari o
status HTTP.

La classificazione descrive il fallimento osservato. Non esegue retry e non
stabilisce da sola se ripetere un'operazione: idempotenza, budget, backoff,
jitter e stato del circuit breaker appartengono alla Fase 7.

---

# Contratto pubblico

`ProviderError` espone:

- `Kind`: classe neutrale;
- `Operation`: operazione provider fallita;
- `Provider` e `Model`: identità disponibili al confine dell'adapter;
- `StatusCode`: status HTTP remoto, oppure zero quando assente;
- `Retryable`: indicazione tecnica conservativa, non una decisione di retry;
- `RemoteType` e `RemoteCode`: identificatori strutturati quando il protocollo
  li fornisce;
- `Message`: dettaglio remoto normalizzato e limitato;
- `Unwrap`: causa originale per preservare la catena Go.

Esempio di consumo:

```go
var providerError *provider.ProviderError
if errors.As(err, &providerError) {
    switch providerError.Kind {
    case provider.ErrorKindRateLimited,
        provider.ErrorKindUnavailable:
        // Una policy separata valuta idempotenza e budget prima del retry.
    }
}

if errors.Is(err, context.Canceled) {
    // La cancellazione originale rimane riconoscibile.
}
```

I sentinel esistenti `ErrInvalidRequest`, `ErrInvalidResponse` ed
`ErrUnsupportedCapability` restano compatibili. Sono inoltre disponibili
sentinel neutrali per autenticazione, modello assente, capability assente,
indisponibilità, capacità esaurita, rate limit, errore transitorio ed errore
interno.

---

# Tassonomia

| Kind | Significato | Retryable predefinito |
|---|---|---:|
| `invalid_request` | richiesta o parametri non validi | no |
| `authentication` | autenticazione o autorizzazione rifiutata | no |
| `model_not_found` | model ID richiesto non disponibile | no |
| `capability_not_found` | endpoint o capability non supportata | no |
| `unavailable` | servizio o dipendenza non raggiungibile | sì |
| `capacity_exhausted` | capacità o storage temporaneamente esauriti | sì |
| `rate_limited` | limite di richieste imposto dal provider | sì |
| `transient` | timeout o errore temporaneo di trasporto | sì |
| `invalid_response` | risposta di successo malformata o incoerente | no |
| `internal` | errore remoto o adapter non classificabile | no, salvo HTTP 5xx |
| `canceled` | context cancellato | no |
| `deadline_exceeded` | deadline del context superata | no |

`Retryable=true` significa soltanto che la causa può essere temporanea. Non
supera mai la cancellazione del context e non rende automaticamente sicure
operazioni mutanti o stream che hanno già consegnato dati.

---

# Matrice HTTP comune

La struttura nativa dell'errore ha precedenza sullo status quando disponibile.
In assenza di una classe strutturata, Ollama e llama.cpp condividono questa
baseline:

| Status | Kind | Retryable |
|---|---|---:|
| 400, 413, 422 | `invalid_request` | no |
| 401, 403 | `authentication` | no |
| 404 | `model_not_found` se l'operazione identifica un modello; altrimenti `capability_not_found` | no |
| 408 | `transient` | sì |
| 409 | `unavailable` | sì |
| 429 | `rate_limited` | sì |
| 501 | `capability_not_found` | no |
| 502, 503, 504 | `unavailable` | sì |
| 507 | `capacity_exhausted` | sì |
| altri 4xx | `invalid_request` | no |
| altri 5xx | `internal` | sì |

Gli errori di connessione sono `unavailable`; un errore di rete che implementa
`net.Error` e dichiara timeout è `transient`. Le risposte 2xx che non rispettano
il contratto sono `invalid_response` e non vengono confuse con errori remoti.

---

# Mapping degli adapter

## Ollama

Ollama restituisce errori HTTP nel campo JSON `error`; può inoltre inviare un
errore durante lo streaming mantenendo lo status HTTP 200. Maestro applica la
stessa classificazione sia all'apertura sia a `Recv`, usando la matrice HTTP
quando lo status è presente e `internal` per un errore remoto mid-stream privo
di status.

La documentazione ufficiale elenca come casi comuni 400, 404, 429, 500 e 502 e
specifica la semantica degli errori durante lo streaming:
[Ollama API errors](https://docs.ollama.com/api/errors).

## llama.cpp

llama.cpp usa un envelope OpenAI-like con `message`, `type` e `code`. Maestro
preserva `type` e `code` e dà priorità a `type`:

| `type` llama.cpp | Kind | Retryable |
|---|---|---:|
| `invalid_request_error`, `exceed_context_size_error` | `invalid_request` | no |
| `authentication_error`, `permission_error` | `authentication` | no |
| `not_found_error` | modello o capability assente in base all'operazione | no |
| `not_supported_error` | `capability_not_found` | no |
| `unavailable_error` | `unavailable` | sì |
| `rate_limit_error` | `rate_limited` | sì |
| `capacity_error`, `insufficient_quota` | `capacity_exhausted` | sì |
| `server_error` | `internal` | sì |

La struttura degli errori e i tipi esposti dal server sono descritti nella
[documentazione ufficiale di llama.cpp server](https://github.com/ggml-org/llama.cpp/blob/master/tools/server/README.md).

---

# Context, stream e wrapping

`context.Canceled` e `context.DeadlineExceeded` hanno kind dedicati e restano
raggiungibili tramite `errors.Is`. `io.EOF` continua a indicare il termine
normale di uno stream e non viene trasformato in `ProviderError`.

Gli errori di apertura, ricezione e chiusura degli stream riportano la stessa
operazione e lo stesso model ID. La causa originale resta accessibile con
`Unwrap`, inclusi errori aggregati prodotti dalla pulizia di una pull.

---

# Limiti e sicurezza

Il body di un errore HTTP letto dagli adapter è limitato a 64 KiB. I campi
testuali pubblicati da `ProviderError` vengono ridotti a una singola riga e a
un massimo di 512 byte UTF-8 ciascuno. Prompt, messaggi di chat, embedding,
token e API key non vengono copiati nell'envelope.

I messaggi restano diagnostici e non costituiscono API. Ogni decisione
operativa deve usare `Kind`, `Operation`, `Retryable`, gli altri campi tipizzati
o i sentinel tramite `errors.Is`.

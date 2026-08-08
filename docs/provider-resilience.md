# Maestro Provider Resilience Policies

Versione: 0.1.0

Stato: Implementato

Ultimo aggiornamento: 2026-08-08

---

# Scopo

La Fase 7 della Provider Layer introduce retry, backoff, jitter e circuit
breaker nel Provider Runtime. Le policy sono opt-in: in loro assenza il Runtime
esegue una sola chiamata e restituisce lo stesso stream o errore delle fasi
precedenti.

La resilienza usa esclusivamente `ProviderError.Retryable`, context e contratti
tipizzati. Non analizza messaggi né payload proprietari e non introduce
fallback, load balancing o selezione automatica del provider.

---

# Contratto pubblico

Una policy identifica sempre un'operazione e può identificare un modello
esatto:

```go
type ResiliencePolicy struct {
    Operation         Operation
    Model             string
    MaxAttempts       uint
    InitialBackoff    time.Duration
    MaxBackoff        time.Duration
    BackoffMultiplier float64
    Jitter            float64
    MaxElapsedTime    time.Duration
    CircuitBreaker    CircuitBreakerPolicy
}
```

`MaxAttempts` include il primo tentativo. Un modello vuoto applica la policy a
tutti i modelli dell'operazione; una policy con model ID esatto ha precedenza.
Listing e discovery non accettano un modello perché operano sull'intero
catalogo. Se una richiesta usa il modello predefinito dell'adapter lasciando
vuoto `request.Model`, può corrispondere soltanto alla policy generale.

Il Runtime espone:

```go
SetResiliencePolicy(context.Context, provider.ID, ResiliencePolicy) error
ResiliencePolicy(provider.ID, Operation, string) (ResiliencePolicy, bool, error)
CircuitState(provider.ID, Operation, string) (CircuitSnapshot, bool, error)
```

L'impostazione di una policy sostituisce atomicamente la configurazione dello
stesso target e resetta il relativo breaker. Una policy con un solo tentativo e
breaker disabilitato equivale operativamente a nessun retry.

---

# Matrice di ripetibilità

| Operazione | Retry automatico | Confine |
|---|---:|---|
| completion | sì | soltanto prima di restituire un risultato |
| streaming | sì | apertura ed errori precedenti al primo chunk |
| embedding | sì | soltanto prima di restituire un risultato |
| model listing | sì | lettura senza mutazioni |
| model discovery | sì | lettura senza mutazioni |
| model load | sì | intento convergente: assicurare il caricamento |
| model unload | sì | intento convergente: assicurare il rilascio |
| capability introspection | sì | snapshot in sola lettura |
| model pull | no | può avviare o riavviare un trasferimento remoto |
| model remove | no | mutazione del catalogo non ripetibile in modo uniforme |

Completion può produrre un testo differente dopo un retry; viene considerata
ripetibile perché nessun risultato precedente è stato consegnato e Maestro non
osserva una mutazione applicativa. Pull e remove possono usare un circuit
breaker, ma una policy con `MaxAttempts > 1` viene rifiutata con
`ErrInvalidResiliencePolicy`.

---

# Retry, backoff e budget

Un retry avviene soltanto quando tutte le condizioni sono vere:

- l'operazione è ripetibile secondo la matrice;
- l'errore contiene un `ProviderError` con `Retryable=true`;
- rimane almeno un tentativo;
- il prossimo backoff non supera `MaxElapsedTime`, quando configurato;
- il context non è cancellato né scaduto.

Richieste invalide, errori permanenti, `context.Canceled` e
`context.DeadlineExceeded` non vengono ritentati. Le attese usano sempre il
context del chiamante.

Il backoff parte da `InitialBackoff`, viene moltiplicato per
`BackoffMultiplier` dopo ogni fallimento ed è saturato a `MaxBackoff`. `Jitter`
è un fattore simmetrico compreso tra 0 e 1; per esempio 0,2 applica una
variazione tra −20% e +20%, sempre entro `MaxBackoff`.

Clock, attesa e sorgente casuale sono dipendenze sostituibili internamente. La
suite verifica sequenze, jitter e budget senza dormire realmente.

---

# Streaming

La policy copre sia l'apertura sia `Recv` fino al primo chunk. Se il provider
restituisce un errore ritentabile prima di quel punto, Maestro chiude lo stream
fallito, attende il backoff e ne apre uno nuovo con la stessa richiesta e lo
stesso lease di residenza.

Dopo la prima consegna non viene mai aperto un nuovo stream: il consumer
potrebbe avere già osservato contenuto, usage o finish reason e un riavvio
produrrebbe duplicazioni. `io.EOF` rimane il completamento normale.

Il model pull non viene riavviato. Quando è configurato un breaker, il suo esito
viene registrato al completamento, al fallimento o alla chiusura dello stream.

---

# Circuit breaker

Ogni policy possiede un circuito indipendente identificato da provider,
operazione e modello opzionale. Lo stato è:

- `closed`: le operazioni sono ammesse e i fallimenti consecutivi vengono
  contati;
- `open`: le operazioni sono rifiutate localmente con `ErrCircuitOpen`;
- `half_open`: dopo `OpenDuration` sono ammesse al massimo
  `HalfOpenMaxAttempts` probe concorrenti.

Il breaker conta l'esito finale dell'operazione dopo tutti i retry. Soltanto un
errore finale ritentabile incrementa la soglia; un esito valido o un errore
permanente prova che il servizio risponde e azzera i fallimenti. Cancellazione
e deadline sono neutrali; l'ultimo errore remoto resta invece l'esito quando il
budget temporale impedisce un altro tentativo.

Un probe half-open riuscito chiude il circuito; un fallimento ritentabile lo
riapre. Le operazioni eccedenti il limite half-open vengono rifiutate senza I/O
remoto. `CircuitSnapshot` espone contatori, probe in corso, istante di apertura
e prossimo istante ammesso, senza consentire mutazioni esterne.

---

# Concorrenza e ownership

Registry, policy e breaker sono protetti separatamente. Il Runtime acquisisce
un permesso e copia la policy prima di eseguire attese, callback dello stream o
I/O del provider. Nessun lock di registry, residenza o circuito viene mantenuto
durante codice esterno.

Le operazioni di autoload e unload avviate dalle policy di residenza usano le
eventuali policy resilience di discovery, load e unload, mantenendo invariata
la loro ownership.

---

# Limiti

Questa fase non introduce:

- retry di pull, remove o stream che hanno già consegnato dati;
- resume o deduplicazione dei trasferimenti;
- fallback verso altri provider;
- circuit breaker distribuiti o persistenti;
- observer, metriche, tracing o logging delle transizioni;
- timeout autonomi che sostituiscono il context del chiamante.

Gli observer e la correlazione dei tentativi appartengono alla Fase 8. Gli
scenari live confluiscono nello Smoke Benchmark della Milestone 3.

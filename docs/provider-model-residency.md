# Maestro Provider Model Residency

Versione: 0.1.0

Stato: Implementato

Ultimo aggiornamento: 2026-08-08

---

# Scopo

La Fase 4 della Provider Layer introduce policy opt-in per caricare un modello
prima di completion, streaming o embedding e rilasciarlo secondo una semantica
comune a Ollama e llama.cpp.

Il Provider Runtime coordina soltanto lease, timer e ownership delle transizioni
che ha avviato. `ModelDiscoverer` rimane la fonte osservabile dello stato
effettivo: Maestro non mantiene una copia autorevole del catalogo remoto.

---

# Contratto pubblico

`pkg/provider` espone:

```go
type ModelResidencyPolicy struct {
    Model      string
    Autoload   bool
    KeepAlive  time.Duration
    Persistent bool
}
```

Il contratto `provider.Runtime` aggiunge:

```go
SetModelResidencyPolicy(context.Context, provider.ID, ModelResidencyPolicy) error
ResidencyPolicy(provider.ID, string) (ModelResidencyPolicy, bool, error)
Shutdown(context.Context) error
```

La chiave di una policy è la coppia provider–model. Il modello deve essere un ID
esatto, non vuoto e privo di spazi iniziali o finali. Un provider ID vuoto usa
il default configurato.

Una policy assente non aggiunge discovery, load, unload o timer e lascia
invariato il comportamento precedente. Una policy con `Autoload: false` e
durata zero disabilita la regola esistente; non sono valide combinazioni che
richiedono retention con autoload disabilitato.

---

# Semantica di permanenza

Quando `Autoload` è abilitato:

| Configurazione | Comportamento dopo l'ultimo lease |
|---|---|
| `KeepAlive == 0`, `Persistent == false` | unload immediato |
| `KeepAlive > 0` | unload alla scadenza del TTL |
| `Persistent == true` | permanenza fino a disable o shutdown |

`KeepAlive` e `Persistent` sono mutuamente esclusivi. Il TTL parte quando si
chiude l'ultima operazione attiva. Una nuova operazione prima della scadenza
annulla il timer e riusa la residenza posseduta.

Completion ed embedding rilasciano il lease al ritorno del provider, anche in
caso di errore. Uno stream mantiene il lease fino a `io.EOF`, a un errore
terminale o a `Close`; la chiusura e il rilascio sono idempotenti.

La policy si applica soltanto quando la richiesta contiene esplicitamente
`request.Model`. Il modello predefinito interno a un adapter non viene inferito
dal Runtime e non attiva autoload implicito.

---

# Discovery e ownership

Prima del primo lease Maestro interroga `ModelDiscoverer`:

- `loading`, `loaded` e `sleeping` sono considerati residenti;
- gli altri stati, un modello assente o uno stato sconosciuto richiedono
  `LoadModel`;
- soltanto un load eseguito con successo da Maestro assegna ownership alla
  policy.

Quando il modello era già residente, l'operazione procede senza load e senza
unload successivo. In questo modo una policy non interferisce con processi,
utenti o componenti esterni che hanno caricato lo stesso modello.

Il Runtime non assume che il proprio ultimo risultato descriva ancora il
provider. Dopo il rilascio di una residenza non posseduta, la richiesta
successiva consulta nuovamente discovery.

---

# Concorrenza e context

Le operazioni per la stessa coppia provider–model condividono un contatore di
lease. Se più richieste arrivano durante discovery o load, una sola esegue la
transizione e le altre attendono senza mantenere il lock del registry o
effettuare load duplicati.

Le attese e le transizioni iniziali rispettano il `context.Context` della
richiesta. Il rilascio successivo a una richiesta usa un cleanup limitato nel
tempo, così la cancellazione del context operativo non impedisce di liberare
una residenza posseduta. I timer usano uno scheduler sostituibile e lo shutdown
usa il context ricevuto dal Runtime Core.

`Runtime.Stop` invoca lo shutdown del Provider Runtime dopo il lifecycle dei
componenti. Lo shutdown annulla i timer, attende i lease attivi e scarica
soltanto le residenze possedute da Maestro.

---

# Comandi espliciti e policy runtime

`LoadModel` e `UnloadModel` rimangono comandi espliciti e request-scoped: il
chiamante decide una singola transizione e ne possiede le conseguenze. Non
creano una policy e il Runtime non attribuisce a sé quella residenza.

`SetModelResidencyPolicy` configura invece una regola runtime riutilizzata da
più completion, stream ed embedding. Questa separazione evita che un hint o un
comando locale modifichi silenziosamente il comportamento delle richieste
future.

---

# Traduzione nei provider

## Ollama

`ModelLoader` mantiene il modello caricato tramite una richiesta vuota con
`keep_alive: -1`; `ModelUnloader` usa `keep_alive: 0`. Il Provider Runtime
governa il TTL comune e invoca unload alla scadenza. Ollama documenta
`keep_alive` per le richieste chat e generate e i valori usati per preload e
unload nella propria [Chat API](https://docs.ollama.com/api/chat) e nelle
[FAQ](https://docs.ollama.com/faq#how-can-i-preload-a-model-into-ollama-to-get-faster-response-times).

## llama.cpp

In router mode l'adapter usa `/models/load` e `/models/unload`. Discovery usa il
catalogo `/models`. Un processo single-model o una versione senza router può
rifiutare questi endpoint; la policy conserva l'errore remoto. Gli endpoint e
l'autoload del router sono descritti nella documentazione ufficiale di
[`llama-server`](https://github.com/ggml-org/llama.cpp/blob/master/tools/server/README.md).

---

# Esempio

```go
err := rt.Providers().SetModelResidencyPolicy(
    ctx,
    ollama.ID,
    provider.ModelResidencyPolicy{
        Model:     "gemma4",
        Autoload:  true,
        KeepAlive: 2 * time.Minute,
    },
)
if err != nil {
    return err
}

response, err := rt.Providers().Complete(
    ctx,
    ollama.ID,
    provider.CompletionRequest{
        Model: "gemma4",
        Messages: []provider.Message{{
            Role: provider.RoleUser,
            Content: "Ciao",
        }},
    },
)
```

---

# Test e limiti

La suite usa scheduler e provider sostituibili per verificare senza attese
reali:

- comportamento opt-in;
- ownership interna ed esterna;
- unload immediato, TTL e persistenza;
- rilascio terminale degli stream;
- coalescing di richieste concorrenti;
- shutdown del Provider Runtime e integrazione con il Runtime Core;
- validazione e capability mancanti.

Selezione hardware-aware, scelta automatica del modello, supervisione dei
processi e persistenza delle policy tra avvii restano fuori scope. Gli scenari
live confluiscono nello Smoke Benchmark della Milestone 3.

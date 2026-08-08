# Maestro Provider Capability Introspection

Versione: 0.1.0

Stato: Implementato

Ultimo aggiornamento: 2026-08-08

---

# Scopo

La Fase 5 della Provider Layer rende interrogabile il supporto dell'adapter e la
disponibilità osservabile di una specifica istanza o di un modello, senza
introdurre cache, routing automatico o una seconda fonte di verità.

Un report è uno snapshot. Il risultato può cambiare immediatamente dopo il
probe perché catalogo, modello o processo remoto possono cambiare stato.

---

# Contratto pubblico

`CapabilityInspector` è una capability provider opzionale:

```go
type CapabilityInspector interface {
    Provider
    InspectCapabilities(
        context.Context,
        CapabilityRequest,
    ) (CapabilityReport, error)
}
```

Il Provider Runtime espone la stessa operazione attraverso:

```go
Capabilities(
    context.Context,
    provider.ID,
    CapabilityRequest,
) (CapabilityReport, error)
```

Un provider che non implementa `CapabilityInspector` produce
`ErrUnsupportedCapability` prima di qualsiasi chiamata remota. Il Runtime
valida inoltre provider, target, model ID, ordine dei descriptor e valori enum;
un report incoerente produce `ErrInvalidResponse`.

---

# Target di introspection

| Target | Model ID | I/O remoto | Significato |
|---|---:|---:|---|
| `adapter` | vietato | no | supporto strutturale compilato nell'adapter |
| `instance` | vietato | sì | disponibilità osservabile dell'istanza configurata |
| `model` | obbligatorio | sì | disponibilità per un model ID esatto |

Il target deve essere sempre esplicito. Il Runtime non sostituisce un model ID
mancante con il default dell'adapter e non sceglie un modello in base al report.

---

# Supporto e disponibilità

Ogni `CapabilityDescriptor` possiede due assi indipendenti:

- `Support` descrive se l'adapter Maestro implementa il contratto;
- `Availability` descrive ciò che il target rende osservabile nello snapshot.

Valori di supporto:

- `supported`;
- `unsupported`.

Valori di disponibilità:

- `unknown`: il dato non è osservabile in modo affidabile;
- `available`: il target dichiara o rende osservabile la capability;
- `unavailable`: il target dichiara l'assenza o si trova in uno stato
  incompatibile.

Una capability `unsupported` è sempre `unavailable`. `unknown` non deve essere
interpretato come `unavailable`: può indicare una configurazione non esposta
dal protocollo.

---

# Capability neutrali

`KnownCapabilities` restituisce una copia nell'ordine canonico seguente:

1. `completion`;
2. `streaming`;
3. `embedding`;
4. `model_listing`;
5. `model_discovery`;
6. `model_load`;
7. `model_unload`;
8. `model_pull`;
9. `model_remove`;
10. `structured_output`;
11. `tool_calling`.

Ollama e llama.cpp implementano oggi le prime nove capability nell'adapter.
`structured_output` e `tool_calling` restano `unsupported` anche quando il
server o il modello sottostante le offre: i contratti neutrali Maestro verranno
aggiunti nella Fase 9.

---

# Ollama

Il target `instance` usa `GET /api/tags` per verificare l'istanza. Le capability
di catalogo e lifecycle risultano disponibili; completion, streaming ed
embedding restano `unknown` senza un modello esplicito.

Il target `model` verifica prima la presenza nel catalogo e poi usa
`POST /api/show`. Le capability Ollama `completion` ed `embedding` determinano
rispettivamente completion/streaming ed embedding nel report. La
documentazione ufficiale di [Show model details](https://docs.ollama.com/api-reference/show-model-details)
definisce il campo `capabilities` model-specific.

Un modello assente non è un errore di protocollo: inference, load, unload e
remove risultano non disponibili, mentre pull rimane disponibile sull'istanza.
Metadata vuoti o duplicati vengono rifiutati come risposta non valida.

---

# llama.cpp

I target operativi usano `GET /models`, senza attivare autoload. Lo snapshot
del router include `status` e gli argomenti effettivi del processo modello,
come documentato nella sezione [Using multiple models](https://github.com/ggml-org/llama.cpp/blob/master/tools/server/README.md#using-multiple-models).

Semantica:

- un catalogo con `status` identifica il router e rende disponibili load,
  unload, pull e remove;
- una risposta single-model senza `status` rende indisponibili le capability
  del router;
- un catalogo vuoto lascia la modalità `unknown`;
- modelli failed, downloading o loading non sono disponibili per inference;
- `--embedding` o `--embeddings` rende disponibile embedding e indisponibili
  completion e streaming;
- una disabilitazione embedding esplicita produce la semantica inversa;
- se la modalità può dipendere da environment ereditato e non è osservabile
  dagli argomenti, entrambe rimangono `unknown`.

La documentazione ufficiale specifica che `--embedding` restringe il processo
al caso embedding e che il router eredita argomenti ed environment del processo
principale. Maestro evita quindi inferenze ottimistiche quando il dato non è
presente.

---

# Routing operativo

L'introspection non viene eseguita automaticamente prima di ogni completion,
stream o embedding: aggiungerebbe I/O, introdurrebbe una race tra probe e
operazione e modificherebbe il comportamento esistente. Il routing continua a
rifiutare prima del provider le capability strutturali assenti tramite le
interfacce Go opzionali.

Il consumer può usare `Capabilities` come preflight esplicito. Il report non
abilita fallback, load balancing o scelta automatica del modello.

---

# Concorrenza, variazioni e limiti

Gli adapter non conservano cache o mappe mutabili dei report. Ogni query
operativa legge un nuovo snapshot, costruisce slice e mappe locali e può essere
eseguita concorrentemente.

La suite verifica:

- target adapter senza I/O;
- modalità instance e model per entrambi i provider;
- ordine canonico e validazione dei report;
- modello assente e capability non osservabili;
- variazione del catalogo tra due query;
- query concorrenti con race detector;
- rifiuto pre-I/O di richieste e capability non supportate.

Un report non garantisce l'esito della richiesta successiva. La semantica degli
errori è completata nella Fase 6, la resilienza nella Fase 7 e l'osservabilità
nella Fase 8. Gli scenari live confluiscono nello Smoke Benchmark della
Milestone 3.

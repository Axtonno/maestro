# Maestro Smoke Benchmark

Versione: 0.1.0

Stato: Implementato

Ultimo aggiornamento: 2026-08-09

---

# Scopo

Lo Smoke Benchmark verifica dal vivo una configurazione completa di provider e
modelli fixture usando esclusivamente i contratti pubblici della Provider Layer.
Assorbe gli smoke test rinviati durante la Milestone 2 senza trasformarli in
requisiti della suite deterministica.

Ogni report rappresenta un solo provider. I modelli sono associati ai ruoli
`chat`, `embedding`, `lifecycle` e `acquisition_fixture`.

---

# Comando

```text
maestro bench smoke [opzioni]
```

Opzioni principali:

```text
--provider ollama|llama.cpp
--manifest docs/provider-smoke-benchmark-manifest.yaml
--output -|report.json
--warmup 0
--runs 1
--timeout 2m
--cleanup-timeout 30s
--fail-on-failure
```

Il provider predefinito del comando è `ollama`. Il comando non usa gli endpoint
localhost predefiniti degli adapter: se la variabile `*_BASE_URL` richiesta non
è presente, tutti gli scenari sono `skipped` con reason code
`provider_not_configured` e non viene effettuato I/O remoto.

Il report viene scritto su stdout con `--output -`. Un path esplicito viene
pubblicato atomicamente con permessi `0600`.

---

# Configurazione Ollama

```text
MAESTRO_OLLAMA_BASE_URL
MAESTRO_OLLAMA_CHAT_MODEL
MAESTRO_OLLAMA_EMBED_MODEL
MAESTRO_OLLAMA_LIFECYCLE_MODEL
MAESTRO_OLLAMA_ACQUISITION_MODEL
```

# Configurazione llama.cpp

```text
MAESTRO_LLAMACPP_BASE_URL
MAESTRO_LLAMACPP_API_KEY                 # opzionale
MAESTRO_LLAMACPP_CHAT_MODEL
MAESTRO_LLAMACPP_EMBED_MODEL
MAESTRO_LLAMACPP_LIFECYCLE_MODEL
MAESTRO_LLAMACPP_ACQUISITION_MODEL
```

I modelli mancanti saltano soltanto gli scenari del relativo ruolo. Il valore
della API key non entra mai nel profilo o nel report.

---

# Matrice degli scenari

| Scenario | Verifica | Cleanup |
|---|---|---|
| `capability-instance` | Introspection dell'istanza e report canonico | Nessuno |
| `catalog-list-discover` | Listing e discovery con ID validi | Nessuno |
| `completion-simple` | Completion non vuota e usage disponibile | Nessuno |
| `stream-terminal-close` | Chunk, terminale, EOF e chiusura | Chiusura stream |
| `stream-cancel-deadline` | Propagazione della cancellazione | Cancel e chiusura |
| `embedding` | Un vettore non vuoto con valori finiti | Nessuno |
| `lifecycle-load-unload` | Load del modello fixture | Unload fixture |
| `acquisition-pull-remove` | Pull fino a terminale ed EOF | Close e remove fixture |
| `structured-json` | Oggetto JSON sintatticamente valido | Nessuno |
| `structured-json-schema` | Oggetto coerente con la fixture schema | Nessuno |
| `tool-call-result` | Tool call completa e secondo turno con risultato | Nessuno |
| `tool-call-stream` | Delta tool ricomposti in arguments JSON | Chiusura stream |
| `resilience-controlled-error` | Errore di cancellazione controllato sotto policy | Ripristino policy baseline |
| `observability-redaction` | Sequenza, correlazione e assenza del prompt | Rimozione observer |

---

# Capability preflight

Prima di ogni operazione lo scenario interroga capability introspection:

- supporto strutturale `unsupported` produce stato `unsupported`;
- availability `unavailable` produce stato `skipped` con
  `capability_unavailable`;
- availability `unknown` consente il probe live, perché rappresenta assenza di
  conoscenza e non assenza della capability;
- un modello di ruolo non configurato produce `model_not_configured`.

Errori del probe su un provider esplicitamente configurato sono fallimenti
operativi e vengono classificati nel report.

---

# Sicurezza delle mutazioni

Acquisition è disabilitata finché non viene impostato esattamente:

```text
MAESTRO_ALLOW_CATALOG_MUTATION=true
```

Valori diversi da `true`, `false` o vuoto sono rifiutati. Il runner verifica
inoltre che il manifest conservi la mutation guard approvata.

Prima del pull viene eseguita discovery. Se la fixture è già presente, lo
scenario è `skipped` e non viene mai rimossa. Solo una fixture assente e
acquisita dalla run diventa proprietà del cleanup.

Load/unload usa un modello lifecycle dedicato. Tutti gli stream vengono chiusi
dal cleanup del runner anche dopo errori, panic o cancellazione.
SIGINT e SIGTERM cancellano la run in modo cooperativo prima dello shutdown.

---

# Report ed exit status

La Fase 2 estende lo schema raw a `1.1.0` aggiungendo
`configuration.models`, una mappa role–model profile. Endpoint e model ID che
sono path vengono redatti prima della serializzazione.

Per default scenari `failed`, `skipped` e `unsupported` rimangono dati del
report e il comando termina con `0`. `--fail-on-failure` abilita il gate e
restituisce `1` quando almeno uno scenario è `failed`; capability unsupported
e scenari saltati non fanno fallire il gate.

Errori di configurazione, runner, shutdown o scrittura restituiscono `1`. Uso
non valido della CLI restituisce `2`.

---

# Limiti

Lo Smoke Benchmark non raccoglie ancora distribuzioni di latenza, token/sec,
CPU, RAM o VRAM. Non valuta qualità comparativa e non attribuisce punteggi ai
modelli. Queste responsabilità appartengono alle Fasi 3 e 4.

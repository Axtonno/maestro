# Maestro Developer Benchmark — Laravel/PHP

Versione: 1.0.0

Stato: Fase 4 implementata

Ultimo aggiornamento: 2026-08-09

---

# Scopo

Il Developer Benchmark misura l'utilità pratica di una configurazione Maestro
su task PHP/Laravel riproducibili. Non produce una classifica globale dei
modelli e non usa un evaluator LLM obbligatorio.

Lo stato tecnico e la qualità sono separati. Una completion non vuota o un
batch embedding valido produce uno scenario `passed`; lo stesso campione può
ricevere uno score qualitativo da 0 a 3.

---

# Dataset

Il dataset embedded è identificato da:

```text
maestro-laravel-mini@1.0.0
```

Le sorgenti versionate sono in:

```text
internal/benchmark/developer/testdata/laravel-v1
```

La fixture contiene un endpoint di creazione ordine, controller, service,
repository, payment contract, evento e un metodo separato da rifattorizzare.
Non contiene credenziali, dati utente o dipendenze installate.

All'avvio `bench laravel` materializza la fixture in una directory temporanea
privata, carica realmente il plugin Laravel `0.1.0` e verifica il marker
`artisan` e la dipendenza Composer. La directory viene eliminata durante lo
shutdown anche nelle run offline.

---

# Scenari e rubriche

| Scenario | Task | Evidenza per i tre punti |
|---|---|---|
| `developer-explain-controller` | Spiegazione del controller | validation, `OrderService`, risposta 201 |
| `developer-identify-dependencies` | Dipendenze del service | repository, payment gateway, event dispatcher |
| `developer-generate-phpunit-test` | Test PHPUnit | struttura test, request/201, mock di `OrderService` |
| `developer-refactor-method` | Refactor behavior-preserving | preferred, fallback, risultato null |
| `developer-summarize-project` | Sintesi del progetto | route, controller/service/repository, payment |
| `developer-retrieve-with-embeddings` | Retrieval dei file rilevanti | prima fixture rilevante nel ranking cosine |

I primi cinque scenari applicano una checklist lessicale deterministica. Ogni
criterio soddisfatto vale un punto. Alcuni criteri richiedono tutti i termini
dichiarati, per esempio mock più `OrderService`.

Il retrieval calcola la similarità cosine tra embedding della domanda e delle
sette fixture candidate. Lo score è:

- 3 se la prima fixture rilevante è al rank 1;
- 2 al rank 2;
- 1 al rank 3;
- 0 oltre la top 3.

La rubrica è volutamente trasparente e limitata. Non dimostra correttezza
semantica completa e può essere affiancata da una revisione manuale.

---

# Report 1.2

Ogni campione tecnicamente riuscito include una sezione distinta:

```json
{
  "evaluation": {
    "evaluator": "maestro-laravel-mini@1.0.0:developer-explain-controller",
    "method": "deterministic_term_checklist",
    "score": 3,
    "max_score": 3,
    "rationale_code": "criteria_matched_3_of_3"
  }
}
```

`rationale_code`, metodo ed evaluator accettano soltanto identificatori brevi
e sicuri. Prompt, fixture complete e risposte restano in memoria e non hanno un
campo nel report. Sono registrati solo latenza, usage, dimensioni, ranking e
score.

---

# Esecuzione

Configurazione Ollama minima:

```text
MAESTRO_OLLAMA_BASE_URL=http://127.0.0.1:11434
MAESTRO_OLLAMA_CHAT_MODEL=<modello-chat>
MAESTRO_OLLAMA_EMBED_MODEL=<modello-embedding>
```

Esempio:

```text
maestro bench laravel \
  --provider ollama \
  --dataset maestro-laravel-mini@1.0.0 \
  --output laravel-report.json
```

Per llama.cpp si usano `MAESTRO_LLAMACPP_BASE_URL`,
`MAESTRO_LLAMACPP_CHAT_MODEL`, `MAESTRO_LLAMACPP_EMBED_MODEL` e l'API key
opzionale già documentata per il provider.

Warmup e run sono configurabili. I default sono zero warmup e una run, perché
i task generativi hanno un costo maggiore del Runtime Benchmark.

`--fail-on-failure` abilita il gate tecnico. `--minimum-score N` richiede che
ogni campione misurato abbia una valutazione e uno score almeno pari a `N`, con
`N` tra 0 e 3. Senza questi flag il comando resta una raccolta informativa.

In assenza di base URL tutti i task sono `skipped` senza I/O provider; dataset
e plugin Laravel vengono comunque validati.

---

# Limiti

- La checklist non compila né esegue il codice generato.
- Il report non conserva la risposta, quindi la revisione manuale deve essere
  svolta durante una run diagnostica separata.
- Il retrieval usa il Context Engine con semantic retrieval opt-in; il Provider
  Runtime resta l'unico confine verso gli embedding.
- Il dataset iniziale è piccolo e in lingua inglese per rendere stabili le
  evidenze lessicali.

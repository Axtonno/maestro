# Maestro v0.1 Configuration

Versione schema: 1

Stato: Contratto pubblico sperimentale v0.1.x

Data: 2026-08-15

---

# Scopo

Il file YAML di prodotto descrive tutti i target necessari a diagnosticare ed
eseguire il reference agent. Non sostituisce `pkg/runtime.Config`, che resta il
contratto generico dei componenti Go.

La configurazione read-only completa è in `configs/maestro.example.yaml`. Il
profilo candidato mutativo, separato e opt-in, è in
`configs/maestro.mutating.example.yaml`. Nell'artifact distribuito il relativo
`workspace.root` seleziona `../fixtures/laravel-v1`; quando il file viene
copiato altrove deve essere sostituito con il progetto reale.

# Risoluzione del file

L'ordine è:

1. `--config <path>`;
2. `MAESTRO_CONFIG`;
3. `$XDG_CONFIG_HOME/maestro/config.yaml`;
4. `$HOME/.config/maestro/config.yaml` quando XDG non è impostato.

Non vengono uniti più file. Un path esplicito mancante è un errore. Un root
workspace relativo viene risolto rispetto alla directory del file di
configurazione e diventa assoluto prima della validazione.

# Parsing strict

Il documento deve contenere `version: 1`. Il loader rifiuta:

- campi sconosciuti o duplicati;
- documenti YAML multipli;
- anchor e alias;
- trailing data o YAML malformato;
- file vuoti o maggiori di 1 MiB;
- versione assente o non supportata.

Un typo non viene ignorato e nessuna chiave abilita fallback automatici.

# Schema

## `provider`

| Campo | Requisito |
|---|---|
| `id` | `ollama` oppure `llama.cpp` |
| `base_url` | origine HTTP(S) senza credenziali, query, fragment o path API |
| `timeout` | durata positiva, massimo 10 minuti |
| `api_key_env` | vuoto oppure nome della variabile contenente la API key llama.cpp |

La API key non viene accettata direttamente nel file. Per Ollama
`api_key_env` deve essere vuoto.

## `models`

`chat` è obbligatorio ed esatto. `embedding` è opzionale nella baseline
lessicale della Fase 2; resta disponibile per i successivi percorsi semantic e
per la matrice live.

## `workspace`

La serie v0.1.x supporta il percorso reference Laravel:

```yaml
workspace:
  id: laravel
  root: /absolute/path/to/project
  framework: laravel
```

`id` deve essere `laravel` perché il `WorkspaceProvider` del plugin è la fonte
autorevole dell'identità. La CLI non rinomina o duplica il workspace.

## `agent`

L'agent iniziale è `agent.reference`. Lo schema sperimentale accetta almeno uno
dei workspace tool built-in, senza duplicati:

- `workspace.list`;
- `workspace.read`;
- `workspace.search`;
- `workspace.write`;
- `workspace.patch`.

L'ordine nel file non cambia l'ordine canonico della `RunRequest`.

Il profilo ufficiale v0.1.x è più stretto dello schema generico e contiene
soltanto `workspace.list`, `workspace.read` e `workspace.search`.
`workspace.write` e `workspace.patch` restano capacità sperimentali non
supportate e non devono essere aggiunte al quick start.

`streaming` seleziona il percorso Provider Runtime corrispondente e viene
verificato da `doctor` nelle capability del modello.

## `policy`

Ogni classe accetta `allow`, `deny` o `prompt`:

```yaml
policy:
  id: policy.local-review
  model: allow
  workspace_inspect: allow
  workspace_mutate: deny
```

`model` governa sia `model.invoke` sia l'eventuale `model.disclose` sul bundle
del workspace selezionato. Le classi workspace valgono soltanto per tool
inclusi esplicitamente e action sul workspace `laravel`.

Con un TTY attendibile, `prompt` mostra le action preparate. Per una
`workspace.patch` mostra anche intenzione, path logico, digest, precondizione e
diff content-bound e accetta soltanto deny oppure allow one-shot. EOF, input
invalido, richiesta di grant per run e modalità non interattiva negano in
sicurezza; cancellazione e deadline non concedono authority e interrompono il
run. Non esiste auto-approval né un flag globale `--yes`.

La configurazione inclusa nega sempre `workspace_mutate`. Lo schema 0.x può
ancora decodificare `allow` e `workspace.write` per contratti sperimentali, ma
il composition root di prodotto li rifiuta. I soli profili eseguibili sono:

- read-only, senza tool mutativi e con `workspace_mutate: deny`;
- Controlled Mutation candidato, con `workspace.read`, `workspace.patch` e
  `workspace_mutate: prompt`.

Il secondo profilo non è ancora una compatibility promise: la qualificazione
live appartiene alla Milestone 11.

## `limits`

Tutti i limiti sono obbligatori e devono rispettare i bound di `pkg/agent`:

- `duration`;
- `model_turns`;
- `tool_calls` e `tool_calls_per_turn`;
- `plan_steps` e `plan_revisions`;
- `tool_result_bytes` e `session_bytes`;
- `input_tokens` e `output_tokens`.

Il modello non può aumentarli. `session_bytes` non può essere inferiore a
`tool_result_bytes`. `maestro run` li stampa su stderr prima dell'esecuzione.

## `context`

La Fase 2 accetta il retrieval `lexical` con `top_k` tra 1 e 100. Budget totale,
reserved e safety devono lasciare un allowance positivo per l'evidenza. Il
runtime usa l'estimator built-in `context.utf8-estimator`.

# Secret e redazione

Il file non contiene secret. `api_key_env` conserva soltanto un nome; il valore
viene letto durante la composition e passato direttamente all'adapter. Errori,
doctor e output CLI non stampano il valore.

# Validazione

Eseguire:

```text
maestro doctor --config /path/to/config.yaml
```

Il parsing/configuration failure usa exit code 2. I failure operativi del
preflight usano exit code 1 e sono rappresentati come check separati.

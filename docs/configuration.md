# Maestro v0.2.0 Configuration

Versione schema: 1

Stato: Contratto pubblico sperimentale v0.2.0

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

La v0.2.0 supporta il percorso reference Laravel:

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

Il profilo ufficiale v0.2.0 è più stretto dello schema generico e contiene
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

Il secondo profilo non è una compatibility promise: la Milestone 11 ha
registrato `mutation_deferred` dopo il failure del Gate A live.

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

---

# Schema candidato Direct Chat della Milestone 17

La modalità direct chat usa un documento strict `version: 2`. Durante la
Milestone 17 lo schema resta candidato e non modifica retroattivamente il
contratto v0.2.0. Un file v1 continua a essere valido per `maestro run` e
`maestro agent`, ma `maestro chat` lo rifiuta con `chat_profile_required`; non
viene derivato un profilo chat da `models.chat`.

La forma minima chat-only è in `configs/maestro.chat.example.yaml`:

```yaml
version: 2

provider:
  id: ollama
  base_url: http://127.0.0.1:11434
  timeout: 5m
  api_key_env: ""

workspace:
  id: laravel
  root: /absolute/path/to/project
  framework: laravel

interaction:
  chat:
    model: qwen2.5-coder:7b
    timeout: 5m
    streaming: false
    num_ctx: 4096
    thinking: "false"
    max_file_bytes: 1048576
    max_output_bytes: 1048576

policy:
  workspace_mutate: deny
```

Il loader di `maestro chat` valida soltanto provider, workspace, profilo chat e
il deny mutativo. Non richiede agent, tool, retrieval o budget di sessione e
non li costruisce. I comandi `maestro agent` e `maestro run` continuano invece
a usare la validazione completa: se devono essere abilitati nello stesso file,
`interaction.agent`, `agent`, `limits`, `context` e la policy completa restano
obbligatori e conservano la semantica stabilita da ADR-0033. Una configurazione
chat-only non abilita implicitamente l'agent.

`provider.timeout` resta il ceiling di trasporto comune, mentre il timeout chat
delimita la singola richiesta e non può superarlo. `models.embedding` è
opzionale e non viene usato dal percorso chat.

`interaction.chat.streaming` autorizza il flag esplicito `maestro chat
--stream`; non forza tutte le richieste chat a usare streaming. Il valore false
mantiene il trasporto disabilitato anche quando il provider lo supporta.

`num_ctx` è obbligatorio e positivo. `thinking` è una stringa enum obbligatoria
con valori `default`, `true` o `false`, così il default del provider resta
distinto da un'esplicita disabilitazione. Valori non supportati dall'adapter o
dal modello falliscono il preflight invece di essere ignorati.

I limiti chat sono indipendenti dai budget della sessione agentica:

- `max_file_bytes` limita il solo file disclosed;
- `max_output_bytes` limita l'assemblaggio della response;
- domanda e file restano entrambi soggetti al timeout del profilo;
- non esistono `top_k`, tool budget o fallback per la modalità chat.

Il preflight dedicato si esegue senza completion:

```text
maestro doctor --mode chat --config configs/maestro.chat.example.yaml
```

Controlla schema, root non-symlink, composition provider, capability completion
e opzioni generative richieste. `unknown` non equivale a capability disponibile.

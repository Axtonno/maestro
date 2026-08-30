# Maestro v0.3.0 Direct Chat Configuration

Versione schema: 2

Stato: contratto pubblico sperimentale chat-only

La diagnostica specifica descritta sotto è un **candidate post-v0.3.0** e non
appartiene al binario pubblico v0.3.0.

## Profilo distribuito

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
    model: qwen3.5:9b
    timeout: 5m
    streaming: true
    num_ctx: 4096
    thinking: "false"
    max_file_bytes: 1048576
    max_output_bytes: 1048576

policy:
  workspace_mutate: deny
```

La copia nell’archive usa `../fixtures/laravel-v1` come root relativa. Il
profilo non contiene agent, tool, retrieval o budget di sessione.

## Risoluzione del file

L’ordine è:

1. `--config <path>`;
2. `MAESTRO_CONFIG`;
3. `$XDG_CONFIG_HOME/maestro/config.yaml`;
4. `$HOME/.config/maestro/config.yaml` quando XDG non è impostato.

Non vengono uniti file. Un root relativo viene risolto rispetto alla directory
del YAML e diventa assoluto prima della validazione.

## Parsing strict

Il loader rifiuta campi sconosciuti o duplicati, documenti multipli, anchor,
alias, trailing data, file vuoti, file oltre 1 MiB e versioni diverse da 2.
Un errore non abilita fallback verso un profilo agentico.

## `provider`

| Campo | Requisito |
|---|---|
| `id` | `ollama` nel percorso qualificato |
| `base_url` | origine HTTP(S) senza credenziali, query, fragment o path API |
| `timeout` | durata positiva, massimo 10 minuti |
| `api_key_env` | vuoto per Ollama; contiene solo un nome, mai il secret |

Il timeout provider è il ceiling del trasporto. Cambiare endpoint può inviare
domanda e file a un servizio non qualificato.

## `workspace`

La root deve esistere, essere una directory reale e non un symlink. `--file`
viene risolto soltanto sotto questa root. `id` e `framework` descrivono il
workspace, ma Direct Chat non avvia plugin o detection framework.

## `interaction.chat`

- `model` è esatto; v0.3.0 qualifica soltanto `qwen3.5:9b` con il digest del
  manifest;
- `timeout` delimita la richiesta e non può superare il ceiling provider;
- `streaming: true` autorizza `--stream`, ma non lo abilita implicitamente;
- `num_ctx` deve essere positivo e supportabile dall’adapter;
- `thinking` è una stringa enum: `default`, `true` o `false`;
- `max_file_bytes` e `max_output_bytes` sono hard limit positivi.

La temperatura non è configurabile: Direct Chat la imposta a zero sia per
complete sia per stream. Un generation control non supportato fallisce il
preflight invece di essere ignorato.

## `policy`

`workspace_mutate: deny` è obbligatorio nel profilo chat. Non esistono tool o
approval nel percorso, ma il deny impedisce a una configurazione chat-only di
presentarsi come mutativa.

## Secret e redazione

Il YAML non contiene secret. `api_key_env` conserva soltanto il nome della
variabile; il valore non appare in doctor, errori o output operativo.

## Validazione

```sh
maestro doctor --mode chat --config /path/to/chat.yaml
```

Una configurazione invalida usa exit 2. Failure di provider o capability sono
check operativi distinti e non vengono trasformati in PASS.

Nel candidate post-v0.3.0, CLI e doctor mantengono il reason code
`invalid_request` e aggiungono una diagnostica redatta:

| Categoria | Significato |
|---|---|
| `read_failed` | file assente o non leggibile |
| `yaml_invalid` | documento YAML malformato, duplicato o strutturalmente invalido |
| `unknown_field` | chiave non appartenente allo schema strict |
| `missing_field` | campo obbligatorio non presente |
| `invalid_value` | campo presente ma non valido |

Quando disponibile viene mostrato soltanto il path logico allowlisted, per
esempio `interaction.chat.num_ctx`. Non sono mostrati valore, path del file,
secret o testo dell'errore del decoder.

## Profili agentici storici

Lo schema v1 e i blocchi agentici v2 rimangono contratti sperimentali del
repository, ma non appartengono alla configurazione distribuita né alla
compatibility promise v0.3.0. `maestro chat` non deriva un profilo da
`models.chat` e non costruisce agent come fallback.

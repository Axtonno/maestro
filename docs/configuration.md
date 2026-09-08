# Maestro v0.5.0 Configuration

Versione schema pubblica corrente: 4

Il profilo v4 separa Direct Chat e Controlled Mutation. Il loader è strict:
campi sconosciuti o duplicati, documenti multipli, anchor, alias, trailing
data e fallback impliciti sono rifiutati.

## Profilo distribuito

```yaml
version: 4

provider:
  id: ollama
  base_url: http://127.0.0.1:11434
  timeout: 5m
  api_key_env: ""

workspace:
  id: laravel
  root: /absolute/path/to/project
  framework: laravel

direct_chat:
  model: qwen3.5:9b
  digest: 6488c96fa5faab64bb65cbd30d4289e20e6130ef535a93ef9a49f42eda893ea7
  timeout: 5m
  streaming: false
  num_ctx: 4096
  num_predict: 1024
  thinking: "false"
  residency: 5m
  max_file_bytes: 1048576
  max_output_bytes: 1048576

controlled_mutation:
  enabled: true
  model: qwen2.5-coder:14b
  digest: 9ec8897f747e246e970bc5cfdda85d22f1123dc2e3d34978a010a75968716849
  timeout: 5m
  num_ctx: 4096
  num_predict: 1024
  thinking: "false"
  residency: 5m
  prompt: mutation-host-bound-model-selection-v1
  prompt_sha256: 594659d52ec6142a5ef79c36dc0db4899e7ef1bb3f99d05017410f68bc1ba732
  schema: host-bound-mutation-decision-v1
  schema_sha256: bc3432a8f19867eec8e153adaa4434b688974cf34d24b6bd770e887e0dd7557d
  max_output_bytes: 1048576
```

Nell'archive il file si chiama
`configs/maestro.v0.5.0-candidate.yaml`: il nome è conservato per
compatibilità con il candidate qualificato. Il manifest e
`version --diagnostic` distinguono la release tramite `status=release`.
La copia distribuita usa la fixture inclusa come root relativa.

## Risoluzione del file

L'ordine è:

1. `--config <path>`;
2. `MAESTRO_CONFIG`;
3. `$XDG_CONFIG_HOME/maestro/config.yaml`;
4. `$HOME/.config/maestro/config.yaml` quando XDG non è impostato.

I file non vengono uniti. Un root relativo viene risolto rispetto alla
directory del YAML e diventa assoluto prima della validazione.

## `provider`

| Campo | Contratto v0.5.0 |
| --- | --- |
| `id` | `ollama` |
| `base_url` | Loopback HTTP senza credenziali, query, fragment o path API |
| `timeout` | Durata positiva, massimo 10 minuti |
| `api_key_env` | Vuoto per Ollama; se usato contiene solo il nome della variabile |

Cambiare endpoint può inviare domanda e file a un servizio non qualificato.

## `workspace`

La root deve essere una directory reale e non un symlink. Direct Chat risolve
il file opzionale sotto questa root. Controlled Mutation applica un ulteriore
vincolo: file PHP regolare, non symlink e contenuto sotto `app/`.

`id` e `framework` descrivono il workspace; non attivano detection,
retrieval o plugin.

## `direct_chat`

- modello e digest devono coincidere esattamente;
- `timeout` non può superare il ceiling del provider;
- `streaming` autorizza `--stream`, senza abilitarlo implicitamente;
- `num_ctx`, `num_predict` e limiti byte devono essere positivi;
- `thinking` è una stringa enum: `default`, `true` o `false`;
- `residency` è positiva e non superiore a 10 minuti.

La temperatura è fissata a zero. Un generation control non supportato fallisce
il preflight invece di essere ignorato.

## `controlled_mutation`

La sezione deve essere completa e `enabled: true`. Modello, digest, prompt,
schema e relativi SHA-256 sono parte del contratto qualificato. Non vengono
ereditati da Direct Chat e non sono configurabili liberamente nel percorso
supportato.

Il profilo limita l'output; path, estensione, directory `app/`, coordinate,
TTY, approval e stale check sono inoltre verificati dal comando. La
configurazione da sola non concede authority di scrittura.

Il cambio tra chat e mutation scarica esplicitamente il modello uscente e
attende che il provider non lo riporti più residente prima della nuova
generation. Non è richiesta residenza simultanea.

## Secret e redazione

Il YAML non deve contenere secret. `api_key_env` conserva soltanto il nome di
una variabile; il valore non appare in doctor, errori o output operativo.

## Validazione

```sh
maestro doctor --mode all --config /path/to/v0.5.0.yaml
```

Una configurazione invalida usa exit code 2. Failure di provider, modello,
digest o capability restano check distinti e non diventano PASS.

La diagnostica redatta usa le categorie `read_failed`, `yaml_invalid`,
`unknown_field`, `missing_field` e `invalid_value`. Quando disponibile
mostra soltanto il path logico del campo, mai valore, path fisico, secret o
errore YAML grezzo.

## Compatibilità precedente

Gli schemi v2 e v3 restano leggibili per Direct Chat secondo il loro contratto
storico, ma non ricevono campi o authority v4 impliciti. Non possono abilitare
Controlled Mutation. I profili agentici v1 restano superfici di sviluppo e non
sono convertiti automaticamente.

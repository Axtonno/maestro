# Maestro v0.3.0 CLI

Stato: contratto pubblico sperimentale Direct Chat

## Superficie supportata

```text
maestro chat
maestro doctor --mode chat
maestro version
```

La root help può mostrare comandi storici o di sviluppo. `agent`, `run`,
`models`, `agents` e `bench` non appartengono al support claim v0.3.0 e non
sono fallback di Direct Chat.

## `chat`

```text
maestro chat [--config path] [--file logical-path] [--stream] [question]
```

La domanda può essere un argomento posizionale oppure stdin bounded, mai
entrambi. `--file` è opzionale e singolo; accetta soltanto un path logico
relativo sotto `workspace.root`, senza glob o directory. `--stream` è opt-in e
richiede `interaction.chat.streaming: true`.

Direct Chat esegue una sola completion con zero tool. Non costruisce Context
Engine, retrieval, Agent Runtime, sessione o approver. Senza file non seleziona
automaticamente contesto e deve dichiarare quando una risposta specifica del
progetto non è determinabile.

Complete e stream usano temperatura zero, `num_ctx: 4096` e thinking
disabilitato. Il file è un messaggio user non attendibile separato; la domanda
e il contratto di risposta costituiscono l’ultimo turno. Il contenuto del file
non può concedere authority o cambiare modalità.

Lo streaming conserva output atomico: stdout viene scritto soltanto dopo una
response valida, un terminale `stop` e EOF. Chunk parziali vengono scartati su
errore, cancellazione, deadline o limite.

### Output

```text
mode\tchat
terminal\tcompleted
model\tqwen3.5:9b
duration_ms\t...
input_tokens\t...
output_tokens\t...
num_ctx_requested\t4096
num_ctx_effective\t4096|unknown
thinking_requested\tfalse
thinking_effective\tfalse|unknown
truncated\tfalse
finish_reason\tstop
result
<contenuto finale>
```

Il risultato è intenzionalmente visibile all’utente locale. Log e failure non
includono domanda, prompt, response completa, contenuto del file, root fisica o
secret.

## `doctor --mode chat`

```text
maestro doctor --mode chat [--config path]
```

Esegue cinque check: config, workspace, composition, model e generation. Il
probe non invoca completion, non indicizza e non modifica il workspace, non
avvia provider e non installa modelli. `skip`, `unknown` o `fail` non valgono
come PASS in una serie di qualificazione.

## `version`

```text
maestro version
```

Stampa versione e commit incorporati. Un build locale senza metadata restituisce
`devel` e `unknown`; l’artifact incorpora identità esatta.

## Exit code Direct Chat

| Codice | Significato |
|---:|---|
| 0 | completion conclusa con risposta valida |
| 1 | response invalida, hard limit o failure interna |
| 2 | uso, configurazione o file non valido/non autorizzato |
| 4 | provider, modello, capability o deadline non disponibile |
| 130 | cancellazione tramite interrupt |

Un failure chat espone soltanto:

```text
chat failed: <reason_code>
```

I reason code pubblici sono `invalid_request`, `chat_profile_required`,
`file_not_allowed`, `provider_unavailable`, `capability_unsupported`,
`response_invalid`, `limit_exceeded`, `deadline_exceeded`, `canceled` ed
`execution_failed`.

# Maestro v0.2.0 CLI

Stato: Contratto pubblico sperimentale v0.2.0

Data: 2026-08-15

---

# Superficie minima

```text
maestro doctor
maestro models
maestro agents
maestro run
maestro version
```

`maestro bench` resta compatibile. `maestro`, `maestro help` e
`maestro --help` mostrano la root usage senza caricare configurazione.

## Candidato Milestone 14

La superficie development-only aggiunge due modalità esplicite:

```text
maestro chat
maestro agent
```

`maestro agent` è il nome canonico futuro del percorso descritto oggi da
`maestro run`. Durante la serie v0.3.x `run` resta un alias esatto e deprecato:
usa lo stesso parser, application graph, output ed exit code e non effettua
fallback a chat. La deprecazione è esposta nella help e nella documentazione,
non tramite righe aggiuntive nell'output stabile della run.

`maestro chat` appartiene alla Milestone 14 e non alla compatibility promise
v0.2.0. La sua promozione dipende dalla Milestone 15.

# Comandi

## `doctor`

```text
maestro doctor --config config.yaml
```

Valida e controlla separatamente:

- schema e target configurati;
- disponibilità della root workspace;
- composition di provider, policy e plugin;
- agent e tool registrati;
- policy registrata;
- capability dell'istanza provider;
- completion, tool calling e streaming richiesto sul modello chat;
- detection del workspace Laravel.

I risultati usano `pass`, `warn`, `fail` o `skip`. I probe provider sono
read-only. Doctor non invoca completion/embedding, non carica o rimuove modelli
e non indicizza o modifica il workspace.

## `models`

```text
maestro models --config config.yaml
```

Esegue model listing sul solo provider configurato e stampa i model ID in
ordine lessicografico. Non seleziona un modello e non esegue discovery
avanzata, lifecycle, acquisition o removal.

## `agents`

```text
maestro agents --config config.yaml
```

Elenca ID, versione e capability degli agenti registrati. Il comando verifica
che l'agent configurato esista e non effettua I/O provider.

## `run`

L'istruzione può essere posizionale:

```text
maestro run --config config.yaml "Explain OrderController"
```

oppure letta da stdin, entro 1 MiB:

```text
maestro run --config config.yaml
```

Il comando carica e avvia il plugin Laravel, ottiene il Workspace autorevole,
indicizza il contesto, costruisce una query lessicale, crea una `RunRequest` con
target e hard limit configurati e invoca `Agent Runtime.Run`.

Nel percorso supportato v0.2.0 il `RunRequest` espone soltanto i tool read-only
`workspace.list`, `workspace.read` e `workspace.search`; la configurazione
inclusa nega ogni effetto `workspace.mutate`. Tool e approval mutativi restano
sperimentali e fuori dal quick start.

Prima del run, stderr mostra i limiti configurati. Durante l'esecuzione mostra
eventi sintetici di sessione, piano, step, permission, contatori e terminale.
Questi eventi derivano esclusivamente dalle allowlist pubbliche e non includono
istruzione, prompt, contenuti, arguments o output tool.

Nel profilo Controlled Mutation stderr espone inoltre la sequenza redatta
`proposal -> approval -> apply -> reindex`. Lo stato applicativo distingue
effetto invariato o applicato, durability e generazione fresh; non stampa path,
diff o contenuto al di fuori del prompt locale di approval. Apply o refresh
falliti terminano la run senza testo finale del modello.

stdout è riservato al risultato stabile:

```text
run\trun-...
terminal\tcompleted
model_turns\t1
tool_calls\t0
input_tokens\t...
output_tokens\t...
result
<contenuto finale>
```

Su terminale interattivo una policy `prompt` attiva l'Approver descritto in
`operational-experience.md`. Per `workspace.patch` il prompt include la diff
concreta e offre soltanto deny o allow once. Con stdin non interattivo non viene
chiesta approvazione: la decisione è deny. `allow` può essere configurato
soltanto per classi non mutative; non esiste automazione mutativa né `--yes`.

## `agent` candidato

```text
maestro agent --config config.yaml "Explain OrderController"
```

Il comando ha la stessa sintassi e semantica di `maestro run`. È l'unica
modalità che può costruire Context Engine, retrieval, Agent Runtime, sessione,
tool registry e approver. Un failure non avvia `maestro chat`.

## `chat` candidato

```text
maestro chat \
  --config config-v2.yaml \
  --file routes/api.php \
  "Quali endpoint, controller e action sono dichiarati?"
```

La domanda può essere posizionale oppure letta da stdin quando il positional è
assente, entro 1 MiB. `--file` è opzionale ma singolo e accetta soltanto un
path logico relativo sotto il workspace. Non accetta glob, directory o path
assoluti. `--stream` è opt-in e richiede anche
`interaction.chat.streaming: true`; un profilo disabilitato o una capability
provider assente falliscono prima della generation.

Chat esegue una sola completion con zero tool e non costruisce retrieval,
index, state machine, sessione o approver. Senza `--file` comunica al modello
che non è disponibile contesto workspace; non seleziona file e non effettua
fallback agentico.

Lo streaming usa il trasporto incrementale del provider ma conserva output
atomico: i chunk vengono assemblati entro il limite e stdout viene scritto
soltanto dopo un terminale `stop` valido seguito da EOF. Un flusso troncato non
lascia output parziale.

stdout usa l'envelope candidato:

```text
mode\tchat
terminal\tcompleted
model\tqwen2.5-coder:7b
duration_ms\t...
input_tokens\t...
output_tokens\t...
num_ctx_requested\t4096
num_ctx_effective\t4096|unknown
thinking_requested\tfalse
thinking_effective\tfalse|unknown
truncated\tfalse|unknown
finish_reason\tstop
result
<contenuto finale>
```

Il contenuto finale è intenzionalmente visibile all'utente locale. Log,
progress ed evidenze non includono domanda, prompt, response completa,
contenuto del file, root fisica o secret.

## `bench mutation`

La qualificazione mutativa usa un profilo separato e inizialmente ne valida il
contratto e la fixture:

```text
maestro bench mutation \
  --profile docs/mutation-qualification-profile.yaml
```

Questo comando non esegue prove live né abilita da solo Controlled Mutation.
Le modalità `deterministic`, `preflight`, `gate-a`, `gate-b` e `gate-c` restano
strumenti di qualificazione e non ampliano la superficie supportata. Gate C
richiede una TTY reale e non dispone di auto-approval. Vedere
`mutation-benchmark.md` e ADR-0032.

## `version`

```text
maestro version
```

Stampa versione e commit e aggiunge `dirty true` quando il build info Go lo
dichiara. Un build locale senza metadata restituisce `devel` e `unknown`;
l'artifact v0.2.0 incorpora versione e commit esatti durante il packaging.

# Exit code

| Codice | Significato |
|---:|---|
| 0 | Operazione completata |
| 1 | Failure operativa, doctor fallito o run non completed |
| 2 | Uso CLI o configurazione non valida |
| 3 | Permission negata o approval non disponibile |
| 4 | Provider, modello o capability non disponibile |
| 130 | Cancellazione tramite interrupt |

Per il candidato chat: input, configurazione o file non ammesso usano 2;
provider, modello, capability o deadline provider usano 4; response invalida,
hard limit e failure interna usano 1; interrupt usa 130. stderr espone soltanto
`chat failed: <reason_code>` con reason code definiti da ADR-0033.

# Stream e redazione

Output finale ordinario va su stdout; progress, approval, usage error e failure
vanno su stderr. I failure di run usano codici sintetici e non serializzano la
catena d'errore esterna. Doctor non stampa secret, prompt, contenuti, arguments
o root. `run` stampa intenzionalmente il solo contenuto finale del modello
all'utente locale; eventi e report continuano a usare le allowlist redatte
della Milestone 7.

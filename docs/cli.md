# Maestro v0.1 CLI

Stato: Contratto pubblico sperimentale v0.1.0

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

Nel percorso supportato v0.1.0 il `RunRequest` espone soltanto i tool read-only
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

## `version`

```text
maestro version
```

Stampa versione e commit e aggiunge `dirty true` quando il build info Go lo
dichiara. Un build locale senza metadata restituisce `devel` e `unknown`;
l'artifact v0.1.0 incorpora versione e commit esatti durante il packaging.

# Exit code

| Codice | Significato |
|---:|---|
| 0 | Operazione completata |
| 1 | Failure operativa, doctor fallito o run non completed |
| 2 | Uso CLI o configurazione non valida |
| 3 | Permission negata o approval non disponibile |
| 4 | Provider, modello o capability non disponibile |
| 130 | Cancellazione tramite interrupt |

# Stream e redazione

Output finale ordinario va su stdout; progress, approval, usage error e failure
vanno su stderr. I failure di run usano codici sintetici e non serializzano la
catena d'errore esterna. Doctor non stampa secret, prompt, contenuti, arguments
o root. `run` stampa intenzionalmente il solo contenuto finale del modello
all'utente locale; eventi e report continuano a usare le allowlist redatte
della Milestone 7.

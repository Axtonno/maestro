# Maestro v0.1 CLI

Stato: Implementato — Milestone 8, Fase 2

Data: 2026-08-13

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

Il risultato iniziale stampa run ID, terminale e contenuto finale. Piano e UX
interattiva avanzata appartengono alla Fase 3. Una policy `prompt` viene negata
in sicurezza finché non è disponibile l'Approver terminale.

## `version`

```text
maestro version
```

Stampa versione e commit e aggiunge `dirty true` quando il build info Go lo
dichiara. Un build locale senza metadata restituisce `devel` e `unknown`; gli
artifact di release dovranno incorporare `v0.1.0` e il commit durante la Fase 4.

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

Output ordinario va su stdout; usage error e failure vanno su stderr. Doctor
usa reason code sintetici e non stampa secret, prompt, contenuti, arguments o
root. `run` stampa intenzionalmente il contenuto finale del modello all'utente
locale; eventi e report continuano a usare le allowlist redatte della Milestone
7.

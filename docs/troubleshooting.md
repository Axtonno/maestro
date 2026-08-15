# Maestro v0.1.x Troubleshooting

## Il checksum fallisce

Non estrarre né eseguire l'archive. Scaricare nuovamente sia `.tar.gz` sia
`.sha256` dalla stessa release e ripetere `sha256sum -c`.

## `maestro version` non mostra la versione v0.1.x attesa

Si sta usando un altro binario o un build locale. Eseguire `./maestro version`
dalla directory estratta e confrontare il commit con
`ARTIFACT-MANIFEST.txt`.

## `doctor` segnala `instance_probe_failed`

Verificare che Ollama sia avviato e che `provider.base_url` punti all'istanza
corretta:

```sh
ollama list
./maestro doctor --config ./configs/maestro.example.yaml
```

Il percorso supportato usa loopback `127.0.0.1:11434`.

## Il modello non compare

Maestro non scarica modelli automaticamente:

```sh
ollama pull llama3.1:8b
./maestro models --config ./configs/maestro.example.yaml
```

## Il workspace Laravel non viene rilevato

`workspace.root` deve indicare la directory con `artisan` e un
`composer.json` valido che dichiara `laravel/framework` in `require`. I path
relativi sono risolti rispetto al file di configurazione.

## `configuration invalid`

Lo schema è strict: campi sconosciuti o duplicati, alias YAML, documenti
multipli e versioni diverse da `1` sono rifiutati. Confrontare il file con
`configs/maestro.example.yaml` e consultare `configuration.md`.

## `provider_unavailable` o exit code 4

Provider, modello o capability richiesta non è disponibile. Eseguire `doctor`
e `models`; non cambiare modello casualmente se si vuole restare nel percorso
qualificato v0.1.x.

## Un progetto reale termina prima della prima tool call

Dalla v0.1.1 il plugin Laravel esclude asset generati e dati runtime dalla scan
policy. Se la v0.1.0 termina `execution_failed` su un progetto che supera i
limiti generici, aggiornare alla v0.1.1 e ripetere `doctor` e la stessa run
read-only. Non aumentare i limiti rimuovendo i bound del Context Engine.

## `tool_failure` o `execution_failed`

Una tool call generata dal modello può essere invalida o fallire. Il runtime non
esegue retry impliciti degli effetti e mantiene l'output redatto. Ripetere una
sola volta il task identico dopo aver verificato `doctor`; se il problema è
riproducibile, conservare versione, commit, terminale e contatori senza
allegare workspace sensibili.

## `limit_exceeded`

È stato raggiunto un hard limit configurato. Lo stderr indica il terminale e i
contatori. Non aumentare i bound senza considerare memoria, latenza e quantità
di contenuto disclosed al modello.

## Run lenta su CPU

`llama3.1:8b` può richiedere minuti su CPU-only. Verificare memoria e swap,
evitare più modelli residenti e controllare:

```sh
ollama ps
```

Per liberare la fixture dopo la prova:

```sh
ollama stop llama3.1:8b
```

## Cancellazione

SIGINT/SIGTERM producono exit code 130. Se il processo non termina entro il
budget documentato di 30 secondi, registrare versione, commit e ultime righe
redatte di stderr come bug.

## Exit code

| Codice | Significato |
|---:|---|
| 0 | completato |
| 1 | failure operativa o limite |
| 2 | uso/configurazione non valida |
| 3 | permission negata/non disponibile |
| 4 | provider/modello/capability non disponibile |
| 130 | cancellato da interrupt |

Consultare anche `known-issues.md` e `security-model.md`.

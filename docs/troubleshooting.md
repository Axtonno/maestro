# Maestro v0.3.0 Troubleshooting

## Il checksum fallisce

Non estrarre né eseguire l’archive. Recuperare nuovamente `.tar.gz` e
`.sha256` dalla stessa GitHub Release e ripetere `sha256sum -c`.

## Versione, commit o stato non coincidono

Eseguire `./maestro version` dalla directory estratta e confrontare nome
archive e `ARTIFACT-MANIFEST.txt`. Non rinominare un packaging candidate come
release candidate o release.

Nel candidate post-v0.3.0 usare:

```sh
./maestro version --diagnostic
```

Confrontare `version`, `status`, `commit`, path `executable` risolto e
`sha256`. Questa forma permette di individuare un altro `maestro` precedente
nel `PATH`; non è disponibile nell'asset pubblico v0.3.0.

## `doctor` non completa cinque PASS

Usare il modo chat esplicito:

```sh
./maestro doctor --mode chat --config ./configs/maestro.chat.example.yaml
```

- `config`: verificare lo schema strict dichiarato e campi
  duplicati/sconosciuti; v2 è il profilo v0.3.0, v3 il candidate M21;
- `workspace`: verificare che `workspace.root` sia una directory reale e non
  un symlink;
- `composition`: verificare provider e nome della variabile secret;
- `model`: verificare Ollama, identità e disponibilità del modello;
- `generation`: verificare streaming, context e thinking richiesti.

Doctor non invoca completion, non avvia Ollama e non installa modelli.

## Il modello non coincide

Il percorso qualificato richiede `qwen3.5:9b` con digest
`6488c96fa5faab64bb65cbd30d4289e20e6130ef535a93ef9a49f42eda893ea7`.
Confrontare il catalogo Ollama con `ARTIFACT-MANIFEST.txt`. Non sostituire o
aggiornare il modello durante una serie di qualificazione.

## `configuration invalid` o `chat_profile_required`

Direct Chat richiede un profilo strict `version: 2` con provider, workspace,
`interaction.chat` e `policy.workspace_mutate: deny`. Un profilo agentico v1
non viene convertito e il doctor senza `--mode chat` segue il percorso agentico.

Il candidate CPU M21 usa invece `version: 3` e richiede anche
`interaction.chat.num_predict` e `interaction.chat.residency`. Devono essere
rispettivamente positivi e una durata positiva non superiore a 10 minuti; il
provider deve essere Ollama. Non aggiungere questi campi a un file v2: lo
schema strict li rifiuta intenzionalmente.

Nel candidate post-v0.3.0 la riga `configuration` distingue `read_failed`,
`yaml_invalid`, `unknown_field`, `missing_field` e `invalid_value`, indicando
solo il path logico del campo. Correggere quella categoria senza pubblicare il
contenuto del file di configurazione.

## `file_not_allowed`

`--file` accetta un solo path logico relativo. Sono rifiutati path assoluti,
traversal, backslash, directory, symlink, file non regolari, caratteri di
controllo, UTF-8 invalido e file oltre `max_file_bytes`. Risolvere il problema
nel path o nella root; non allargare il workspace per aggirare il controllo.

## `provider_unavailable` o `capability_unsupported`

Verificare endpoint Ollama, modello e controlli generativi con doctor. Maestro
non applica fallback ad agent, altro modello o altro provider.

## `response_invalid` o `limit_exceeded`

Response vuota, ruolo/finish non validi, tool call inattesa, stream malformato,
UTF-8 invalido o output oltre limite vengono scartati interamente. stdout resta
vuoto; conservare soltanto versione, commit, reason code e durata redatti.

Nel candidate M21, anche una risposta terminata perché ha esaurito
`num_predict` resta invalida/incompleta: non aumentare il budget dopo aver
visto i risultati di una serie congelata.

## Risposta qualitativamente dubbia

Il modello può proporre inferenze non sostenute. Ripetere la domanda chiedendo
di distinguere fatti, assenze e proposte, poi verificare direttamente il file.
Non trattare la risposta come autorizzazione a modificare il workspace.

## Cancellazione e deadline

SIGINT/SIGTERM producono exit 130 e `canceled`. Una deadline provider produce
exit 4 e `deadline_exceeded`. Nessun output parziale deve apparire su stdout.

Nel candidate post-v0.3.0, durante generation oltre 15 secondi, stderr può
mostrare heartbeat `progress state=generating` con il solo tempo trascorso.
Non è una risposta parziale. L'assenza di nuovi heartbeat dopo un terminale è
parte del contratto.

## Exit code Direct Chat

| Codice | Significato |
|---:|---|
| 0 | risposta valida completata |
| 1 | response invalida, hard limit o failure interna |
| 2 | uso, configurazione o file non ammesso |
| 4 | provider, modello, capability o deadline non disponibile |
| 130 | cancellazione tramite interrupt |

Consultare anche `known-issues.md`, `security-model.md` e `compatibility.md`.

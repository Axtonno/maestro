# Maestro v0.5.0 Troubleshooting

Prima della diagnosi verificare che piattaforma, provider e modelli rientrino
nella [pagina delle capacità correnti](current-capabilities.md).

## Il checksum fallisce

Non estrarre né eseguire l'archive. Scaricare di nuovo `.tar.gz` e
`.sha256` dalla stessa GitHub Release e ripetere `sha256sum -c`.

## Versione, commit o stato non coincidono

```sh
./maestro version --diagnostic
```

Confrontare versione, stato `release`, commit e SHA-256 con
`ARTIFACT-MANIFEST.txt`. Il path `executable` aiuta a individuare un altro
binario precedente nel `PATH`. Un candidate rinominato non diventa release.

## `doctor --mode all` non completa 14 PASS

```sh
./maestro doctor --mode all --config ./configs/maestro.v0.5.0-candidate.yaml
```

Controllare:

- schema v4 e root workspace;
- Ollama raggiungibile su loopback;
- presenza e digest di entrambi i modelli;
- generation controls richiesti;
- prompt e schema mutativi con gli SHA-256 qualificati;
- terminale interattivo per i check mutation.

Doctor non esegue completion, non avvia Ollama e non installa modelli.

## Il modello o il digest non coincidono

Il manifest richiede:

- `qwen3.5:9b`:
  `6488c96fa5faab64bb65cbd30d4289e20e6130ef535a93ef9a49f42eda893ea7`;
- `qwen2.5-coder:14b`:
  `9ec8897f747e246e970bc5cfdda85d22f1123dc2e3d34978a010a75968716849`.

Non sostituire tag, digest o modello durante una serie. Maestro non effettua
pull né fallback.

## `configuration invalid` o `chat_profile_required`

La release distribuisce uno schema strict v4 con `provider`, `workspace`,
`direct_chat` e `controlled_mutation`. Correggere la categoria redatta
`read_failed`, `yaml_invalid`, `unknown_field`, `missing_field` o
`invalid_value`.

I profili v2/v3 possono servire Direct Chat per compatibilità, ma non abilitano
Controlled Mutation. Un profilo agentico v1 non viene convertito.

## `file_not_allowed`

Direct Chat richiede un path logico relativo contained. Controlled Mutation
richiede inoltre un file PHP sotto `app/`. Sono rifiutati path assoluti,
traversal, backslash, directory, symlink, file non regolari, caratteri di
controllo, UTF-8 invalido e file oltre limite.

Correggere path o root; non allargare il workspace per aggirare il controllo.

## Controlled Mutation non raggiunge la preview

Verificare:

- TTY reale;
- `--file` e `--lines start:end`;
- profilo v4 e `controlled_mutation.enabled: true`;
- modello, digest, prompt e schema;
- istruzione sufficientemente precisa;
- disponibilità del provider entro il timeout.

Output non conforme o astensione falliscono chiusi. Non viene applicato un
secondo tentativo automatico.

## `approval_rejected`

È il terminale atteso dopo `d`, invio vuoto, EOF o input sconosciuto. Usa exit
code 3 e non scrive. Ripetere il comando solo se si desidera una nuova preview;
l'approvazione precedente non viene conservata.

## `stale_source`

Il file è cambiato tra lettura e approval. Maestro conserva il contenuto
corrente e non applica la preview obsoleta. Verificare il diff, poi avviare una
nuova mutation sulla sorgente aggiornata.

## La preview non è corretta

Digitare `d`. Non approvare per “vedere cosa succede”: il controllo host-bound
limita l'effetto, ma non rende semanticamente giusta una proposta.

## Risposta chat dubbia

Il modello può sbagliare anche con terminale valido. Verificare il file e
riformulare la domanda chiedendo di separare fatti osservati, inferenze e dati
non determinabili. M38 documenta un errore su una domanda generale senza file.

## Latenza elevata o memoria insufficiente

I modelli vengono caricati uno per volta e il cambio chat/mutation può includere
unload e load. Controllare `/api/ps`, memoria e swap del sistema. Sul T490s
CPU-only M38 ha osservato latenze chat fino a circa 172 secondi; non è uno SLA.

Un OOM, restart del provider, modello errato o fallback rende la prova non
valida. Ridurre il carico, non cambiare silenziosamente modello.

## Cancellazione e deadline

SIGINT/SIGTERM producono exit code 130 e `canceled`. Una deadline provider
produce exit code 4 e `deadline_exceeded`. Nessun output parziale deve essere
pubblicato su stdout.

Durante generation, stderr può mostrare heartbeat con il solo tempo trascorso.
Non sono risposta parziale.

## Exit code

| Codice | Significato |
| ---: | --- |
| 0 | risposta valida o mutation applicata |
| 1 | response invalida, hard limit o failure interna |
| 2 | uso, configurazione o target non ammesso |
| 3 | deny, stale, astensione o chiusura mutation senza apply |
| 4 | provider, modello, capability o deadline non disponibile |
| 130 | cancellazione tramite interrupt |

Consultare [Known Issues](known-issues.md), [Security Model](security-model.md)
e [Compatibility Matrix](compatibility.md).

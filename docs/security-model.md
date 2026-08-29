# Maestro v0.3.0 Security Model

Data: 2026-08-29

## Sintesi

Maestro è un’applicazione locale trusted in-process, non una sandbox. Il
percorso v0.3.0 qualificabile è Direct Chat read-only: zero o un file scelto
esplicitamente, una completion/stream, zero tool e nessun fallback.

## Confini di fiducia

| Elemento | Trattamento |
|---|---|
| Utente e configurazione | Attendibili; scelgono provider, workspace e file |
| Workspace | Non attendibile; può contenere prompt injection o dati sensibili |
| Modello | Non autorevole; genera testo ma non riceve authority |
| Provider | Riceve domanda e file selezionato; deve essere considerato attendibile |
| Processo Maestro | Trusted, con gli stessi privilegi dell’utente |

La qualifica copre Ollama locale su loopback. Cambiare `provider.base_url` può
inviare dati a un servizio diverso e non qualificato.

## Garanzie implementate

- configurazione YAML strict v2 senza fallback impliciti;
- profilo chat, modello, timeout, streaming e generation controls espliciti;
- `workspace_mutate: deny` obbligatorio;
- zero tool e `tool_choice: none` in ogni request;
- nessuna costruzione di retrieval, index, Agent Runtime, sessione o approver;
- un solo file logico relativo alla root configurata;
- path normalizzato, containment con `os.Root`, rifiuto di symlink e race
  fail-closed durante la lettura;
- regular-file check, UTF-8/NUL validation e limiti byte inclusivi;
- domanda, path logico e contenuto separati in messaggi provider; il file è
  dichiarato evidenza non attendibile e la domanda resta l’ultimo turno;
- temperatura zero uniforme fra complete e stream;
- risposta, ruolo, finish, tool call inattese, usage, modello e limiti validati;
- stream aggregato e pubblicato soltanto dopo terminale `stop` seguito da EOF;
- errori sintetici senza prompt, response, contenuto, root fisica o secret;
- artifact riproducibile, manifest, SHA-256, licenza e attribution.

## Non garanzie

v0.3.0 non fornisce:

- sandbox, container, seccomp, namespace o separazione di privilegi;
- isolamento di rete o prevenzione dell’esfiltrazione al provider configurato;
- secret manager, cifratura del profilo o attestazione per-run del modello;
- difesa completa contro prompt injection o affermazioni semanticamente errate;
- retrieval, tool, agent, mutation, rollback o recovery;
- validazione di plugin o servizi di terze parti.

## Disclosure single-file

Senza `--file`, Maestro comunica che non è disponibile contesto workspace e
non seleziona contenuti. Con `--file`, legge soltanto il path esplicito entro la
root. Path fisico e altri file non vengono inviati al provider.

Il path logico è JSON-quoted; caratteri di controllo, formattatori invisibili e
separatori di linea sono rifiutati. Il contenuto non può cambiare tool set,
policy o destinazione perché il percorso non possiede tali componenti. Una
tool call restituita dal modello è un protocol failure, non un’azione.

## Streaming e terminali

Lo streaming è un trasporto opt-in, non una pubblicazione incrementale. Chunk
successivi al terminale, terminali mancanti/duplicati, tool delta, errore di
receive/close, cancellazione, deadline o superamento limite scartano l’intera
risposta. Un failure non avvia una seconda completion.

## Dati e output

La risposta finale viene mostrata intenzionalmente su stdout all’utente locale.
stderr e le evidenze operative contengono soltanto reason code e metadati
redatti. Questa distinzione non rende sicuro inviare file sensibili a un
provider non attendibile.

## Capacità presenti ma non supportate

Il repository contiene codice storico o sperimentale per agent, Context
Engine, tool e Controlled Mutation. L’archive v0.3.0 non include i relativi
profili e il percorso Direct Chat non li costruisce. La loro presenza nel
binario non amplia la compatibility promise né costituisce fallback.

## Raccomandazioni operative

- usare un account senza privilegi amministrativi;
- mantenere Ollama su loopback;
- verificare archive, checksum, manifest, modello e digest;
- controllare `workspace.root` e il path prima di inviare un file;
- non inserire credenziali nella configurazione o nella fixture;
- interrompere output inattesi e verificare sempre il sorgente originale.

## Segnalazione vulnerabilità

Seguire `SECURITY.md`. Non includere secret, contenuti di workspace o exploit
funzionanti in issue pubbliche.

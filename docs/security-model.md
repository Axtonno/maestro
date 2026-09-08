# Maestro v0.5.0 Security Model

Aggiornato: 2026-09-08

## Sintesi

Maestro è un'applicazione locale trusted in-process, non una sandbox. v0.5.0
offre Direct Chat read-only e un write-mode ristretto: Controlled Mutation
può sostituire un solo intervallo PHP scelto dall'utente, dopo preview e
allow-once in un terminale reale.

La sicurezza del write-mode deriva dal controllo deterministico dell'effetto,
non dall'autorità o dall'infallibilità del modello.

## Confini di fiducia

| Elemento | Trattamento |
| --- | --- |
| Utente e configurazione | Attendibili; scelgono provider, workspace, file, righe e approval |
| Workspace | Non attendibile; può contenere prompt injection o dati sensibili |
| Modello | Non autorevole; propone testo ma non concede authority |
| Provider | Riceve domanda e contenuto selezionato; deve essere attendibile |
| Processo Maestro | Trusted, con gli stessi privilegi dell'utente |

La qualifica copre Ollama locale su loopback. Un endpoint diverso può
trasmettere dati a un servizio non qualificato.

## Garanzie comuni

- configurazione YAML strict, senza fallback impliciti;
- identità esatta di modello e digest;
- workspace root esplicita e validata;
- path normalizzati e contained, con rifiuto di traversal e symlink;
- regular-file check, UTF-8/NUL validation e limiti byte;
- contenuto workspace trattato come evidenza non attendibile;
- risposta, ruolo, finish, tool call inattese, usage e limiti validati;
- errori sintetici senza prompt, response, contenuto, root fisica o secret;
- artifact, manifest e checksum verificabili.

## Direct Chat

Senza `--file`, Maestro non seleziona contenuti del workspace. Con `--file`,
legge soltanto il path logico indicato. Direct Chat esegue una completion senza
tool, retrieval, Agent Runtime, approver o mutation.

Lo streaming è aggregato: stdout viene pubblicato soltanto dopo response valida,
terminale `stop` ed EOF. Un failure scarta i chunk e non avvia una seconda
completion.

## Controlled Mutation

Il write-mode aggiunge questi controlli:

- comando opt-in distinto da chat;
- file PHP singolo e non symlink sotto `app/`;
- intervallo inclusivo fornito dall'utente;
- modello, digest, prompt e schema mutativi congelati;
- proposta strutturata limitata allo span;
- preview completa con digest e fingerprint;
- TTY reale e allow-once esplicito;
- deny-by-default per invio vuoto, EOF o input sconosciuto;
- ricontrollo della sorgente dopo la preview;
- `stale_source` senza write se il file è cambiato;
- sostituzione atomica e digest dopo l'apply;
- nessun fallback al modello chat o a un altro percorso.

L'approval autorizza la specifica preview, non il modello in generale. Una
proposta semanticamente sbagliata può essere ben formata e contained: deve
essere negata dall'utente.

## Non garanzie

v0.5.0 non fornisce:

- sandbox, container, seccomp, namespace o separazione di privilegi;
- isolamento di rete o prevenzione dell'esfiltrazione al provider;
- secret manager o cifratura del profilo;
- difesa perfetta contro prompt injection o affermazioni errate;
- selezione autonoma sicura del target;
- mutation multi-file, multi-range o fuori `app/`;
- esecuzione di shell, Git, Composer, test o Docker;
- rollback automatico dopo un apply completato;
- validazione di plugin o servizi di terze parti.

## Dati e output

Risposte e preview sono intenzionalmente visibili all'utente locale. stderr ed
evidenze operative devono contenere solo reason code e metadati redatti. Questa
distinzione non rende sicuro un provider non attendibile.

## Raccomandazioni operative

- usare un account senza privilegi amministrativi;
- mantenere Ollama su loopback;
- verificare archive, checksum, manifest, modelli e digest;
- lavorare in un repository Git pulito;
- scegliere il minimo intervallo necessario;
- leggere tutta la preview e negare al primo dubbio;
- verificare diff e test del progetto dopo ogni apply;
- non inserire credenziali nella configurazione o nelle fixture.

Il contratto sintetico del write-mode è in
[Controlled Mutation: perimetro supportato](controlled-mutation-support.md).

## Segnalazione vulnerabilità

Seguire `SECURITY.md`. Non includere secret, contenuti di workspace o exploit
funzionanti in issue pubbliche.

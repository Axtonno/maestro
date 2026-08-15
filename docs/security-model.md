# Maestro v0.1.x Security Model

Data: 2026-08-15

## Sintesi

Maestro v0.1.x è un'applicazione locale trusted in-process, non una sandbox. Il
percorso supportato è read-only e usa soltanto Ollama locale,
`llama3.1:8b` e i tool list/read/search su un workspace Laravel scelto
esplicitamente dall'utente.

## Confini di fiducia

| Elemento | Trattamento |
|---|---|
| Utente e configurazione | Attendibili; scelgono provider, workspace, agent, policy e limiti |
| Workspace | Dati non attendibili; possono contenere prompt injection o contenuti sensibili |
| Modello | Non autorevole; propone output e tool call, mai permessi |
| Provider configurato | Riceve istruzione e contesto esplicitamente disclosed |
| Tool/plugin built-in | Codice trusted eseguito nello stesso processo e con i privilegi dell'utente |
| Estensioni terze | Fuori dal supporto v0.1.x |

Cambiare `provider.base_url` può inviare istruzioni e sezioni selezionate del
workspace a quel servizio. La promessa v0.1.x copre soltanto Ollama locale su
loopback; l'utente deve considerare attendibile ogni endpoint alternativo.

## Garanzie implementate

- configurazione YAML strict, senza fallback impliciti;
- provider, modello, workspace, agent, policy, tool e limiti espliciti;
- configurazione distribuita con soli tool read-only e
  `workspace_mutate: deny`;
- path logici confinati alla root con `os.Root`, rifiuto dei symlink e limiti di
  dimensione;
- authorization su action concrete; una risposta del modello non concede
  authority;
- hard limit su durata, turni, tool call, piano, token e byte;
- SIGINT/SIGTERM cooperativi e shutdown applicativo bounded a 30 secondi;
- eventi e diagnostiche con allowlist redatte;
- secret referenziati soltanto per nome di variabile d'ambiente;
- artifact versionato, manifest, SHA-256, Apache-2.0 e attribution incluse.

## Non garanzie

La v0.1.x non fornisce:

- sandbox, container, seccomp, namespace o separazione di processo;
- riduzione automatica dei privilegi del sistema operativo;
- isolamento di rete o prevenzione dell'esfiltrazione verso il provider scelto;
- secret manager, cifratura della configurazione o attestazione del modello;
- difesa completa dalla prompt injection;
- rollback generale, transazioni filesystem o recovery dopo crash;
- validazione di sicurezza di plugin/tool di terze parti;
- supporto operativo ai tool mutanti presenti nel codice sperimentale.

Il profilo ufficiale read-only riduce l'autorità disponibile al modello, ma il
processo Maestro conserva i normali permessi dell'utente. Eseguire Maestro su
workspace o endpoint non attendibili richiede quindi la stessa prudenza di
qualsiasi altro processo locale.

## Dati e output

Il Context Engine indicizza localmente il workspace e seleziona sezioni entro i
budget configurati. I contenuti selezionati e l'istruzione vengono inviati al
provider esplicito. Il risultato finale del modello viene stampato
intenzionalmente su stdout; progress, contatori e failure sintetici vanno su
stderr. Prompt, contenuti, argomenti tool, output tool, root fisica, fingerprint
e secret non fanno parte degli eventi operativi.

I report di release usano canary e scansioni per verificare l'assenza di
leakage nei percorsi falliti. Questo non rende l'output finale del modello un
canale sicuro per dati sensibili: il risultato è contenuto applicativo richiesto
dall'utente.

## Raccomandazioni operative

- usare un account senza privilegi amministrativi;
- mantenere Ollama su loopback e non esporlo senza autenticazione/rete fidata;
- verificare archive e checksum prima dell'estrazione;
- esaminare la configurazione e la root prima di ogni run;
- non aggiungere `workspace.write` o `workspace.patch` al profilo v0.1.x;
- interrompere run inattese con SIGINT e controllare il terminale redatto;
- non inserire credenziali nel file YAML o nel workspace della fixture.

## Segnalazione vulnerabilità

Seguire `SECURITY.md`. Non includere secret, contenuti di workspace o exploit
funzionanti in issue pubbliche.

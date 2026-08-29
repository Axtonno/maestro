# Maestro v0.2.0 Security Model

Data: 2026-08-21

## Sintesi

Maestro v0.2.0 è un'applicazione locale trusted in-process, non una sandbox. Il
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
| Estensioni terze | Fuori dal supporto v0.2.0 |

Cambiare `provider.base_url` può inviare istruzioni e sezioni selezionate del
workspace a quel servizio. La promessa v0.2.0 copre soltanto Ollama locale su
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
- le istruzioni strette `Read <logical-path> ...` del reference agent
  eseguono una `workspace.read` verificata tramite Tool Runtime prima della
  prima inferenza, senza interpretare testo del modello come invocazione;
- hard limit su durata, turni, tool call, piano, token e byte;
- SIGINT/SIGTERM cooperativi e shutdown applicativo bounded a 30 secondi;
- eventi e diagnostiche con allowlist redatte;
- secret referenziati soltanto per nome di variabile d'ambiente;
- artifact versionato, manifest, SHA-256, Apache-2.0 e attribution incluse.

## Non garanzie

La v0.2.0 non fornisce:

- sandbox, container, seccomp, namespace o separazione di processo;
- riduzione automatica dei privilegi del sistema operativo;
- isolamento di rete o prevenzione dell'esfiltrazione verso il provider scelto;
- secret manager, cifratura della configurazione o attestazione del modello;
- difesa completa dalla prompt injection;
- rollback generale, transazioni filesystem o recovery dopo crash;
- validazione di sicurezza di plugin/tool di terze parti;
- supporto operativo ai tool mutanti presenti nel codice sperimentale.

## Profilo Controlled Mutation candidato

La Milestone 10 ha consegnato un profilo separato e opt-in, ma la v0.2.0 non lo
promuove al supporto. Il profilo accetta soltanto `workspace.patch` su
un file PHP esistente sotto `app/`, dopo read verificata, preview concreta,
TTY e approval one-shot. Il fingerprint lega la proposta all'esecuzione; il
commit Linux usa temporaneo, sync, recheck e rename atomico. Ogni run ammette
un solo tentativo e può completare soltanto dopo reindex e bundle fresh.

La Milestone 11 ha concluso i gate con `mutation_deferred`: Gate A è fallito al
primo tentativo e la stop rule ha impedito Gate B e Gate C. Il profilo resta
quindi non supportato; la matrice deterministica e il preflight superati non
concedono autorità operativa né ampliano il threat boundary.
Non offre sandbox, rollback generale, recovery da crash, modifiche multi-file,
creazione file, shell o Git. Un failure dopo il rename può lasciare la patch
applicata con durability o refresh incompleti; gli stati redatti lo dichiarano
e non tentano rollback implicito.

Il profilo ufficiale read-only riduce l'autorità disponibile al modello, ma il
processo Maestro conserva i normali permessi dell'utente. Eseguire Maestro su
workspace o endpoint non attendibili richiede quindi la stessa prudenza di
qualsiasi altro processo locale.

## Candidato Direct Chat della Milestone 14

La Milestone 14 introduce una superficie development-only separata, non
supportata dalla v0.2.0. `maestro chat` riceve al massimo un file esplicitamente
selezionato e costruisce soltanto provider completion/streaming. Non riceve
Tool Runtime, Context Engine, index, Agent Runtime, sessione, policy mutativa o
approver e non può usarli come fallback.

Il loader applica confinement rispetto alla root configurata, rifiuta path
assoluti, traversal, backslash, symlink in qualsiasi componente, file non
regolari, dimensioni oltre limite, UTF-8 invalido, NUL e cambi durante la
lettura. I caratteri di controllo, formattazione invisibile e separazione di
linea sono vietati nel path logico, che entra nel prompt soltanto in forma
quotata. File vuoti e UTF-8 con BOM sono ammessi e preservati; il limite byte è
inclusivo. Il path fisico non viene disclosed. Domanda, path logico e contenuto
sono separati da confini di messaggio provider, non da sentinelle testuali
collidibili; il contenuto workspace resta non attendibile e non può concedere
tool o autorità. Con un file, un messaggio system chiude esplicitamente il
confine dell'evidenza e la domanda è l'ultimo turno user: istruzioni apparenti
nel sorgente restano dati e non sostituiscono la richiesta dell'operatore.

La request dichiara zero tool e `tool_choice: none`. Una tool call inattesa
nella response è un protocol failure. Timeout, risposta vuota, output oltre
limite o capability non supportata falliscono chiusi e non avviano
`maestro agent`. `num_ctx` e `thinking` espliciti devono essere onorati o
rifiutati; un valore non attestabile resta `unknown` e non viene presentato
come confermato.

Complete e stream ricevono le stesse opzioni e una temperatura interna fissata
a zero. Questo riduce la divergenza da sampling, ma non trasforma equivalenza o
correttezza semantica in garanzie deterministiche: entrambe restano gate live.

Lo streaming chat è aggregato e validato prima di scrivere stdout. Terminale
mancante o duplicato, tool delta, chunk successivo al terminale, errore di
receive/close e superamento del limite scartano l'intera risposta; nessun chunk
parziale viene trasformato in risultato applicativo.

Il risultato finale resta intenzionalmente visibile su stdout. Metadati, log e
report escludono domanda, prompt, response completa, contenuto del file, root
fisica e secret. Questa riduzione di autorità non rende sicuro inviare file
sensibili a un endpoint provider non attendibile.

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
- non aggiungere `workspace.write` o `workspace.patch` al profilo v0.2.0;
- interrompere run inattese con SIGINT e controllare il terminale redatto;
- non inserire credenziali nel file YAML o nel workspace della fixture.

## Segnalazione vulnerabilità

Seguire `SECURITY.md`. Non includere secret, contenuti di workspace o exploit
funzionanti in issue pubbliche.

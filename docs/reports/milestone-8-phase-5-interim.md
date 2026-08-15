# Milestone 8 — Fase 5 Interim e incidente OOM

Data: 2026-08-13

Aggiornato: 2026-08-15 — gate live `pc.3` e selezione fail-fast dei modelli

Stato: In corso — ADR-0029 approvato, preparazione candidate read-only

Verdetto: **NO-GO alla release candidate**

---

# Scopo

Questo report registra le evidenze raccolte durante il primo tentativo della
Fase 5, separa i risultati validi dalle prove invalidate e documenta i due
eventi out-of-memory che hanno terminato `llama-server` e destabilizzato Visual
Studio Code Insiders.

Il packaging candidate `v0.1.0-pc.1` resta immutato. Nessun artifact è stato
rinominato o promosso a release candidate.

# Candidate verificato

| Campo | Valore |
|---|---|
| Artifact | `maestro-v0.1.0-pc.1-linux-amd64.tar.gz` |
| Versione | `v0.1.0-pc.1` |
| Commit incorporato | `4578c132682e6b715317a6b4d1de958459cfc086` |
| SHA-256 | `18d67a2a6bbeb3db2e46c8a99229fc36346b7de2567f395e3323c91bb75d8e97` |
| Piattaforma | Linux `amd64` |

Il preflight ripetuto sullo stesso archive ha confermato checksum, versione,
commit, help, configurazione e fixture. Il probe `doctor` con endpoint
deliberatamente non disponibile ha mantenuto verdi configurazione, workspace,
composition, agent, tool, policy e riconoscimento Laravel, fallendo soltanto i
check provider e modello come previsto.

# Nuova iterazione `v0.1.0-pc.2`

Gli hardening emersi dal gate live non appartengono a `pc.1`. Dopo regression
test e benchmark a sistema stabile è stata quindi prodotta una nuova
iterazione, che diventa l'unico input ammesso per la ripresa della Fase 5:

| Campo | Valore |
|---|---|
| Artifact | `maestro-v0.1.0-pc.2-linux-amd64.tar.gz` |
| Versione | `v0.1.0-pc.2` |
| Commit incorporato | `b9f571ac5914d2565e2a7bd28f4d5d6fc14a2710` |
| SHA-256 | `91ef1bb196e9904ef3f3f0fefccf3a80acba22f14da43cdccbf9a83680fa41bc` |
| Dimensione | `3586821` byte |
| Piattaforma | Linux `amd64` |
| Licenza | Apache-2.0 |

`pc.2` include envelope JSON strutturato, protocollo del reference agent
aggiornato, timeout e retrieval rivisti per il percorso CPU-only e una guida
d'installazione renderizzata con la versione esatta del candidate.

Il gate di packaging è stato ripetuto integralmente senza riaprire la Fase 4.
Due build indipendenti dallo stesso commit sono byte-identiche; checksum,
manifest, archive paths, licenze, assenza di credential-shaped data,
configurazione, fixture, `version`, help, `doctor` offline e installazione da
directory vuota sono verdi. Un controllo aggiunto rifiuta archive che
contengano il token di versione non risolto.

Un primo build `pc.2` preliminare dal commit precedente è stato scartato prima
di qualsiasi test live perché la guida inclusa citava ancora `pc.1`. I relativi
file sono stati spostati fuori da `dist/` e non costituiscono artifact validi.

## Gate live esatto di `pc.2`

Il candidate definitivo è stato estratto nuovamente in `/tmp` e usato senza
checkout. Checksum, `maestro version`, `doctor`, `models` e `agents` sono
risultati positivi. Il quick start Laravel read-only, con la configurazione
inclusa e streaming abilitato, ha completato correttamente:

| Misura | Risultato |
|---|---:|
| Terminale | `completed` |
| Model turns | 2 |
| Tool calls | 1 |
| Input tokens | 2905 |
| Output tokens | 104 |
| Durata | 363027 ms |

Il run ha letto `OrderController.php` e ha identificato correttamente
`OrderService`. Questa prova sostituisce, per il percorso read-only, la prova
precedente basata sul profilo temporaneamente ridotto.

Il primo run mutativo da una nuova estrazione ha invece emesso due call nello
stesso turno: lettura e patch dipendente. Il runtime di `pc.2` ha terminato con
`tool_failure` prima di mostrare un prompt di approval. Il confronto con il
file estratto direttamente dall'archive conferma in entrambi i casi SHA-256
`4826abe9c6c5d701133817a9dcb565f0b84f760da57e1b518d430b601520b1bd`:
nessuna modifica è avvenuta. La serie mutativa registra 0 successi su 1
tentativo eseguito; il gate previsto richiedeva 3 successi consecutivi e i
tentativi 2–3 non sono stati eseguiti. Non è stato eseguito ulteriore prompt
tuning e `pc.2` è classificato non promuovibile.

ADR-0028 registra l'hardening conseguente: una sola esecuzione tool per turno,
risultato recuperabile correlato per le call eccedenti, `workspace.patch`
nascosto fino a una read verificata, digest/path/testo `old` vincolati
all'ultima osservazione e rifiuto di un terminale testuale mentre una mutazione
proposta resta incompleta. I test al confine applicativo dimostrano che una
patch prematura non raggiunge né approval né esecuzione. L'hardening non è
presente in `pc.2` e richiede un nuovo packaging candidate.

# Nuova iterazione `v0.1.0-pc.3`

L'hardening ADR-0028 è incorporato in un nuovo packaging candidate, prodotto
soltanto dopo il completamento dei gate deterministici:

| Campo | Valore |
|---|---|
| Artifact | `maestro-v0.1.0-pc.3-linux-amd64.tar.gz` |
| Versione | `v0.1.0-pc.3` |
| Commit incorporato | `d362b9910f68e5aecae3a489eb5852e339bc3939` |
| SHA-256 | `8fbdfbf9b207c8c984f295240bcb6345d32fcbfa42f5869dd27a39acc158fe26` |
| Dimensione | `3595670` byte |
| Piattaforma | Linux `amd64` |
| Licenza | Apache-2.0 |

Due build indipendenti sono byte-identiche. Checksum, manifest, archive paths,
licenze, assenza di dati credential-shaped e path del checkout, configurazione,
fixture, guida renderizzata, `version`, help, `doctor` offline e installazione
da directory vuota sono verdi. `pc.3` diventa l'unico input ammesso per la
ripresa live; resta un packaging candidate e non una release candidate.

## Gate live di `pc.3`

Dal candidate estratto sono positivi `doctor` (9 check passed), `models`,
`agents` e il quick start Laravel read-only con configurazione esatta:

| Misura | Risultato |
|---|---:|
| Terminale | `completed` |
| Model turns | 2 |
| Tool calls | 1 |
| Input/output tokens | 2831 / 61 |
| Durata | 340820 ms |

La risposta identifica correttamente `OrderService` e il run non richiede
approval. Il primo tentativo della successiva serie mutativa, da una nuova
estrazione, non ha però emesso tool call. `llama3.1:8b` ha descritto pseudo-call
come contenuto assistant e il runtime ha accettato un terminale `completed` con
1 model turn, 0 tool calls e durata 171978 ms. Nessuna approval è stata
presentata e il target conserva SHA-256
`4826abe9c6c5d701133817a9dcb565f0b84f760da57e1b518d430b601520b1bd`,
identico all'archive.

La coreografia ADR-0028 governa una mutazione dopo che il provider propone una
call mutante; non può inferire in sicurezza l'effetto richiesto da testo libero
quando nessuna call viene emessa. La serie registra quindi 0 successi su 1
tentativo eseguito; il gate previsto richiedeva 3 successi consecutivi e i
tentativi 2–3 non sono stati eseguiti. Non è stato eseguito altro prompt
tuning. `llama3.1:8b` resta positivo per tool calling diretto e
reference agent read-only, ma non è dichiarato supportato per il reference
agent mutante. La release rimane bloccata in attesa di un modello alternativo
validato oppure di un contratto operativo esplicito più stretto. Dopo il test
il modello è stato scaricato dalla RAM con `ollama stop`.

## Selezione economica successiva

Il candidato locale `rnj-1:8b-instruct-q4_K_M` ha superato il Gate A
provider-level con 3 sequenze read-result-patch valide su 3. Il Gate non ha
eseguito tool o modificato file. Al Gate B, il primo run read-only dal binario
`pc.3` ha invocato una read reale ma ha terminato `provider_failure` nel secondo
turno dopo 535537 ms, con exit code 4. Il requisito era `completed` per almeno
due run consecutivi: il modello è stato escluso in fail-fast, il tentativo 2 e
il Gate C non sono stati eseguiti.

Il report completo è `reports/milestone-8-model-selection.md`. Non esiste ancora
un modello vincente e non viene prodotto `pc.4`.

Il candidato successivo `ibm/granite4.1:8b` ha superato `maestro doctor`, il
Gate A provider-level 3/3 e il Gate B read-only 2/2 dal binario `pc.3`. Le due
esecuzioni read-only hanno invocato una read reale, identificato correttamente
`OrderService::create` e terminato `completed` rispettivamente dopo 431979 ms
e 47606 ms.

Il Gate C è stato avviato da una nuova estrazione dell'archive. Il primo run ha
eseguito la read e raggiunto 3 turni e 2 tool call, ma è terminato
`deadline_exceeded` al limite pubblico complessivo, dopo 600077 ms e con exit
code 130. Nessun prompt di approval è stato mostrato, nessun grant è stato
emesso e il controller conserva il digest originale
`4826abe9c6c5d701133817a9dcb565f0b84f760da57e1b518d430b601520b1bd`.
Il Gate C registra quindi 0 successi su 1 tentativo eseguito; i tentativi 2–3
non sono stati eseguiti in fail-fast. Anche Granite è escluso dalla fixture
v0.1.0 sul profilo CPU-only pubblico. Il modello è stato scaricato dalla RAM e
non viene prodotto `pc.4`.

L'ultimo candidato `qwen3:8b` è stato configurato in modalità non-thinking con
una riga finale `/no_think` fissata prima dei gate e applicata uniformemente al
messaggio utente iniziale. Il preflight dal binario `pc.3` ha superato tutti i
9 check. La prima sequenza del Gate A non ha però emesso `workspace_read`: è
terminata `tool_call_count` dopo 100977 ms, con 227/256 token e zero tool call.
Le sequenze 2–3 e i Gate B/C non sono stati eseguiti in fail-fast. L'harness
non ha eseguito Tool Runtime, la fixture conserva il digest originale e Qwen3 è
stato scaricato dalla RAM.

La matrice 8B corrente è conclusa senza un modello vincente. Nessun `pc.4`
viene prodotto e la selezione non prosegue casualmente con altri modelli dello
stesso profilo.

# Ambiente live rilevato

| Componente | Evidenza |
|---|---|
| Ollama | `0.32.5`, endpoint locale |
| Chat positiva | `llama3.1:8b`, GGUF Q4_K_M |
| Embedding | `embeddinggemma:latest`, GGUF BF16 |
| Caso negativo conservato | `qwen2.5-coder:7b`, GGUF Q4_K_M |
| Candidati esclusi | `rnj-1:8b-instruct-q4_K_M`, `ibm/granite4.1:8b`, `qwen3:8b` |
| llama.cpp | `llama-server` versione `1` (`51eae8cfc`) |
| Host | Linux `amd64`, Intel Core i5-8365U, 8 logical CPU |
| Memoria | 15 GiB RAM, 4 GiB swap |

I blob locali dei modelli usati erano leggibili. I candidati successivi erano
già presenti prima del rispettivo gate; nessun modello è stato scaricato,
copiato o modificato durante le esecuzioni di validazione.

# Evidenze Ollama valide

La Smoke matrix eseguita dal binario estratto dal candidate, con manifest
versionato nel repository, ha terminato con exit code `0`:

| Stato | Numero |
|---|---:|
| Passed | 13 |
| Skipped | 1 |
| Failed | 0 |

Lo scenario `acquisition-pull-remove` è stato saltato perché la mutation guard
non era abilitata. Capability introspection, catalogo, completion, streaming,
cancellazione stream, embedding, lifecycle, structured output, tool calling,
resilienza e redazione sono risultati positivi.

Dal candidate installato sono inoltre risultati positivi `doctor`, `models` e
`agents` contro Ollama. Il primo `run` Laravel read-only ha evidenziato che il
timeout provider di due minuti e il bundle di contesto d'esempio erano troppo
aggressivi per un host CPU-only. Con un profilo di prova ridotto (`top_k: 2`,
budget contesto 2048 token) e percorso non-streaming, il run ha completato con:

| Misura | Risultato |
|---|---:|
| Terminale | `completed` |
| Model turns | 2 |
| Tool calls | 1 |
| Input tokens | 2701 |
| Output tokens | 82 |
| Durata | circa 70 secondi |

Il risultato finale ha identificato correttamente il servizio invocato dal
controller Laravel. Poiché il profilo differiva dalla configurazione inclusa
nel candidate, questa evidenza è utile ma non chiude da sola il quick start
ufficiale.

# Finding agentico mutativo

Lo scenario `read -> patch -> reindex -> final` non ha superato il gate live.
Le prove hanno osservato, in momenti diversi:

1. lettura e patch dipendente emesse nello stesso turno, prima che il digest
   della lettura fosse disponibile;
2. un placeholder al posto del digest autorevole;
3. newline ricostruite come testo escaped invece che copiate dal contenuto;
4. una patch descritta come contenuto assistant anziché invocata tramite
   `tool_calls`.

In tutti i casi il Runtime o il tool guardato hanno negato l'esecuzione non
valida. Il digest SHA-256 del file target è rimasto
`c2ae5b667100867423a3822fed0bba2ff64603a63fbb2c802fe785a319021cab`:
nessuna mutazione è avvenuta senza una invocazione valida e nessun prompt di
approvazione è stato aggirato.

Il finding ha prodotto due hardening interni, senza modificare contratti
pubblici o aumentare l'autorità del runtime:

- protocollo del reference agent esplicito su tool sequenziali, invocazione
  tramite interfaccia dichiarata, digest e testo `old` copiati dalle evidenze;
- risultati tool `application/json` inseriti come JSON strutturato
  nell'envelope agent-provider, evitando una doppia codifica ambigua.

La configurazione d'esempio è stata inoltre ridimensionata per uso CPU-only:
timeout provider di cinque minuti, durata run di dieci minuti e retrieval più
compatto. Queste modifiche sono incrementali; non costituiscono una
certificazione del percorso mutativo, che deve essere rieseguito.

# Matrice llama.cpp e incidente

Per riutilizzare esclusivamente i modelli già installati è stato avviato
`llama-server` in router mode con `llama3.1:8b` ed
`embeddinggemma:latest`. Il preflight `/health` e il listing dei due modelli
erano positivi. La matrice completa non è però valida: due tentativi hanno
esaurito la memoria dell'host.

Il kernel registra:

| Ora locale | Processo terminato | RSS anonima registrata |
|---|---|---:|
| 22:18:13 | `llama-server` PID 341826 | 7.948.576 KiB |
| 22:21:53 | `llama-server` PID 344018 | 14.007.056 KiB |

Entrambi gli eventi sono classificati `global_oom`; il kernel ha terminato
`llama-server`. Nel secondo episodio la memoria del processo era prossima alla
RAM fisica totale, mentre erano attivi anche VS Code, Ollama e i processi della
sessione di sviluppo.

Poiché il server era stato avviato dalla sessione integrata in VS Code, la
pressione memoria ha destabilizzato anche il renderer Electron. La finestra ha
mostrato due volte `launch-failed`, codice `1002`. Il crash dell'IDE è quindi
correlato temporalmente e causalmente agli eventi OOM, non a un failure
funzionale di VS Code dimostrato separatamente.

I report generati durante i due tentativi llama.cpp sono invalidi: il primo
contiene due soli preflight passati e dodici scenari cancellati dopo la morte
del server; il secondo è stato interrotto durante il ripristino dell'IDE. Non
devono essere usati per chiudere la Milestone 3.

# Contenimento applicato

- nessun `llama-server` o `maestro bench smoke` resta attivo;
- la memoria disponibile è tornata a circa 12 GiB;
- nessun artifact o modello è stato modificato;
- i candidate `pc.1` e `pc.2` non sono stati promossi;
- la matrice llama.cpp e la Fase 5 restano aperte.

# Gate repository dopo l'hardening

```text
go test -count=3 ./...
go test -race ./...
go vet ./...
go test -run '^$' -bench BenchmarkAgentLoopDeterministic -benchmem ./internal/agent
git diff --check
```

Tutti i gate sono verdi. Il benchmark locale post-modifica osservava circa
`105363 ns/op`, `15517 B/op` e `219 allocs/op` immediatamente dopo l'incidente.
Ripetuto cinque volte a sistema stabile, osserva `36394–41950 ns/op`,
`15517–15518 B/op` e `219 allocs/op`, rientrando nell'intervallo storico. È un
indicatore dell'host, non un budget di release. Nessun packaging candidate è
stato generato da un worktree contenente modifiche o file non tracciati.

Dopo ADR-0028, suite ripetuta tre volte, race detector e vet sono verdi. Il
benchmark ripetuto cinque volte osserva `42385–44923 ns/op`, `15581 B/op` e
`220 allocs/op`; l'allocazione aggiuntiva corrisponde allo stato interno della
coreografia e non modifica il carattere bounded del loop.

# Strategia sicura di ripresa

Il router multi-modello non deve essere rieseguito su questo host. La prossima
validazione llama.cpp deve usare una delle seguenti condizioni:

1. host dedicato con almeno 24–32 GiB di RAM; oppure
2. processi single-model sequenziali per chat ed embedding, fuori dal cgroup di
   VS Code, con Ollama scaricato dalla memoria, `parallel=1`, contesto e batch
   ridotti e limite di memoria imposto dal service manager.

Ogni step deve eseguire un preflight di RAM/swap, monitorare RSS e interrompere
il server prima che inizi pressione sostenuta sullo swap. Chat, streaming,
structured output e tool calling vanno provati con il solo modello chat;
embedding in un processo separato. Lifecycle/router va eseguito soltanto su un
host con memoria sufficiente o classificato esplicitamente come non validato.

ADR-0029 sceglie la v0.1.0 ufficialmente read-only. Il percorso Ollama
supportato usa `llama3.1:8b`, list/read/search e policy mutativa deny; llama.cpp,
tool mutanti e reference agent mutante sono sperimentali/non supportati. La
validazione mutativa viene rinviata almeno alla v0.2.0. Timeout, budget e
criteri delle matrici concluse non vengono ampliati retroattivamente.

# Verdetto

**Fase 5 non conclusa. NO-GO alla release candidate e alla release.**

Sono valide la Smoke matrix Ollama provider-level e la prova Laravel read-only
con configurazione esatta di `pc.3`. La matrice mutativa 8B non produce un
vincitore e ADR-0029 restringe il contratto v0.1.0 al read-only. Restano da
produrre e validare `pc.4`, completare l'installazione pulita e ripetere due
quick start consecutivi. La Milestone 3 resta formalmente aperta, ma llama.cpp
è sperimentale e il relativo debito non blocca la release read-only. `pc.1`,
`pc.2` e `pc.3` restano baseline storiche non promuovibili.

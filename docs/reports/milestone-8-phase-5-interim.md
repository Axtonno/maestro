# Milestone 8 — Fase 5 Interim e incidente OOM

Data: 2026-08-13

Stato: In corso, validazione live sospesa

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

# Ambiente live rilevato

| Componente | Evidenza |
|---|---|
| Ollama | `0.32.5`, endpoint locale |
| Chat positiva | `llama3.1:8b`, GGUF Q4_K_M |
| Embedding | `embeddinggemma:latest`, GGUF BF16 |
| Caso negativo conservato | `qwen2.5-coder:7b`, GGUF Q4_K_M |
| llama.cpp | `llama-server` versione `1` (`51eae8cfc`) |
| Host | Linux `amd64`, Intel Core i5-8365U, 8 logical CPU |
| Memoria | 15 GiB RAM, 4 GiB swap |

I blob locali dei tre modelli erano leggibili. Nessun modello è stato
scaricato, copiato o modificato durante la validazione.

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
- il candidate `pc.1` non è stato promosso;
- la matrice llama.cpp e la Fase 5 restano aperte.

# Gate repository dopo l'hardening

```text
go test -count=3 ./...
go test -race ./...
go vet ./...
go test -run '^$' -bench BenchmarkAgentLoopDeterministic -benchmem ./internal/agent
git diff --check
```

Tutti i gate sono verdi. Il benchmark locale post-modifica osserva circa
`105363 ns/op`, `15517 B/op` e `219 allocs/op`; è un indicatore dell'host, non
un budget di release. Non è stato rigenerato alcun packaging candidate mentre
il worktree conteneva le modifiche della Fase 5.

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

Il percorso Ollama deve essere ripreso da una copia pulita della fixture e deve
dimostrare una patch reale, approval one-shot, reindex e risposta finale prima
di procedere a deny, EOF, no-TTY, SIGINT, hard limit e installazione pulita.

# Verdetto

**Fase 5 non conclusa. NO-GO alla release candidate e alla release.**

Sono valide la Smoke matrix Ollama provider-level e la prova Laravel read-only
con profilo ridotto. Restano blocker lo scenario mutativo Ollama, la matrice
llama.cpp, l'installazione pulita completa e tutti i gate operativi finali. La
Milestone 3 resta formalmente aperta.

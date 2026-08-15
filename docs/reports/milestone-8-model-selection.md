# Milestone 8 — Selezione economica del modello mutativo

Data: 2026-08-15

Stato: Matrice corrente conclusa — nessun modello vincente

Verdetto: **rnj escluso al Gate B, Granite al Gate C e Qwen3 al Gate A;
decisione sul contratto v0.1.0 richiesta, nessun `pc.4`**

---

# Scopo

Questo report applica la matrice fail-fast successiva al gate live di `pc.3`.
La selezione precede qualsiasi nuovo packaging candidate: usa il comportamento
provider e agentico incorporato in `pc.3`, configurazioni temporanee esplicite
e una fixture Laravel immutata.

Il gate mutativo precedente va letto come 0 successi su 1 tentativo eseguito;
il criterio previsto era 3 successi consecutivi e i tentativi 2–3 non sono
stati eseguiti. La forma abbreviata `0/3` non viene più usata.

# Candidate `rnj-1:8b-instruct-q4_K_M`

| Campo | Valore |
|---|---|
| Modello | `rnj-1:8b-instruct-q4_K_M` |
| Ollama model ID | `d20e29ab8d0f` |
| Dimensione locale | 5,1 GB |
| Provider | Ollama `0.32.5` |
| Baseline Maestro | `v0.1.0-pc.3` |
| Commit codice | `d362b9910f68e5aecae3a489eb5852e339bc3939` |
| Piattaforma | Linux `amd64`, CPU-only |

Prima del test, `maestro doctor` dal candidate con il modello selezionato ha
superato tutti i 9 check. Il sorgente Go usato dall'harness provider-level è
byte-identico al commit incorporato in `pc.3`: non risultano differenze in
file `*.go`, `go.mod` o `go.sum`.

# Gate A — Protocollo diretto

Il gate usa temperatura `0` e tre conversazioni indipendenti con input
identico. Ogni sequenza:

1. dichiara soltanto `workspace_read` e richiede una call reale;
2. valida nome, cardinalità, schema e path;
3. restituisce un risultato JSON con contenuto e SHA-256 autorevoli;
4. dichiara soltanto `workspace_patch`;
5. valida schema strict, path, digest e presenza esatta del testo `old`;
6. verifica che la patch proposta realizzi la sola sostituzione richiesta;
7. non invoca Tool Runtime e non scrive il file.

Una prima esecuzione di calibrazione non è conteggiata: il modello aveva emesso
una tool call reale accompagnata da contenuto testuale, mentre l'harness
iniziale rifiutava qualunque contenuto. Il criterio è stato riallineato a quello
del gate, che vieta pseudo-call JSON/fenced/tagged ma non prosa accessoria.

| Sequenza ufficiale | Read call | Patch call | Esito |
|---:|---:|---:|---|
| 1 | 1 | 1 | Passed |
| 2 | 1 | 1 | Passed |
| 3 | 1 | 1 | Passed |

Gate A: **superato 3/3**. Nessuna patch è stata eseguita.

# Gate B — Reference agent read-only

Il Gate B usa il binario estratto da `pc.3`, streaming abilitato e lo stesso
profilo pubblico di timeout, tool, limiti e contesto, modificando esplicitamente
il solo modello chat e il path della fixture temporanea.

| Tentativo | Terminale | Model turns | Tool calls | Durata | Exit code |
|---:|---|---:|---:|---:|---:|
| 1 | `provider_failure` | 2 | 1 | 535537 ms | 4 |

La `workspace_read` è stata realmente invocata. La completion finale non ha
però concluso entro il profilo operativo: il terminale e la durata sono
coerenti con l'esaurimento del timeout provider di cinque minuti durante il
secondo turno. Il requisito era `completed`, risposta corretta e almeno 2
esecuzioni consecutive.

Gate B: **0 successi su 1 tentativo eseguito**. Il tentativo 2 non è stato
eseguito per la regola fail-fast.

# Integrità e contenimento

- Gate C non avviato;
- nessuna approval presentata;
- nessuna patch o altra mutazione eseguita;
- `OrderController.php` conserva SHA-256
  `4826abe9c6c5d701133817a9dcb565f0b84f760da57e1b518d430b601520b1bd`,
  identico all'archive;
- il modello è stato scaricato dalla RAM dopo il gate;
- la memoria disponibile è tornata a circa 11 GiB;
- `qwen2.5-coder:7b` resta il caso negativo canonico e non è stato rieseguito.

# Verdetto

`rnj-1:8b-instruct-q4_K_M` dimostra un protocollo tool diretto affidabile nel
campione 3/3, ma non soddisfa l'esperienza operativa read-only sul profilo
CPU-only e timeout pubblico di `pc.3`. È escluso dalla selezione v0.1.0 prima
del Gate C.

Non esiste ancora un modello vincente, `pc.3` resta una baseline storica non
promuovibile e `pc.4` non deve essere prodotto.

# Candidate `ibm/granite4.1:8b`

## Identità e preflight

| Campo | Valore |
|---|---|
| Modello | `ibm/granite4.1:8b` |
| Ollama model ID | `444af1c4b2fe` |
| Dimensione locale | 5,3 GB |
| Provider | Ollama `0.32.5` |
| Baseline Maestro | `v0.1.0-pc.3` |
| Commit codice | `d362b9910f68e5aecae3a489eb5852e339bc3939` |
| Piattaforma | Linux `amd64`, CPU-only |

Il modello era già presente localmente. `maestro doctor`, eseguito dal binario
estratto da `pc.3` con configurazione temporanea identica al profilo pubblico
salvo il modello chat e il path esplicito della fixture, ha superato tutti i 9
check. Nessun artifact è stato ricostruito o modificato.

## Gate A — Protocollo diretto

È stato riutilizzato lo stesso harness strict descritto per `rnj-1`, con
temperatura zero, conversazioni indipendenti e nessuna esecuzione dei tool. Per
ogni turno sono stati registrati anche durata e token.

| Sequenza | Read | Patch | Token read in/out | Token patch in/out | Esito |
|---:|---:|---:|---:|---:|---|
| 1 | 43145 ms | 86741 ms | 258 / 28 | 584 / 83 | Passed |
| 2 | 9065 ms | 28166 ms | 258 / 28 | 584 / 83 | Passed |
| 3 | 9264 ms | 28555 ms | 258 / 28 | 584 / 83 | Passed |

Ogni sequenza ha prodotto esattamente una call `workspace_read`, ha consumato
il risultato autorevole con contenuto e digest e ha poi prodotto esattamente
una call `workspace_patch` valida per la sostituzione richiesta.

Gate A: **superato 3/3**. Nessun effetto è stato eseguito.

## Gate B — Reference agent read-only

Il Gate B ha usato il binario e la fixture estratti da `pc.3`, streaming e
profilo pubblico invariati. Entrambi i run hanno letto realmente
`app/Http/Controllers/OrderController.php`, hanno identificato correttamente
`App\Services\OrderService::create` e hanno terminato senza mutazioni.

| Tentativo | Terminale | Model turns | Tool calls | Token in/out | Durata | Exit code |
|---:|---|---:|---:|---:|---:|---:|
| 1 | `completed` | 2 | 1 | 3148 / 102 | 431979 ms | 0 |
| 2 | `completed` | 2 | 1 | 3148 / 115 | 47606 ms | 0 |

Gate B: **superato 2/2**.

## Gate C — Reference agent mutante

Il primo tentativo è partito da una nuova estrazione dell'archive, con digest
iniziale del controller
`4826abe9c6c5d701133817a9dcb565f0b84f760da57e1b518d430b601520b1bd`.
Il task richiedeva di leggere il file e sostituire esclusivamente la chiave JSON
`data` con `order` nel metodo `store`.

Il run ha eseguito una read reale e ha raggiunto 3 turni e 2 tool call, ma non
ha completato la coreografia entro il limite pubblico di dieci minuti. È
terminato `deadline_exceeded` dopo 600077 ms con exit code 130, prima di
mostrare un prompt di approval o eseguire la patch. Il controller conserva il
digest originale: nessuna mutazione è avvenuta e nessun grant è stato emesso.

Gate C: **0 successi su 1 tentativo eseguito**. I tentativi 2–3 non sono stati
eseguiti per la regola fail-fast.

## Matrice dopo Granite

| Modello | Tool diretto | Agent read-only | Agent mutante | Fixture v0.1.0 |
|---|---|---|---|---|
| `rnj-1:8b-instruct-q4_K_M` | Compatibile, 3/3 | Non supportato sul profilo pubblico | Non valutato | Escluso |
| `ibm/granite4.1:8b` | Compatibile, 3/3 | Compatibile, 2/2 | Non supportato sul profilo pubblico | Escluso |

Granite dimostra un progresso sostanziale rispetto a `rnj-1`, ma non soddisfa
il requisito release-oriented di tre mutazioni complete consecutive entro i
limiti pubblici. Il modello è stato scaricato dalla RAM; `ollama ps` non
riportava modelli residenti e la memoria disponibile era tornata a circa 11
GiB.

Non esiste ancora un modello vincente. `pc.3` resta una baseline storica non
promuovibile, `pc.4` non viene prodotto e il candidato successivo previsto era
`qwen3:8b` in modalità non-thinking, da sottoporre senza modifiche agli stessi
Gate A, B e C.

# Candidate `qwen3:8b`

## Profilo non-thinking fissato prima dei gate

| Campo | Valore |
|---|---|
| Modello | `qwen3:8b` |
| Ollama model ID | `500a1f067a9f` |
| Dimensione locale | 5,2 GB |
| Provider | Ollama `0.32.5` |
| Baseline Maestro | `v0.1.0-pc.3` |
| Commit codice | `d362b9910f68e5aecae3a489eb5852e339bc3939` |
| Piattaforma | Linux `amd64`, CPU-only |
| Thinking control | riga finale `/no_think` nell'istruzione utente iniziale |

`pc.3` non espone il campo Ollama `think` nella configurazione pubblica. Il
profilo Qwen3 usa quindi il soft switch documentato dal modello: ogni
conversazione indipendente del Gate A e ogni istruzione iniziale dei Gate B/C
terminano con una sola riga `/no_think`. La direttiva resta nello stesso punto
del protocollo, non viene aggiunta ai risultati tool o a turni selezionati e
non modifica system prompt, timeout, budget, fixture, temperatura del Gate A o
criteri di accettazione. Riferimenti: [Ollama Thinking](https://docs.ollama.com/capabilities/thinking)
e [Qwen3-8B model card](https://huggingface.co/Qwen/Qwen3-8B).

Questo meccanismo è fissato prima dell'esecuzione. Se Qwen3 supera tutti i gate,
il profilo dovrà diventare parte esplicita del successivo `pc.4`; in caso di
fallimento non verranno provati casualmente altri modelli 8B nella matrice
corrente e il contratto di release dovrà essere rivalutato.

## Preflight e Gate A

Il modello era già presente localmente. `maestro doctor` dal binario `pc.3`,
con configurazione temporanea identica al profilo pubblico salvo il modello
chat, ha superato tutti i 9 check. L'harness Gate A testato e sottoposto a vet
aveva SHA-256
`9223362b49384946d2a02ff75eaf1ad4f5505b53512b6899c4178c98644e4184`.

La prima conversazione ufficiale ha applicato temperatura zero, tool choice e
schema strict invariati e la riga finale `/no_think`. Il completamento read non
ha emesso tool call:

| Sequenza | Stage | Tool call | Token in/out | Durata | Esito |
|---:|---|---:|---:|---:|---|
| 1 | read | 0 | 227 / 256 | 100977 ms | Failed: `tool_call_count` |

L'output ha raggiunto il limite di 256 token senza produrre la singola
`workspace_read` richiesta. Il Gate A richiedeva tre sequenze valide
consecutive.

Gate A: **0 successi su 1 sequenza eseguita**. Le sequenze 2–3 non sono state
eseguite per la regola fail-fast; i Gate B e C non sono stati avviati.

## Integrità e contenimento finali

- Gate A non ha invocato Tool Runtime né eseguito effetti;
- nessuna approval è stata presentata;
- `OrderController.php` conserva SHA-256
  `4826abe9c6c5d701133817a9dcb565f0b84f760da57e1b518d430b601520b1bd`;
- timeout, budget, fixture, prompt di protocollo e criteri non sono stati
  modificati dopo l'esito;
- Qwen3 è stato scaricato dalla RAM, `ollama ps` era vuoto e la memoria
  disponibile era tornata a circa 11 GiB;
- nessun packaging candidate è stato prodotto.

# Matrice finale e decisione richiesta

| Modello | Tool diretto | Agent read-only | Agent mutante | Fixture v0.1.0 |
|---|---|---|---|---|
| `rnj-1:8b-instruct-q4_K_M` | Compatibile, 3/3 | Non supportato sul profilo pubblico | Non valutato | Escluso |
| `ibm/granite4.1:8b` | Compatibile, 3/3 | Compatibile, 2/2 | Non supportato sul profilo pubblico | Escluso |
| `qwen3:8b` non-thinking | Non compatibile, 0 successi su 1 sequenza | Non valutato | Non valutato | Escluso |

La ricerca dei modelli 8B della matrice corrente termina senza un vincitore.
Le evidenze disponibili non giustificano `pc.4` né un aumento retroattivo dei
limiti. Prima di proseguire deve essere scelta e formalizzata una delle seguenti
direzioni di prodotto:

1. v0.1.0 ufficialmente read-only, con mutazioni rinviate;
2. mutazioni conservate con un requisito hardware superiore che includa la
   capacità computazionale, non soltanto la RAM;
3. release rinviata finché non viene identificata una fixture adeguata.

Fase 5 resta aperta e la release candidate rimane in NO-GO fino a questa
decisione. `pc.3` è una baseline storica non promuovibile e `pc.4` non viene
prodotto.

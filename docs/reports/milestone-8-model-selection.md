# Milestone 8 — Selezione economica del modello mutativo

Data: 2026-08-15

Stato: In corso — nessun modello vincente

Verdetto: **`rnj-1:8b-instruct-q4_K_M` escluso al Gate B e
`ibm/granite4.1:8b` escluso al Gate C; nessun `pc.4`**

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

## Matrice e verdetto aggiornati

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
promuovibile, `pc.4` non viene prodotto e il prossimo candidato previsto è
`qwen3:8b` in modalità non-thinking, sottoposto senza modifiche agli stessi
Gate A, B e C.

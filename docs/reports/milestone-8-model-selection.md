# Milestone 8 — Selezione economica del modello mutativo

Data: 2026-08-15

Stato: In corso — nessun modello vincente

Verdetto: **`rnj-1:8b-instruct-q4_K_M` escluso al Gate B; nessun `pc.4`**

---

# Scopo

Questo report applica la matrice fail-fast successiva al gate live di `pc.3`.
La selezione precede qualsiasi nuovo packaging candidate: usa il comportamento
provider e agentico incorporato in `pc.3`, configurazioni temporanee esplicite
e una fixture Laravel immutata.

Il gate mutativo precedente va letto come 0 successi su 1 tentativo eseguito;
il criterio previsto era 3 successi consecutivi e i tentativi 2–3 non sono
stati eseguiti. La forma abbreviata `0/3` non viene più usata.

# Candidate esaminato

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
promuovibile e `pc.4` non deve essere prodotto. La selezione può riprendere
soltanto con un nuovo modello candidato sottoposto agli stessi Gate A, B e C.

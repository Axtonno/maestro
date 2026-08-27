# Milestone 13 — Direct/Chat Differential Diagnostic

Data: 2026-08-27

Stato: esperimento development-only completato; nessun valore di
qualificazione

## Scopo

Isolare il primo livello di failure fra modello nativo, adapter Maestro,
contesto esplicito e loop agentico usando lo stesso modello, la stessa domanda
circoscritta e lo stesso file Laravel. La prova non riapre la matrice ufficiale
di 22 run, non modifica choreography o prompt di prodotto e non autorizza un
artifact `v0.2.1`.

## Profilo congelato

| Campo | Valore |
|---|---|
| Modello | `qwen3.5:9b` |
| Digest | `6488c96fa5faab64bb65cbd30d4289e20e6130ef535a93ef9a49f42eda893ea7` |
| Ollama | `0.32.14` |
| Contesto effettivo | 4096 |
| Thinking | default del modello |
| Residenza osservata | 100% CPU; `ollama ps` conferma context 4096 dopo la run |
| Timeout completion | 10 minuti |
| Limite agent run | 15 minuti |
| File | fixture versionata `routes/api.php` |
| SHA-256 file | `7e224d7e57bf0be6d2618e668d8515b07d332f94b7960de4640e4388b31bbc39` |
| SHA-256 domanda | `1565130107f8bac656423082176daf7fde8a1709da3d3e223b147f9d96e7a97a` |
| SHA-256 system direct/completion | `70bf0cc52cd1b7548167164fc266710f43e00968576331b32a572765794aa3d8` |

La domanda chiede endpoint HTTP, controller e action dichiarati dal file e
impone di usare soltanto fatti dimostrati dal file. Nei percorsi senza file, il
comportamento corretto è dichiarare che la risposta non è determinabile.

Le quattro completion senza tool hanno usato lo stesso system prompt e la
stessa domanda. I due percorsi “preloaded” aggiungono lo stesso contenuto del
file al messaggio utente. Il loop agentico usa necessariamente il system prompt
e la choreography del reference agent: questa differenza è la variabile
osservata, non viene nascosta come contesto identico byte per byte.

## Limite del confronto Continue

Continue è installato come estensione VS Code, non come CLI automatizzabile. Il
profilo locale osservato usa `qwen2.5-coder:7b`, quindi eseguirlo così com'è
avrebbe violato il freeze sul modello. Non è stata modificata la configurazione
Continue e non viene dichiarata una run Continue.

Il secondo punto è pertanto un **file-attached direct proxy**: API Ollama
nativa, `qwen3.5:9b`, stesso system prompt e file incluso esplicitamente. Isola
il valore del contesto fornito, ma non misura template o UX specifici di
Continue.

## Risultati

| Percorso | Esito | Durata ms | Token in/out | Qualità / osservazione |
|---|---|---:|---:|---|
| Ollama diretto, nessun file | `context deadline exceeded` | 600.100 | non disponibili | nessuna risposta terminale |
| File-attached direct proxy | HTTP 200 / `stop` | 368.908 | 123/844 | `correct`: `POST /orders`, `OrderController`, `store`; 3.392 caratteri thinking |
| Maestro completion, nessun file | `stop` | 561.237 | 80/1.344 | `correct`: dichiara che il contenuto non è disponibile; zero tool call |
| Maestro completion, file pre-caricato | `deadline_exceeded` | 600.100 | non disponibili | nessuna risposta terminale; zero tool dichiarati |
| Maestro progressive agent loop | `deadline_exceeded` | 900.001 | ultimo totale completo 12.688/878 | 6 turni, 5 tool call, nessuna finale; workspace invariato |

Le richieste sono state eseguite in sequenza. La prima include il cold load;
le successive beneficiano della residenza warm. Non essendoci ripetizioni, le
durate non costituiscono una distribuzione e la differenza fra i due percorsi
con file non può essere attribuita causalmente al solo adapter.

## Traccia redatta del loop agentico

Il preflight ha superato 9/9 controlli. Il loop ha poi osservato:

1. bootstrap `workspace.read` di `routes/api.php` riuscito e stato `route`
   chiuso;
2. prima risposta modello con call nativa `workspace.read` del controller,
   riuscita;
3. tentativo di finalizzazione al secondo turno, correttamente respinto con
   `incomplete_evidence`;
4. tre ulteriori read riuscite nello stato `controller_action`, senza
   dichiarazione `covered`;
5. sesto turno interrotto dalla deadline complessiva.

Il terminale registra 6 turni e 5 tool call. L'ultimo contatore pubblicato dopo
un turno completo è 12.688/878 token. Il fixture è ancora identico alla
versione Git e nessun tool mutativo era dichiarato.

## Interpretazione

La matrice non produce una singola frontiera “passa/fallisce” monotona:

- un failure è osservabile già in Ollama diretto senza file, quindi latenza e
  convergenza con default thinking precedono l'agent runtime;
- il file-attached direct proxy risponde correttamente, quindi il modello può
  svolgere il task single-file quando riceve l'evidenza esplicita;
- Maestro completion senza file conserva correttamente la disciplina
  epistemica, quindi l'adapter non mostra un failure sistematico;
- Maestro con file pre-caricato scade dove il proxy nativo completa, ma una
  sola coppia e la forte variabilità osservata non distinguono adapter da
  modello/template;
- nel loop completo read, permission e tool contract funzionano, mentre il
  modello non chiude lo stato dopo evidenza positiva e consuma la deadline.

La diagnosi più precisa è quindi:

```text
modello/default thinking: operativamente instabile anche senza agent loop
adapter/context pre-caricato: non qualificato; causa non isolata su n=1
tool execution e containment: positivi nel run osservato
verified agent: failure di convergenza/choreography dopo retrieval positivo
```

Non emerge evidenza per la conclusione generica “Maestro non funziona”. Emerge
invece che il runtime sicuro e rigoroso richiede al reference agent capacità di
progressione che i modelli eseguibili stabilmente sull'hardware attuale non
hanno dimostrato. La CLI di prodotto non offre inoltre una modalità
conversazionale esplicita per domande circoscritte con contesto già selezionato.

## Decisione di prodotto

Il confronto sostiene la progettazione di due modalità distinte:

| Modalità | Scopo | Autorità |
|---|---|---|
| `direct/chat` | domanda circoscritta su contesto esplicitamente fornito | nessun tool; nessun retrieval o fallback agentico implicito |
| `verified agent` | esplorazione e sintesi con evidenza verificata | tool read-only, state machine e stop rule |

`direct/chat` non deve essere un fallback silenzioso del verified agent. Deve
avere comando/configurazione, prompt, limiti, metriche e verdict distinti. Il
primo gate futuro misura almeno risposta epistemicamente corretta senza file,
risposta single-file corretta con file, latenza, token, `num_ctx` e `thinking`
effettivi. Il profilo Qwen/default-thinking osservato non è già qualificato per
questa modalità: una completion corretta ha richiesto oltre sei minuti e il
percorso Maestro equivalente ha raggiunto la deadline.

Il verdetto finale della Milestone 13 resta
`field_validation_completed_with_limitations` con
`adoption_no_go_on_reference_profile`, ora con una distinzione causale più
precisa e un requisito di prodotto concreto prima di nuova selezione modelli.

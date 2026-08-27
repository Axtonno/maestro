# Milestone 13 — Field Validation Task Matrix

Versione: 0.2.0

Stato: Chiuso anticipatamente — 5/22 run ufficiali, 17 `not_run`

Data: 2026-08-21; chiusura 2026-08-27

La matrice sotto conserva il disegno originario. La stop rule ha chiuso la
campagna dopo il primo blocco su `project-b`; nessuna cella mancante viene
simulata o reinterpretata. Il verdetto è in
`reports/milestone-13-field-validation.md`.

Questa matrice definisce i task minimi della Field Validation. Prima della
prima run, i placeholder vengono istanziati localmente con simboli verificati
nel repository e le relative rubriche vengono congelate. La copia istanziata,
i prompt esatti, le risposte e i riferimenti al sorgente non vengono committati
quando il progetto non è pubblico.

## Matrice minima

| ID | Ambito | Istruzione template | Evidenza obbligatoria | Frequenza minima |
|---|---|---|---|---:|
| `FV-01` | Mappa progetto | Spiega la struttura applicativa rilevante di questo progetto Laravel e indica gli entry point principali, citando i path logici che supportano la risposta. | entry point reale, almeno due componenti applicativi, relazioni coerenti, nessun simbolo inventato | 2 run per progetto |
| `FV-02` | Controller | Spiega come `<Controller>::<method>` gestisce `<request/use-case>`, dalla validazione alla risposta, e indica le collaborazioni chiamate. | controller e metodo esatti, validazione/input, collaboratore reale, response/terminale | 2 run per progetto |
| `FV-03` | Service e dipendenze | Elenca le dipendenze dirette e indirette rilevanti di `<Service>` per `<use-case>` e spiega il ruolo di ciascuna con evidenza dai file. | constructor/factory binding, interfacce e implementazioni quando presenti, ruolo corretto, path logici | 2 run per progetto |
| `FV-04` | Flusso multi-file | Traccia `<use-case>` dall'entry point fino all'effetto applicativo finale attraversando controller, service e gli altri componenti effettivamente coinvolti. Distingui ciò che il codice dimostra da ciò che resta incerto. | ordine del flusso, almeno tre file quando esistono, branch/errori rilevanti, incertezze esplicite | 2 run per progetto |
| `FV-05` | Verifica di una tesi | Verifica la tesi `<claim preparato>` sul comportamento di `<symbol/flow>`. Confermala o confutala usando soltanto il repository e indica l'evidenza. | verdetto corretto, evidenza a favore/contro, nessuna accettazione della premessa senza verifica | 1 claim per progetto, 2 run |
| `FV-06` | Docker senza authority | Usando soltanto i file del repository, spiega come la configurazione Docker/Compose supporta `<service/flow>` e quali aspetti runtime non possono essere determinati senza interrogare Docker. Non eseguire container o comandi. | file Docker reali, servizi/config coerenti, limite statico dichiarato, nessun claim sullo stato runtime | 2 run su almeno 1 progetto con Docker |

Il minimo della campagna è 22 run: 16 per i quattro task core su due progetti,
quattro per `FV-05` e due per `FV-06`. Un progetto senza configurazione Docker
produce `not_applicable`, non PASS; la coorte deve comunque includere almeno un
progetto idoneo a `FV-06` per chiudere il relativo gate.

## Parametri da congelare

Per ogni coppia progetto/task, prima della run vengono compilati localmente:

| Campo | Regola |
|---|---|
| `project_alias` | alias non reversibile, per esempio `project-a` |
| `task_id` | uno degli ID sopra |
| `target_symbols` | simboli esistenti verificati dall'operatore |
| `instruction_sha256` | digest del prompt istanziato; il prompt resta locale |
| `required_evidence_codes` | codici della rubrica, senza contenuto sorgente |
| `allowed_tools` | esattamente list/read/search |
| `limits_profile` | identità della configurazione congelata |
| `repetitions` | `2`, salvo deviazione dichiarata |
| `stop_rule` | security incident, leak o prerequisito perso |

La seconda ripetizione usa lo stesso `instruction_sha256`, artifact,
configurazione e stato applicativo iniziale. Cache o residenza del modello
vengono annotate, non alterate per migliorare un risultato.

## Rubrica per task

Ogni task ha da tre a cinque criteri obbligatori, decisi prima della campagna.
La classe finale segue il piano della Milestone 13:

| Esito | Regola operativa |
|---|---|
| `correct` | tutti i criteri obbligatori soddisfatti e zero falsità materiali |
| `partial` | almeno un criterio utile soddisfatto, uno o più omessi, zero falsità materiali |
| `incorrect` | falsità materiale, simbolo inventato, flusso contraddetto o task non risposto |
| `unevaluable` | ground truth insufficiente o accesso autorizzato incompleto |

La rubrica valuta anche la disciplina epistemica: una risposta perde la classe
`correct` se presenta come certo un legame che il repository non dimostra. Non
si richiede una forma testuale particolare e non si premia la lunghezza.

## Controlli associati a ogni run

1. verificare identità dell'artifact e profilo;
2. acquisire snapshot redatto pre-run e stato Docker osservato dall'operatore
   per `FV-06`;
3. eseguire una sola istruzione senza follow-up interattivi;
4. acquisire stdout e stderr separati in storage locale `0600`;
5. registrare exit code, terminale, reason code, durata, turni, token e tool
   call disponibili;
6. confrontare snapshot e stato post-run;
7. valutare la risposta contro la rubrica, con seconda revisione quando
   richiesta;
8. trasferire nel dataset pubblico soltanto campi redatti e codici;
9. eseguire anti-leak prima di committare qualsiasi report.

## Prove operative fuori dal completion rate

SIGINT, deadline e hard limit sono controlli intenzionalmente terminali e non
entrano nel denominatore dei task applicativi. Vengono eseguiti almeno una
volta per profilo e registrati separatamente con terminale atteso,
immutabilità, tempo di shutdown e anti-leak.

La prova Docker non concede authority aggiuntiva: nessun socket viene esposto,
nessun tool process/container viene registrato e l'operatore osserva lo stato
esterno prima e dopo. Maestro può leggere soltanto i file indicizzati già
ammessi dal profilo ufficiale.

# Milestone 25 — v0.3.1 Direct Chat Field Adoption

Versione osservata: v0.3.1

Stato: Completata — `field_adoption_negative`

Data: 2026-09-02

Prerequisito: M24 chiusa con `v0.3.1_released_and_verified`.

## Obiettivo

Misurare se l'esatto asset pubblico v0.3.1 è utile nel lavoro quotidiano su
uno o due progetti Laravel reali, usando esclusivamente Direct Chat
single-file sul reference hardware già qualificato.

La milestone è osservativa. Non aggiunge funzionalità, non modifica il
support claim e non riqualifica hardware, provider o modello. Una futura
qualifica CPU moderna resta un'attività separata, con GPU disabilitata e zero
offload dimostrato; il ThinkPad T490s conserva la classificazione di lower
bound legacy respinto.

## Identità congelata

| Campo | Valore |
|---|---|
| release | `v0.3.1`, asset pubblico GitHub |
| commit incorporato | `bd0e902c8d7ef01c01117537fceed76845a33732` |
| SHA-256 archive | `2420ba89ada7b0b9cf3de8bd62d7f97dc32868aa342e44e5c3dacbaa94b3a6b6` |
| SHA-256 binario | `0d5e068019e5187c517f9ff0bc7966b5f3123be933b6d858f2f2fa16978c36ed` |
| piattaforma | WSL2, Ubuntu 24.04, Linux `amd64`, NVIDIA RTX 5070 |
| provider | Ollama 0.33.1, endpoint loopback |
| modello | `qwen3.5:9b` |
| digest modello | `6488c96fa5faab64bb65cbd30d4289e20e6130ef535a93ef9a49f42eda893ea7` |
| profilo | schema v3, context 4096, thinking false, temperatura zero, `num_predict: 1024`, residency `5m` |

Archive e checksum devono essere riscaricati dalla release in una directory
nuova. Il binario viene estratto fuori dal checkout e invocato mediante path
assoluto. Non sono ammessi rebuild, asset locali, model pull, cambio di
provider, retry qualitativi o tuning.

## Progetti e riservatezza

La serie usa almeno `project-a`; `project-b` è ammesso come estensione
precongelata. Devono essere progetti Laravel reali e distinti dalle fixture di
Maestro. Nomi, path, remote, contenuti, prompt completi e risposte complete non
entrano nei documenti versionati.

Prima della prima inferenza viene creato fuori dal repository un manifest
privato con permessi `0600` contenente:

- alias progetto, commit Git e stato iniziale;
- file selezionati, dimensione e SHA-256;
- classe di complessità `simple` o `articulated` e motivazione;
- prompt esatto e oracolo scritto esclusivamente dal file;
- digest completo del workspace e regole di esclusione, se necessarie.

La selezione dei file, gli oracoli e i prompt vengono congelati insieme. Non è
consentito sostituire un file o riscrivere un prompt dopo aver visto una
risposta. `.env`, secret, credenziali, dati cliente, `vendor/`, `storage/` e
file generati non sono eleggibili.

## Configurazione di esecuzione

Per ogni progetto si crea fuori dal repository una copia privata del profilo
distribuito. L'unico delta ammesso è `workspace.root`, impostato al path reale
del progetto. Versione schema, provider, endpoint, modello, digest, context,
thinking, temperatura, `num_predict`, residency, timeout e policy read-only
devono restare byte-equivalenti nei rispettivi valori.

Il delta viene verificato automaticamente o con un diff redatto prima delle
run. Nessuna configurazione reale o path fisico viene committato.

## Matrice core

| ID | Attività quotidiana | Classe file | Evidenza richiesta |
|---|---|---|---|
| `M25-C0` | domanda sul progetto senza file | nessuno | dichiara il contesto insufficiente senza inventare fatti |
| `M25-C1` | spiegazione di servizio o controller semplice | simple | responsabilità, dipendenze e flusso visibili |
| `M25-C2` | analisi di un controller con validazione | simple/articulated | regole, chiamate e risposta senza dedurre route o persistenza |
| `M25-C3` | individuazione degli effetti di un service | articulated | ordine ed effetti visibili; semantica esterna marcata ignota |
| `M25-C4` | proposta di refactoring | articulated | fatti e proposte nettamente separati; nessuna modifica applicata |
| `M25-C5` | suggerimenti di test | simple/articulated | casi ancorati ai branch e contratti osservabili |
| `M25-C6` | domanda deliberatamente oltre il file | simple | limiti dell'evidenza espliciti; nessuna ricostruzione certa |
| `M25-C7` | spiegazione di un file articolato | articulated | flusso completo, incertezze e output senza troncamento |

Se `project-b` è disponibile, la matrice aggiunge `M25-B1`, `M25-B2` e
`M25-B3`, repliche rispettivamente di spiegazione semplice, validazione e file
articolato. Le soglie sono calcolate sul numero totale di task congelati.

`M25-C1` viene ripetuto una sola volta con `--stream` per verificare coerenza
semantica dei trasporti; la replica non sostituisce il risultato core e non
entra nel denominatore qualitativo.

## Metodo di esecuzione

1. riscaricare e verificare archive e checksum pubblici;
2. estrarre fuori checkout e verificare identità binaria;
3. attestare hardware, provider e digest del modello senza modificarli;
4. congelare manifest privato, oracoli, prompt e digest workspace;
5. eseguire `doctor --mode chat` sul profilo di ogni progetto;
6. eseguire ogni task una sola volta nell'ordine della matrice;
7. registrare stdout e stderr separati con permessi `0600`;
8. valutare le risposte contro gli oracoli prima di aggregare le metriche;
9. ripetere stato Git e digest completo del workspace;
10. produrre soltanto evidenza redatta e aggregata nel report pubblico.

Un failure infrastrutturale può essere ripetuto una sola volta soltanto dopo
averne documentato la causa; entrambe le run restano nell'evidenza operativa.
Non sono ammessi retry per migliorare qualità, latenza o risposta.

## Metriche

Per ogni run si registrano:

- exit code, terminale e reason code;
- token input/output e presenza di `length`;
- durata end-to-end e modalità complete/stream;
- dimensione e classe del file;
- heartbeat attesi/osservati e conformità alla forma redatta;
- qualità `correct`, `partial`, `incorrect` o `unevaluable`;
- falsità materiali e assunzioni non dimostrate;
- utilità percepita da 1 a 5, annotata prima della run successiva;
- stato Git e digest workspace pre/post.

Il report aggrega completion rate, correct rate sulle run valutabili, falsità
materiali, terminali `length`, p50/p95 della latenza, distribuzione
dell'utilità, copertura heartbeat e risultato della diagnostica. Le run senza
risposta sono `unevaluable`, non vengono trasformate in errori qualitativi e
restano failure nel completion rate.

Per ogni run durata almeno 15 secondi è atteso almeno un heartbeat. Stderr può
contenere soltanto righe `progress\tstate=generating elapsed_ms=N`; contenuto,
prompt, risposta, modello, endpoint, path e secret sono vietati.

## Diagnostica osservata

Oltre al profilo valido vengono create in una directory temporanea tre copie
sintetiche e non sensibili: versione schema ignota, `num_predict` invalido e
residency invalida. Ogni caso deve fallire chiuso con categoria e campo utili,
senza riportare valori, path o contenuti. Queste prove misurano l'efficacia
operativa della diagnostica e non modificano i progetti.

## Gate e verdetti

I gate di sicurezza sono assoluti:

- identità dell'asset pubblico e configurazione congelata coerenti;
- nessuna mutazione del workspace;
- nessun leak o ampliamento di autorità;
- nessun tool, retrieval, agent o fallback.

Una violazione produce stop immediato e verdetto `field_adoption_incident`.

| Verdetto | Regola |
|---|---|
| `field_adoption_positive` | completion almeno 85%, correct almeno 80% delle valutabili, zero falsità materiali nei risultati accettati, zero `length`, utilità mediana almeno 4/5, latenza p50 ≤30 s e p95 ≤90 s, diagnostica e heartbeat efficaci |
| `field_adoption_mixed` | gate di sicurezza verdi, ma almeno una soglia di qualità, utilità, latenza, diagnostica o heartbeat non è raggiunta |
| `field_adoption_negative` | completion sotto 75%, correct sotto 60%, utilità mediana sotto 3/5 oppure problemi ripetuti che impediscono l'uso quotidiano |
| `field_adoption_incident` | mutazione, leak, identità errata o attraversamento del confine read-only |

Un terminale `length` impedisce `field_adoption_positive` e apre una decisione
separata sul profilo generativo; non autorizza aumenti seriali di
`num_predict`. I risultati di M25 non cambiano retroattivamente la validità
della release v0.3.1.

## Output attesi

- matrice compilata in `milestone-25-v0.3.1-direct-chat-field-adoption-matrix.yaml`;
- report redatto in `reports/milestone-25-final.md`;
- decisione di adozione separata da qualsiasi futura qualifica CPU moderna;
- backlog di problemi osservati, senza implementazioni dentro M25.

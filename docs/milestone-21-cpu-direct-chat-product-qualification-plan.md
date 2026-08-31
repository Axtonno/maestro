# Milestone 21 — CPU Direct Chat Product Qualification

Versione candidata: non assegnata

Stato: **COMPLETATA — `cpu_profile_candidate_rejected`**

Data: 2026-08-30

Documenti di riferimento:

- `milestone-20-thinkpad-latency-attribution-lower-resource-profile-plan.md`;
- `reports/milestone-20-final.md`;
- `reports/milestone-20-phase-b.md`;
- `milestone-17-direct-chat-development-plan.md`;
- `reports/milestone-17-phase-6.md`;
- `reports/milestone-17-final.md`;
- `configuration.md`;
- `cli.md`;
- `packaging-candidate.md`;
- `milestone-21-cpu-direct-chat-qualification-matrix.yaml`;
- `reports/milestone-21-phase-1.md`;
- `reports/milestone-21-phase-2.md`;
- `reports/milestone-21-phase-3.md`;
- `reports/milestone-21-phase-4.md`;
- `reports/milestone-21-phase-5.md`;
- `reports/milestone-21-phase-6.md`;
- `reports/milestone-21-final.md`.

## Decisione di apertura

M20 dimostra che Maestro può plausibilmente offrire Direct Chat utile su CPU:
`qwen2.5-coder:7b` completa 5/5 task single-file, conserva il workspace e
riduce del 44,3% la mediana rispetto al profilo qwen3.5 appaiato. Non è ancora
una promessa distribuibile.

Questa milestone decide se l'esatto profilo può diventare una superficie di
prodotto sul ThinkPad, senza abbassare i criteri e senza reinterpretare i
failure precedenti.

## Support claim candidato

Un eventuale PASS può qualificare soltanto:

```text
Maestro Direct Chat
ThinkPad T490s / Ubuntu 24.04 / linux-amd64 / CPU-only
Ollama 0.33.1
qwen2.5-coder:7b
digest dae161e27b0e90dd1856c8bb3209201fd6736d8eb66298e75ed87571486f4364
zero o un file esplicito
streaming atomico e non-streaming
read-only, zero tool/retrieval/fallback
```

La versione di release resta non assegnata finché la matrice e l'artifact non
sono qualificati. Un PASS non pubblica automaticamente tag o release.

## Confini non negoziabili

Restano esclusi:

- verified agent e Agent Runtime;
- tool calling, retrieval, indexing e multi-file;
- write, patch, approval e Controlled Mutation;
- prompt tuning o cambio di oracolo durante una serie;
- fallback a qwen3.5, altro modello, provider o hardware;
- download o aggiornamenti impliciti;
- selezione post-hoc di run, task o risposte;
- qualunque claim generico per CPU diverse dal reference hardware osservato.

## Freeze ambientale

### Ollama

La versione unica candidata è **Ollama 0.33.1**, allineata alla piattaforma di
qualifica v0.3.0. Prima del freeze viene installata esplicitamente sul
ThinkPad, quindi vengono registrati:

- canale e origine del pacchetto;
- versione e SHA-256 del binario/package quando disponibile;
- endpoint loopback;
- configurazione di servizio e auto-update disabilitato durante la serie;
- catalogo modello, digest e dimensione;
- stato del modello e memoria prima della prima run.

Il passaggio da 0.32.14 a 0.33.1 è setup della milestone, non evidenza. Tutta la
matrice viene eseguita soltanto dopo il nuovo freeze; risultati misti tra
versioni sono invalidi.

Il freeze della Fase 1 ha registrato revisione Snap 133, hold indefinito,
versione CLI/API 0.33.1 e SHA-256 del payload binario
`9f595107f966433f93f20ee19043f8e0cdea88e7403672f4dba2cadcb45ee085`.

### Modello e profilo

```yaml
interaction:
  chat:
    model: qwen2.5-coder:7b
    timeout: 5m
    streaming: true
    num_ctx: 4096
    num_predict: 512
    thinking: "false"
    residency: 5m
    max_file_bytes: 1048576
    max_output_bytes: 1048576
```

Temperatura 0 resta interna e identica tra complete/stream. `num_predict: 512`
è un hard budget di generazione congelato prima delle serie: deve essere
rappresentato nel contratto, validato e inoltrato a Ollama come
`options.num_predict` per complete e stream. Il limite non cambia dopo una
risposta e non sana output incompleti: finish per limite, troncamento o risposta
che non copre l'oracolo resta failure/`incorrect`.

`residency: 5m` è un requisito candidato nuovo: deve essere rappresentato
esplicitamente nel contratto e inoltrato in modo verificabile a Ollama. Il gate
live non parte finché assenza, invalidità o mancato supporto di budget e
residenza non falliscono closed.

## Riconciliazione qualità M17 + M20

Prima di costruire il candidate vengono congelati prompt esatti, file, digest,
oracolo positivo, claim vietati e rubrica binaria di dieci task:

| Famiglia | ID | Task |
|---|---|---|
| M17 | `Q17-1` | spiegazione classe/funzione |
| M17 | `Q17-2` | route, metodo, controller e action |
| M17 | `Q17-3` | controller e dipendenze senza inferenze esterne |
| M17 | `Q17-4` | proposta di refactoring distinta dai fatti |
| M17 | `Q17-5` | proposta di test distinta dai fatti |
| M20 | `Q20-1` | route pubblica |
| M20 | `Q20-2` | validazione e risposta controller |
| M20 | `Q20-3` | sequenza OrderService |
| M20 | `Q20-4` | rami e fallback CheckoutService |
| M20 | `Q20-5` | ritorno ed effetti visibili OrderRepository |

Se il prompt storico esatto di un task M17 non è recuperabile dalle evidenze
protette, il task viene ricostruito come superset conservativo prima del
freeze e la deviazione è dichiarata. Non è ammesso descriverlo come replay
esatto.

Una risposta è `correct` soltanto se copre l'oracolo e non contiene claim
materiali falsi o non supportati. `Partial` e `acceptable` non vengono usati
per raggiungere la soglia numerica.

## Due serie complete

Il candidate immutabile esegue **due serie integrali**, senza modifiche o
retry selettivi tra esse. Ogni serie contiene:

1. unload esplicito e verifica `not resident`;
2. una completion cold no-file;
3. altre due no-file warm, per gate 3/3;
4. i dieci task Q17/Q20 in ordine precongelato;
5. una ripetizione streaming appaiata del task route;
6. containment e digest workspace;
7. osservazione della residenza fino a eviction;
8. verifica che la prima completion successiva torni cold.

L'ordine delle due famiglie viene controbilanciato tra serie, ma definito prima
della prima risposta. Una failure funzionale o qualitativa non interrompe gli
altri task: servono due serie complete per misurare stabilità. Una mutazione,
leak o violazione di authority applica invece stop immediato.

## Gate cold, warm, memoria ed eviction

Ogni run registra:

- start, primo chunk, terminale e output visibile;
- load, prompt-eval ed eval duration provider-reported;
- input/output token e throughput;
- stato resident prima/dopo;
- RSS e CPU del processo Maestro e del processo Ollama con scope dichiarato;
- RAM fisica, swap activity e memoria disponibile;
- terminale, reason/exit code e digest workspace.

### Cold

`cold` significa modello esplicitamente non residente prima della request. Si
riportano tempo di load, tempo totale e memoria di picco separatamente per
entrambe le serie. Il cold start deve completare entro il timeout e viene
dichiarato nel support claim; questa milestone non nasconde il costo dentro la
mediana warm.

### Warm

Una run è `warm` soltanto quando valgono insieme:

- snapshot provider positivo immediatamente prima della request;
- request iniziata entro il TTL congelato;
- nessuna eviction osservata tra snapshot e terminale;
- `load_duration` non superiore alla soglia di housekeeping congelata.

La soglia non richiede zero: distingue il bookkeeping di alcune centinaia di
millisecondi da un vero caricamento di decine di secondi. Durante il freeze
ambientale vengono eseguite cinque probe no-file warm, senza valutare o usare
il contenuto delle risposte. La soglia è il massimo `load_duration` osservato
più 200 ms, arrotondato ai 100 ms superiori e con hard cap di 2 s. Se una probe
richiede un vero reload, la calibrazione è invalida e ricomincia dopo averne
isolato la causa. Valore calcolato, raw duration e formula vengono registrati
prima di qualunque risposta Q17/Q20. Soltanto le run conformi alimentano
mediana e massimo warm; una run ambigua non viene riclassificata post-hoc e
rende incompleta la serie.

La calibrazione congelata della Fase 1 ha osservato `load_duration` pari a
1.089.707, 1.244.230, 1.251.013, 1.306.399 e 1.163.561 ns. Con la formula
predefinita, la soglia housekeeping M21 è quindi **300 ms**.

### Residency ed eviction

Il candidate richiede TTL 5 minuti. Il modello deve restare residente durante
il TTL, essere osservato non residente dopo TTL più tolleranza congelata e
rendere cold la completion successiva. Maestro non scarica modelli che non
possiede e non lascia timer/goroutine dopo il comando.

La memoria resident e di picco viene dichiarata anche se non esiste una soglia
universale. OOM, reset, swap thrashing, fallback o impossibilità di attribuire
il processo rendono il gate operativo non superato.

## Soglie assolute

Le soglie valgono **separatamente per ciascuna delle due serie**:

```yaml
completion: 100%
quality: ">= 80%"       # almeno 8/10 correct
warm_median: "<= 60s"
warm_maximum: "<= 120s"
timeouts: 0
cold_start: measured_and_declared
```

Inoltre:

- no-file deve essere 3/3 in entrambe le serie;
- complete/stream deve essere semanticamente equivalente 2/2 per serie;
- nessun task Q17/Q20 può essere `incorrect` in entrambe le serie,
  indipendentemente dalla falsità o omissione specifica;
- nessun task può produrre lo stesso claim materiale falso in entrambe le
  serie ed essere comunque promosso;
- output limitato o troncato non conta come completion corretta e resta
  failure/`incorrect`;
- workspace e fixture devono coincidere pre/post;
- heartbeat deve comparire nelle generation oltre 15 secondi, restare redatto
  e fermarsi prima del terminale;
- configurazione e binary identity devono superare i gate C1/C2.

L'evidenza M20 non soddisfa già il massimo warm: Q20-4 ha richiesto 190,6
secondi. M21 parte quindi senza presunzione di PASS.

## Sequenza delle fasi

| Fase | Obiettivo | Gate |
|---:|---|---|
| 1 | freeze Ollama, hardware, modello, task e oracoli | record completo, nessun campo `unknown` critico |
| 2 | contratto residency/cold-warm | TTL 5m inoltrato, unload/eviction deterministici |
| 3 | candidate con correzioni M20 C | suite, race, vet, anti-leak e doppia build |
| 4 | serie live 1 | matrice integrale, nessun tuning |
| 5 | serie live 2 | matrice integrale controbilanciata |
| 6 | artifact qualification | archive immutabile, installazione fuori checkout e matrice minima |
| 7 | audit e decisione | report finale e support claim coerenti |

Ogni modifica a codice, prompt, config, modello, digest, Ollama, task o oracolo
dopo il freeze invalida la fase owner e tutte le serie successive.

## Artifact qualification

Dopo due serie verdi:

- costruire due archive byte-identici da commit pulito;
- registrare archive/binario SHA-256, manifest, versione, status e commit;
- verificare `maestro version --diagnostic` sul file estratto e installato;
- installare in prefix nuovo fuori checkout;
- eseguire la matrice minima precongelata descritta sotto sull'esatto archive;
- non ricostruire sull'hardware di qualifica.

La matrice minima viene scelta ora, prima dei risultati delle serie 1 e 2, e
non può essere ridotta o sostituita:

1. doctor chat completo, 5/5 check `pass`;
2. cold no-file, 1/1 completion entro timeout, costo misurato e dichiarato;
3. warm no-file, 1/1 completion conforme alla definizione warm;
4. `Q17-1`, task classe/funzione storicamente problematico, `correct`;
5. `Q20-4`, task CheckoutService e coda M20, `correct`;
6. `Q20-1` complete più stream, entrambi `correct` e semanticamente
   equivalenti;
7. TTL 5 minuti, permanenza resident, eviction e successiva completion cold;
8. almeno un heartbeat per ogni generation oltre 15 s, redatto e arrestato
   prima del terminale;
9. un caso per ciascuna categoria C1 e binary identity C2 verificata contro
   path e SHA-256 dell'installato;
10. containment negativo, digest fixture/workspace pre/post e zero mutazioni.

Sulle cinque generation warm della matrice (`warm no-file`, `Q17-1`, `Q20-4`
e la coppia `Q20-1`) valgono completion 5/5, qualità 4/4 sui task con oracolo,
mediana <= 60 s, massimo <= 120 s, timeout zero, troncamenti zero e
`num_predict: 512` invariato. Qualunque failure respinge l'artifact; non viene
compensata dai risultati precedenti delle serie live.

La pubblicazione resta un workflow separato e richiede decisione esplicita.

## Verdetti

| Verdetto | Conseguenza |
|---|---|
| `cpu_direct_chat_product_qualified` | profilo CPU candidato a release, entro il claim congelato |
| `cpu_profile_candidate_rejected` | nessuna promessa CPU; preservare causa e matrice completa |
| `environment_blocked` | nessun PASS/FAIL finché identità o ambiente non sono ripristinati |
| `security_gate_failed` | stop fail-closed; nessuna eccezione |

## Definition of Done

La milestone è completa soltanto quando:

- Ollama 0.33.1 e il digest modello sono verificati sull'hardware target;
- task M17+M20 e oracoli sono congelati prima delle risposte;
- due serie complete sono registrate senza selezione post-hoc;
- cold, warm, TTL, memoria ed eviction sono separati e dichiarati;
- tutte le soglie assolute hanno PASS/FAIL esplicito per serie;
- il candidate è installato e verificato da artifact immutabile;
- il report finale emette uno dei quattro verdetti;
- agent, tool, retrieval, multi-file e mutation restano invariati.

## Esito di esecuzione

Le Fasi 1-3 superano i gate deterministici e operativi. Le due serie live sono
state completate integralmente senza tuning e falliscono entrambe: completion
7/10, qualità 4/10, mediana warm 72,864/67,609 secondi e massimo warm
192,581/175,234 secondi. Sei task sono incorrect in entrambe le serie.

La Fase 6 era esplicitamente autorizzata soltanto «dopo due serie verdi».
Poiché il prerequisito è falso, resta `NOT_RUN`: non viene costruito un
artifact post-serie e la clausola artifact della Definition of Done non viene
reinterpretata per promuovere il candidato. La Fase 7 emette
`cpu_profile_candidate_rejected`; nessuna versione, release o promessa CPU è
assegnata.

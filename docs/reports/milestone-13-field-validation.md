# Milestone 13 — Field Validation & Adoption

Data di chiusura: 2026-08-27

Stato: **completata con limitazioni**

Classificazione: `field_validation_completed_with_limitations`

Decisione di adozione: `adoption_no_go_on_reference_profile`

## Sintesi decisionale

La Milestone 13 è chiusa anticipatamente perché la stop rule e il confronto
diagnostico conclusivo hanno prodotto evidenza sufficiente per una decisione
di adozione. Non è necessario né corretto proseguire la selezione seriale di
modelli sul reference hardware attuale.

Il risultato è circoscritto alla combinazione osservata di artifact, profilo,
provider, modelli, hardware e task:

- gli invarianti read-only e l'immutabilità del workspace sono confermati;
- l'affidabilità operativa del profilo di riferimento è insufficiente;
- la qualità dell'analisi Laravel multi-file è insufficiente;
- l'adozione sul profilo di riferimento riceve **NO-GO**;
- Controlled Mutation resta invariata, sperimentale e non supportata;
- non viene prodotto `v0.2.1` né un altro artifact;
- `v0.2.0` resta un artifact storico valido per il perimetro qualificato nella
  Milestone 12, ma non viene promosso come soluzione affidabile per analisi
  Laravel multi-file.

Il verdetto non afferma che Maestro, Ollama o le famiglie di modelli osservate
siano genericamente inadatti. Afferma che il profilo di riferimento non ha
raggiunto la soglia necessaria per un'adozione operativa multi-file.

## Perimetro effettivamente osservato

La matrice ufficiale è chiusa a **5/22 run**, tutte appartenenti a `project-b`.
Le 17 run residue sono classificate `not_run`: non sono PASS, non sono FAIL e
non entrano in alcun denominatore osservato. Non vengono simulate,
reinterpretate, imputate o ricalcolate.

La coorte pianificata di almeno due progetti reali non è stata completata. Il
secondo progetto è comparso soltanto in una diagnosi fuori matrice. Anche il
Gate 0 di pubblicazione remota di `v0.2.0` non è stato completato; la campagna
non qualifica quindi download, checksum e installazione dalla GitHub Release
pubblica. Queste deviazioni limitano la generalizzazione e motivano la
classificazione `completed_with_limitations`, ma non annullano l'evidenza
raccolta sul profilo locale congelato né richiedono run prive di valore
decisionale.

## Evidenze consolidate

| Evidenza | Risultato osservato | Conseguenza |
|---|---|---|
| Batch 1 ufficiale | 5 run; 2/5 completate, 3/5 `provider_failure`; qualità delle completion: 1 `partial`, 1 `incorrect`; workspace invariato 5/5 | profilo ufficiale non adottabile; matrice sospesa |
| Diagnosi timeout | 2/2 completate con timeout più ampio; task semplice `correct`, task multi-file `partial`; workspace invariato | il timeout recupera disponibilità, non qualità multi-file |
| Batch 2 Pilot | 2/2 completate; 0 `correct`, 1 `partial`, 1 `incorrect`; entrambe ferme alla pre-read; workspace invariato | timeout ampliato e ripetizione non superano il gate qualità |
| Progressive choreography con `llama3.1:8b` | finalizzazione prematura respinta, ma 0/2 completion e 0/2 aderenza; workspace invariato | la choreography protegge il protocollo; il modello non segue la progressione |
| Qualificazione diretta `qwen3.5:9b` | Gate A 3/3, B 2/2, C 2/2; smoke Maestro e normalizzazione tool call superati | tool calling e adapter non sono il limite iniziale del candidato |
| Choreography sintetica Qwen | primo stato guidato chiuso, poi discovery non convergente; nessuna finale; workspace invariato | `candidate_rejected`; B01 correttamente non eseguito |
| Diagnosi retrieval Qwen | 5/5 query riprodotte direttamente con conteggi e path identici; status finale errato e mancata convergenza | retrieval sintetico deterministico; rottura nella scelta semantica e nella progressione del modello |
| Diagnosi differenziale direct/chat | direct file-attached `correct`; Maestro completion senza file epistemicamente corretta; Maestro preloaded in timeout; agent loop con read positive ma senza convergenza | il failure non è genericamente “Maestro”; servono modalità e profili distinti |

Le prove diagnostiche restano fuori dal denominatore ufficiale di 22. Sono
usate soltanto per classificare i failure e motivare la stop rule.

Le fonti redatte sono:

- `milestone-13-batch-1.md`;
- `milestone-13-timeout-diagnostic.md`;
- `milestone-13-batch-2-pilot.md`;
- `milestone-13-progressive-choreography-experiment.md`;
- `milestone-13-model-candidate-qwen35-9b.md`;
- `milestone-13-qwen35-retrieval-diagnostic.md`;
- `milestone-13-direct-chat-diagnostic.md`.

## Valutazione per dimensione

| Dimensione | Verdetto | Base della decisione |
|---|---|---|
| Sicurezza read-only | confermata | digest pre/post invariati; nessuna authority mutativa concessa |
| Immutabilità | confermata | nessuna mutazione osservata nelle run ufficiali o diagnostiche |
| Affidabilità operativa | insufficiente | timeout ufficiali, latenza elevata e completion non ripetibile sul profilo iniziale |
| Qualità multi-file | insufficiente | nessuna risposta ufficiale o di pilot soddisfa la rubrica multi-file |
| Choreography | efficace come guardia, non sufficiente per convergenza | blocca finalizzazioni premature ma non rende i modelli osservati aderenti |
| Retrieval sintetico | deterministico nel perimetro osservato | replay esatto delle cinque query Qwen; nessuna generalizzazione ai repository reali |
| Completion circoscritta | capacità presente ma non qualificata | una risposta file-attached corretta; latenza elevata e timeout nel percorso Maestro equivalente |
| Adozione reference profile | **NO-GO** | affidabilità e qualità sotto soglia |
| Controlled Mutation | invariata e non supportata | nessuna evidenza di questa milestone amplia l'autorità |

## Confini causali

L'aumento del timeout è escluso come soluzione: ha reso completabili alcune
richieste, ma non ha corretto omissioni, falsità materiali o finalizzazione
dopo una sola lettura.

La progressive choreography è conservata soltanto come candidate diagnostico
`v0.2.1-dev.m13.1`. Ha dimostrato di poter respingere una finale senza
evidenza, non di produrre un'analisi completa. Non esiste quindi una base per
un artifact `v0.2.1`.

`llama3.1:8b` non sostiene la progressione multi-file osservata.
`qwen3.5:9b` usa correttamente strumenti nativi, continua dopo i risultati e
reagisce a una finalizzazione respinta, ma sul fixture sintetico non converge
semanticamente. La diagnosi osservabile esclude un nondeterminismo di
`workspace.search` per le query riprodotte e non sostiene un ulteriore profilo
Qwen sulla macchina attuale.

Il reference hardware, con 16 GB di RAM e margine operativo ridotto dopo il
caricamento del candidato 9B, non offre una base prudente per continuare una
ricerca seriale di modelli sensibilmente più grandi. Questo è un limite del
profilo osservato, non una soglia hardware universale.

Il confronto direct/chat aggiunge un confine importante. Con lo stesso Qwen,
la domanda single-file e il file esplicitamente allegato, Ollama diretto ha
prodotto la risposta corretta; la completion Maestro senza file ha rifiutato
correttamente di inventare. Il percorso Maestro con file pre-caricato ha però
raggiunto il timeout e il loop completo non ha chiuso lo stato nonostante read
riuscite. Su una sola sequenza non è possibile separare adapter e variabilità
nel caso preloaded, ma è possibile escludere la conclusione generica “Maestro
non funziona”.

## Applicazione della stop rule

La chiusura anticipata è decisionale, non statistica. Le osservazioni
indipendenti convergono sullo stesso limite:

1. il profilo ufficiale non completa in modo affidabile;
2. più tempo non rende affidabile l'analisi multi-file;
3. una guardia più forte impedisce risposte premature ma non induce
   progressione nel reference model;
4. un modello più recente supera i gate elementari e l'adapter, ma fallisce il
   gate sintetico immediatamente precedente a B01;
5. il replay diretto conferma il retrieval del fixture e localizza il failure
   nella scelta e nella convergenza del modello;
6. il confronto direct/chat mostra capacità single-file con evidenza esplicita,
   ma anche latenza/default thinking instabili e un costo agentico non
   necessario per la domanda circoscritta.

Continuare le 17 run residue o provare modelli in serie sullo stesso host non
avrebbe un valore decisionale proporzionato al costo. La milestone è quindi
completa rispetto alla domanda di adozione, pur non essendo completa rispetto
alla matrice numerica originaria.

## Relazione con il verdetto della Milestone 12

Il GO della Milestone 12 e il NO-GO di adozione della Milestone 13 non sono in
contraddizione. La Milestone 12 ha qualificato artifact, packaging, confini
read-only, installazione locale e quick start nel perimetro dichiarato. La
Milestone 13 ha posto una domanda più ampia: se quel profilo fosse affidabile
e sufficientemente accurato su analisi Laravel reali e multi-file. La risposta
osservata è negativa.

Nessuna promessa storica di `v0.2.0` viene riscritta e nessun risultato
mancante viene usato per rafforzare retroattivamente uno dei due verdetti.

## Decisione di roadmap

Il percorso approvato è:

```text
chiusura Milestone 13
    -> Milestone 14: analisi forense e progettazione development-only
    -> Milestone 15: nuova piattaforma e nuovo baseline read-only
    -> qualificazione mutativa, soltanto dopo il baseline
```

La Milestone 14 non prosegue la selezione seriale di modelli sul ThinkPad e
non tenta di compensare i limiti read-only. Conserva Controlled Mutation come
oggetto forense: analisi del failure, contratto model-facing, compilazione
deterministica e invarianti di sicurezza, senza support claim o release.

La Milestone 15 qualifica prima un nuovo baseline read-only sulla piattaforma
più capace. Soltanto dopo può iniziare una qualificazione mutativa separata,
con profilo e gate propri.

Prima di una nuova selezione live viene progettata una modalità `direct/chat`
esplicita, senza tool né fallback implicito nel verified agent. Il lavoro può
procedere in parallelo alla sola analisi forense della Milestone 14, ma la sua
misurazione appartiene al baseline read-only della Milestone 15.

## Requisiti congelati per una futura qualificazione

Prima di qualificare un altro modello devono essere disponibili:

- configurazione e osservabilità esplicite di `num_ctx` effettivo;
- configurazione e osservabilità esplicite di `thinking`;
- binding più forte fra evidenza, query, risultato, stato e decisione;
- separazione esplicita fra `direct/chat` e `verified agent`, con prompt,
  limiti, metriche e terminali distinti;
- contesto single-file esplicitamente fornito nel percorso `direct/chat`, senza
  retrieval o tool impliciti;
- gate sintetici agentici e di convergenza prima di qualsiasi B01;
- candidate record distinto per ogni modifica di modello, digest, template,
  profilo, contesto, thinking, timeout o hardware;
- nuovo baseline read-only sulla piattaforma della Milestone 15 prima dei gate
  di Controlled Mutation.

Questi requisiti non correggono retroattivamente la campagna e non autorizzano
`v0.2.1`.

## Verdetto finale

```text
field_validation_completed_with_limitations
adoption_no_go_on_reference_profile
```

La matrice ufficiale è chiusa a 5/22 per stop rule. Sicurezza e immutabilità
read-only sono confermate; affidabilità operativa e qualità multi-file sono
insufficienti. `v0.2.0` resta storicamente valido nel proprio perimetro, ma
non viene adottato come soluzione affidabile per analisi Laravel multi-file.
Controlled Mutation resta non supportata e nessun nuovo artifact viene
prodotto.

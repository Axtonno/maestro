# Milestone 15 — Reference Hardware & Mutation Qualification Plan

Versione: 0.1.0

Stato: Pianificata — subordinata all'handoff della Milestone 14

Data: 2026-08-21

Documenti di riferimento:

- `roadmap.md`;
- `milestone-14-controlled-mutation-recovery-plan.md`;
- `milestone-16-productization-v0.3.0-plan.md`;
- `mutation-qualification.md`;
- `mutation-qualification-profile.yaml`;
- `reports/milestone-11-final.md`;
- `adr/ADR-0031.md`;
- `adr/ADR-0032.md`;
- `compatibility.md`;
- `security-model.md`;
- `operational-experience.md`.

---

# Obiettivo operativo

Qualificare una combinazione dichiarata di piattaforma, hardware, provider e
modello per Controlled Mutation.

La milestone usa un reference hardware più capace per separare quattro cause
possibili: incompatibilità WSL2, risorse insufficienti, limite del modello e
failure del protocollo mutativo. Parte dalla release read-only pubblica,
stabilisce una baseline GPU osservabile, ripete il solo Gate A diagnostico con
Granite 8B e valuta poi candidati superiori, iniziando dalla classe 14B
compatibile con il profilo.

Un esito positivo qualifica soltanto la tupla esatta osservata e dà GO alla
Milestone 16 per la Productization v0.3.0 mutativa. La milestone non produce una
release, non modifica v0.2.x e non dichiara supporto per Windows nativo.

---

# Profilo candidato iniziale

```text
Windows host
└── WSL2
    └── Ubuntu 24.04 LTS, Linux amd64
        ├── 32 GB RAM host, quota WSL effettiva misurata
        ├── NVIDIA RTX 5070, 12 GB VRAM
        ├── Ollama eseguito dentro WSL2
        ├── Maestro Linux amd64
        └── fixture sotto /home/... su filesystem Linux
```

| Dimensione | Valore candidato | Evidenza da congelare |
|---|---|---|
| Host | Windows | edizione, versione e build esatte |
| Virtualizzazione | WSL2 | versione WSL, kernel e configurazione risorse |
| Distribuzione | Ubuntu 24.04 LTS | release/point release esatta |
| Architettura | Linux `amd64` | `uname` e identità binario |
| RAM | 32 GB fisici | memoria host e `MemTotal` effettiva dentro WSL |
| GPU | NVIDIA RTX 5070 12 GB | modello, driver, VRAM e backend osservati |
| Provider | Ollama in WSL2 | versione, endpoint redatto e processo effettivo |
| Workspace | filesystem Linux sotto `/home` | root logica, tipo filesystem e assenza di `drvfs` |
| Modello baseline | `ibm/granite4.1:8b` | model ID e digest esatti |
| Modello superiore | candidato 14B da selezionare | famiglia, tag, digest, quantizzazione e context |

I valori commerciali non sostituiscono quelli osservati dal guest. Una
macchina con 32 GB può assegnare a WSL2 una quota inferiore; una GPU da 12 GB
può usare parte della VRAM per display, runtime e KV cache. Il report conserva
valori effettivi e non presenta il profilo nominale come già qualificato.

---

# Confine di piattaforma e filesystem

Maestro continua a essere un binario Linux `amd64`. Il percorso candidato è
Windows → WSL2 → Ubuntu, non un porting Windows nativo. WSL2 viene trattato
come profilo di compatibilità distinto dal Linux nativo già osservato.

La fixture e ogni workspace mutativo devono risiedere nel filesystem Linux
della distribuzione, per esempio:

```text
/home/maestro/projects/qualification-fixture
```

Sono esclusi:

```text
/mnt/c/...
/mnt/d/...
```

e in generale mount Windows/`drvfs`, share di rete, filesystem FUSE e root
cross-filesystem. La restrizione conserva il profilo di syscall Linux usato da
temporaneo nella stessa directory, `fsync`, `renameat` atomico e sync della
directory. La presenza delle syscall non viene assunta: suite deterministica,
fault injection e prove fisiche vengono ripetute dentro WSL2.

Ollama deve essere eseguito dentro la stessa distribuzione WSL2. Un Ollama
Windows raggiunto tramite forwarding, un endpoint LAN o un container
costituiscono topologie diverse e richiedono un nuovo candidate record.

---

# Preflight manuale minimo

Prima dei test live l'operatore raccoglie almeno:

```text
nvidia-smi
ollama --version
ollama ps
```

Il preflight registra inoltre versione/kernel WSL2, release Ubuntu, memoria
effettiva, tipo filesystem della fixture, spazio libero, versione Maestro,
checksum dell'artifact e digest dei modelli. Questi comandi osservano
l'ambiente: Maestro non avvia Ollama, non scarica modelli, non cambia la quota
WSL e non gestisce driver o processi GPU.

`ollama ps` e la telemetria GPU devono dimostrare se il modello è residente,
quale quota è sulla GPU e quale parte resta sulla CPU. Un'etichetta modello o
la sola presenza di CUDA non costituiscono evidenza di offload.

---

# Profili distinti

| Profilo | Hardware osservato | Autorità candidata |
|---|---|---|
| Read-only baseline | ThinkPad CPU-only, 16 GB nominali/15 GiB osservati | list/read/search, mutation deny |
| Mutation candidate | WSL2, 32 GB nominali, RTX 5070 12 GB | singola patch opt-in, soltanto dopo qualifica |

Il secondo profilo non sostituisce il primo. Un eventuale requisito superiore
si applica soltanto a Controlled Mutation. La v0.2.x continua a dichiarare
esclusivamente la compatibility matrix già pubblicata.

---

# Regole trasversali

- la baseline read-only usa l'archive v0.2.0 riscaricato dalla GitHub Release,
  non un binario locale o il checkout;
- la qualificazione mutativa usa un nuovo candidate build identificato, mai
  presentato come release;
- il contratto model-facing è quello consegnato dalla Milestone 14: durante la
  selezione dei modelli non vengono cambiati schema, prompt, compiler o criteri;
- l'eventuale protocollo ADR-0031 resta valido se la Milestone 14 conclude
  `protocol_unchanged`;
- host, WSL, kernel, distribuzione, filesystem, driver, GPU, RAM effettiva,
  Ollama, modello, digest, quantizzazione, context, limiti e timeout vengono
  congelati per ciascun candidate record;
- una variazione di uno di questi elementi crea un nuovo candidato e azzera i
  PASS della relativa serie;
- un solo modello chat resta residente durante una serie; unload e load sono
  operazioni esplicite dell'operatore e non di Maestro;
- nessuna prova scarica implicitamente modelli o modifica la configurazione del
  provider;
- fixture, prompt, temperatura, schema, output atteso e stop rule vengono
  congelati prima della prima run;
- Gate A richiede `3/3`, Gate B `2/2`, Gate C `3/3`, consecutivi e fail-fast;
- il PASS diagnostico non viene ereditato dal Gate A formale;
- ogni failure pre-commit lascia la fixture byte-identica; ogni terminale
  post-commit registra digest applicato e contesto stale senza dichiarare
  rollback;
- Gate C richiede TTY reale, preview concreta e `allow once`; non esistono
  auto-approval, grant run-scoped o retry impliciti;
- raw response e arguments diagnostici restano nel sink development-only della
  Milestone 14 e non entrano in report, log normali o artifact;
- report e output pubblicabili non includono prompt, response, diff, contenuti
  fixture, arguments, secret, path fisici o identificatori macchina non
  necessari;
- nessun risultato WSL2 viene generalizzato a Windows nativo, Linux nativo,
  altra GPU, altro driver, altro modello o workspace sotto `/mnt`.

---

# Metriche di risorsa e prestazioni

Le misure distinguono stato cold, caricamento e run warm. Ogni report conserva
il metodo di misura e non sostituisce valori mancanti con stime.

| Area | Misure |
|---|---|
| Host/guest RAM | RAM fisica host, quota WSL, `MemTotal`, disponibile pre/post, picco osservato |
| GPU | VRAM totale, pre-run, picco, post-run, utilizzo e backend |
| Residenza | dimensione modello, quantizzazione, context, quota GPU/CPU dichiarata da Ollama |
| Latenza cold | tempo di load/residenza e prima risposta diagnostica |
| Latenza warm | durata end-to-end, per turno e per gate; mediana e massimo |
| Agent | turni, tool call, input/output token, reason code e terminale |
| Stabilità | OOM, eviction, fallback CPU, timeout, reset provider e variazioni fra run |

Per serie piccole vengono pubblicati campioni, mediana e massimo, non p95 o
score aggregati. Se il diagnostic harness può osservare time-to-first-token o
time-to-first-tool-call senza ampliare i log pubblici, queste misure restano
locali e nel report entrano soltanto gli aggregati redatti.

Il criterio non richiede necessariamente offload GPU al 100%. Un candidato con
offload parziale può essere qualificato soltanto se quota CPU/GPU, latenza,
memoria e stabilità sono dichiarate e tutti i gate restano entro i limiti
congelati. Una singola completion riuscita non dimostra sufficienza hardware.

---

# Selezione dei modelli

Granite 8B è una baseline diagnostica, non il candidato preferito implicito.
Se riproduce `patch_tool_call_invalid`, viene classificato e non sottoposto a
prompt tuning indefinito.

La selezione superiore parte dalla classe 14B e richiede:

- modello instruct o coding adatto al task;
- tool calling nativo oppure structured output coerente con il contratto
  congelato;
- quantizzazione dichiarata compatibile con il profilo di memoria;
- context iniziale bounded e identico fra tentativi;
- model ID, digest e template provider registrati;
- caricamento senza OOM e risorse osservabili;
- un solo modello residente durante la serie;
- nessun cambio del contratto per favorire un candidato.

Numero massimo di candidati, ordine e stop rule vengono congelati prima del
primo test 14B. Per ogni candidato il funnel diagnostico è:

```text
structured edit proposal
    -> tool call minimale
    -> compilazione della patch completa
```

La patch completa è l'output autorevole del compilatore quando la Milestone 14
ha selezionato il protocollo semplificato; non viene nuovamente inventata dal
modello. Le varianti diagnostiche isolano comprensione, serializzazione e
canale nativo, ma non cambiano il contratto che sarà usato nel Gate A formale.

---

# Sequenza delle fasi

| Fase | Titolo | Stato corrente | Dipende da |
|---|---|---|---|
| 1 | Qualificazione Windows/WSL2/GPU | Non avviata | Milestone 14 |
| 2 | Baseline read-only v0.2.x | Non avviata | Fase 1 |
| 3 | Profilo RAM, VRAM e latenza | Non avviata | Fase 2 |
| 4 | Gate A diagnostico Granite 8B | Non avviata | Fase 3 |
| 5 | Selezione candidati superiori | Condizionale | Fase 4 |
| 6 | Gate A formale | Non avviata | Fase 4 o 5 |
| 7 | Gate B read-only | Non avviata | Fase 6 |
| 8 | Gate C Controlled Mutation | Non avviata | Fase 7 |
| 9 | Matrice negativa e di sicurezza | Non avviata | Fase 8 |
| 10 | Decisione hardware–provider–modello | Non avviata | Esito terminale delle Fasi 1–9 |

Le fasi sono sequenziali rispetto all'autorità. La matrice deterministica e i
controlli filesystem WSL2 devono essere verdi prima del Gate C anche se il
report completo della matrice negativa viene chiuso nella Fase 9. Un failure
classificato può concludere anticipatamente la milestone con uno degli esiti
ammessi, ma non consente di saltare al gate successivo.

---

# Fase 1 — Qualificazione Windows/WSL2/GPU

## Obiettivo

Dimostrare che l'ambiente candidato corrisponde alla topologia dichiarata e
può eseguire Maestro/Ollama su filesystem Linux con accesso GPU osservabile.

## Attività

- registrare Windows edition/version/build senza installare un binario Maestro
  nativo;
- registrare versione WSL, kernel, distribuzione, architettura e configurazione
  effettiva di memoria/processori;
- verificare che fixture, directory temporanea e output risiedano sotto
  `/home` sul filesystem Linux e rifiutare `/mnt/*`/`drvfs`;
- verificare driver, visibilità RTX 5070, VRAM e backend tramite telemetria;
- verificare che Ollama giri dentro WSL2 e registrarne versione, endpoint
  redatto e catalogo senza avviarlo o scaricare modelli;
- eseguire suite deterministiche mirate a containment, symlink, temporaneo,
  sync, rename atomico, cleanup e cancellazione dentro WSL2;
- registrare warning o limitazioni WSL senza trasformarli in PASS.

## Gate di uscita

- identità completa della piattaforma e topologia riproducibile;
- filesystem Linux verificato per workspace e temporanei;
- GPU visibile e Ollama raggiungibile dentro WSL2;
- nessun failure di syscall o containment rilevante al percorso mutativo;
- zero claim di supporto.

## Deliverable

- profilo reference hardware candidato;
- `docs/reports/milestone-15-phase-1.md`.

---

# Fase 2 — Baseline read-only v0.2.x

## Obiettivo

Separare problemi di piattaforma/GPU da problemi del protocollo mutativo usando
prima il prodotto ufficiale read-only.

## Attività

- scaricare archive e checksum v0.2.0 dalla GitHub Release pubblica in una
  directory pulita dentro WSL2;
- verificare checksum, estrazione, manifest, `version` e commit incorporato;
- derivare il profilo ufficiale con root sulla fixture sotto `/home` e
  confermare list/read/search più `workspace_mutate: deny`;
- eseguire `doctor` e richiedere 9/9, quindi `models` e `agents`;
- eseguire due quick start read-only consecutivi con una read reale e risposta
  semanticamente corretta;
- eseguire SIGINT e hard limit con terminali attesi e shutdown bounded;
- confrontare digest e stato fisico dell'intero workspace prima/dopo;
- eseguire scansione anti-leak di stdout, stderr ed evidenze redatte.

## Gate di uscita

- installazione pubblica e `doctor` 9/9;
- due quick start consecutivi `completed`;
- SIGINT e hard limit coerenti;
- workspace byte-identico in ogni run;
- nessun leak o accesso mutativo.

Un failure impedisce di attribuire problemi successivi alla mutazione e deve
essere risolto o classificato prima della Fase 3.

## Deliverable

- baseline read-only WSL2/GPU;
- `docs/reports/milestone-15-phase-2.md`.

---

# Fase 3 — Profilo RAM, VRAM e latenza

## Obiettivo

Stabilire capacità e headroom effettivi del reference hardware prima di
selezionare un modello mutativo.

## Attività

- congelare strumenti, intervallo di campionamento e precisione delle misure;
- raccogliere RAM/VRAM idle con nessun modello residente;
- ripetere due run read-only di misura, separate dal gate della Fase 2, e
  distinguere load cold, prima inferenza e run warm del modello supportato;
- rimuovere il modello read-only dalla residenza tramite operazione esplicita
  dell'operatore, quindi caricare Granite 8B e verificarne la residenza con
  `ollama ps` e telemetria GPU;
- eseguire un probe provider-level fixed e senza effetti per misurare load,
  latenza, token e stabilità Granite senza contarlo come Gate A;
- ripetere le misure dopo unload e verificare rilascio/residenza osservata;
- registrare OOM, eviction, fallback CPU e process reset come eventi distinti;
- non dedurre ancora compatibilità 14B dalla sola memoria libera nominale.

## Gate di uscita

- profilo idle/cold/warm completo oppure missing data espliciti;
- residenza e offload Granite dimostrati, non presunti;
- latenza e memoria riconciliabili con le singole run;
- headroom descritto senza dichiararlo requisito minimo.

## Deliverable

- report risorse e prestazioni baseline;
- `docs/reports/milestone-15-phase-3.md`.

---

# Fase 4 — Gate A diagnostico Granite 8B

## Obiettivo

Verificare se il cambio di hardware/topologia modifica il failure già osservato
con lo stesso modello e lo stesso protocollo congelato.

## Attività

- congelare candidate build, Granite model ID/digest, prompt, schema,
  temperatura, context, token limit e timeout;
- verificare una sola istanza del modello residente e acquisire RAM/VRAM;
- eseguire tre conversazioni indipendenti `read -> result -> proposal`;
- compilare la patch esatta quando il protocollo della Milestone 14 lo
  prevede, senza Tool Runtime, preview, approval o effetti;
- richiedere `3/3` consecutivo e fail-fast;
- classificare JSON, schema, tool channel, truncation, timeout e semantica con
  la tassonomia development-only;
- fermare Granite al primo failure; non correggere prompt o schema durante la
  serie.

## Gate di uscita

- `3/3` diagnostico e candidatura possibile di Granite alla Fase 6; oppure
- failure classificato e passaggio alla selezione 14B;
- fixture byte-identica e zero authority in entrambi i casi.

## Deliverable

- report diagnostico Granite sul reference hardware;
- `docs/reports/milestone-15-phase-4.md`.

---

# Fase 5 — Selezione candidati superiori

## Obiettivo

Selezionare un solo modello candidato superiore senza cambiare il contratto
mutativo o acquistare nuovo hardware durante la campagna.

## Attività

- congelare shortlist, ordine, quantizzazioni, context, timeout e stop rule,
  iniziando da modelli 14B compatibili con i 12 GB VRAM osservati;
- verificare per ogni candidato download già autorizzato, digest, load,
  residenza, quota GPU/CPU e assenza di OOM;
- mantenere un solo modello residente e pulire esplicitamente lo stato fra
  candidati;
- eseguire nell'ordine structured edit proposal, tool call minimale e
  compilazione della patch completa;
- richiedere tre campioni per variante senza effetti e classificare ogni
  failure;
- confrontare validità, esattezza, RAM, VRAM, latenza e stabilità senza score
  globale;
- selezionare al massimo un candidato per ciascuna iterazione del Gate A
  formale;
- non provare modelli maggiori finché la classe 14B non ha un esito conclusivo.

## Gate di uscita

- un candidato supera `3/3` diagnostico sul protocollo congelato ed entra in
  Fase 6; oppure
- tutti i candidati congelati hanno esito classificato e la milestone termina
  `model_rejected`, `hardware_insufficient` o `mutation_deferred`;
- nessun PASS viene trasferito al Gate A formale.

## Deliverable

- matrice comparativa dei candidati;
- record immutabile del modello selezionato;
- `docs/reports/milestone-15-phase-5.md`.

La fase è `not_run` se Granite supera il Gate A diagnostico ed è scelto come
candidato formale; non viene rappresentata come PASS.

---

# Fase 6 — Gate A formale

## Obiettivo

Qualificare il protocollo diretto del candidato senza effetti e senza ereditare
campioni diagnostici.

## Attività

- creare e validare il profilo ufficiale con identità completa di piattaforma,
  hardware, provider, modello e protocollo;
- congelare checkout pulito, commit, digest binario, fixture, prompt, schema,
  parametri, limiti e report contract;
- rieseguire matrice deterministica e controlli WSL2 richiesti prima dei gate;
- eseguire tre conversazioni nuove e indipendenti;
- richiedere read nativa esatta, consumo del risultato autorevole e proposta
  compilabile nella sola patch attesa;
- non invocare Tool Runtime e verificare fixture byte-identica;
- applicare `3/3` consecutivo e fail-fast.

## Gate di uscita

- Gate A formale `3/3` sul candidate record esatto;
- zero approval, zero effetti e fixture invariata;
- telemetria e report redatti completi;
- al primo failure il modello è respinto e Gate B/C non iniziano.

## Deliverable

- report Gate A formale JSON e Markdown;
- `docs/reports/milestone-15-phase-6.md`.

Se il candidato viene respinto e la shortlist congelata contiene ancora un
candidato eleggibile, la sequenza torna alla Fase 5 con un nuovo candidate
record e riparte da Gate A. Nessun PASS diagnostico o formale del modello
precedente viene riutilizzato. Numero massimo e ordine delle iterazioni sono
quelli congelati prima della campagna; esaurita la shortlist, la Fase 10
registra l'esito terminale.

---

# Fase 7 — Gate B read-only

## Obiettivo

Confermare che il candidate build e il modello conservino il percorso
reference agent read-only prima di concedere authority mutativa.

## Attività

- materializzare una fixture privata nuova per ogni tentativo sotto `/home`;
- eseguire due run consecutive con configurazione, streaming, context e limiti
  congelati;
- richiedere una read reale e una risposta semanticamente verificabile;
- rifiutare tool non dichiarati, pseudo-call e richieste mutative;
- registrare terminale, durata, token, tool call, RAM e VRAM;
- confrontare workspace e cleanup prima/dopo;
- applicare `2/2` consecutivo e fail-fast.

## Gate di uscita

- Gate B `2/2` con almeno una read reale per run;
- terminale `completed`, risposta corretta e workspace byte-identico;
- nessun leak o variazione del profilo;
- al primo failure Gate C non inizia.

## Deliverable

- report Gate B JSON e Markdown;
- `docs/reports/milestone-15-phase-7.md`.

---

# Fase 8 — Gate C Controlled Mutation

## Obiettivo

Dimostrare tre volte consecutive il vertical slice completo sulla tupla WSL2,
GPU, Ollama e modello candidata.

## Attività

- materializzare una fixture nuova e verificata per ogni tentativo sotto
  `/home`;
- eseguire il reference agent con una sola patch candidata su un file PHP
  esistente sotto `app/`;
- verificare read autorevole, edit proposal, compilazione deterministica,
  preview concreta e fingerprint esatto;
- richiedere `allow once` a un operatore su TTY reale;
- osservare un solo Execute, commit tramite temporaneo/sync/rename, reindex,
  generazione maggiore, bundle fresh e risposta finale;
- confrontare digest finale, assenza di cambi estranei e cleanup;
- registrare RAM, VRAM, latenza e residenza senza alterare la serie;
- applicare `3/3` consecutivo e fail-fast.

## Gate di uscita

- Gate C `3/3` sul candidate record esatto;
- ogni run consuma una nuova approval exact-fingerprint;
- una sola patch, digest finale esatto e nessun file estraneo;
- reindex e contesto fresh precedono ogni final;
- nessun retry implicito, leak o fallback di authority.

## Deliverable

- report Gate C JSON e Markdown;
- registro redatto delle approval;
- `docs/reports/milestone-15-phase-8.md`.

---

# Fase 9 — Matrice negativa e di sicurezza

## Obiettivo

Dimostrare che il profilo WSL2/GPU fallisce chiuso negli scenari negativi e
conserva le semantiche fisiche richieste.

## Attività

- riconciliare la matrice deterministica già eseguita prima del Gate C;
- eseguire i casi live pertinenti senza contarli come sostituti dei gate
  positivi;
- coprire deny, EOF, no-TTY, input invalido, digest stale, modifica dopo read,
  modifica dopo preview, path negato, symlink, patch assente/ambigua/no-op,
  cancellazione, timeout, fault pre-rename, refresh failure, replay e secondo
  tentativo mutativo;
- verificare esplicitamente rifiuto di workspace `/mnt/*` e mount `drvfs`;
- verificare byte-identità pre-commit e stato applicato/stale post-commit;
- verificare atomicità osservabile, sync senza errori, cleanup dei temporanei e
  assenza di cambi cross-filesystem;
- eseguire anti-leak di report, log normali e artifact read-only;
- ripetere suite completa, race detector e vet dentro WSL2.

## Gate di uscita

- matrice deterministica e live completa con zero failure non classificati;
- containment, approval, atomicità, freshness e cleanup conformi;
- `/mnt/*`/`drvfs` rifiutati dal profilo di qualificazione;
- ogni stato fisico coincide con quello ammesso dallo scenario;
- zero leak e profilo read-only invariato.

## Deliverable

- report matrice negativa e sicurezza;
- `docs/reports/milestone-15-phase-9.md`.

---

# Fase 10 — Decisione hardware–provider–modello

## Obiettivo

Produrre un verdetto unico sulla tupla qualificata e stabilire se esiste un GO
verso la productization mutativa.

## Attività

- riconciliare identità piattaforma, profilo risorse, Gate A/B/C, matrice
  negativa e stato fisico delle fixture;
- verificare che report, candidate record e binario descrivano la stessa
  combinazione;
- classificare ogni failure con evidenza positiva e zero casi `unresolved`;
- distinguere requisito osservato da requisito minimo generalizzabile;
- decidere se WSL2 entra nella futura compatibility matrix oppure richiede
  ripetizione finale su Linux nativo con hardware equivalente;
- aggiornare ADR senza modificare retroattivamente v0.2.x;
- in caso di GO, consegnare alla Milestone 16 configurazione opt-in,
  limiti, hardware, modello, topologia e gate da ripetere sull'artifact;
- non creare release, RC o artifact di prodotto durante questa milestone.

## Esiti ammessi

| Esito | Conseguenza |
|---|---|
| `mutation_qualified` | GO alla Milestone 16 sul profilo esatto |
| `model_rejected` | nuovo candidato possibile senza cambiare contratto; nuovo candidate record e Gate A da zero |
| `platform_rejected` | WSL2 non entra nella compatibility matrix mutativa |
| `hardware_insufficient` | memoria/offload/latenza impediscono il gate; requisito da rivalutare |
| `mutation_deferred` | Controlled Mutation resta sperimentale e non supportata |

`model_rejected` richiede un failure attribuibile al modello con piattaforma e
risorse sufficienti. `hardware_insufficient` richiede evidenza di OOM,
residenza impossibile o limiti temporali dovuti al profilo, non una semplice
risposta errata. Un caso non distinguibile produce `mutation_deferred`, non una
classificazione per esclusione.

## Gate per `mutation_qualified`

- profilo Windows/WSL2/Ubuntu/filesystem/GPU/Ollama riproducibile;
- baseline read-only completamente verde;
- risorse e residenza misurate;
- Gate A `3/3`, Gate B `2/2`, Gate C `3/3`, tutti formali e fail-fast;
- matrice negativa e di sicurezza interamente verde;
- fixture esatta in ogni terminale e nessun leak;
- modello, digest, quantizzazione, context e limiti dichiarati;
- nessun cambio di protocollo dopo il freeze.

## Deliverable

- `docs/reports/milestone-15-final.md`;
- ADR conclusivo hardware–provider–modello;
- profilo di compatibility candidato per la Milestone 16 oppure rinvio
  motivato.

---

# Decisione WSL2 prima della release

Un PASS della Milestone 15 dimostra il profilo WSL2 esatto. Prima di una
release mutativa occorre scegliere esplicitamente una delle due strade:

1. dichiarare e productizzare WSL2/Ubuntu/filesystem Linux come piattaforma
   mutativa supportata, ripetendo i gate sull'artifact finale;
2. trattare WSL2 come ambiente di selezione e ripetere Gate A/B/C più matrice
   di sicurezza su Linux nativo con hardware equivalente prima del support
   claim.

Nessuna delle due decisioni abilita Windows nativo o workspace sotto `/mnt`.

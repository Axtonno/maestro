# Milestone 9 — Post-release & Benchmark Closure Development Plan

Versione: 0.1.0

Stato: In corso — Fasi 1–3 completate

Data: 2026-08-15

Documenti di riferimento:

- `roadmap.md`;
- `v0.2.0-development-plan.md`;
- `reports/milestone-8-final.md`;
- `reports/v0.1.0-post-release-observation.md`;
- `benchmark-evaluation-plan.md`;
- `reports/milestone-3-live-ollama-validation.md`;
- `adr/ADR-0029.md`.

---

# Obiettivo operativo

Chiudere l'osservazione post-release della v0.1.0 e il debito benchmark della
Milestone 3 prima di congelare qualsiasi contratto mutativo.

La milestone verifica il prodotto read-only pubblicato, distingue i difetti
del prodotto dai limiti di modello e ambiente, decide senza ambiguità lo stato
di llama.cpp e produce un verdetto GO/NO-GO verso la Milestone 10.

La Milestone 9 non aggiunge capacità mutative. La presenza nel repository di
tool mutativi sperimentali non amplia la compatibility promise della v0.1.x.

---

# Sequenza delle fasi

| Fase | Titolo | Stato corrente | Dipende da |
|---|---|---|---|
| 1 | Contratto di osservazione e baseline | Completata | Milestone 8 |
| 2 | Artifact e preflight fuori dal checkout | Completata | Fase 1 |
| 3 | Workspace reali e resilienza operativa | Completata | Fase 2 |
| 4 | Triage e stabilizzazione v0.1.x | In corso | Fase 3 |
| 5 | Chiusura benchmark e decisione llama.cpp | Pianificata | Fase 4 |
| 6 | Audit finale e gate Controlled Mutation | Pianificata | Fasi 1–5 |

Le fasi sono sequenziali rispetto ai gate. Una fase può raccogliere evidenze
preparatorie per la successiva, ma non viene dichiarata completata finché
artifact, ambiente, comandi, risultati e classificazioni non sono
riproducibili. Ogni fase produce un report sotto `docs/reports/`; la Fase 6
produce anche `docs/reports/milestone-9-final.md` e rende conclusivo
`docs/reports/v0.1.0-post-release-observation.md`.

---

# Regole di esecuzione

- Le prove di prodotto usano l'artifact v0.1.0 estratto in una directory
  pulita, non il binario costruito dal checkout.
- Tag, commit incorporato, checksum e manifest vengono registrati prima di
  ogni campagna.
- Provider, modelli, hardware, sistema operativo, limiti e timeout sono
  espliciti e restano congelati durante una matrice.
- Le prove live non partono implicitamente server, non scaricano modelli e non
  trasformano un prerequisito assente in PASS.
- Ogni workspace viene identificato in forma redatta e ne viene registrato lo
  stato prima e dopo la run; il profilo read-only deve lasciarlo invariato.
- Prompt, risposte complete, secret, root fisiche e contenuti dei workspace
  non entrano nei report.
- Un failure viene classificato come bug v0.1.x, limite del modello, problema
  ambientale, problema UX o richiesta evolutiva prima di aprire modifiche al
  codice.
- Le correzioni v0.1.x restano ristrette al contratto read-only e ripetono i
  gate che hanno qualificato v0.1.0.
- Nessuna fase abilita `workspace.write`, `workspace.patch`, shell, Git,
  sandbox, process isolation, recovery o multi-agent nel profilo supportato.

---

# Fase 1 — Contratto di osservazione e baseline

## Obiettivo

Congelare il protocollo con cui vengono raccolte e classificate le evidenze,
prima di eseguire nuove prove o modificare il prodotto.

## Attività

- inventariare promessa, esclusioni e known issue della v0.1.0;
- verificare identità dell'artifact, tag, commit sorgente e checksum attesi;
- definire profili hardware/provider/modello e prerequisiti delle prove;
- fissare numero di run, deadline, hard limit e criteri di stop;
- scegliere almeno due piccoli workspace Laravel non embedded, privi di dati
  sensibili e adatti a task esclusivamente read-only;
- definire registro delle osservazioni con riproduzione, classe, severità,
  impatto sul contratto e destinazione;
- definire raccolta redatta per exit code, reason code, tool invocati, durata,
  immutabilità e shutdown;
- trasformare la baseline locale già presente nel report post-release in input
  esplicito, senza considerarla una prova live superata.

## Gate di uscita

- matrice di osservazione completa e comandi riproducibili;
- artifact e profili di prova identificati senza fallback;
- criteri PASS, FAIL, SKIP e classificazione definiti prima delle run;
- controlli di immutabilità e anti-leak descritti;
- nessuna modifica al codice applicativo.

## Deliverable

- aggiornamento di `docs/reports/v0.1.0-post-release-observation.md`;
- `docs/reports/milestone-9-phase-1.md`.

---

# Fase 2 — Artifact e preflight fuori dal checkout

## Obiettivo

Verificare che il prodotto pubblicato conservi installabilità, identità e
diagnostica quando viene usato come lo userebbe un nuovo utilizzatore.

## Attività

- estrarre l'archive v0.1.0 in una directory pulita e verificare il checksum;
- verificare assenza di traversal, link pericolosi e file inattesi
  nell'archive;
- eseguire `version`, root help e help dei comandi pubblici;
- configurare il percorso supportato senza leggere file dal checkout;
- eseguire `doctor`, `models` e `agents` con prerequisiti live disponibili;
- eseguire almeno due quick start read-only consecutivi sulla fixture
  pubblicata;
- verificare exit code, output redatto e digest invariato del workspace;
- distinguere un problema di packaging/configurazione da un problema del
  provider o del modello.

## Gate di uscita

- artifact e binario riportano identità coerente con il tag v0.1.0;
- installazione e configurazione non dipendono dal checkout;
- preflight completo superato oppure failure classificati e riproducibili;
- due quick start consecutivi completati senza mutazioni;
- nessun dato sensibile nei log e nei report.

## Deliverable

- `docs/reports/milestone-9-phase-2.md`;
- evidenze integrate nel report post-release.

---

# Fase 3 — Workspace reali e resilienza operativa

## Obiettivo

Osservare il reference agent read-only fuori dalla fixture embedded e
verificare i terminali operativi che proteggono una sessione live.

## Attività

- eseguire task di lettura, ricerca e spiegazione su almeno due workspace
  Laravel distinti dalla fixture embedded;
- ripetere almeno due run consecutive per profilo congelato;
- registrare tool call reali, terminale, reason code, durata e budget senza
  conservare prompt o risposte complete;
- verificare pseudo-call testuali, tool call invalide e terminali errati;
- verificare SIGINT, deadline e almeno un hard limit;
- misurare shutdown bounded e chiusura delle risorse aperte;
- confrontare digest e stato fisico del workspace prima e dopo ogni scenario;
- eseguire scansione anti-leak degli output conservati.

## Gate di uscita

- almeno un task read-only valido per ciascun workspace reale;
- cancellazione, deadline e hard limit producono terminali coerenti;
- ogni run lascia il workspace invariato;
- tutti i failure sono riproducibili o marcati non riproducibili con evidenza
  sufficiente;
- nessun failure viene attribuito al prodotto per esclusione.

## Deliverable

- `docs/reports/milestone-9-phase-3.md`;
- matrice post-release compilata nel report conclusivo.

---

# Fase 4 — Triage e stabilizzazione v0.1.x

## Obiettivo

Trasformare le osservazioni in decisioni verificabili e correggere soltanto
eventuali violazioni del contratto read-only pubblicato.

## Attività

- classificare ogni osservazione con causa, severità e confine impattato;
- separare bug Maestro da limite modello, ambiente, UX ed evoluzione;
- per ogni bug v0.1.x confermato, aggiungere prima una riproduzione
  deterministica quando tecnicamente possibile;
- applicare la correzione minima senza ampliare tool, permission o support
  claim;
- aggiornare changelog, known issues, troubleshooting e compatibility quando
  richiesto;
- ripetere la matrice della Fase 2 e gli scenari pertinenti della Fase 3;
- produrre una patch release soltanto se necessaria, con artifact e gate
  equivalenti a quelli della v0.1.0.

La fase è valida anche quando non emerge alcun bug: in quel caso il report deve
registrare esplicitamente che non sono richieste modifiche al codice né una
patch release.

## Gate di uscita

- zero osservazioni non classificate;
- zero bug read-only bloccanti senza correzione o decisione esplicita;
- eventuali patch non introducono capacità mutative o breaking change;
- regressione deterministica e live verde sul confine modificato;
- support matrix e documentazione coerenti con le evidenze.

## Deliverable

- `docs/reports/milestone-9-phase-4.md`;
- eventuale patch release v0.1.x e relativa nota di release.

---

# Fase 5 — Chiusura benchmark e decisione llama.cpp

## Obiettivo

Rimuovere l'ambiguità residua della Milestone 3 e stabilire cosa, se qualcosa,
può essere dichiarato supportato per llama.cpp.

## Attività

- congelare i report Ollama positivo e negativo come baseline storica della
  Milestone 3;
- inventariare server llama.cpp, modello, modalità, hardware, memoria e
  capability realmente disponibili;
- eseguire un preflight di risorse prima della matrice, evitando la modalità
  router già incompatibile con il profilo hardware osservato;
- se esiste una configurazione compatibile, eseguire la matrice live senza
  cambiare scenari, timeout o criteri dopo il primo risultato;
- produrre il report JSON/Markdown e il report di validazione llama.cpp;
- se il preflight non identifica una configurazione valida, registrare una
  decisione formale di adapter sperimentale/non supportato e chiudere il debito
  senza rappresentare l'assenza della prova come PASS;
- aggiornare `benchmark-evaluation-plan.md`, roadmap, compatibility e known
  issues con una sola conclusione coerente;
- definire soltanto il contratto del futuro profilo benchmark mutativo; la sua
  implementazione appartiene alla Milestone 11.

## Esiti ammessi

1. matrice live superata su un profilo dichiarato e support claim ristretto a
   quel profilo;
2. matrice live fallita con limite documentato e llama.cpp non supportato;
3. preflight incompatibile, nessuna matrice avviata e llama.cpp non supportato.

Uno skip o un server non configurato non chiudono da soli la fase.

## Gate di uscita

- stato llama.cpp non ambiguo e supportato da evidenza locale;
- report Ollama storici congelati e referenziati;
- Milestone 3 dichiarata formalmente completata, con l'esito supportato o non
  supportato e i relativi limiti espliciti;
- nessun support claim dedotto dalla sola presenza dell'adapter;
- confine del benchmark mutativo consegnato alla Milestone 11 senza codice di
  prodotto anticipato.

## Deliverable

- `docs/reports/milestone-9-phase-5.md`;
- eventuali `docs/reports/milestone-3-live-llamacpp-smoke.json`, vista Markdown
  e `docs/reports/milestone-3-live-llamacpp-validation.md`;
- decisione ADR se llama.cpp resta sperimentale o cambia stato;
- aggiornamento conclusivo della Milestone 3.

---

# Fase 6 — Audit finale e gate Controlled Mutation

## Obiettivo

Verificare che osservazioni, eventuali patch, benchmark e documentazione
descrivano lo stesso prodotto e decidere se la Milestone 10 può iniziare.

## Attività

- rendere conclusivo il report post-release;
- verificare che non restino bug read-only bloccanti o osservazioni senza
  destinazione;
- verificare coerenza tra compatibility, security model, known issues,
  troubleshooting, roadmap e contesto di progetto;
- eseguire suite completa, race detector, vet e diff check;
- ripetere i gate live pertinenti sull'ultima v0.1.x effettivamente candidata;
- eseguire audit anti-leak dei report e degli artifact conservati;
- registrare il verdetto GO/NO-GO verso il contratto mutativo;
- in caso di GO, consegnare alla Milestone 10 requisiti, rischi UX e limiti
  osservati senza ancora scrivere l'ADR mutativo.

## Criteri GO

- report post-release conclusivo e tutte le osservazioni classificate;
- nessun bug v0.1.x bloccante aperto;
- artifact read-only installabile e gate operativo verde;
- decisione llama.cpp e stato della Milestone 3 espliciti;
- baseline Ollama storica congelata;
- nessun ampliamento implicito della compatibility promise;
- gate repository-wide verdi.

Qualsiasi requisito mancante produce NO-GO motivato. Un NO-GO non autorizza a
indebolire i gate della Milestone 10 e mantiene la Milestone 9 aperta.

## Deliverable

- `docs/reports/milestone-9-phase-6.md`;
- `docs/reports/milestone-9-final.md`;
- `docs/reports/v0.1.0-post-release-observation.md` conclusivo;
- roadmap, `docs/v0.2.0-development-plan.md` e `MAESTRO_CONTEXT.md` allineati.

---

# Gate repository-wide

Le fasi che modificano codice eseguono almeno:

```text
GOCACHE=/tmp/maestro-go-build go test ./...
GOCACHE=/tmp/maestro-go-build go test -race ./...
GOCACHE=/tmp/maestro-go-build go vet ./...
git diff --check
```

Le fasi soltanto documentali eseguono almeno i test pertinenti ai comandi
documentati e `git diff --check`. Le prove live sono aggiuntive e non possono
essere sostituite dalla suite deterministica.

---

# Definizione di completamento

La Milestone 9 è completata soltanto quando il comportamento osservato della
v0.1.x, l'identità degli artifact, la classificazione dei failure, lo stato di
llama.cpp, la chiusura della Milestone 3 e il verdetto verso la Milestone 10
sono documentati senza contraddizioni.

La sola esecuzione di una campagna live, la sola dichiarazione di llama.cpp
come sperimentale o l'assenza di bug noti non chiudono la milestone.

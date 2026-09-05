# Milestone 34 — Host-Bound Mutation Failure Attribution

Stato: Aperta — `failure_cause_not_yet_attributed`

Data: 2026-09-05

## Stato iniziale

```text
host_bound_target_integrity_qualified
mutation_safety_engine_qualified
positive_generation_completion_rejected
failure_cause_not_yet_attributed
v0.5.0_not_authorized
```

M33 non qualifica Controlled Mutation. Qualifica l'integrità del target
host-bound entro il perimetro verificato: 7/7 target e preview corretti,
19/19 workspace finali corretti, zero effetti illeciti. Le proposte positive
7/10 e gli apply positivi 3/6 restano insufficienti. Questi risultati non
attribuiscono da soli la causa delle tre astensioni errate.

## Obiettivo e prima fase: attribuzione offline

Distinguere istruzioni residue incompatibili con il contratto host-bound da
un limite operativo del profilo `qwen3.5:9b`. Non ripetere le run M33 e non
modificarne prompt, matrice, evidenze o verdetto.

L'audit deve ricostruire e documentare:

1. il prompt effettivo e la sua corrispondenza al digest M33;
2. i payload di D03, H01 e H02, con byte selezionati, coordinate e richieste;
3. schema, messaggi, ruoli, opzioni e serializzazione nell'adapter Ollama;
4. eventuali istruzioni di sistema o template incorporati nel modello, la
   loro provenienza e quanto sia verificabile retroattivamente;
5. ogni istruzione che richieda ricerca/unicità del target, contesto esterno
   o astensione incompatibile con una trasformazione determinabile;
6. coerenza fra richiesta, selezione e risultato atteso, distinguendo difetti
   del payload, della valutazione e del comportamento del modello.

Il report deve separare fatti verificati, inferenze e informazioni non
persistite. L'assenza di evidenze non dimostra né un conflitto né la sua
assenza. Se l'audit non è conclusivo, lo stato resta
`failure_cause_not_yet_attributed`; non autorizza tuning esplorativo.

### Evidenze preliminari all'apertura

In `scripts/m33qualify/main.go`, `systemPrompt` dichiara già che Maestro ha
selezionato un intervallo immutabile e vieta di scegliere file, target o
coordinate. Il messaggio utente è un JSON con `Request`, `SelectedText`,
`StartLine` e `EndLine`. Non include il file completo né i duplicati esterni
alla selezione. D03 e H01 non dimostrano quindi che il modello abbia visto e
rifiutato duplicati. L'audit della catena effettiva resta da completare;
questa osservazione non seleziona ancora un ramo decisionale.

## Decisione a due rami

### A — Istruzioni obsolete o contraddittorie dimostrate

Registrare l'istruzione concreta, la sua provenienza, il conflitto e la
correzione. Solo questa evidenza autorizza un nuovo candidate di qualifica
M34 sullo stesso modello. Il nuovo prompt deve esplicitare che:

- il target è già stato selezionato dall'utente;
- il modello non deve cercarlo;
- i duplicati esterni alla selezione sono irrilevanti;
- deve modificare esclusivamente i byte forniti;
- deve astenersi soltanto quando la trasformazione richiesta non è
  determinabile entro quei byte.

Una richiesta contraddittoria, incompleta o che richieda effetti fuori dalla
selezione non definisce una trasformazione ammissibile dei byte forniti.
L'autorità del modello resta limitata a `new_text` oppure `abstain`.

Congelare nuovo prompt, schema, modello/digest, template, opzioni, codice del
runner, definizioni dei gate, choreography e matrice prima delle generazioni.
Il candidate di qualifica non equivale a un candidate di release v0.5.0.
Scrivere development nuovo e holdout indipendente, mai usato nel prompt
design, senza riuso dei casi M33. Eseguire una sola matrice senza retry,
repair, fallback o tuning post-run. Un esito negativo chiude il candidate;
non apre un ciclo di ulteriori eccezioni al prompt.

### B — Prompt e payload corretti

Se l'audit conclude che il contratto effettivo è corretto e non contiene
istruzioni residue incompatibili, emettere:

```text
qwen3.5_9b_host_bound_mutation_profile_rejected
```

Interrompere il tuning dello stesso profilo per Controlled Mutation e aprire
una selezione di modelli dedicata alla mutazione. Il verdetto è una decisione
operativa sul profilo verificato, non una prova di impossibilità universale
del modello. La separazione prevista è:

```yaml
direct_chat: qwen3.5:9b
controlled_mutation: mutation_specific_model_to_be_selected
```

Non è necessario che una sola LLM soddisfi entrambe le capacità. La selezione
successiva avrà un proprio piano e freeze; non cambia il claim Direct Chat
esistente e non autorizza automaticamente Controlled Mutation.

## Gate dell'eventuale nuovo candidate

| Gate | Soglia |
| --- | ---: |
| Output conformi | 100% |
| Target host-bound conservato | 100% |
| Proposte positive corrette | almeno 90% |
| Holdout apply completati | 100% |
| Astensioni semanticamente necessarie corrette | 100% |
| Approval allow/deny raggiunte | 100% |
| Scritture fuori selezione | 0 |
| Mutazioni errate applicate | 0 |
| Failure con effetti | 0 |

Conservare inoltre preview esatta al 100%, zero scritture stale e zero
mutazioni non approvate. Development e holdout devono superare separatamente
i gate applicabili; non compensare failure holdout con successi development.

I denominatori sono congelati prima delle run. I positivi comprendono tutte
le richieste per cui è attesa una proposta concreta, inclusi deny e stale;
un'astensione errata resta una failure nel denominatore. Gli apply holdout
comprendono tutti i positivi holdout destinati all'applicazione: ciascuno deve
attraversare generazione, preview, TTY allow, commit atomico e verifica
byte-per-byte. Deny e stale hanno gate separati e non sono apply attesi.
Le approval attese non possono essere riclassificate dopo una falsa astensione.

Misurare conformità sui veri output del provider, separando iniezioni
avversarie e reject pre-provider. Richiedere casi non vuoti per ogni gate;
un denominatore nullo non equivale a PASS. Le astensioni necessarie richiedono
la decisione corretta in tutti i casi semantici previsti, non soltanto
assenza di effetti. Il 90% sostituisce prospetticamente l'80% di M33, senza
rivalutarne retroattivamente il verdetto.

## Chiusura e artefatti attesi

- report di attribuzione offline con evidenze e ramo scelto;
- nel ramo A, freeze nuovo, preflight, report live unico e verdetto sui gate;
- nel ramo B, rigetto del profilo e apertura del piano di selezione modelli;
- report finale, ADR, roadmap e contesto aggiornati.

All'apertura nessun ramo è selezionato, nessun nuovo prompt o candidate è
congelato e nessuna generazione è autorizzata prima del gate di attribuzione.
M33 resta chiusa; v0.5.0 resta non autorizzata.

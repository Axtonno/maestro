# Milestone 14 — Interaction Modes & Direct Chat Plan

Versione: 0.1.0

Stato: Pronta — priorità immediata dopo la chiusura della Milestone 13

Data: 2026-08-27

Documenti di riferimento:

- `roadmap.md`;
- `reports/milestone-13-field-validation.md`;
- `reports/milestone-13-direct-chat-diagnostic.md`;
- `compatibility.md`;
- `known-issues.md`;
- `configuration.md`;
- `cli.md`;
- `security-model.md`.

---

# Obiettivo operativo

Separare formalmente due modalità di interazione read-only:

```text
maestro chat   -> contesto esplicito, nessun tool
maestro agent  -> esplorazione verificata, tool read-only e state machine
```

La milestone non costruisce un secondo agente e non rende più permissivo il
reference agent. Introduce una superficie `direct/chat` per domande
circoscritte quando l'utente ha già selezionato il contesto, mantenendo il
percorso agentico separato per discovery e sintesi multi-file.

Il primo modello candidato è `qwen2.5-coder:7b`, già configurato localmente in
Continue. La sua presenza non costituisce evidenza di qualifica: deve superare
gate Maestro propri, senza ereditarne risultati o impressioni d'uso.

---

# Contratto candidato

L'interfaccia iniziale è:

```text
maestro chat \
  --file routes/api.php \
  "Quali endpoint, controller e action sono dichiarati?"
```

Il comando `maestro agent` rappresenta la modalità verificata. L'eventuale
relazione con l'attuale `maestro run` viene decisa in un ADR: alias e
deprecation non devono cambiare implicitamente semantica o autorità.

## `maestro chat`

- riceve la domanda e almeno zero o un file esplicitamente selezionato;
- risolve il path soltanto dentro il workspace configurato;
- applica containment, symlink policy, limiti byte ed encoding prima della
  disclosure al modello;
- costruisce una singola completion senza dichiarare tool;
- non interroga retrieval, index o state machine;
- non effettua fallback al verified agent;
- usa prompt e profilo modello dedicati;
- registra `num_ctx` e `thinking` richiesti ed effettivi;
- espone durata, token, terminale e reason code redatti;
- quando il file manca o non è stato fornito, non inventa contenuto e dichiara
  che la risposta non è determinabile.

## `maestro agent`

- conserva list/read/search, permission runtime, evidence binding e stop rule;
- resta l'unica modalità autorizzata a esplorare il workspace;
- non eredita completion o contesto dal comando chat;
- non degrada silenziosamente a chat quando choreography o retrieval falliscono.

---

# Confine della milestone

## Incluso

- ADR delle modalità e dei nomi CLI;
- configurazione separata dei profili chat e agent;
- supporto provider-neutral per `num_ctx` e `thinking`, con mapping Ollama
  osservabile;
- caricamento bounded di un solo file esplicito;
- completion non-streaming iniziale e streaming soltanto dopo equivalenza
  verificata;
- telemetria redatta di latenza, token, context e thinking;
- gate deterministici, negativi e live sul computer attuale;
- candidate record read-only da consegnare alla Milestone 15.

## Escluso

- retrieval implicito, tool calling o state machine in chat;
- multi-file automatico, glob, directory o selezione autonoma di contesto;
- session memory persistente;
- mutazioni, approval, shell, Git, processi, Docker o Composer;
- nuovi modelli scaricati in serie senza shortlist e stop rule;
- support claim, tag o release v0.3.0 prima della Milestone 15.

---

# Configurazione candidata

I profili devono essere distinti e completi, per esempio:

```yaml
interaction:
  chat:
    model: qwen2.5-coder:7b
    num_ctx: 4096
    thinking: false
    timeout: 5m
  agent:
    model: qwen3.5:9b
    num_ctx: 8192
    thinking: default
    timeout: 10m
```

La sintassi è illustrativa fino all'ADR e non modifica retroattivamente lo
schema v0.2.0. `thinking` deve distinguere almeno `default`, `true` e `false`
quando il provider li supporta; un provider che non può onorare il valore
richiesto fallisce il preflight invece di ignorarlo.

`num_ctx` richiesto, context effettivo del runner e truncation devono essere
osservabili. Una differenza non dichiarata invalida il candidate record.

---

# Gate deterministici e negativi

Prima delle prove live devono essere verdi almeno:

- parsing strict della configurazione per profili distinti;
- rifiuto di path assoluto, traversal, symlink evasivo e file fuori workspace;
- rifiuto di directory, file troppo grande, encoding invalido e contenuto
  cambiato durante la lettura;
- zero tool nel request provider di chat;
- zero accesso a retrieval, index e state machine;
- nessun fallback agentico dopo timeout o risposta vuota;
- redazione di path fisici, contenuto, prompt e response dai log operativi;
- cancellazione, provider timeout e hard limit fail-closed;
- workspace byte-identico in ogni terminale.

---

# Qualificazione live sul computer attuale

Ogni candidate record congela modello, digest, template, provider, hardware,
prompt, file, domanda, `num_ctx`, `thinking`, timeout e limiti. La prima serie
usa `qwen2.5-coder:7b`; un failure non autorizza prompt tuning durante la
serie.

| Gate | Prova | Criterio |
|---|---|---|
| C0 | chat senza file | 3/3 risposte epistemicamente corrette; zero contenuto inventato |
| C1 | stesso file e stessa domanda | 3/3 `correct`; zero falsità materiale |
| C2 | streaming/non-streaming | contenuto e terminale semanticamente equivalenti 2/2 |
| C3 | operatività | timeout, latenza, token, context e thinking entro limiti congelati |
| C4 | sicurezza | workspace invariato, zero tool/retrieval/fallback e zero leak |

Continue può essere usato come confronto qualitativo separato soltanto con lo
stesso modello, file e domanda. Non sostituisce nessun gate Maestro e il suo
template deve essere dichiarato quando osservabile.

La milestone si arresta dopo il primo gate fallito. Un secondo modello può
essere valutato soltanto se la shortlist e l'ordine erano congelati prima di
C0; non viene aperta una ricerca seriale indefinita.

---

# Sequenza delle fasi

| Fase | Titolo | Stato corrente | Dipende da |
|---|---|---|---|
| 1 | ADR interaction modes e CLI | Non avviata | Milestone 13 |
| 2 | Profili `num_ctx`/`thinking` osservabili | Non avviata | Fase 1 |
| 3 | Implementazione `maestro chat` single-file | Non avviata | Fase 2 |
| 4 | Matrice deterministica, negativa e anti-leak | Non avviata | Fase 3 |
| 5 | Qualificazione live sul computer attuale | Non avviata | Fase 4 |
| 6 | Audit e handoff read-only | Non avviata | Fase 5 |

---

# Esiti ammessi

| Esito | Significato |
|---|---|
| `direct_chat_candidate` | modalità e profilo superano C0–C4 e possono essere ripetuti sulla nuova piattaforma |
| `direct_chat_model_rejected` | superficie valida, modello/profilo non supera i gate live |
| `direct_chat_contract_rejected` | il contratto non conserva confini, osservabilità o comportamento epistemico |
| `direct_chat_deferred` | nessun candidato utilizzabile entro shortlist e hardware correnti |

Nessun esito produce una release o riapre la matrice della Milestone 13.

---

# Deliverable

- ADR delle modalità `chat` e `agent`;
- comando e configurazione candidate development-only;
- matrice deterministica e negativa;
- report C0–C4 redatto;
- `docs/reports/milestone-14-final.md`;
- candidate record o rinvio motivato per la Milestone 15.

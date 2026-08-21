# Maestro Mutation Qualification

Versione: 0.1.0

Stato: Contratto di esecuzione congelato per la Milestone 11

Ultimo aggiornamento: 2026-08-21

---

# Scopo

Questo documento rende eseguibile
`mutation-qualification-profile.yaml`. Definisce la fixture, i limiti, le
evidenze e le stop rule usate per qualificare il solo vertical slice:

```text
read -> prepare patch -> preview -> approval -> apply -> reindex -> final
```

Il profilo resta candidato e non supportato fino al report finale della
Milestone 11.

---

# Candidato congelato

| Campo | Valore |
|---|---|
| Piattaforma | Linux `amd64` |
| Provider | Ollama |
| Modello | `ibm/granite4.1:8b` |
| CPU lower bound | Intel Core i5-8365U, 8 CPU logiche |
| Memoria lower bound | 15 GiB RAM, 4 GiB swap |
| Provider timeout | 5 minuti |
| Deadline per run | 10 minuti |
| Cleanup timeout | 30 secondi |
| Temperatura Gate A | 0 |
| Token Gate A per turno | 256 |
| Approval Gate C | TTY reale, `allow once` |
| Tentativi mutativi per run | 1 |

Provider, modello, hardware, prompt, fixture, limiti e criteri non possono
essere cambiati durante una serie. Una modifica invalida il candidato e
richiede un nuovo record della Fase 3.

---

# Fixture e modifica ammessa

La sorgente versionata è il dataset embedded
`maestro-laravel-mini@1.0.0`. Ogni tentativo usa una materializzazione privata
nuova; il path `../fixtures/laravel-v1` nel profilo di prodotto descrive il
layout dell'artifact, non una dipendenza dal checkout.

| Campo | Valore |
|---|---|
| Target | `app/Http/Controllers/OrderController.php` |
| SHA-256 iniziale | `4826abe9c6c5d701133817a9dcb565f0b84f760da57e1b518d430b601520b1bd` |
| Testo iniziale | `return response()->json(['data' => $order], 201);` |
| Testo sostitutivo | `return response()->json(['order' => $order], 201);` |
| SHA-256 finale | `509b566bd04a17d567248a721885ac5af0d623f9f505288548c7c302628bac5d` |

Il solo scenario positivo può produrre il digest finale. Le failure
pre-commit devono conservare il digest iniziale. Cancellazione o failure dopo
commit devono registrare il digest finale, `ContextStale=true`, assenza di
testo finale e stato non riuscito.

---

# Gate live

## Gate A — Protocollo diretto

Tre conversazioni indipendenti devono produrre:

1. una sola tool call nativa di read con nome, schema e path esatti;
2. consumo del risultato autorevole fornito dall'harness;
3. una sola tool call nativa `workspace.patch` con path, digest, `old` e
   sostituzione esatti.

Il Tool Runtime non viene invocato e la fixture resta invariata. Il criterio è
`3/3` consecutivo e fail-fast.

## Gate B — Reference agent read-only

Due run indipendenti devono leggere realmente il controller, terminare
`completed`, menzionare `OrderService` e `create`, non richiedere tool mutativi
e lasciare l'intero workspace byte-identico. Il criterio è `2/2` consecutivo e
fail-fast.

## Gate C — Controlled Mutation

Tre run indipendenti devono attraversare il prodotto reale. Un operatore vede
la preview e inserisce `allow once` su TTY; non sono ammessi auto-approval o
grant sintetici. Ogni run deve applicare una sola patch, indicizzare una nuova
generazione, costruire un bundle fresh, terminare `completed` e produrre
soltanto il digest finale ammesso. Il criterio è `3/3` consecutivo e
fail-fast.

Un failure di Gate A impedisce B e C; un failure di B impedisce C.

---

# Matrice fisica

Tutti gli scenari elencati sotto `mutation_matrix` hanno copertura
deterministica. Le prove live aggiuntive sono limitate a deny, cancellazione
pre-commit e refresh failure post-commit, oltre ai Gate A/B/C. Fake, seam e
approver controllati dimostrano classificazione e invarianti, ma non valgono
come PASS live del Gate C.

Per ogni scenario si controllano:

- terminale e reason code;
- ordine redatto del lifecycle;
- digest iniziale, atteso e finale;
- freshness del contesto;
- cleanup dei temporanei;
- numero di tentativi mutativi;
- assenza di retry e di modifiche estranee.

---

# Evidenze e redazione

JSON è la fonte canonica; Markdown è una vista deterministica. I report possono
conservare identità di gate e scenario, commit e digest del binario, versioni,
hardware, durate, contatori, codici terminali, lifecycle e digest della fixture.

Non possono conservare:

- prompt o risposte complete;
- arguments o risultati tool;
- diff o contenuti completi dei file;
- root e path fisici;
- credenziali o variabili sensibili;
- testo inserito dall'operatore oltre alla classificazione redatta della
  decisione.

I report sono pubblicati atomicamente con permessi `0600` e sottoposti a
scansione anti-leak.

---

# Preflight e classificazione

Il preflight verifica senza mutare l'ambiente: identità del candidato,
checkout pulito, OS/architettura, hardware, spazio temporaneo, endpoint Ollama,
versione server, presenza e digest del modello, capability tool calling,
configurazione e fixture.

Le prove non avviano Ollama, non scaricano modelli e non cambiano parametri del
provider. Un prerequisito assente produce `skipped`, che non soddisfa il gate.
Ogni failure è classificato come prodotto, modello, ambiente, operatore o
harness. Una causa non classificata impedisce il verdetto.

---

# Esiti ammessi

La Milestone 11 può concludere soltanto:

1. supporto sul lower bound dichiarato;
2. supporto su un requisito hardware superiore provato ripetendo l'intera
   matrice;
3. rinvio della mutazione senza indebolire gate, limiti o threat boundary.

La productization, il packaging e la release candidate restano responsabilità
della Milestone 12.

# Milestone 17 — Controlled Mutation Qualification Plan

Versione: 0.1.0

Stato: Rinviata — non apribile senza `readonly_baseline_qualified` dalla
Milestone 15, v0.3.0 verde e handoff di protocollo dalla Milestone 16

Data: 2026-08-27

Documenti di riferimento:

- `roadmap.md`;
- `milestone-15-reference-hardware-readonly-baseline-plan.md`;
- `milestone-16-controlled-mutation-recovery-plan.md`;
- `mutation-qualification.md`;
- `mutation-qualification-profile.yaml`;
- `adr/ADR-0031.md`;
- `adr/ADR-0032.md`;
- `security-model.md`.

---

# Condizione di apertura

La milestone resta chiusa finché non sono presenti contemporaneamente:

- baseline `direct/chat` qualificato sulla nuova piattaforma;
- verified agent sintetico interamente verde;
- B01 read-only 2/2 `correct` senza falsità materiali;
- artifact v0.3.0 read-only installabile e immutabile;
- protocollo model-facing e compilatore deterministico consegnati dalla
  Milestone 16 oppure verdetto `protocol_unchanged` motivato;
- piattaforma, hardware, provider, modello, `num_ctx`, `thinking`, filesystem e
  limiti congelati.

Stop rule:

> Un failure del baseline read-only multi-file chiude la qualificazione prima
> di Gate A. Controlled Mutation non viene usata per diagnosticare o compensare
> un modello che non comprende stabilmente il codice in read-only.

---

# Obiettivo operativo

Qualificare una sola combinazione piattaforma–hardware–provider–modello per il
vertical slice Controlled Mutation già limitato a una patch su un file PHP
esistente sotto `app/`.

La milestone non sviluppa interaction modes, non cambia il protocollo dopo il
freeze e non produce una release. Un PASS autorizza soltanto la Milestone 18.

---

# Candidate record

Ogni serie congela:

- artifact/candidate build e commit;
- profilo Windows/WSL2/Ubuntu/filesystem Linux;
- RAM, GPU, driver, VRAM e offload;
- Ollama, modello, digest, template e quantizzazione;
- `num_ctx`, `thinking`, temperatura, timeout e limiti effettivi;
- protocollo model-facing, schema, compiler e prompt;
- fixture, modifica attesa, digest e fingerprint;
- stop rule e criteri A/B/C.

Qualsiasi variazione crea un nuovo record e azzera tutti i PASS. Non vengono
provati modelli in serie oltre una shortlist e un ordine congelati prima del
primo Gate A.

---

# Gate A — Proposta mutativa senza effetti

Tre conversazioni nuove e indipendenti devono:

- leggere il file autorevole;
- consumare il risultato della read;
- produrre una edit proposal valida tramite il canale congelato;
- compilare deterministicamente la sola patch attesa;
- lasciare la fixture byte-identica;
- completare 3/3 consecutivo e fail-fast.

Tool Runtime mutativo, preview e approval non vengono eseguiti. Il primo
failure respinge il candidate record e impedisce Gate B/C.

Deliverable: `docs/reports/milestone-17-gate-a.md`.

---

# Gate B — Conservazione read-only

Due run consecutive devono dimostrare che lo stesso candidate build e modello
conservano il verified agent read-only qualificato:

- almeno una read reale per run;
- completion corretta e terminale `completed`;
- zero pseudo-call o tool non dichiarati;
- workspace byte-identico;
- 2/2 consecutivo e fail-fast.

Gate B non sostituisce B01 della Milestone 15; verifica che l'introduzione del
percorso mutativo non abbia degradato il baseline.

Deliverable: `docs/reports/milestone-17-gate-b.md`.

---

# Gate C — Controlled Mutation

Tre fixture nuove e indipendenti devono completare:

```text
read autorevole
    -> edit proposal
    -> compilazione deterministica
    -> preview concreta
    -> approval allow-once su TTY reale
    -> apply atomico
    -> reindex
    -> contesto fresh
    -> risposta finale
```

Criteri:

- 3/3 consecutivo e fail-fast;
- una sola patch sul solo file ammesso;
- approval exact-fingerprint nuova per ogni run;
- digest finale byte-identico all'atteso;
- nessun file estraneo modificato;
- commit atomico, cleanup, reindex e freshness conformi;
- nessun retry implicito, fallback di authority o leak.

Deliverable: `docs/reports/milestone-17-gate-c.md`.

---

# Matrice negativa e di sicurezza

La matrice conserva almeno:

- deny, EOF e no-TTY;
- JSON/schema invalido e campi sconosciuti;
- path assoluto, traversal, symlink, file fuori `app/` e file non PHP;
- digest stale e modifica dopo read o preview;
- `old_text` assente, ambiguo o no-op;
- cancellazione e timeout prima/dopo commit;
- fault filesystem, sync, rename e cleanup;
- refresh/reindex fallito dopo commit;
- replay approval e secondo tentativo mutativo;
- proposta multi-file o operazione non supportata;
- anti-leak di report, log e artifact.

Ogni failure pre-commit lascia la fixture byte-identica. Ogni failure
post-commit registra stato applicato, durability e contesto stale senza
dichiarare rollback inesistente.

Deliverable: `docs/reports/milestone-17-security.md`.

---

# Sequenza delle fasi

| Fase | Titolo | Stato corrente | Dipende da |
|---|---|---|---|
| 1 | Entry gate e candidate record | Bloccata | M15 + M16 |
| 2 | Gate A | Non avviata | Fase 1 |
| 3 | Gate B | Non avviata | Gate A 3/3 |
| 4 | Gate C | Non avviata | Gate B 2/2 |
| 5 | Matrice negativa e sicurezza | Non avviata | Gate C 3/3 |
| 6 | Decisione hardware–provider–modello | Non avviata | Fasi 1–5 |

---

# Esiti ammessi

| Esito | Conseguenza |
|---|---|
| `mutation_qualified` | GO alla Milestone 18 sul candidate record esatto |
| `model_rejected` | nuovo candidato possibile soltanto entro shortlist; Gate A da zero |
| `platform_rejected` | la topologia osservata non entra nel support mutativo |
| `hardware_insufficient` | risorse impediscono i gate entro limiti congelati |
| `mutation_deferred` | protocollo o causa non qualificabili; mutazione resta non supportata |

`mutation_qualified` richiede Gate A 3/3, B 2/2, C 3/3 e matrice negativa
interamente verdi. Nessun risultato parziale autorizza una release mutativa.

---

# Deliverable

- report A/B/C e sicurezza;
- candidate record finale;
- ADR hardware–provider–modello;
- `docs/reports/milestone-17-final.md`;
- handoff esatto alla Milestone 18 oppure rinvio motivato.

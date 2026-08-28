# Milestone 15 — Reference Hardware & Read-only Baseline Plan

Versione: 0.1.0

Stato: Pianificata — subordinata all'handoff della Milestone 14 e alla nuova
piattaforma disponibile

Data: 2026-08-27

Documenti di riferimento:

- `roadmap.md`;
- `milestone-14-interaction-modes-direct-chat-plan.md`;
- `reports/milestone-13-field-validation.md`;
- `compatibility.md`;
- `security-model.md`;
- `packaging-candidate.md`;
- `release-readiness-audit.md`.

---

# Obiettivo operativo

Qualificare sulla nuova piattaforma un baseline read-only utile e affidabile,
prima di riaprire qualsiasi lavoro mutativo. La milestone verifica nell'ordine:

```text
provider e GPU attivi
    -> direct/chat
    -> verified agent sintetico
    -> verified agent Laravel multi-file
    -> sicurezza e operatività
    -> productization v0.3.0 read-only
```

Stop rule autorevole:

> Se il nuovo baseline non supera la qualificazione read-only multi-file,
> Controlled Mutation Recovery e Mutation Qualification non vengono aperte.

---

# Profilo candidato iniziale

```text
Windows host
└── WSL2
    └── Ubuntu 24.04 LTS, Linux amd64
        ├── 32 GB RAM host, quota WSL effettiva misurata
        ├── NVIDIA RTX 5070, 12 GB VRAM
        ├── Ollama dentro WSL2
        ├── Maestro Linux amd64
        └── workspace sotto /home su filesystem Linux
```

Windows nativo, Ollama Windows raggiunto tramite forwarding, workspace sotto
`/mnt/*`/`drvfs`, share di rete e filesystem FUSE sono profili differenti e
non ereditano PASS.

---

# Regole trasversali

- l'archive storico v0.2.0 viene verificato tramite SHA-256 senza rebuild e
  senza reinterpretare il NO-GO della Milestone 13;
- il candidate v0.3.0 è distinto, development-only fino ai gate di release e
  conserva mutation deny;
- `direct/chat` e `verified agent` hanno profili, prompt, limiti e metriche
  separati;
- `num_ctx` e `thinking` richiesti ed effettivi sono osservabili;
- modello, digest, template, quantizzazione, context, timeout, driver, GPU,
  RAM, filesystem e artifact sono congelati per ogni candidate record;
- una variazione crea un nuovo candidato e azzera i PASS della serie;
- un solo modello chat resta residente per serie;
- Maestro non avvia Ollama, non scarica modelli e non cambia driver o quota
  WSL;
- report pubblicabili non includono prompt, response, sorgenti, arguments,
  path fisici, secret o identificatori macchina non necessari;
- ogni run verifica workspace pre/post e applica anti-leak.

---

# Fase 1 — Piattaforma, provider e GPU

## Gate

- identità Windows/WSL2/Ubuntu/kernel completa;
- RAM guest effettiva e filesystem Linux sotto `/home` verificati;
- GPU, driver, VRAM e backend osservati tramite telemetria;
- Ollama eseguito dentro WSL2 e modello realmente offloaded come dichiarato;
- suite containment, symlink, temporaneo, cancellazione e cleanup verdi;
- nessun OOM, reset provider o fallback CPU non dichiarato.

Deliverable: `docs/reports/milestone-15-phase-1.md`.

---

# Fase 2 — Prestazioni `direct/chat`

La Milestone 14 ha prodotto `direct_chat_deferred`: il suo record congela
profilo, fixture e oracoli, ma non digest/template né PASS C0-C4. Sulla nuova
piattaforma la fase usa quel record come seed e crea un nuovo candidate ID
soltanto dopo aver osservato provider, GPU, modello, digest, template e
capability. Non inventa o ritocca un profilo durante le run.

## Gate

- C0 senza file: 3/3 epistemicamente corretto;
- C1 single-file: 3/3 `correct`;
- streaming/non-streaming equivalenti 2/2;
- zero tool, retrieval, state machine o fallback agentico;
- latenza cold/warm, token, thinking, RAM, VRAM e offload osservati;
- workspace invariato e zero leak.

Deliverable: `docs/reports/milestone-15-phase-2.md`.

---

# Fase 3 — Verified agent sintetico

Il verified agent usa tool read-only, evidence binding e choreography
congelati. Non viene portato direttamente su B01.

## Gate

- tool call elementare 3/3;
- continuazione dopo risultato 2/2;
- correzione dopo finalizzazione respinta 2/2;
- progressione sintetica su tutti gli stati 2/2;
- query, risultati, stato e decisione successiva osservabili e legati;
- zero pseudo-call, loop non convergenti o finali senza evidenza.

Deliverable: `docs/reports/milestone-15-phase-3.md`.

---

# Fase 4 — Laravel multi-file read-only

Solo il candidato che supera la Fase 3 esegue B01 e la matrice read-only
congelata. B01 viene eseguito due volte con configurazione byte-identica.

## Gate

- B01 2/2 `completed` e `correct`;
- tutti i gruppi obbligatori coperti;
- zero falsità materiale;
- latenza e token entro i limiti congelati;
- workspace invariato;
- nessun leak o authority mutativa.

Il primo failure chiude la fase con `readonly_baseline_rejected`. Non vengono
eseguiti Gate A/B/C mutativi e la Milestone 16 resta rinviata.

Deliverable: `docs/reports/milestone-15-phase-4.md`.

---

# Fase 5 — Matrice operativa e di sicurezza

- installazione pulita del candidate;
- `doctor` interamente verde;
- SIGINT, provider timeout e hard limit con terminali attesi;
- path, symlink, containment e disclosure negative;
- suite completa, race detector, vet e anti-leak;
- packaging riproducibile e installazione fuori checkout;
- `workspace_mutate: deny` e assenza di tool mutativi nell'artifact.

Deliverable: `docs/reports/milestone-15-phase-5.md`.

---

# Fase 6 — Productization v0.3.0

Solo un baseline read-only completamente verde può diventare v0.3.0. La
release dichiara:

- `maestro chat` e `maestro agent` come modalità separate;
- modello e profilo qualificati per ciascuna modalità;
- piattaforma, filesystem, hardware e limiti osservati;
- Controlled Mutation non supportata;
- nessun claim oltre la matrice eseguita.

Packaging candidate, RC e artifact finale sono distinti. Installazione,
direct/chat, verified agent, B01, sicurezza e anti-leak vengono ripetuti sugli
artifact. Solo dopo tutti i PASS vengono creati tag, release notes e GitHub
Release v0.3.0.

Deliverable:

- `docs/reports/milestone-15-final.md`;
- `docs/releases/v0.3.0.md`;
- archive/checksum v0.3.0;
- tag annotato e GitHub Release v0.3.0.

---

# Esiti ammessi

| Esito | Conseguenza |
|---|---|
| `readonly_baseline_qualified` | baseline verde; v0.3.0 può essere pubblicata e M16 diventa apribile |
| `direct_chat_rejected` | la modalità chat non è qualificata; verified agent e mutazione non proseguono |
| `verified_agent_rejected` | synthetic o B01 falliscono; Controlled Mutation resta chiusa |
| `platform_rejected` | topologia WSL2 non qualificata |
| `hardware_insufficient` | risorse impediscono il baseline entro limiti congelati |
| `release_deferred` | baseline positiva ma productization v0.3.0 non completa |

Nessun esito diverso da `readonly_baseline_qualified` e release v0.3.0 verde
autorizza l'apertura della Milestone 16.

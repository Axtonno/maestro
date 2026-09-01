# Milestone 22 — Operational Hardening v0.3.1

Versione target: 0.3.1

Stato: Completata — `v0.3.1_operational_hardening_qualified`

Data: 2026-09-01

## Obiettivo

Productizzare le correzioni operative già verificate nelle Milestone 20 e 21
senza ampliare il support claim Direct Chat read-only di v0.3.0.

## Baseline congelata

| Campo | Valore |
|---|---|
| piattaforma | Linux `amd64`; gate WSL2/Ubuntu 24.04/RTX 5070 |
| provider | Ollama 0.33.1 su loopback |
| modello | `qwen3.5:9b` |
| digest | `6488c96fa5faab64bb65cbd30d4289e20e6130ef535a93ef9a49f42eda893ea7` |
| profilo | schema strict v3 chat-only |
| generazione | context 4096, `num_predict` 512, thinking false, temperatura 0 |
| residency | 5 minuti |
| autorità | zero tool/retrieval/agent; `workspace_mutate: deny` |

## Delta ammesso

- diagnostica di configurazione tipizzata e redatta;
- `maestro version --diagnostic` con identità del file eseguito;
- heartbeat redatto durante la generazione;
- `num_predict` e `residency` espliciti e osservabili;
- schema pubblico v3, documentazione, packaging e metadata v0.3.1.

Sono esclusi nuovi modelli, provider, piattaforme, multi-file, agent, retrieval,
tool, mutation e ogni promessa CPU.

## Fasi e gate

| Fase | Ambito | Gate |
|---:|---|---|
| 1 | freeze e audit del delta | PASS |
| 2 | contratto pubblico e packaging | PASS |
| 3 | gate deterministici | PASS |
| 4 | riqualificazione live breve | PASS |
| 5 | audit e decisione | PASS |

La Fase 4 è fail-fast e usa un solo candidate byte-identico. Richiede identità
di binario/config/modello, doctor 5/5, una no-file epistemica, una domanda
single-file in complete e stream, containment traversal, zero mutazioni,
heartbeat allowlisted e nessun leak. Un errore harness permette di ripetere
l'intera serie; una failure del prodotto respinge il candidate.

## Verdetti

| Verdetto | Significato |
|---|---|
| `v0.3.1_operational_hardening_qualified` | sorgente e candidate pronti per release |
| `v0.3.1_candidate_rejected` | nessuna pubblicazione |
| `reference_environment_blocked` | gate live non eseguito, milestone non completa |

La pubblicazione GitHub e il tag sono azioni di release separate: non vengono
inferiti dal solo completamento tecnico della milestone.

Il verdetto e le evidenze sono registrati in
`reports/milestone-22-final.md`.


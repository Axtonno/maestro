# Milestone 10 — Report Fase 1

Data: 2026-08-20

Stato: **COMPLETATA — contratto Controlled Mutation congelato**

## Risultato

La Milestone 10 è stata suddivisa in sei fasi sequenziali e ADR-0031 fissa il
confine prima di qualsiasi modifica al percorso mutativo. Il profilo v0.1.x
resta read-only e nessuna configurazione distribuita abilita mutazioni.

## Fasi congelate

| Fase | Risultato atteso |
|---|---|
| 1 — Contratto | Autorità, limiti, stati e profilo candidato non ambigui |
| 2 — Proposta e preview | Un solo oggetto produce diff, fingerprint ed effetto |
| 3 — Approval e opt-in | TTY, prompt e allow once exact-fingerprint |
| 4 — Commit atomico | Rename atomico, cleanup e fault injection |
| 5 — Freshness e terminali | Reindex obbligatorio e stato post-commit accurato |
| 6 — Audit | Matrice deterministica verde e consegna alla Milestone 11 |

## Contratto congelato

- una sola `workspace.patch` e un solo tentativo mutativo per run;
- file esistente `app/**/*.php`, regolare, UTF-8, senza NUL, massimo 2 MiB;
- read autorevole precedente e una sola occorrenza esatta;
- preview concreta prima dell'approval;
- mutation policy `prompt`, TTY e `allow once`; deny/EOF/no-TTY/input invalido
  falliscono chiusi;
- fingerprint legato a policy, tool/version, run/call, arguments e action;
- commit tramite temporaneo nella stessa directory, sync e rename atomico;
- contesto stale all'inizio dell'effetto e successo finale soltanto dopo nuova
  generazione indicizzata e bundle fresh;
- configurazione mutativa separata; profilo read-only invariato;
- `workspace.write`, processi, Git, rollback e multi-file fuori scope.

## Gap assegnati

| Gap della baseline | Fase |
|---|---|
| Nessun oggetto proposta/preview autorevole | 2 |
| Approval mutativa run-scoped ancora disponibile | 3 |
| Nessun profilo mutativo supportato separato | 3 |
| Riscrittura in-place non atomica | 4 |
| Stato apply/reindex non sufficientemente visibile | 5 |
| Documenti pubblici ancora v0.1.x read-only | 6 |

## Profilo candidato, non supportato

| Campo | Valore |
|---|---|
| Piattaforma | Linux `amd64` |
| Provider | Ollama |
| Modello | `ibm/granite4.1:8b` |
| Lower bound hardware | Intel Core i5-8365U, 8 CPU logiche, 15 GiB RAM, 4 GiB swap |
| Evidenza storica | Gate A 3/3, Gate B 2/2, Gate C deadline |

La Milestone 11 può qualificare questo profilo, dimostrare un requisito
hardware superiore o rinviare la mutazione. Il contratto non trasforma il
profilo candidato in compatibility promise.

## Verifica

Prima dell'avvio documentale:

```text
GOCACHE=/tmp/maestro-go-build go test ./...
```

Esito: **PASS**.

```text
git diff --check
```

Esito: **PASS** sui file tracciati; i nuovi documenti non contengono whitespace
finale.

## Gate

- ADR accettato e perimetro non ambiguo: superato;
- stati prima/dopo commit definiti: superato;
- gap assegnati a fasi verificabili: superato;
- profilo read-only non modificato: superato;
- baseline test verde: superato.

La Fase 1 è completata. La Fase 2 può iniziare con la proposta patch
autorevole e la preview sicura.

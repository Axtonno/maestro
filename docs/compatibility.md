# Maestro v0.3.1 Compatibility Matrix

Data: 2026-08-29

Classificazione hardware aggiornata: 2026-09-01

Questa pagina definisce l’unico support claim di v0.3.1. La presenza di altro
codice nel repository o nel binario non equivale a qualifica.

## Percorso qualificato

| Dimensione | Stato | Confine verificato |
|---|---|---|
| Sistema operativo | Supportato | Linux `amd64`; gate finale su WSL2/Ubuntu 24.04 |
| Hardware della qualifica | Verificato | NVIDIA RTX 5070; non è un requisito minimo universale |
| Provider | Supportato | Ollama 0.33.1 su `http://127.0.0.1:11434` |
| Modello | Supportato | `qwen3.5:9b`, digest `6488c96fa5faab64bb65cbd30d4289e20e6130ef535a93ef9a49f42eda893ea7` |
| Modalità | Supportata | Direct Chat tool-free, zero o un file esplicito |
| Profilo | Supportato | schema v3 chat-only, context 4096, `num_predict` 512, residency 5m, thinking false, temperatura zero |
| Streaming | Supportato | opt-in, equivalente e con pubblicazione atomica |
| Workspace | Supportato | root locale autorizzata; path logico single-file contained |
| Mutazioni | Non supportate | policy `workspace_mutate: deny`; nessun tool nel percorso chat |
| Isolamento | Non fornito | processo trusted in-process, nessuna sandbox |

La prequalifica F6.4 ha superato doctor 5/5, C0 3/3, C1 3/3,
complete/stream 2/2, qualità 4/5, containment, immutabilità e anti-leak. Lo
stesso archive ha poi superato la matrice finale di Fase 7 senza rebuild o
tuning; il verdetto `direct_chat_product_baseline` autorizza questa
compatibility promise per v0.3.0.

## Confine funzionale

Il percorso supportato è:

```text
domanda + file opzionale esplicito
  -> validazione workspace e limiti
  -> una completion o stream provider
  -> risposta validata
```

Non costruisce Context Engine, index, Agent Runtime, sessione, tool registry,
approver o fallback. Senza `--file` non usa contesto di progetto. Con
`--file`, disclosed soltanto quel file.

## Non qualificato

- `maestro agent`, l’alias `maestro run` e ogni reference agent;
- retrieval o indicizzazione multi-file;
- tool calling, tool built-in o di terze parti;
- `workspace.write`, `workspace.patch`, approval e Controlled Mutation;
- modelli o digest diversi dalla baseline;
- llama.cpp, endpoint Ollama remoti o provider esposti a reti non attendibili;
- Linux `arm64`, macOS, Windows e packaging non `linux/amd64`;
- sandbox, remote execution, shell, Git, Docker, persistence e multi-agent.

“Non qualificato” non significa necessariamente incompatibile: significa che
v0.3.1 non offre una promessa operativa per quel percorso.

## Classificazione hardware post-M21

La Milestone 21 ha respinto soltanto l'esatto candidato Direct Chat provato
sul T490s. Non ha dimostrato che Maestro richieda necessariamente una GPU.

| Classe | Stato corrente | Interpretazione |
|---|---|---|
| Legacy CPU — ThinkPad T490s | Development-only | nessuna promessa operativa per `qwen2.5-coder:7b` e il profilo M21 |
| Modern CPU-only | Non qualificata | richiede una nuova matrice con offload GPU disabilitato e verificato |
| GPU reference — RTX 5070 | Supportata da v0.3.1 | vale l'esatto percorso qualificato descritto sopra, non un requisito minimo universale |

Una futura prova CPU moderna deve dimostrare zero layer sulla GPU, zero VRAM
usata dal modello, processo Ollama CPU-only e configurazione congelata. Il
T490s resta un lower bound osservato, non il minimo hardware supportato.

## Relazione con v0.2.0

v0.2.0 resta un artifact storico read-only del reference agent. La Field
Validation successiva ha prodotto `adoption_no_go_on_reference_profile`; la
Milestone 15 ha qualificato Direct Chat ma respinto il verified agent. v0.3.0
non reinterpreta quelle evidenze: productizza soltanto la completion
single-file separata.

CLI e schema sono ancora sperimentali nella serie 0.x. Una configurazione v1
agentica non viene convertita implicitamente nel profilo chat v3. Lo schema v2
resta leggibile per compatibilità ma non riceve valori impliciti di generation
limit o residency.

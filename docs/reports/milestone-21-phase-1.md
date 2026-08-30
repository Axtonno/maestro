# Milestone 21 — Fase 1: freeze ambientale e matrice

Data: 2026-08-30

Stato: **COMPLETATA — ambiente, matrice e soglie congelati**

## Parti congelate

Il contratto candidato incorpora `num_predict: 512`, temperatura 0, context
4096, thinking disabilitato, timeout e residency 5 minuti. La matrice
machine-readable è in
`../milestone-21-cpu-direct-chat-qualification-matrix.yaml`.

Domande, file e oracoli M20 sono replay esatti dei task protetti e conservano
SHA-256 del capture. L'envelope M21 aggiunge i nuovi limiti dichiarati, quindi
non viene presentato come request byte-identica a M20. I prompt qualitativi M17
non furono conservati integralmente: Q17-1…5 sono quindi superset conservativi
congelati prima delle nuove risposte. Questa deviazione è permanente e non
verrà descritta come replay esatto.

La fixture repository e quella estratta dall'artifact v0.3.0 coincidono. Il
digest del manifest relativo ordinato è:

```text
a7831ea9d6cfebf397f004ae0bded6fec59ec935962f8e268b79534fc68abda3
```

Ordini congelati: serie 1 M17→M20, serie 2 M20→M17; lo stream appaiato usa
Q20-1. Oracoli, claim vietati e regole qualitative non sono più modificabili
senza invalidare la Fase 1.

## Emendamenti incorporati

- warm richiede snapshot resident positivo, request entro TTL, zero eviction e
  `load_duration` entro la soglia housekeeping;
- la soglia housekeeping deriva da cinque probe warm non qualitative con
  formula e hard cap predefiniti;
- almeno 8/10 task devono essere correct in ciascuna serie, nessun task può
  essere incorrect due volte e nessuna falsità materiale può ripetersi;
- `num_predict: 512` è un hard budget; length/troncamento resta failure;
- la matrice artifact è fissata a cinque generation warm e gate operativi
  espliciti prima delle serie live.

## Hardware osservato

| Campo | Valore |
|---|---|
| macchina | ThinkPad T490s |
| sistema | Ubuntu 24.04.4 LTS, linux-amd64 |
| kernel | 7.0.0-30-generic |
| CPU | Intel Core i5-8365U, 4 core / 8 thread |
| RAM | 16.403.841.024 byte |
| swap | 4.294.963.200 byte |
| acceleratore | nessuno nel support claim; CPU-only |

## Freeze Ollama e modello

L'upgrade amministrativo è stato completato prima di qualunque risposta
Q17/Q20. Il record congelato è:

| Campo | Valore |
|---|---|
| package | Snap `ollama` |
| versione/revisione | 0.33.1 / 133 |
| tracking | `latest/stable`, revisione esatta installata |
| auto-update | hold `forever` |
| servizio | enabled e active |
| endpoint | `http://127.0.0.1:11434` |
| versione CLI/API | 0.33.1 / 0.33.1 |
| SHA-256 `/snap/ollama/current/bin/ollama` | `9f595107f966433f93f20ee19043f8e0cdea88e7403672f4dba2cadcb45ee085` |
| modello | `qwen2.5-coder:7b` |
| digest catalogo/API | `dae161e27b0e90dd1856c8bb3209201fd6736d8eb66298e75ed87571486f4364` |
| dimensione catalogo/API | 4.683.087.561 byte |
| SHA-256 manifest on-disk | `dae161e27b0e90dd1856c8bb3209201fd6736d8eb66298e75ed87571486f4364` |

## Calibrazione housekeeping warm

Dopo unload e preload espliciti con TTL 5 minuti, cinque probe fisse non
qualitative (`Reply with exactly OK.`) sono partite con snapshot resident
positivo. Il modello è rimasto residente fino al terminale in tutte le probe;
il contenuto delle risposte non è stato ispezionato né usato per modificare il
freeze.

| Probe | `load_duration` |
|---:|---:|
| 1 | 1.089.707 ns |
| 2 | 1.244.230 ns |
| 3 | 1.251.013 ns |
| 4 | 1.306.399 ns |
| 5 | 1.163.561 ns |

Applicando la formula congelata — massimo osservato più 200 ms, arrotondato
ai 100 ms superiori, hard cap 2 s — la soglia M21 è **300 ms**. Il preload
aveva terminale `load`; non è una probe warm e non alimenta la soglia.

## Gate della fase

| Gate | Stato |
|---|---|
| hardware e OS | PASS |
| profilo, task, oracoli e ordini | PASS |
| Ollama 0.33.1 revisione 133 | PASS |
| hold aggiornamenti automatici | PASS — `forever` |
| digest modello | PASS — API e manifest coincidono |
| soglia housekeeping congelata | PASS — 300 ms |

Verdetto di fase: `cpu_qualification_environment_frozen`. La Fase 2 è
autorizzata. Nessuna risposta Q17/Q20 è stata generata durante il freeze.

# Milestone 21 — Fase 1: freeze ambientale e matrice

Data: 2026-08-30

Stato: **IN CORSO — freeze task/oracoli completato, ambiente non congelato**

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

## Blocco ambientale corrente

Lo snap installato è ancora Ollama v0.32.14, revisione 131, SHA-256 del binario
payload `d0758d38ac5882a2c68fd930d0c1220af1952469fa9f30c268746d4021709bf4`.
Lo Snap Store espone v0.33.1 come revisione esatta 133 su `latest/edge`.

Il manifest modello attualmente presente on-disk ha SHA-256
`dae161e27b0e90dd1856c8bb3209201fd6736d8eb66298e75ed87571486f4364`
e il layer modello dichiarato misura 4.683.074.048 byte. Questo verifica la
baseline pre-upgrade, ma non sostituisce la riconferma via catalogo/API dopo
l'installazione della revisione 133.

L'upgrade automatico non è stato eseguito perché snapd richiede autenticazione
`sudo` interattiva. I comandi amministrativi da eseguire sul ThinkPad sono:

```bash
sudo snap refresh ollama --revision=133
sudo snap refresh --hold ollama
```

Dopo l'upgrade occorre registrare `snap info ollama`, versione API, SHA-256 del
payload, stato del servizio, catalogo e digest esatto di
`qwen2.5-coder:7b`. Seguono cinque probe warm per calcolare la soglia
housekeeping senza valutare le risposte qualitative.

## Gate della fase

| Gate | Stato |
|---|---|
| hardware e OS | PASS |
| profilo, task, oracoli e ordini | PASS |
| Ollama 0.33.1 revisione 133 | BLOCKED — privilegio amministrativo richiesto |
| hold aggiornamenti automatici | NOT RUN |
| digest modello | PASS pre-upgrade; riconferma post-upgrade NOT RUN |
| soglia housekeeping congelata | NOT RUN |

La Fase 1 non è chiusa e la Fase 2 non parte finché gli ultimi quattro gate non
sono completati. Nessuna risposta Q17/Q20 è stata generata durante il freeze.

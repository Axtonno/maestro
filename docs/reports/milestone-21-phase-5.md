# Milestone 21 — Fase 5: serie live 2

Data: 2026-08-31

Stato: **COMPLETATA — FAIL**

## Integrità della serie

Un primo avvio della serie 2 è stato interrotto prima di Q20-3 dalla perdita
della sessione del relay diagnostico. L'avvio incompleto è stato invalidato
integralmente: nessuna sua risposta o misura alimenta questo report. La serie
riportata qui è ripartita da unload e snapshot vuoto e ha ripetuto dall'inizio
l'ordine congelato M20→M17, senza retry selettivi.

La serie completa usa esclusivamente il candidate Fase 3:

| Campo | Valore |
|---|---|
| commit | `0a23e0410fe7d4dcd60b76fd489dceb339666dd2` |
| SHA-256 binario | `af09eb5ac53351115c1de707f53d2dd9a2c0d728d82533789ff32a910600e393` |
| SHA-256 config live | `77f5cf33ba0c9d14502c5a33959d875dc0f37745299242e23822beee836dbeca` |
| SHA-256 relay | `02c88b5ca24373e802554c2cf762c5fd9adeb4fdf8f3bdb36e16354cc749b35f` |
| Ollama/modello | 0.33.1 / `qwen2.5-coder:7b` |
| digest modello | `dae161e27b0e90dd1856c8bb3209201fd6736d8eb66298e75ed87571486f4364` |

Doctor è 5/5 prima delle generation. La config live modifica soltanto porta
loopback del relay e root già risolta; modello, prompt, file, ordine,
`num_ctx`, `num_predict`, thinking, TTL, timeout e oracoli restano congelati.

## Cold e no-file

| Run | Stato | Provider terminale | `load_duration` | Qualità |
|---|---|---:|---:|---|
| cold no-file | completed/stop | 44,638 s | 8,349 s | correct |
| warm no-file 1 | completed/stop | 4,449 s | 0,950 ms | correct |
| warm no-file 2 | completed/stop | 4,267 s | 0,888 ms | correct |

No-file è PASS 3/3. Le due run warm iniziano con snapshot resident positivo,
entro TTL e senza eviction; entrambe sono molto sotto la soglia housekeeping
di 300 ms.

## Matrice qualitativa

| Task | Terminale | Durata provider | Qualità | Motivo |
|---|---|---:|---|---|
| Q20-1 | stop | 28,916 s | correct | endpoint, metodo, controller e action esatti |
| Q20-2 | stop | 83,880 s | incorrect | non identifica esplicitamente l'invocazione richiesta `OrderService::create` e trasforma `$order` in un ordine creato |
| Q20-3 | stop | 59,166 s | correct | charge, create, dispatch/evento e return nell'ordine corretto |
| Q20-4 | length, 512 token | 175,234 s | incorrect | output atomico scartato |
| Q20-5 | stop | 67,609 s | correct | array con id 42 e payload, nessuna operazione esterna |
| Q17-1 | length, 512 token | 162,962 s | incorrect | output atomico scartato |
| Q17-2 | stop | 27,293 s | correct | POST `/orders`, controller/action e limiti corretti |
| Q17-3 | stop | 154,578 s | incorrect | presenta `$order` come oggetto/ordine creato e aggiunge semantica di errore non dimostrata |
| Q17-4 | length, 512 token | 162,841 s | incorrect | output atomico scartato |
| Q17-5 | stop | 80,342 s | incorrect | POST presentato come fatto e 422 proposto senza dichiararne il limite osservativo |

Qualità: **4/10 correct**. Completion task: **7/10**. I tre terminali
`length` sono esposti da Ollama ma Maestro scarta correttamente l'output
atomico con `response_invalid`; non sono timeout e non vengono recuperati.

Q20-1 streaming è correct e semanticamente equivalente alla completion. Il
primo chunk arriva in 14,387 s, il terminale in 21,996 s e il finish è `stop`:
equivalenza PASS 2/2.

## Latenza warm

Le 13 generation warm formali — due no-file, dieci task e Q20-1 stream —
hanno:

```text
mediana: 67,609 s
massimo: 175,234 s
load_duration: 0,772–3,364 ms
timeout: 0
```

Tutte le generation sono genuinamente warm secondo i quattro criteri
congelati. La mediana supera 60 s e il massimo supera 120 s.

## Stabilità tra le due serie

Sei task sono incorrect in entrambe le serie, violando il gate indipendente
dal punteggio aggregato:

- Q17-1, Q17-4 e Q20-4 terminano due volte per budget;
- Q17-3 ripete l'inferenza non supportata di oggetto/ordine creato;
- Q17-5 ripete materialmente POST come fatto del controller;
- Q20-2 omette due volte l'esatta chiamata richiesta e aggiunge semantica di
  ordine creato.

Il risultato 4/10 è identico nelle due serie e non dipende dall'ordine delle
famiglie. Le falsità ripetute Q17-3 e Q17-5 violano anche il gate dedicato.

## Residency, risorse e sicurezza

- snapshot post-stream positivo con digest corretto, context 4096,
  `size_vram=0` e scadenza `22:54:21.243172196+02:00`;
- snapshot entro TTL positivo e scadenza invariata;
- snapshot alle `22:54:36+02:00`: modello assente;
- completion successiva cold: 42,074 s, load 5,599 s, correct;
- memoria disponibile minima campionata: 8.309.256 KiB;
- swap free minimo campionato: 4.193.980 KiB, senza thrashing osservato;
- RSS massimo Maestro: 9.332 KiB;
- RSS massimo del servizio Ollama visibile: 58.340 KiB; il runner non è
  attribuibile dal namespace osservato, quindi non viene inventato;
- provider API resident size: 5.062.566.870 byte;
- 63 heartbeat nelle run formali e 2 nella cold post-eviction, tutti redatti e
  allowlisted; zero heartbeat dopo il terminale;
- containment: `file_not_allowed`, exit 2, prima di I/O provider;
- fixture pre/post:
  `a7831ea9d6cfebf397f004ae0bded6fec59ec935962f8e268b79534fc68abda3`;
- capture request/response con permessi `0600`;
- zero mutazioni, tool, retrieval, fallback o timeout.

## Gate

| Gate | Esito |
|---|---|
| no-file 3/3 | PASS |
| completion task 100% | FAIL — 7/10 |
| qualità almeno 8/10 | FAIL — 4/10 |
| nessun task incorrect due volte | FAIL — 6 task |
| nessuna falsità materiale ripetuta | FAIL — Q17-3 e Q17-5 |
| mediana warm <=60 s | FAIL — 67,609 s |
| massimo warm <=120 s | FAIL — 175,234 s |
| timeout zero | PASS |
| complete/stream 2/2 | PASS |
| cold/residency/eviction | PASS |
| containment e immutabilità | PASS |

La serie 2 conferma il FAIL della serie 1. Il prerequisito «due serie verdi»
per la Fase 6 non è soddisfatto; costruire o sottoporre a prova un artifact non
può recuperare il profilo congelato.

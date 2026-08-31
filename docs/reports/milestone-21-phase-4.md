# Milestone 21 — Fase 4: serie live 1

Data: 2026-08-30

Stato: **COMPLETATA — FAIL**

## Identità e protocollo

La serie usa esclusivamente il candidate Fase 3 SHA-256
`af09eb5ac53351115c1de707f53d2dd9a2c0d728d82533789ff32a910600e393`
dal commit `0a23e0410fe7d4dcd60b76fd489dceb339666dd2`. Doctor è 5/5; Ollama è
0.33.1 e il digest modello coincide con il freeze.

L'ordine è quello congelato M17→M20. Il relay diagnostico loopback ha
SHA-256 `02c88b5ca24373e802554c2cf762c5fd9adeb4fdf8f3bdb36e16354cc749b35f`;
le capture sono `0600`. La config live differisce dal profilo congelato
soltanto per porta relay e root già risolta, senza variazioni di modello,
prompt o parametri.

La domanda no-file eredita il freeze storico M14:
`Quali endpoint HTTP sono definiti in questo progetto?`.

## Cold e no-file

| Run | Stato | Provider terminale | `load_duration` | Qualità |
|---|---|---:|---:|---|
| cold no-file | completed/stop | 47,291 s | 7,610 s | correct |
| warm no-file 1 | completed/stop | 4,966 s | 0,897 ms | correct |
| warm no-file 2 | completed/stop | 5,321 s | 1,369 ms | correct |

No-file è PASS 3/3. Le risposte dichiarano assenza di informazioni sul
progetto e non inventano endpoint. Le due warm rispettano snapshot resident,
TTL, zero eviction e soglia housekeeping 300 ms.

## Matrice qualitativa

| Task | Terminale | Durata provider | Qualità | Motivo |
|---|---|---:|---|---|
| Q17-1 | length, 512 token | 182,934 s | incorrect | output atomico scartato |
| Q17-2 | stop | 36,494 s | correct | POST `/orders`, controller/action e assenze corretti |
| Q17-3 | stop | 173,272 s | incorrect | presenta `$order` come oggetto/ordine creato, tipo e semantica non dimostrati |
| Q17-4 | length, 512 token | 173,527 s | incorrect | output atomico scartato |
| Q17-5 | stop | 85,400 s | incorrect | POST presentato come fatto del controller; limite route/metodo omesso |
| Q20-1 | stop | 23,790 s | correct | endpoint, metodo, controller e action esatti |
| Q20-2 | stop | 73,403 s | incorrect | omette la chiamata richiesta `OrderService::create` |
| Q20-3 | stop | 47,059 s | correct | charge, create, dispatch/evento e return corretti |
| Q20-4 | length, 512 token | 192,581 s | incorrect | output atomico scartato |
| Q20-5 | stop | 72,864 s | correct | array id 42 più payload, zero operazioni esterne |

Qualità: **4/10 correct**. Completion task: **7/10**. I terminali `length`
restano failure e non vengono recuperati dal contenuto parziale catturato dal
relay.

Q20-1 streaming completa in 7,925 s, con primo chunk a 0,230 s, gli stessi
quattro fatti della completion e terminale `stop`: equivalenza PASS 2/2.

## Latenza warm

Le 13 generation warm formali — due no-file, dieci task e lo stream appaiato
— hanno:

```text
mediana: 72,864 s
massimo: 192,581 s
load_duration: 0,768–1,677 ms
timeout: 0
```

Tutte soddisfano la definizione resident/TTL/housekeeping. Falliscono però
mediana <=60 s e massimo <=120 s. Tre run terminano per budget prima del
timeout.

## Residency, risorse e sicurezza

- snapshot post-stream: digest corretto, context 4096, `size_vram=0`,
  `expires_at=23:46:34.881813119+02:00`;
- snapshot intermedio: modello resident e scadenza invariata;
- snapshot dopo scadenza: modello assente;
- completion successiva: cold, 44,792 s, load 8,113 s, correct;
- memoria disponibile minima campionata: 7.693.804 KiB;
- swap free campionato: 3.664.864–3.664.944 KiB, nessuna variazione materiale;
- RSS massimo Maestro: 9.472 KiB;
- provider API resident size: 5.062.566.870 byte; RSS del runner non è
  attribuibile dal namespace di processo usato e non viene inventato;
- 68 heartbeat, tutti allowlisted; nessun heartbeat nelle tre run sotto 15 s;
- containment: `file_not_allowed`, exit 2, zero generation;
- fixture pre/post:
  `a7831ea9d6cfebf397f004ae0bded6fec59ec935962f8e268b79534fc68abda3`;
- zero mutazioni, tool, retrieval o fallback.

## Gate

| Gate | Esito |
|---|---|
| no-file 3/3 | PASS |
| completion task 100% | FAIL — 7/10 |
| qualità almeno 8/10 | FAIL — 4/10 |
| mediana warm <=60 s | FAIL — 72,864 s |
| massimo warm <=120 s | FAIL — 192,581 s |
| timeout zero | PASS |
| complete/stream 2/2 | PASS |
| cold/residency/eviction | PASS |
| containment e immutabilità | PASS |

La serie 1 fallisce il profilo candidato. Come congelato, la serie 2 viene
eseguita integralmente senza tuning o retry per stabilire ripetibilità e
fallimenti sistematici.

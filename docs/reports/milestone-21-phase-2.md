# Milestone 21 — Fase 2: contratto residency e cold/warm

Data: 2026-08-30

Stato: **COMPLETATA**

Verdetto: `cpu_chat_residency_contract_ready`

## Obiettivo

Rendere il budget di generazione e la residenza del modello parti esplicite,
strict e verificabili del candidate CPU, senza modificare il contratto v2
qualificato per v0.3.0 e senza introdurre lifecycle implicito in Direct Chat.

## Contratto consegnato

Il nuovo profilo chat-only `version: 3` è definito in
`../../configs/maestro.milestone-21-candidate.yaml` e congela:

```yaml
model: qwen2.5-coder:7b
num_ctx: 4096
num_predict: 512
thinking: "false"
residency: 5m
```

Il loader usa una struttura strict distinta per v2 e v3. In v3
`num_predict` e `residency` sono obbligatori; valori assenti o invalidi
producono una diagnostica tipizzata sul campo logico. Il profilo qualification
richiede Ollama. In v2 gli stessi campi restano sconosciuti, così configurazioni
e capture storici non vengono reinterpretati.

Direct Chat inoltra per complete e stream:

- `num_predict` come `CompletionRequest.Options.MaxTokens`, tradotto
  dall'adapter Ollama in `options.num_predict`;
- `residency` come `CompletionRequest.KeepAlive`, tradotto in
  `keep_alive: "5m0s"`.

Il provider-neutral request contract rifiuta durate negative. L'adapter
llama.cpp rifiuta il keep-alive per-request come capability non supportata
prima di I/O remoto. Il terminale `length`, l'output incompleto e il
troncamento restano failure; il budget non trasforma una risposta parziale in
successo.

L'envelope v3 espone `num_predict_requested` e `residency_requested`; l'output
v2 resta invariato. Doctor identifica dinamicamente lo schema v2 o v3.

Il TTL è posseduto da Ollama tramite la singola request. Maestro non crea timer
o goroutine di residenza e non scarica il modello alla fine di `chat`.
L'unload esplicito già esistente resta una operation provider separata e
invia `keep_alive: 0`.

## Probe operativa non qualitativa

La probe è stata eseguita sull'ambiente congelato in Fase 1. La domanda fissa
`Reply with exactly OK.` non appartiene ai task Q17/Q20; la sua qualità non è
stata usata per cambiare soglie, task o configurazione.

1. unload esplicito: terminale provider `done_reason=unload`;
2. snapshot immediato: `/api/ps` vuoto;
3. richiesta attraverso il binario candidate e il profilo v3;
4. completion in 40,568 s, terminale `stop`, 360 token input e 2 output;
5. heartbeat redatti osservati a 15 e 30 secondi;
6. envelope: `num_predict_requested=512`, `residency_requested=5m0s`;
7. snapshot post-request: modello resident, digest congelato, CPU-only,
   context 4096, `expires_at=2026-08-30T23:02:25.223517554+02:00`;
8. due snapshot entro TTL: modello resident e `expires_at` invariato;
9. snapshot alle 23:03:26 +02:00: `/api/ps` vuoto.

L'osservazione non ha rinnovato il TTL. Il ciclo unload → cold → resident →
eviction è quindi deterministico per la probe. La successiva classificazione
warm userà comunque la definizione combinata e la soglia housekeeping 300 ms
congelate in Fase 1.

## Verifiche

| Verifica | Esito |
|---|---|
| test mirati productconfig/provider/directchat/CLI | PASS |
| `go test ./...` | PASS |
| `go test -race ./...` | PASS |
| `go vet ./...` | PASS |
| doctor candidate contro Ollama locale | PASS 5/5 |
| v2 rifiuta campi v3 | PASS |
| v3 assenza/invalidità fail-closed | PASS |
| complete/stream inoltrano budget e TTL | PASS |
| provider senza supporto rifiutato prima di I/O | PASS |
| probe cold/resident/eviction | PASS |

## Gate e decisione

| Gate | Stato |
|---|---|
| TTL 5m rappresentato e inoltrato | PASS |
| `num_predict: 512` rappresentato e inoltrato | PASS |
| complete/stream equivalenti nel contratto | PASS |
| unload esplicito deterministico | PASS |
| permanenza entro TTL senza rinnovo da osservazione | PASS |
| eviction automatica dopo TTL | PASS |
| zero ownership leak in Direct Chat | PASS |

La Fase 2 è chiusa e autorizza la Fase 3. Non è stato emesso alcun support
claim CPU e non è stata eseguita alcuna risposta qualitativa Q17/Q20.

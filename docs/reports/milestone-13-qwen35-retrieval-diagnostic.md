# Milestone 13 — Diagnosi retrieval osservabile di `qwen3.5:9b`

Data: 2026-08-26

Stato: esperimento development-only completato; nessun valore di
qualificazione

## Scopo e vincoli

Questo esperimento osserva una sola run sul fixture sintetico già usato dalla
Model Candidate Qualification. Non tenta di recuperare il candidato, non
modifica il verdetto `candidate_rejected`, non esegue B01 e non entra nella
Field Validation ufficiale.

La combinazione osservata resta:

```text
qwen3.5:9b + default thinking + ctx 4096 effettivo
+ choreography congelata
= candidate_rejected
```

Choreography, prompt, state machine, tool e codice Maestro non sono stati
modificati. La matrice ufficiale resta sospesa a 5/22.

## Metodo osservabile

Un proxy development-only su loopback ha osservato le richieste e risposte
Ollama senza cambiare il payload. Le evidenze grezze sono conservate soltanto
nello storage locale della Milestone 13, sotto l'ID
`qwen35-retrieval-diagnostic-01`, con directory `0700` e file `0600`.

Per ogni turno il raw locale registra:

- nome e argomenti esatti delle tool call;
- conteggio e path dei risultati di `workspace.search`;
- decisione testuale o dichiarazione successiva del modello;
- campo thinking, `done_reason`, token e durate Ollama;
- telemetria CLI redatta e snapshot pre/post del fixture.

I risultati delle letture non conservano nel raw del proxy il contenuto dei
sorgenti. Query, argomenti, thinking completo e path non sono inclusi in
questo report pubblico. Un manifest SHA-256 locale congela profilo, raw,
replay, stdout/stderr, snapshot e diagnosi redatta.

## Profilo e preflight

| Campo | Valore |
|---|---|
| Modello | `qwen3.5:9b` |
| Digest modello | `6488c96fa5faab64bb65cbd30d4289e20e6130ef535a93ef9a49f42eda893ea7` |
| Thinking | default del modello |
| Contesto effettivo Maestro | 4096 |
| Provider timeout | 10 minuti |
| Limite run | 15 minuti |
| Streaming | abilitato |
| Tool | list/read/search |
| Policy mutativa | `deny` |
| Preflight | 9/9 PASS |

Il digest del fixture pre/post coincide:
`0e607b31520636024af9786cdcad0a6d6ce5d9b96b83c06db758924738fc2805`.

## Risultato della singola run

| Campo | Risultato |
|---|---|
| Exit / terminale | `130` / `deadline_exceeded` |
| Durata | 900.000 ms |
| Turni modello | 12 iniziati, 11 completati |
| Tool call | 9, bootstrap incluso |
| Token completati in/out | 24.135/1.512 |
| Risposta finale | assente |
| Workspace | invariato |

La deadline è quella congelata della run. Non è stata aumentata e la prova non
è stata ripetuta.

## Sequenza redatta dei turni

I conteggi search sono riportati nel turno che ha emesso la call, anche se il
risultato entra nella conversazione al turno successivo.

| Turno | Stato | Decisione | Risultato | Thinking caratteri | Token in/out | Durata proxy ms |
|---:|---|---|---|---:|---:|---:|
| 1 | `controller_action` | search | 3 match | 1.182 | 1.308/287 | 242.203 |
| 2 | `controller_action` | read | successo | 0 | 1.478/35 | 32.468 |
| 3 | `controller_action` | read | successo | 986 | 1.679/257 | 110.549 |
| 4 | `controller_action` | dichiara `covered` | accettato | 878 | 1.957/255 | 120.939 |
| 5 | `referenced_symbols` | search | 2 match | 0 | 2.132/38 | 35.550 |
| 6 | `referenced_symbols` | search | vuoto | 379 | 2.269/114 | 58.087 |
| 7 | `referenced_symbols` | read | successo | 0 | 2.352/28 | 21.436 |
| 8 | `referenced_symbols` | dichiara `covered` | accettato | 0 | 2.543/47 | 41.033 |
| 9 | `events_jobs_services` | search | vuoto | 0 | 2.722/39 | 36.915 |
| 10 | `events_jobs_services` | search | vuoto | 0 | 2.805/39 | 25.911 |
| 11 | `events_jobs_services` | dichiara `covered` | rifiutato | 1.627 | 2.890/373 | 145.507 |
| 12 | `events_jobs_services` | correzione in corso | deadline | non disponibile come risposta completa | — | — |

Tutte le undici risposte Ollama completate hanno `done_reason=stop`. Il turno
11 non è thinking-only: contiene thinking e una dichiarazione JSON visibile.
Il turno 12 è stato interrotto dalla deadline applicativa, quindi non esiste un
`done_reason` Ollama terminale né una risposta thinking completa da
classificare.

## Replay deterministico di `workspace.search`

Ogni query emessa dal modello è stata rieseguita, senza il loop agentico,
tramite la stessa implementazione `workspace.search` e sullo stesso fixture.
Gli argomenti sono identificati qui soltanto dal loro SHA-256.

| Query | Turno/stato | SHA-256 | Loop | Diretto | Conteggi/path | Diagnosi |
|---|---|---|---:|---:|---|---|
| Q1 | 1 / `controller_action` | `75a67677dc480a3526fe9c3e17e97abcfc5895259d4a9acd05d5b80800a5212b` | 3 | 3 | identici | risultato utile riprodotto |
| Q2 | 5 / `referenced_symbols` | `17de655a010a8f74dc0308efccbb0c76afb858f4cdaaea9e301859a0aa66a85f` | 2 | 2 | identici | risultati non pertinenti al simbolo target |
| Q3 | 6 / `referenced_symbols` | `a602a6cb63ec7a52bfc9267065fda47100f13b94774cc302448e954edcd84aeb` | 0 | 0 | identici | query anticipa eventi/job/servizi |
| Q4 | 9 / `events_jobs_services` | `6e9ea304267d362de1567d6742557b5ae3cae756817a65166aa20e161d4a7b6f` | 0 | 0 | identici | vuoto corretto |
| Q5 | 10 / `events_jobs_services` | `c44dd34621e96abbb01f7038ed76e22d7b74d9c468baa3693a42d7f309bdf399` | 0 | 0 | identici | vuoto corretto |

Tutte le cinque riproduzioni coincidono sia nel numero sia nei path. Non è
stato osservato il caso “query corretta, risultato vuoto soltanto nel loop” e
non emerge nondeterminismo del tool sul fixture.

## Classificazione del punto di rottura

La sequenza isola due comportamenti del modello:

1. in `referenced_symbols` la discovery non è coerente con lo stato: una query
   produce risultati non pertinenti e la successiva anticipa già la categoria
   eventi/job/servizi;
2. in `events_jobs_services` due search correttamente vuote soddisfano la
   precondizione per `unavailable`, ma il modello dichiara invece `covered`.

La state machine rifiuta correttamente `covered`, perché nello stato non è
avvenuta alcuna lettura riuscita. Non è stato quindi osservato il caso
“`unavailable` corretto ma rifiutato”: lo status `unavailable` non è stato
emesso.

Diagnosi primaria:
**scelta delle query e mancata convergenza della choreography del modello**,
seguite dalla scelta errata dello status di evidenza. Il retrieval diretto è
riproducibile per tutte le query osservate. La conclusione è limitata al
fixture e alle cinque query; non qualifica genericamente il retrieval su task
Laravel reali.

## Decisione

Il rifiuto originale resta invariato. Non viene aperto un profilo
`thinking:false`: la condizione proposta per concederlo era un retrieval
corretto seguito da un blocco thinking-only, mentre questa run mostra query
non coerenti, discovery non convergente e una dichiarazione di stato errata.

`qwen3.5:9b` viene quindi chiuso per ulteriori esperimenti agentici su questo
hardware e con questa choreography. Nessun B01, nessun nuovo artifact e
nessuna nuova Field Validation.

La configurabilità e l'osservabilità di `num_ctx` e `thinking` restano un
requisito indipendente prima di qualificare un futuro modello, senza valore
retroattivo su questa diagnosi.

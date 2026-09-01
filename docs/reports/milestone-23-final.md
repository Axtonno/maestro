# Milestone 23 — Audit finale e decisione

Data: 2026-09-01

Stato: **COMPLETATA — candidate respinto**

Verdetto: `v0.3.1_candidate_rejected_length_regression`

## Decisione

v0.3.1 non viene pubblicata. Il contratto v2/v3 e i gate deterministici sono
verdi, ma la regressione qualitativa appaiata incontra esattamente la stop rule
congelata: l'asset pubblico v0.3.0 completa Q17-1 correttamente oltre 512
token, mentre il candidate v3 viene troncato dal nuovo limite.

Non sono stati creati commit release, RC pubblicabile, artifact finale, tag o
GitHub Release. Le fasi successive sono `NOT_RUN`, non PASS impliciti.

## Compatibilità v2

La configurazione estratta dall'asset pubblico v0.3.0 è byte-identica al
fixture congelato:

- SHA-256 config: `1c5bbe79edf125485d14518e58ff18c48156eaa0fb91faf82fcf3cd97375d0ee`;
- caricamento con il candidate: PASS;
- file non riscritto: PASS;
- `num_predict` implicito: assente;
- residency implicita: assente;
- versioni ignote e configurazioni invalide: failure tipizzato e fail-closed.

Un relay locale ha catturato la request `/api/chat` dell'asset v0.3.0 e del
candidate avviato con lo stesso profilo v2. I due body sono byte-identici:

```text
f0626f1adb6488c5abaffc536fba6ee9e365c148c3e134ce25af32b2b8eb3d35
```

Entrambe le run no-file terminano `stop`, con 23 output token e risposta
semanticamente e testualmente coincidente. Il gate upgrade v2 è PASS.

## Contratto v3

Test, request capture e output M22/M23 confermano:

- `options.num_predict: 512` realmente inoltrato;
- `keep_alive: "5m0s"` realmente inoltrato;
- context 4096, thinking false, temperatura zero;
- assenza o invalidità dei campi v3 respinta;
- nessun fallback alla semantica v2;
- manifest e profilo di packaging coerenti.

Il gate contrattuale v3 è PASS. È il valore qualificato dal contratto a non
preservare la baseline qualitativa, non un errore di inoltro.

## Regressione qualitativa

La serie appaiata usa gli esatti prompt e oracoli Q17-1…Q17-5 congelati nella
matrice. L'ordine è baseline/candidate per ogni task e la stop rule è
fail-fast.

### Q17-1

| Campo | Asset pubblico v0.3.0 | Candidate v3 |
|---|---|---|
| exit | 0 | 1 |
| terminale provider | `stop` | `length` |
| output token | 535 | 512 |
| risposta pubblicata | sì, completa | no, `response_invalid` |
| workspace | invariato | invariato |

La risposta baseline copre dipendenze, ordine `charge` → `create` →
`dispatch`, evento, ritorno e limiti dell'evidenza senza rendere certa la
persistenza. Il candidate non può pubblicare una risposta perché Ollama attesta
`done_reason: length`.

La cattura diretta candidate registra:

```text
request_sha256  f810588c17640621942bd23fda5d7808c3180d58610e220b13594a0655c573af
response_status 200
response_sha256 f5e580744b2e69dcb49a94e55081abdba86c466a99652928dd0bfb58c36a8bac
done            true
done_reason     length
eval_count      512
num_predict     512
keep_alive      5m0s
```

Q17-2…Q17-5 sono `NOT_RUN` per stop rule. Non vengono eseguiti per cercare un
conteggio aggregato favorevole dopo una regressione già dimostrata.

## Gate deterministici e LF

Su una materializzazione temporanea LF dei blob Git sono verdi:

- `go test ./...`;
- `go test -race ./...`;
- `go vet ./...`.

Il repository aggiunge `.gitattributes` per Go, YAML, script e fixture. Il
fixture v2 byte-identico e i test M23 rendono permanenti compatibilità, assenza
di default impliciti e diagnostica fail-closed. Il relay conserva un harness
ripetibile per confrontare request e terminali provider.

## Conseguenze

- M22 resta chiusa come qualifica tecnica dell'hardening;
- v0.3.0 resta l'ultima release pubblica e il suo support claim non cambia;
- v0.3.1 non è release-ready e non riceve tag o asset;
- un nuovo tentativo deve rivalutare `num_predict` prima di riaprire la
  pubblicazione;
- non è ammesso reinterpretare `response_invalid` come qualità accettabile o
  aumentare il limite post-hoc dentro questa matrice.


# Milestone 25 — Report finale

Data: 2026-09-02

Stato: **COMPLETATA**

Verdetto: **`field_adoption_negative`**

L'esatto asset pubblico v0.3.1 è stato osservato su due progetti Laravel reali
con Direct Chat single-file read-only, senza rebuild, tuning, retry qualitativi
o mutazioni dei workspace. I gate di sicurezza sono rimasti verdi, ma la
completion (63,64%) e il correct rate sulle risposte valutabili (57,14%) sono
sotto le soglie negative definite dal protocollo.

## Identità e ambiente

| Campo | Evidenza |
|---|---|
| release | `v0.3.1` pubblica |
| commit incorporato | `bd0e902c8d7ef01c01117537fceed76845a33732` |
| SHA-256 archive | `2420ba89ada7b0b9cf3de8bd62d7f97dc32868aa342e44e5c3dacbaa94b3a6b6` |
| SHA-256 binario | `0d5e068019e5187c517f9ff0bc7966b5f3123be933b6d858f2f2fa16978c36ed` |
| ambiente | WSL2 Ubuntu 24.04, Linux amd64, NVIDIA RTX 5070 |
| provider/modello | Ollama 0.33.1, `qwen3.5:9b`, digest congelato |
| profilo | schema v3, context 4096, thinking false, temperatura zero, `num_predict: 1024`, residency `5m` |

Archive e binario sono stati riscaricati ed estratti fuori dal checkout. Le due
configurazioni private differivano dal profilo distribuito soltanto per
`workspace.root`.

## Risultati aggregati

| Metrica | Risultato | Soglia positiva |
|---|---:|---:|
| task core e project-b | 11 | — |
| completati | 7/11 (63,64%) | almeno 85% |
| correct sulle valutabili | 4/7 (57,14%) | almeno 80% |
| partial | 3 | — |
| unevaluable | 4 | — |
| terminali `length` | 0 | 0 |
| falsità materiali nei risultati accettati | 0 | 0 |
| utilità mediana | 3/5 | almeno 4/5 |
| latenza p50 | 6,97 s | massimo 30 s |
| latenza p95 | 20,26 s | massimo 90 s |
| heartbeat eleggibili conformi | 2/2 (100%) | 100% |
| diagnostica | 4/4 | 4/4 |
| mutazioni workspace | 0 | 0 |

La replica `M25-S1`, esclusa dal denominatore qualitativo, è semanticamente
coerente con `M25-C1` e termina `stop` senza troncamento.

## Osservazioni qualitative

`M25-C0`, `M25-C2`, `M25-C3` e `M25-B1` soddisfano gli oracoli. `M25-C1`,
`M25-C6` e `M25-B2` sono parziali per assunzioni non dimostrate: gestione di
un errore di validazione, tabella effettiva associata a un modello e
conversione di un model binding in identificatore di route.

`M25-C4`, `M25-C5`, `M25-C7` e `M25-B3` non producono una risposta valutabile:
il comando termina con `response_invalid`. Non è stato eseguito alcun retry.
I sette completamenti terminano tutti `stop`, con `truncated=false`.

## Diagnostica e heartbeat

Il profilo valido supera 5 controlli su 5. Versione schema ignota,
`num_predict` invalido e residency semanticamente invalida (`0s`) falliscono
chiusi con categoria e campo redatti. Una prima preparazione del caso residency
con il valore non decodificabile `never` aveva misurato correttamente
`yaml_invalid`, ma non il caso semantico richiesto; l'evidenza correttiva non
esegue alcuna generazione.

Il ticker parte dopo il preflight, immediatamente prima della chiamata di
generazione. Le due generazioni che raggiungono 15 secondi emettono almeno un
heartbeat. La replica streaming dura 15,81 secondi end-to-end, ma soltanto
12,927 secondi nella finestra di generazione: l'assenza di heartbeat è corretta.
Le quattro terminazioni `response_invalid` aggiungono una riga di errore allo
stderr oltre alla forma heartbeat consentita.

## Sicurezza e immutabilità

Commit e digest SHA-256 dei file tracciati coincidono prima e dopo la matrice
per entrambi i progetti. Gli elementi locali esclusi (`.env*` e `storage/`)
non sono stati letti né inclusi nei digest. Non sono stati osservati path,
prompt, contenuti, endpoint o secret negli output; nessun tool, retrieval,
agent, fallback o ampliamento di autorità è stato usato.

## Decisione e backlog

v0.3.1 non viene adottato per l'uso quotidiano osservato. La release resta
valida nel proprio support claim; il verdetto non modifica retroattivamente
M24 e non qualifica hardware CPU.

Backlog separato:

- diagnosticare le quattro risposte Ollama rifiutate come `response_invalid`;
- chiarire la politica stderr per gli errori terminali;
- migliorare il vincolo epistemico su errori, mapping ORM e route model binding.

Manifest, prompt, oracoli e risposte complete restano nell'evidenza privata
fuori dal repository.

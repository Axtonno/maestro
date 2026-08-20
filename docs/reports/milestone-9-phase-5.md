# Milestone 9 — Report Fase 5

Data: 2026-08-20

Stato: **COMPLETATA — Milestone 3 chiusa, llama.cpp non supportato**

## Baseline congelata

- Ollama positivo: `milestone-3-live-ollama-validation.md`, fixture
  `llama3.1:8b`, Smoke finale 13 passed, 1 skipped, 0 failed;
- Ollama negativo: `qwen2.5-coder:7b`, tool calling serializzato come testo;
- llama.cpp storico: due tentativi router mode con preflight iniziale positivo,
  poi OOM dell'host; report incompleti e non validi.

Timeout, scenari e criteri di queste campagne non vengono modificati
retroattivamente.

## Preflight llama.cpp

| Voce | Evidenza |
|---|---|
| Host | Linux `amd64`, Intel Core i5-8365U, 8 CPU logiche |
| Memoria | 15 GiB RAM, 4 GiB swap |
| `llama-server` / `llama-cli` | Assenti dal `PATH` |
| Endpoint locale atteso | `127.0.0.1:8080` non raggiungibile |
| Server/versione corrente | Non disponibile |
| Modello/profilo single-model | Non configurato |
| Router multi-modello | Vietato su questo host dopo due OOM |

Il preflight non identifica una configurazione valida da sottoporre alla
matrice. Nessun server viene avviato, nessun modello viene scaricato e nessuno
scenario skipped viene contato come PASS.

## Decisione

Si applica l'esito 3 previsto dal piano: **preflight incompatibile, matrice non
avviata, llama.cpp non supportato**. ADR-0030 rende la decisione permanente per
la serie v0.1.x e chiude formalmente la Milestone 3.

La presenza dell'adapter e dei test deterministici conferma il confine
architetturale, non la compatibilità live. Una futura promozione richiede un
nuovo report su un profilo dichiarato e sicuro; non riapre retroattivamente la
Milestone 3.

## Handoff benchmark mutativo

La Milestone 11 riceve soltanto questo contratto, senza codice anticipato:

- configurazione hardware–provider–modello congelata prima delle run;
- Gate A `3/3`, Gate B `2/2`, Gate C `3/3`, fail-fast;
- scenario, terminale, tool call, approval, digest e stato fisico registrati;
- nessuna modifica di timeout, temperatura o criteri dopo il primo risultato;
- nessun support claim se manca un profilo live completo.

## Gate

- stato llama.cpp non ambiguo: superato;
- baseline Ollama positiva/negativa congelata: superato;
- report OOM non promossi a evidenza valida: superato;
- Milestone 3 formalmente completata: superato;
- nessuna capacità mutativa o installazione implicita introdotta: superato.

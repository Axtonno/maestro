# Milestone 6 — Phase 5 Report

Stato: Completata

Data: 2026-08-11

---

# Risultato

Il Context Engine riusa analysis, embedding e stime tra refresh, retrieval e
build mantenendo cache e snapshot nettamente separati. La cache è bounded,
in-memory, content-addressed e semanticamente trasparente.

---

# Implementazione

- `CachePolicy` pubblica e validata;
- `CacheStats` e `CacheInspector`;
- LRU deterministico per entry e byte;
- cloni difensivi dei vettori;
- chiavi analysis con digest, path, media, language, analyzer e versione;
- chiavi embedding con provider, modello, dimensione e digest testo;
- chiavi estimator con ID, versione e digest testo;
- riuso delle analysis dopo refresh invariato;
- cache embedding per query e candidati;
- cache delle stime e dei prefissi usati dal troncamento;
- purge atomico del target su cambio dimensione;
- pubblicazione soltanto dopo validazione.

---

# Invalidazione

La matrice coperta include:

- contenuto invariato e path stabile: hit analysis;
- rename: miss e nuova provenance;
- analyzer ID/versione: chiave distinta;
- provider/modello/dimensione: chiave distinta o purge;
- estimator ID/versione: chiave distinta;
- testo query o sezione: digest distinto;
- entry oversized: non inserita;
- pressione entries/bytes: eviction LRU.

La generazione non fa parte della chiave degli artefatti; un refresh invariato
può quindi riusarli senza rendere la cache fonte dello snapshot.

---

# Concorrenza

Due semantic retrieval cold vengono bloccati contemporaneamente e producono
due chiamate provider distinte. Questo conferma l'assenza di singleflight
implicito e conserva ownership dei due context.

Cache, engine registry e snapshot usano lock distinti. Nessuna callback esterna
viene eseguita sotto il lock della cache.

---

# Test

Coperti:

- analysis cold, rename e warm stabile;
- semantic retrieval cold/warm equivalente;
- dimension change e retry dopo purge;
- bundle cold/warm byte-equivalente nei campi pubblici;
- estimator non reinvocato su hit;
- LRU e limiti;
- entry oversized;
- richieste cold concorrenti indipendenti;
- policy invalida e minima.

Comando:

```text
GOCACHE=/tmp/maestro-go-build go test ./internal/contextengine ./pkg/contextengine
```

Esito: superato.

---

# Gate

Hit e miss producono risultati equivalenti, limiti ed eviction sono verificati,
failure non diventano entry e nessun contenuto viene persistito. La Fase 5 è
completata; resta l'integrazione pubblica, osservabilità e hardening della Fase
6.

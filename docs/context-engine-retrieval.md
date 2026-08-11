# Context Engine — Retrieval and Context Builder

Versione: 0.1.0

Stato: Fase 4 implementata

Data: 2026-08-11

---

# Scopo

Descrivere retrieval lessicale, strutturale e semantico, fusione dei ranking,
stima dei costi e costruzione dei context bundle.

---

# Candidati

Documenti senza analysis producono un candidato sull'intero contenuto. Quando
sono disponibili chunk strutturali, ogni chunk diventa un candidato con path e
intervallo originali. Documenti opachi non partecipano al retrieval testuale.

Filtri di path e linguaggio vengono applicati prima del ranking. Tutti gli
ordinamenti usano score discendente e poi path/intervallo come tie-break
dichiarato.

---

# Retrieval lessicale

La baseline `lexical` tokenizza in lowercase lettere, cifre e underscore. Lo
score è la quota di termini unici della query presenti nel candidato:

```text
matched_query_terms / unique_query_terms
```

Il reason code è `term_coverage`. La formula è offline, deterministica e non
usa ordine di registrazione o frequenza del path come preferenza.

---

# Retrieval strutturale

`structured` consulta i simboli prodotti dagli analyzer. `Query.Symbol`
abilita un match case-insensitive esatto con score 1 e reason
`symbol_exact`. Senza filtro simbolo, i termini della query possono produrre
`symbol_term_match` con score 0,9.

Il risultato usa l'intervallo del simbolo, non l'intero documento.

---

# Retrieval semantico

`semantic` richiede sempre `EmbeddingTarget` con provider e modello esatti.
L'engine invia in un unico batch query e candidati attraverso il Provider
Runtime iniettato.

Sono validati:

- cardinalità del batch;
- dimensione uniforme e non vuota;
- valori finiti;
- norme non nulle.

Lo score è la cosine similarity e il reason code `cosine_similarity`. Errori
provider conservano la causa e `ErrEmbeddingFailure`; cancellazione e deadline
restano idiomatiche.

Senza Provider Runtime il metodo restituisce `ErrUnsupported`. Non esiste
fallback silenzioso al retrieval lessicale.

---

# Fusione

Una query con più metodi deve dichiarare `FusionReciprocalRank`. La Reciprocal
Rank Fusion usa la costante 60 e combina risultati con stesso path e intervallo.
L'output dichiara metodo `fused` e reason `reciprocal_rank_fusion`.

Una query multi-metodo senza strategia, o una strategia senza più metodi, è
invalida. La baseline non espone pesi nascosti o ranking appreso.

---

# Estimator

Gli estimator sono registrabili con ID e versione. Devono restituire un costo
non negativo e monotono sui prefissi UTF-8.

`context.utf8-estimator@1` è la baseline offline conservativa:

```text
ceil(UTF-8 bytes / 3)
```

Il valore è sempre dichiarato come stima, non come token count model-specific.
Errori, costi zero per testo non vuoto, valori negativi e panic producono
`ErrEstimatorFailure`.

---

# Context Builder

Il builder acquisisce una sola generazione di snapshot e usa la stessa vista
per retrieval e sezioni. Procede nell'ordine del ranking:

1. scarta intervalli sovrapposti già selezionati;
2. stima il costo dell'intervallo;
3. include integralmente quando possibile;
4. altrimenti trova il prefisso UTF-8 più lungo entro l'allowance;
5. arresta la selezione quando il budget è esaurito.

Ogni sezione conserva path, intervallo, testo esatto, ruolo `evidence`, metodo,
reason code, costo e indicatore di troncamento. Il testo viene verificato contro
lo slice originale del documento.

Il budget separa `MaxTokens`, `ReservedTokens` e `SafetyTokens`; soltanto
`EvidenceTokens` è disponibile al codice. Il bundle non può superarlo.

Il builder non invoca completion, non genera riassunti e non riscrive la query.

---

# Developer Benchmark

Lo scenario `developer-retrieve-with-embeddings` usa ora il percorso pubblico:

```text
dataset documents -> Context Engine index -> semantic retrieval -> ranking
```

Dataset, target provider/model, rubrica e report restano invariati. Un wrapper
locale osserva soltanto la dimensione embedding per la metrica già esistente;
prompt, contenuti e vettori non vengono serializzati.

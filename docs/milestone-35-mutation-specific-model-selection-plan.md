# Milestone 35 — Mutation-Specific Model Selection

Stato: Aperta — `mutation_model_selection_open`.

Data: 2026-09-05. Prerequisito: M34 conclusa nel ramo B, ADR-0039.

## Obiettivo

Selezionare un modello dedicato alla Controlled Mutation host-bound,
mantenendo `qwen3.5:9b` per Direct Chat. Nessun modello mutativo è ancora
selezionato, scaricato o qualificato. Il profilo respinto da M34 è escluso
dal tuning e dalla ripetizione di benchmark per cercarne un esito diverso.

## Fasi

1. Inventariare modelli candidati e requisiti verificabili: licenza,
   disponibilità locale, dimensioni, quantizzazione, memoria sul target
   RTX 5070, structured output e comportamento tool-free. Definire shortlist
   e criteri di esclusione prima del confronto; nessuna scelta per reputazione.
2. Congelare profili completi per candidato: digest, provider/versione,
   renderer/template, parametri espliciti e default, prompt host-bound,
   budget, codice e regole del confronto. Il modello non ottiene autorità
   aggiuntiva rispetto al contratto M33.
3. Usare una matrice di selezione nuova e uguale per tutti i candidati;
   dichiarare in anticipo ordine, ripetizioni e criterio di ranking. Non usare
   M33 né il futuro holdout di qualifica per prompt design o selezione.
4. Congelare il modello selezionato e qualificare una matrice nuova con holdout
   indipendente non visto. Registrare ogni tentativo, senza retry selettivi,
   repair o fallback. Distinguere selezione da qualifica di prodotto.
5. Pubblicare evidenze, esito dei gate e decisione. Se nessun modello supera
   i gate, mantenere Controlled Mutation fuori dal claim.

## Gate di qualifica

Ereditare le definizioni dei denominatori M34: output conformi 100%, target
conservato 100%, proposte positive almeno 90%, holdout apply completati 100%,
astensioni necessarie corrette 100%, approval allow/deny raggiunte 100%.
Preview esatta 100%; zero scritture stale, fuori selezione, mutazioni errate
o non approvate e failure con effetti. Development e holdout devono passare
separatamente. Gli apply attraversano TTY reale, commit e verifica byte.

La selezione è aperta, non congelata. Un esito futuro positivo consentirà una
decisione esplicita di productization; oggi v0.5.0 resta non autorizzata.

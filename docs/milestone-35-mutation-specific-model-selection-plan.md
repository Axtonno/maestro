# Milestone 35 — Mutation-Specific Model Selection

Stato: Completata — `mutation_specific_model_qualified`.

Data: 2026-09-05. Prerequisito: M34 conclusa nel ramo B, ADR-0039.

## Obiettivo

Selezionare e qualificare un modello dedicato alla Controlled Mutation
host-bound, mantenendo `qwen3.5:9b` per Direct Chat. Il profilo respinto da
M34 è escluso dal tuning e dalla ripetizione di benchmark per cercarne un
esito diverso.

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

## Chiusura — 2026-09-05

La selezione congelata ha confrontato `qwen2.5-coder:7b`,
`qwen2.5-coder:14b` e `granite-code:8b-instruct` su 12 casi nuovi per profilo.
Il solo candidato eleggibile è `qwen2.5-coder:14b`, digest
`9ec8897f747e246e970bc5cfdda85d22f1123dc2e3d34978a010a75968716849`,
con 12/12 output conformi, 9/9 positivi e 3/3 astensioni corrette.

La successiva qualifica indipendente ha ottenuto, separatamente su sviluppo e
holdout, 12/12 output provider conformi, 10/10 positivi, 2/2 astensioni,
10/10 target e preview, 10/10 approval e 7/7 apply. Tutti i 15 terminali per
set sono corretti; scritture stale o fuori selezione, mutazioni errate o non
approvate e failure con effetti sono zero.

Verdetto `mutation_specific_model_qualified`. M35 autorizza l'avvio della
productization in M36 con routing per capacità: `qwen3.5:9b` resta il modello
Direct Chat e il profilo congelato `qwen2.5-coder:14b` è quello mutativo.
Il support claim pubblico resta invariato e v0.5.0 non è pubblicata né
autorizzata come release.

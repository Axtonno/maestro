# Milestone 32 — Report finale

Data: 2026-09-04

Stato: **COMPLETATA — QUALIFICATION RESPINTA**

Verdetto: **`binary_mutation_decision_rejected`**

| Gate | Development | Holdout | Globale | Esito |
|---|---:|---:|---:|---|
| output validi | 11/11 | 7/7 | 18/18 | PASS |
| proposte positive corrette | 4/5 | 4/4 | 8/9 | FAIL |
| insufficienze semantiche astenute | 2/2 | 1/1 | 3/3 | PASS |
| target assenti/duplicati e scope bloccati come previsto | 2/4 | 0/2 | 2/6 | FAIL |
| approval allow/deny raggiunta | 4/5 | 4/4 | 8/9 | FAIL |
| terminali corretti | 8/11 | 5/7 | 13/18 | FAIL |
| mutazioni semanticamente errate applicate | 0 | 0 | 0 | PASS |
| failure con effetti | 0 | 0 | 0 | PASS |
| mutazioni fuori scope | 0 | 0 | 0 | PASS |

D02 ha prodotto una proposta formalmente valida ma con `old_text` inesistente,
fermato dal compiler come `target_not_found`: il positivo non ha raggiunto il
TTY. Nei quattro casi absent/duplicate il modello ha invece sostituito il
target esplicito con un frammento presente (o con un blocco più ampio), creando
una preview semanticamente non ammissibile. Le quattro preview sono state
negate al TTY e il workspace è rimasto invariato.

Terminologia conclusiva:

```yaml
proposte semanticamente non ammissibili: 5
mutazioni inventate applicate: 0
failure con effetti: 0
```

I percorsi positivi qualificati hanno attraversato proposta, preview, TTY,
applicazione e verifica byte-per-byte. Entrambi i denial previsti hanno
raggiunto realmente il TTY e terminato `approval_rejected`; stale ha preservato
le modifiche concorrenti.

La stop rule `controlled_mutation_model_profile_rejected` non scatta perché
8/9 (88,9%) supera la soglia congelata dell'80%. Non sono autorizzati ulteriori
tuning o repliche della matrice M32. Nessun candidate v0.5.0 è autorizzato.

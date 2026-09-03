# Milestone 28 – Report finale

Data: 2026-09-03

Stato: **COMPLETATA SENZA QUALIFICAZIONE**

Verdetto: **`controlled_mutation_transport_unresolved`**

## Risultato

Il recupero deterministico della Controlled Mutation è completo. Maestro dispone di un contratto JSON v1 strict, di un compilatore provider-neutral, di una superficie opt-in separata `workspace.replace`, di preview completa, fingerprint canonico, approval one-shot, stale check e replace atomico. Il claim pubblico v0.4.0 non è cambiato.

La qualification v0.5.0 non è autorizzata: nessuno dei due trasporti può essere selezionato senza le run semantiche/live congelate. Non viene prodotto candidate, package, tag o release.

## Gate deterministici

- proposta positiva, determinismo e convergenza dei trasporti: PASS;
- schema strict, campi duplicati/sconosciuti/mancanti e output misto: PASS;
- path mismatch, traversal, target sensibili, old text assente/ambiguo e no-op: PASS;
- preview non mutativa con digest pre/post e fingerprint: PASS;
- approvazione one-shot e binding dell'esatta prepared invocation: PASS;
- fingerprint preview/risultato identico: PASS;
- stale write: respinto, zero effetti;
- fault pre-commit: zero effetti e nessun temporaneo residuo;
- containment e symlink matrix del filesystem layer: PASS;
- suite completa Linux LF, race, vet e `git diff --check`: PASS.

## Confronto trasporti

Tool calling nativo e structured output compilano la stessa proposta valida nello stesso candidato. Nessun fallback o repair è presente. Il confronto semanticamente decisivo è `not_run`, perché provider e modello target non sono disponibili.

## Handoff

Non esiste handoff a release readiness v0.5.0. Una nuova milestone di qualification potrà riusare il protocollo congelato soltanto predisponendo Linux `amd64`, Ollama, modello e digest congelati, task set/holdout, approval TTY reale e installazione fuori checkout. Qualunque modifica a schema, compiler, policy o prompt richiederà un nuovo freeze.

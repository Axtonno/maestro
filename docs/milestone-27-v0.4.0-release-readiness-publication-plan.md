# Milestone 27 — v0.4.0 Release Readiness & Publication

Versione candidata: v0.4.0

Stato: Completata — `v0.4.0_released_and_verified`

Data: 2026-09-03

Prerequisito: M26 chiusa con `v0.4.0_candidate_field_qualified`.

## Obiettivo

Portare il candidate qualificato da M26 alla pubblicazione v0.4.0 attraverso
una validazione indipendente e una catena release riproducibile. Il gate non
può ampliare capability o autorità e non può usare il nuovo holdout per
selezionare o riscrivere il candidate.

Restano fuori scope CPU qualification, altri modelli, multi-file, agent,
retrieval e Controlled Mutation.

## Identità congelata

- candidate di partenza: `v0.4.0-rc.4`;
- provider/modello/profilo: identici al gate conclusivo M26;
- `num_predict`: 1024;
- validator e adapter: invariati rispetto a M26;
- autorità: Direct Chat single-file read-only, zero tool call;
- stato iniziale: `release_not_yet_authorized`.

Una modifica funzionale a prompt, validator, adapter, profilo, modello o
autorità invalida il candidate. Correzioni esclusivamente a fixture o test
devono essere motivate e dimostrare che non cambiano il comportamento live.

## Sequenza vincolante

1. classificare ogni failure della suite globale e risolverlo oppure aggiornare
   motivatamente l'aspettativa congelata;
2. ottenere suite completa, race detector, vet e `git diff --check` verdi su
   una materializzazione Linux LF pulita;
3. verificare esplicitamente caricamento e comportamento compatibile dei
   profili v2 e v3;
4. congelare una matrice holdout single-file mai usata per selezionare
   `v0.4.0-rc.4`;
5. eseguire ogni task holdout una sola volta e richiedere almeno 80% correct,
   zero falsità materiali e zero `response_invalid`;
6. creare dal commit reale due package indipendenti e verificarne
   l'identità byte per byte;
7. estrarre e installare l'archive fuori da ogni checkout Git;
8. eseguire dall'archive installato un breve gate live read-only;
9. pubblicare v0.4.0 soltanto dopo tutti i gate precedenti, riscaricare asset e
   checksum e verificarne l'identità con quelli qualificati.

## Gate holdout

Il set deve contenere casi single-file reali e non sovrapposti ai file, prompt
o oracoli usati nella selezione M26. Manifest, prompt, file, oracoli e risposte
complete restano in evidenza privata; il repository conserva soltanto identità
e risultati redatti.

Soglie:

- completion almeno 85%;
- correct almeno 80% delle risposte valutabili;
- `response_invalid` zero;
- falsità materiali zero;
- mutazioni, tool call e ampliamenti di autorità zero.

Il failure di una soglia respinge il candidate. Non sono ammessi retry del
singolo task, sostituzioni post-risposta o tuning sul holdout.

## Gate packaging e pubblicazione

Il commit qualificato deve essere pulito e incorporato nell'identità del
binario. Due build indipendenti devono produrre archive byte-identici. Il gate
fuori checkout verifica almeno identità, doctor, no-file, complete con file,
stream, containment, redazione, parametri generativi e immutabilità.

La pubblicazione richiede tag annotato v0.4.0 sul commit qualificato, release
non draft e non prerelease, asset e checksum. Una nuova directory deve
riscaricare gli asset pubblici e dimostrare corrispondenza byte per byte prima
del verdetto finale.

## Output attesi

- matrice `milestone-27-v0.4.0-release-readiness-publication-matrix.yaml`;
- report dei test globali e della compatibilità v2/v3;
- report holdout redatto;
- report packaging/installazione/gate live;
- report finale con identità di commit, tag, archive e asset riscaricato.

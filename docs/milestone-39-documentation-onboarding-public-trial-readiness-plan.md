# Milestone 39 — Documentation, Onboarding & Public Trial Readiness

Stato: **completata — `documentation_onboarding_public_trial_ready`**

## Obiettivo

Rendere Maestro v0.5.0 comprensibile e provabile da una persona che arriva
dal README, senza supervisione dell'autore e senza conoscere l'architettura
interna. M39 non introduce feature, non modifica il runtime e non ricostruisce
l'asset v0.5.0: congela una baseline pubblica credibile e ne verifica il
percorso operativo.

La fonte sintetica autorevole è
[Capacità correnti](current-capabilities.md). README, installazione, quick
start, supporto Controlled Mutation, compatibility matrix, security model,
known issues, troubleshooting e release note devono restare coerenti con
quella pagina.

## Regole della prova

- matrice e criteri sono congelati prima dell'esecuzione;
- si usa esclusivamente l'archive pubblico v0.5.0 e il relativo checksum;
- niente checkout, rebuild, patch del binario o sostituzione dei modelli;
- i passaggi seguono la documentazione come farebbe un nuovo utente;
- la fixture estratta viene inizializzata come repository Git pulito;
- chat e mutation hanno un solo tentativo ciascuno;
- l'allow-once è dato solo se la preview coincide con l'atteso;
- output, hash, diff e stato Git diventano evidenza preservata.

## Casi congelati

| ID | Prova | Esito richiesto |
| --- | --- | --- |
| P01 | Navigazione pubblica | Link locali risolti e nessuna contraddizione di supporto |
| P02 | Identità asset | Archive, checksum, manifest, release, commit e binario coerenti |
| P03 | Identità modelli | Nomi e digest esatti per chat e mutation |
| P04 | Doctor | 14/14 check `pass` |
| P05 | Direct Chat | Controller esplicito, risposta corretta, un tentativo |
| P06 | Controlled Mutation | Riga 22: `201` → `202`, preview esatta e allow-once |
| P07 | Verifica Git | Un solo file modificato e diff esattamente atteso |
| P08 | Contratti repository | Test, race, vet, link e packaging riproducibile verdi |

I valori letterali di asset, modelli, prompt e modifica sono in
[milestone-39-public-trial-matrix.yaml](milestone-39-public-trial-matrix.yaml).

## Gate

M39 può chiudere con verdetto
`documentation_onboarding_public_trial_ready` soltanto con:

- zero contraddizioni tra i documenti pubblici;
- zero link locali irrisolti e zero posizionamento corrente v0.3 legacy;
- asset pubblico e due modelli con identità esatta;
- doctor 14/14, chat 1/1 e mutation 1/1;
- un solo file selezionato e modificato, con diff esatto;
- zero scritture non approvate o passaggi operativi non documentati;
- test, race detector, vet e doppio packaging riproducibile verdi.

Qualunque divergenza produce `public_trial_not_ready` e deve essere corretta
prima di dichiarare chiusa la milestone.

## Chiusura

La prova pubblica del 2026-09-08 ha superato P01–P08 al primo tentativo:
archive e modelli esatti, doctor 14/14, chat corretta, preview esatta,
allow-once e un solo file nel diff Git. Suite, race detector, vet e doppio
packaging sono verdi. Evidenza e verdetto sono in
[reports/milestone-39-public-trial.json](reports/milestone-39-public-trial.json)
e [reports/milestone-39-final.md](reports/milestone-39-final.md).

## Fuori scope

- nuove capability runtime;
- pubblicazione o sostituzione dell'archive v0.5.0;
- qualifica di hardware, provider o modelli aggiuntivi;
- multi-file, agent autonomi, Windows nativo o integrazioni editor.

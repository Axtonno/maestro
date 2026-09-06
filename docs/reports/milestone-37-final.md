# Report finale — Milestone 37

Data: 2026-09-06  
Titolo: **v0.5.0 Release Readiness & Publication**

## Obiettivo

M37 congela la productization qualificata di M36 in un release Linux amd64
riproducibile, installabile fuori checkout e verificabile anche dopo il
download pubblico. Il claim è limitato al reference hardware e alla capability
Controlled Mutation esplicitamente opt-in.

## Controlli preliminari

Il confronto con `cb2a408` ha mostrato modifiche documentali e un solo
allineamento del test contrattuale M36: il verdetto finale
`controlled_mutation_productization_qualified` aveva reso obsoleto il test che
accettava soltanto il checkpoint residency intermedio. Non sono state
modificate le regole runtime di chat, selezione host-bound o apply atomico.

Il release checkout è stato ricreato nel filesystem Linux WSL sotto
`/home/cafeo/src/maestro`, con working tree pulito e Ollama eseguito dentro
WSL. L'ambiente congelato è registrato in
`docs/reports/milestone-37-environment.yaml`.

## Qualificazione tecnica

Nel checkout Linux sono passati:

- `go test ./...`;
- `go test -race ./...`;
- `go vet ./...`;
- `git diff --check`.

La compatibilità operativa è stata verificata sui profili v2, v3 e v4:

- v2 e v3 restano profili Direct Chat compatibili;
- v4 passa il doctor chat e abilita separatamente Controlled Mutation;
- il doctor mutation v4 passa soltanto in una TTY reale;
- non esiste attivazione implicita della mutation dai profili v2/v3.

La configurazione v4 conserva i digest qualificati di `qwen3.5:9b` e
`qwen2.5-coder:14b`, il prompt host-bound e lo schema strict già verificati in
M36.

## Claim pubblico

Su Linux amd64 e sul reference hardware qualificato, Maestro può modificare un
intervallo esplicitamente selezionato in un singolo file PHP sotto `app/`, dopo
preview e approvazione esplicita, utilizzando il profilo dedicato
`qwen2.5-coder:14b`.

Restano fuori dal claim l'individuazione autonoma del target, il multi-file,
gli inserimenti arbitrari, gli agent, le mutazioni senza TTY, altri linguaggi o
directory e modelli o hardware non qualificati.

## Stato della release

Il packaging release ha superato il doppio archive byte-identico, checksum,
installazione fuori checkout e `doctor --mode all`. I gate live eseguiti dalla
copia installata hanno superato allow, deny, abstain, stale e il roundtrip
`chat → mutation → chat`. Il manifest congela il commit, Go 1.24.5, i due
digest qualificati e il profilo `mutation-productization`.

Il tag annotato `v0.5.0` e gli asset pubblici vengono creati soltanto dopo
questo report; il riscaricamento anonimo e la prova esclusiva della copia
pubblica chiudono M37.

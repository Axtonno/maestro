# Report finale — Milestone 37

Data: 2026-09-06  
Titolo: **v0.5.0 Release Readiness & Publication**

## Obiettivo

M37 congela la productization qualificata di M36 in un release artifact Linux
amd64 riproducibile, installabile fuori checkout e verificabile anche dopo il
download pubblico. La qualification è stata eseguita su Windows → WSL2 →
filesystem Linux, con Ollama dentro WSL e RTX 5070 da 12 GB. Il claim è quindi
limitato alla piattaforma WSL2 qualificata e alla capability Controlled Mutation
esplicitamente opt-in.

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

Su Linux amd64 in WSL2, sul reference hardware qualificato con RTX 5070 da
12 GB, Ollama locale e i modelli/digest distribuiti, Maestro può modificare un
intervallo esplicitamente selezionato in un singolo file PHP sotto `app/`, dopo
preview e approvazione esplicita, utilizzando il profilo dedicato
`qwen2.5-coder:14b`.

Target artifact: `linux_amd64`. Piattaforma qualificata: `WSL2`. Linux amd64
nativo non è ancora validato sul campo.

Restano fuori dal claim l'individuazione autonoma del target, il multi-file,
gli inserimenti arbitrari, gli agent, le mutazioni senza TTY, altri linguaggi o
directory e modelli o hardware non qualificati.

## Stato della release

Il packaging release ha superato il doppio archive byte-identico, checksum,
installazione fuori checkout e `doctor --mode all`. I gate live eseguiti dalla
copia installata hanno superato allow, deny, abstain, stale e il roundtrip
`chat → mutation → chat`. Il manifest congela il commit, Go 1.24.5, i due
digest qualificati e il profilo `mutation-productization`.

Il tag annotato `v0.5.0` è stato pubblicato sul commit
`86ed92495c5ce5bd2d0ec8d3c8a11b8c29a316f2`. L'archive pubblico
`maestro-v0.5.0-linux-amd64.tar.gz` e il checksum sono stati riscaricati senza
autenticazione. Lo SHA-256 verificato è
`0afcfe4d648edcde3caf4327c4f995606fb4c3974c05606e13f90dd8cff321d9`.

La copia pubblica installata ha superato `doctor --mode all` e il roundtrip
`chat → mutation → chat`, con gli stessi modelli e digest del builder. Questo
chiude M37 con il verdetto `v0.5.0_released_and_verified`.

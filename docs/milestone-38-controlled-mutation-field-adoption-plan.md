# Milestone 38 — v0.5.0 Controlled Mutation Field Adoption

Stato: completata — `native_linux_field_adoption_qualified`.

## Scopo

Verificare l'uso reale dell'asset pubblico `v0.5.0` su un PC Linux nativo,
senza checkout, rebuild o binari provenienti dal repository. La prova deve
misurare sicurezza della scrittura, qualità semantica, completamento,
latency/residency e utilità percepita.

La qualification M38 è distinta da M37: M37 ha verificato l'artifact Linux
amd64 in Windows → WSL2 → filesystem Linux; M38 deve verificare Linux amd64
nativo sul campo.

## Regole dell'esecuzione

1. Scaricare archive e checksum soltanto dalla release pubblica `v0.5.0`.
2. Verificare il checksum e installare in una directory nuova, fuori da ogni
   checkout Maestro.
3. Seguire esclusivamente `docs/installation.md` e la configurazione v4
   distribuita nell'archive.
4. Registrare sistema operativo, kernel, CPU, GPU, driver, RAM, Ollama,
   modelli e digest.
5. Usare un workspace Laravel dedicato sotto `app/` e un repository Git
   inizializzato per la sola prova.
6. Catturare `git status --short` e `git diff --exit-code` prima e dopo ogni
   operazione; un effetto fuori selezione interrompe la milestone.
7. Non usare checkout Maestro, sorgenti locali, `go run`, rebuild o modelli
   diversi da quelli dichiarati nel manifest pubblico.

## Matrice minima

| Caso | Capacità osservata | Esito richiesto |
| --- | --- | --- |
| F01 | `doctor --mode all` | tutti i check verdi |
| F02 | Direct Chat senza file | completion corretta |
| F03 | Direct Chat con file esplicito | file unico e risposta corretta |
| F04 | sostituzione semplice | preview = diff applicato |
| F05 | selezione multilinea | preview = diff applicato |
| F06 | testo Unicode | byte e diff corretti |
| F07 | codice ripetuto | solo lo span selezionato cambia |
| F08 | approval allow | una scrittura autorizzata |
| F09 | approval deny | zero scritture |
| F10 | sorgente stale | `stale_source`, zero scritture |
| F11 | chat → mutation → chat ripetuto | modello e risposta corretti |
| F12 | latenza e utilità | soglie M38 raggiunte |

Ogni caso mutativo deve conservare preview completa, risposta del modello,
approval, terminale, digest before/after, fingerprint e diff Git.

## Gate di qualifica

- target conservati: 100%;
- diff coincidenti con la preview: 100%;
- mutazioni non approvate: 0;
- modifiche fuori selezione: 0;
- failure con effetti: 0;
- modifiche semanticamente corrette: almeno 90% delle richieste valutabili;
- completion delle richieste valide: almeno 90%;
- utilità mediana: almeno 4/5;
- nessun fallback incrociato, OOM, offload indesiderato o modello errato.

Un singolo effetto fuori selezione, una mutazione non approvata o un failure con
effetti produce `field_adoption_incident` e interrompe la serie. Un gate di
qualità sotto soglia produce `field_adoption_mixed` o `field_adoption_negative`
senza modificare il claim precedente.

## Evidenza

Il report deve includere l'hash dell'archive pubblico, il manifest, l'identità
del binario, l'ambiente nativo, i modelli/digest, la matrice caso per caso,
latency cold/warm, residency, diff Git before/after e i punteggi di utilità.
Il verdetto può essere `native_linux_field_adoption_qualified`,
`field_adoption_mixed`, `field_adoption_negative` o `field_adoption_incident`.

La serie è stata conclusa il 2026-09-07. Risultati ed evidenze sono in
`reports/milestone-38-final.md`, `reports/milestone-38-environment.yaml` e
`reports/milestone-38-live-runs.json`. Hash e policy di immutabilità sono in
`milestone-38-field-adoption-freeze.yaml`.

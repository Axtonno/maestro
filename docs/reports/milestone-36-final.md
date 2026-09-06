# Report finale — Milestone 36

Data di chiusura: 2026-09-06  
Verdetto: **`controlled_mutation_productization_qualified`**  
Release v0.5.0: **non autorizzata**

## Obiettivo

La Milestone 36 aveva il compito di trasformare la combinazione qualificata in
una capability installabile e verificabile. Il profilo Direct Chat mantiene
`qwen3.5:9b`; Controlled Mutation usa il profilo dedicato
`qwen2.5-coder:14b`. I due modelli non devono restare residenti insieme: il
runtime deve scaricare il profilo uscente, attendere la residency vuota e
caricare quello entrante senza fallback silenziosi.

## Implementazione consegnata

- configurazione v4 con sezioni separate `direct_chat` e
  `controlled_mutation`, digest e prompt/schema congelati;
- routing per capability senza fallback incrociati;
- lifecycle esplicito con eviction e verifica della residency;
- CLI opt-in `maestro workspace replace --file <path> --lines <start:end>`;
- selezione host-bound, preview completa, approval esplicita e apply atomico;
- terminali distinti per allow, deny, abstain e sorgente stale;
- `doctor --mode chat`, `doctor --mode mutation` e `doctor --mode all`;
- packaging candidate con manifest delle identità dei due modelli;
- installazione verificata fuori dal checkout.

## Residency sulla RTX 5070

Il reference environment usa una RTX 5070 da 12 GB e WSL con `swap=0`.
Sono stati eseguiti tre cicli indipendenti `chat → mutation → chat`, per 18
richieste complessive.

| Misura | Risultato |
| --- | ---: |
| Direct Chat cold | 3.368–5.699 ms |
| Direct Chat warm | 64–75 ms |
| Chat → mutation | 6.049–7.766 ms |
| Mutation warm | 362–364 ms |
| Mutation → chat | 4.315–4.770 ms |
| Chat warm dopo il ritorno | 65–73 ms |
| Picco GPU osservato | 9.228 MiB |
| RAM WSL libera minima | 13.893.772 KiB |

Ogni completion ha mantenuto esclusivamente il modello richiesto. Non sono
emersi fallback incrociati, offload CPU, OOM, crescita dello swap, restart del
provider o transizioni fallite. Il report residency accettato è
`docs/reports/milestone-36-residency-runs-v4.json`, SHA-256
`9cd7b4ae60fcacc9a97c62faabb2877a0ba3e0a4dc3336f910fb0cffc768c8b1`.

## Gate end-to-end

Il candidate costruito dal commit reale `cb2a408` ha superato:

- doppio archive byte-identico e checksum;
- installazione fuori checkout;
- `doctor --mode all` con identità, schema, prompt, TTY, workspace e
  capability verificati;
- prova allow con apply durevole;
- prova deny senza scrittura;
- prova abstain per informazione insufficiente senza scrittura;
- prova stale con modifica concorrente, terminata con `stale_source`;
- roundtrip Direct Chat → Controlled Mutation → Direct Chat, con sequenza
  osservata `qwen3.5:9b` → `qwen2.5-coder:14b` → `qwen3.5:9b`;
- suite Go, `go vet` e race detector già eseguiti sul codice M36.

L'archive verificato è
`dist/m36-candidate-cb2a408-1/maestro-v0.5.0-pc.1-linux-amd64.tar.gz`, con
SHA-256
`86e84498797305c848224550970f3eb7bbc012fbc87ff82643b02fe94a4d327f`.

## Decisione

La productization è qualificata. L'autorità sul target resta nel runtime, il
modello mutativo non può cercare o ampliare la selezione, e ogni scrittura è
vincolata a un singolo file, all'intervallo selezionato, al digest precedente e
all'approvazione esplicita.

La release pubblica non è inclusa in questo verdetto. Prima di autorizzare
v0.5.0 occorrono la build finale dal commit di release, il tag, la pubblicazione
dell'asset e il suo riscaricamento con verifica indipendente. Il rerun del
packaging dopo gli ultimi commit documentali è rimasto differito perché Go non
era disponibile nell'ambiente WSL; quei commit non modificano il codice o il
contenuto funzionale già qualificato del candidate.

Riferimenti: `docs/milestone-36-residency-decision.yaml`,
`docs/reports/milestone-36-residency-v4.md`,
`docs/releases/v0.5.0.md`.

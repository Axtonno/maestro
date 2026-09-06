# Milestone 36 — Controlled Mutation Productization

Stato: Aperta — `controlled_mutation_productization_open`.

Data: 2026-09-05. Prerequisito: M35 conclusa con
`mutation_specific_model_qualified`, ADR-0040.

## Obiettivo e confine

Portare nel prodotto la combinazione già qualificata, senza cercare altri
modelli né modificare il protocollo host-bound:

- `qwen3.5:9b` per Direct Chat;
- `qwen2.5-coder:14b` per Controlled Mutation.

M36 deve rendere questa capability installabile, fail-closed, osservabile e
verificabile sulla RTX 5070 da 12 GB. Non assume che i due modelli possano
restare residenti insieme. Un'eventuale eviction è accettabile quando il
cambio è deterministico, esplicito nei dati diagnostici e compatibile con i
limiti di latenza congelati prima delle prove.

M36 non ripete M33–M35, non usa le loro run come gate di productization e non
concede al modello autorità su path, coordinate, filesystem o approvazione.

## Configurazione v4 separata

Il nuovo schema deve rappresentare le capacità in sezioni indipendenti:

```yaml
version: 4

direct_chat:
  model: qwen3.5:9b
  digest: 6488c96fa5faab64bb65cbd30d4289e20e6130ef535a93ef9a49f42eda893ea7
  residency: 5m

controlled_mutation:
  enabled: true
  model: qwen2.5-coder:14b
  digest: 9ec8897f747e246e970bc5cfdda85d22f1123dc2e3d34978a010a75968716849
  residency: 5m
  prompt: mutation-host-bound-model-selection-v1
  schema: host-bound-mutation-decision-v1
```

La configurazione deve fallire prima di qualsiasi generazione quando:

- la mutazione è abilitata senza profilo dedicato;
- modello o digest non coincidono con le identità qualificate;
- il modello mutativo coincide con quello Direct Chat;
- prompt o schema mutativi non coincidono con gli artefatti congelati;
- un campo sconosciuto tenta di introdurre fallback o routing implicito.

Non esiste fallback silenzioso fra i profili. Direct Chat non può risolvere il
modello mutativo e Controlled Mutation non può usare il modello chat.

## Superficie CLI opt-in

La sola superficie pubblica prevista è:

```text
maestro workspace replace --file <path> --lines <start:end> <istruzione>
```

Il comando accetta un solo path logico relativo al workspace e un intervallo
inclusivo. Prima della chiamata provider verifica TTY reale, path sotto
`app/`, estensione `.php`, file regolare non symlink, intervallo valido,
provider, identità del modello mutativo, prompt e schema.

Il comando fallisce chiuso per TTY assente, intervallo invalido, path esterno
ad `app/`, symlink, modello indisponibile o difforme, output non conforme,
astensione, approval negata e file cambiato dopo la preview. Non prova il
modello Direct Chat, non ripara output e non ripete la generazione.

## Approvazione

La preview deve mostrare prima dell'input:

- file coinvolto e indicazione esplicita `single_file: true`;
- righe iniziale e finale selezionate;
- diff completo;
- digest dei byte prima, della selezione, della sostituzione e del diff;
- fingerprint che vincola target, coordinate, byte e preview;
- scelta esplicita deny oppure allow-once.

Input vuoto, semplice invio, EOF, input sconosciuto e modalità non interattiva
equivalgono a deny. Non è ammessa un'approvazione persistente per la mutazione.

## Doctor e identità

`maestro doctor --mode mutation` deve produrre check distinti e stabili per:

1. configurazione;
2. provider;
3. modello Direct Chat e relativo digest;
4. modello Controlled Mutation e relativo digest;
5. prompt mutativo e SHA-256;
6. schema mutativo e SHA-256;
7. TTY;
8. workspace;
9. capability host-bound disponibile.

`maestro doctor --mode chat` continua a verificare solo Direct Chat.
`maestro doctor --mode all` aggrega entrambi senza eseguire completion. Un
check mancante, `skip` inatteso o identità difforme impedisce l'esecuzione
mutativa.

## Convivenza e residency

Il profilo `milestone-36-residency-profile.yaml` congela tre cicli indipendenti
con la sequenza seguente:

1. unload esplicito di entrambi i modelli e snapshot baseline;
2. Direct Chat cold, snapshot, Direct Chat warm, snapshot;
3. prima mutation dopo chat, snapshot, mutation warm, snapshot;
4. ritorno a Direct Chat, snapshot, Direct Chat warm, snapshot;
5. unload finale e verifica del ritorno alla baseline.

Ogni snapshot registra `/api/ps`, `size`, `size_vram`, scadenza residency,
VRAM usata/libera/totale tramite NVML, RAM disponibile, swap usata, RSS del
provider e stato dei processi. Ogni richiesta registra latenza client,
`load_duration`, prompt/eval duration, token e terminale.

Il report deve distinguere cold, warm e switch. Deve indicare per ogni
transizione se il modello precedente è rimasto residente o è stato espulso.
Sono failure: esito diverso tra cicli, modello inatteso residente, entrambi i
modelli assenti dopo una completion riuscita, `size_vram < size` non previsto,
incremento dello swap, OOM, crash/restart del provider o fallback CPU. Non è
una failure l'eviction coerente del modello precedente.

Le soglie temporali definitive devono essere congelate nel preflight prima
della prima run formale. Nessun valore può essere rilassato dopo aver visto il
report.

## Packaging e release

La productization procede in queste fasi:

1. implementazione e contract test di configurazione, routing, CLI e doctor;
2. preflight e misura formale di residency sul reference hardware;
3. matrice end-to-end del binario installato con allow, deny, stale e abstain;
4. candidate costruito dal commit reale e due archive byte-identici;
5. installazione fuori checkout e ripetizione dei gate dall'archive;
6. suite completa, race detector e vet su Linux `amd64`;
7. documentazione di supporto e requisiti hardware;
8. eventuale pubblicazione separata, riscaricamento dell'asset pubblico e
   verifica di hash, versione, commit e gate installato.

## Gate conclusivi

- Direct Chat mantiene integralmente claim e gate v0.4.0;
- modello, digest, prompt, schema e parametri mutativi coincidono con M35;
- routing corretto chat/mutation: 100%, fallback incrociati: 0;
- doctor completo senza completion e con identità distinte: 100%;
- selezione host-bound, preview completa e approval esplicita: 100%;
- allow, deny, stale e abstain raggiungono tutti il terminale previsto;
- apply attesi nel binario installato: 100%;
- scritture stale, fuori selezione, errate o non approvate: 0;
- failure con effetti, OOM, swap aggiuntivo e fallback CPU: 0;
- comportamento di residency identico nei tre cicli e interamente osservato;
- doppio archive byte-identico e installazione fuori checkout: PASS;
- suite, race, vet e gate `chat → mutation → chat`: PASS.

M36 parte senza candidate, package, tag o release v0.5.0 autorizzato. Il PASS
di productization può autorizzare soltanto una distinta decisione di release;
non pubblica automaticamente alcun asset.

## Checkpoint residency — 2026-09-05

L'eviction automatica v2 e l'handoff esplicito v3 mantengono un solo modello
interamente in VRAM, con pattern identico nei tre cicli, 18/18 richieste
corrette, latenze sotto soglia e zero fallback, offload CPU, OOM o restart del
provider. L'handoff esplicito è la strategia productizzata preferita.

Entrambi i profili sono però respinti dal gate swap. Il v3 raggiunge circa
487 MiB di swap WSL nonostante l'unload preventivo e la RAM disponibile. M36
resta aperta con stato `residency_gate_rejected_swap_growth`; prima di una
nuova run occorre congelare un reference profile WSL distinto. Non sono
autorizzati tuning di modello/protocollo o rilassamenti retroattivi dei gate.

## Checkpoint residency v4 — 2026-09-05

Il reference profile con `.wslconfig` `swap=0` è congelato in
`docs/milestone-36-residency-profile-v4.yaml` e qualificato dal report
`docs/reports/milestone-36-residency-runs-v4.json`. Il verdetto è
`dual_model_residency_qualified`.

Le 18 richieste dei tre cicli `chat -> mutation -> chat` completano con modello
osservato uguale al modello richiesto, output corretto e latenza sotto soglia.
Il pattern di residency è identico in ogni ciclo: un solo modello residente
dopo ciascuna completion. Non sono osservati fallback incrociati, offload CPU,
OOM, restart del provider, transizioni fallite o crescita dello swap WSL.

La strategia productizzabile è handoff esplicito, non residenza simultanea:
unload del profilo uscente, attesa di `/api/ps` vuoto e generazione con il
profilo entrante. M36 prosegue sui gate di packaging, installazione fuori
checkout, suite/race/vet e decisione release separata.

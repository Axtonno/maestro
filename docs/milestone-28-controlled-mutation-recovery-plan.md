# Milestone 28 — Controlled Mutation Recovery

Linea di versione candidata: v0.5.0

Stato: Aperta — `mutation_recovery_not_yet_qualified`

Data: 2026-09-03

Prerequisito: M27 chiusa con `v0.4.0_released_and_verified`.

## Decisione di priorità

La validazione sul campo post-release di v0.4.0 può continuare informalmente,
usando esclusivamente l'asset pubblico e senza modificare il relativo support
claim. Il prossimo lavoro principale è il recupero di Controlled Mutation.

Il feedback operativo sposta la priorità dalla sola osservazione read-only a
una capacità mutativa piccola e verificabile. Controlled Mutation è una nuova
capability pubblica e non può entrare in una patch release: la destinazione
naturale è v0.5.0.

## Claim candidato

> Maestro può proporre e applicare, dopo approvazione esplicita, una modifica
> atomica a un singolo file all'interno del workspace autorizzato.

Il claim non comprende modifica autonoma, più file, create/delete/rename,
shell, Git, processi, agent autonomi o recupero euristico di patch ambigue.
Fino al superamento di tutti i gate, v0.4.0 resta l'unica release supportata e
rimane esclusivamente read-only.

## Flusso normativo

```text
richiesta esplicita di modifica
    -> lettura autorevole del singolo file indicato
    -> proposta non fidata del modello
    -> parsing e validazione deterministica
    -> preview completa del diff e fingerprint
    -> approvazione esplicita dell'utente
       |-> rifiuta: nessuna modifica
       `-> approva: controllo file invariato
                     -> applicazione atomica
                     -> verifica del file e del diff finale
                     -> risultato verificabile
```

L'approvazione autorizza una sola applicazione dell'esatto oggetto mostrato in
preview. Ogni nuova proposta, variazione del diff, stato stale o secondo
tentativo richiede una nuova approvazione.

## Perimetro iniziale

- un solo file regolare esistente, indicato esplicitamente dall'utente;
- una richiesta esplicita di modifica, distinta da richieste consultive;
- produzione di una patch o proposta strutturata, mai riscrittura opaca;
- read autorevole non troncata e digest pre-modifica;
- patch completa, bounded e deterministicamente validabile;
- preview integrale del diff prima dell'approvazione;
- approvazione interattiva obbligatoria e legata al fingerprint;
- verifica anti-stale immediatamente prima del commit;
- applicazione atomica all-or-nothing;
- containment fisico nel workspace, inclusi controlli sui symlink;
- verifica post-apply di digest, contenuto e diff finale;
- massimo un effetto per approvazione.

Sono esclusi `.env`, credenziali, secret, file generati, file non regolari,
path assoluti o con traversal e ogni target definito dalla policy come
sensibile. Una richiesta multi-file fallisce chiusa prima di qualsiasi effetto.

## Due trasporti da confrontare

La capacità non deve dipendere rigidamente dalla precisione del tool calling
dei modelli locali. La milestone confronta, sullo stesso task set congelato:

1. tool calling nativo con schema minimo;
2. output strutturato vincolato contenente una proposta di patch.

Entrambi i trasporti terminano nello stesso compilatore deterministico e nello
stesso oggetto immutabile di applicazione. Il modello non possiede autorità sul
filesystem: propone soltanto dati non fidati. Maestro analizza, respinge output
ambiguo, verifica path e digest, costruisce il diff, raccoglie l'approvazione e
applica atomicamente.

Il confronto misura validità sintattica, correttezza semantica, terminali,
latenza e failure class. Non è ammesso scegliere il trasporto vincente dopo
aver cambiato task, prompt, schema, modello o oracolo. Un fallback fra i due
trasporti nella stessa run è vietato: maschererebbe il failure originario.

## Contratto della proposta

La rappresentazione candidata minima è strict, versionata e bounded:

```json
{
  "version": 1,
  "path": "relative/path.ext",
  "operation": "replace",
  "old_text": "testo esatto e unico",
  "new_text": "sostituzione proposta"
}
```

Il compilatore:

- rifiuta campi mancanti, sconosciuti, duplicati o tipi errati;
- richiede che `path` coincida con la read autorevole della stessa run;
- accetta inizialmente soltanto `replace` esatto e una singola occorrenza;
- rifiuta no-op, matching fuzzy, encoding invalido, overflow e output misto
  incompatibile;
- deriva autonomamente contenuto risultante, diff, digest e fingerprint;
- non corregge, completa o interpreta euristicamente l'output del modello;
- consegna a preview, approval ed apply lo stesso oggetto immutabile.

## Approvazione e applicazione

La preview mostra path logico, diff completo, digest iniziale e fingerprint
dell'azione. L'approvazione deve provenire da un TTY valido, essere esplicita,
valere una volta sola e non essere sintetizzabile tramite EOF, pipe o default.

Dopo l'approvazione Maestro riapre il file in modo contained e verifica che
identità fisica, tipo, digest e contenuto coincidano con la read autorizzata.
Solo allora scrive un file temporaneo nello stesso filesystem, ne applica i
permessi previsti e usa un commit atomico. Qualsiasi failure precedente al
commit lascia il workspace byte-identico; non sono ammesse patch parziali.

Il risultato terminale riporta esclusivamente identità redatte e verificabili:
stato, path logico, digest pre/post, fingerprint e corrispondenza del diff.
Non espone contenuti, prompt, credenziali o path fisici nei normali log.

## Casi negativi obbligatori

- percorso fuori dal workspace o traversal;
- symlink verso l'esterno;
- file cambiato tra proposta e approvazione;
- patch malformata o schema non valido;
- patch applicabile soltanto parzialmente;
- richiesta o proposta multi-file;
- rifiuto, EOF, input non interattivo o approval non valida;
- terminale provider `length`;
- output con testo e patch incompatibili;
- tentativo su `.env`, credenziali, secret o file esclusi;
- `old_text` assente o presente più volte;
- replay dell'approvazione o secondo effetto;
- cancellazione o fault filesystem prima del commit.

Ogni caso negativo deve terminare con zero effetti e workspace identico.

## Gate minimi

```yaml
patch_valida: 1.00
modifica_semanticamente_corretta_minima: 0.80
applicazioni_senza_approvazione_massime: 0
modifiche_fuori_scope_massime: 0
file_aggiuntivi_modificati_massimi: 0
patch_parziali_applicate_massime: 0
stale_write_accettati_massimi: 0
failure_con_workspace_alterato_massimi: 0
```

Sono inoltre obbligatori:

- matrice deterministica positiva e negativa interamente verde;
- preview, apply e risultato legati allo stesso fingerprint;
- zero terminali `length` nelle run accettate;
- zero falsità materiali nelle modifiche promosse;
- zero leak di contenuto o credenziali in diagnostica ed eventi;
- suite completa, race detector, vet e `git diff --check` verdi;
- installazione e gate live fuori checkout sull'eventuale candidate v0.5.0.

## Sequenza di lavoro

1. congelare threat model, stato macchina e reason code;
2. definire schema strict e compilatore deterministico senza provider;
3. implementare preview, fingerprint e approval one-shot;
4. irrobustire containment, anti-stale e commit atomico;
5. completare la matrice deterministica, inclusi tutti i casi negativi;
6. congelare modello, prompt, fixture, task e i due trasporti;
7. confrontare i trasporti con una sola run per task e senza fallback;
8. qualificare semanticamente il trasporto candidato su casi nuovi;
9. eseguire gate end-to-end con approval reale e failure injection;
10. produrre un candidate v0.5.0 soltanto se ogni soglia è verde.

Una modifica a schema, prompt, compiler, validator, policy, modello o profilo
invalida il candidate e richiede un nuovo freeze. I failure non autorizzano
retry qualitativi o rilassamento delle soglie.

## Verdetti

| Verdetto | Significato |
|---|---|
| `controlled_mutation_candidate_qualified` | tutti i gate verdi; è autorizzata una successiva milestone di release v0.5.0 |
| `controlled_mutation_transport_unresolved` | sicurezza deterministica verde, ma nessun trasporto raggiunge la soglia semantica |
| `controlled_mutation_rejected` | fallisce un gate di integrità, approval, containment, atomicità o sicurezza |

M28 non pubblica automaticamente v0.5.0. La pubblicazione richiede una distinta
release readiness con package riproducibile, installazione pulita, live gate e
verifica degli asset pubblici.

## Output attesi

- matrice `milestone-28-controlled-mutation-recovery-matrix.yaml`;
- ADR del protocollo e del trasporto selezionato;
- schema versionato della proposta;
- suite deterministica positiva e negativa;
- report comparativo dei due trasporti;
- report live redatto e verdetto finale;
- handoff esplicito alla release readiness v0.5.0 oppure decisione di rinvio.

# Milestone 26 — Response Validity & Field Quality Recovery

Versione di partenza: v0.3.1

Stato: Aperta — diagnosi iniziale, nessun candidate

Data: 2026-09-02

Prerequisito: M25 chiusa con `field_adoption_negative`.

## Obiettivo

Recuperare affidabilità e disciplina epistemica di Direct Chat senza ampliare
le capability. Prima di cambiare prompt o profilo, la milestone deve spiegare
i quattro `response_invalid` osservati in M25 e distinguere difetti di
validator, adapter, contratto provider e output del modello.

Restano fuori scope CPU qualification, altri modelli, multi-file, agent,
retrieval e Controlled Mutation.

## Evidenza disponibile e limite del replay

M25 conserva manifest, stdout, stderr e tempi, ma non i body HTTP grezzi di
Ollama. Nei quattro fallimenti stdout è vuoto e stderr contiene soltanto il
reason code redatto, con heartbeat ove previsto. Anche i log del servizio
Ollama non contengono i payload.

Di conseguenza non è possibile attribuire retroattivamente, per quei quattro
eventi, HTTP, `message.content`, `message.thinking`, `done_reason`, dimensione
o usabilità semantica del body. Un replay offline fedele richiede i byte
originali e non può essere ricostruito con una nuova generazione.

Il codice restringe comunque il punto di rifiuto: `response_invalid` viene
emesso dal validator Direct Chat dopo che l'adapter ha restituito una
completion. La regola aggregata rifiuta ruolo diverso da `assistant`, tool
call, finish reason diverso da `stop`, content vuoto/whitespace, UTF-8 non
valido, NUL, model mismatch, usage negativa o durata negativa. L'adapter
decodifica `message.thinking` ma non lo propaga nel tipo normalizzato; un body
con thinking presente e content vuoto è quindi una causa possibile, non ancora
dimostrata.

## Sequenza vincolante

1. aggiungere un harness diagnostico locale che catturi status HTTP e body
   grezzo con permessi `0600`, fuori dal repository e senza log applicativi;
2. validare il replay offline su fixture sintetiche per ciascuna regola;
3. eseguire una nuova matrice diagnostica separata, una sola volta, sui quattro
   input congelati di M25; non chiamarla replay M25;
4. classificare ogni payload per content, thinking, whitespace, byte,
   `done`, `done_reason`, terminale e contenuto semanticamente utilizzabile;
5. rendere il validator diagnostico internamente per regola, mantenendo
   redazione e compatibilità del reason code pubblico;
6. correggere adapter o validator soltanto se l'evidenza dimostra contenuto
   valido rifiutato;
7. definire stderr: heartbeat durante generation e un solo terminale redatto
   sono entrambi consentiti, con canali e forme esplicite;
8. conservare il ticker ancorato alla sola finestra di generation già
   dimostrata, aggiungendo test temporali di preflight e stop;
9. progettare un contratto epistemico con sezioni `Fatti osservati`,
   `Inferenze possibili`, `Informazioni non determinabili`;
10. se cambia il prompt, creare un candidate v0.4.0 separato; altrimenti un
    fix di adapter/validator/diagnostica può candidarsi come v0.3.2;
11. eseguire una matrice pre-release che include tutti i task reali M25 prima
    di qualsiasi pubblicazione;
12. ripetere Field Adoption sul candidate una sola volta.

## Classificazione causale

| Evidenza | Diagnosi |
|---|---|
| content valido rifiutato | difetto software/validator |
| content vuoto e thinking vuoto | instabilità del modello |
| contenuto utile solo in thinking o altro campo | difetto adapter/contratto di normalizzazione |
| `done` o `done_reason` incoerenti | contratto provider |
| contenuto correttamente non pubblicabile | limite modello/profilo |

## Gate pre-release

Nessuna nuova versione viene pubblicata finché la matrice reale non raggiunge:

- completion almeno 85%;
- correct almeno 80% delle risposte valutabili;
- `response_invalid` zero;
- falsità materiali zero;
- utilità mediana almeno 4/5;
- zero mutazioni, leak o ampliamenti di autorità.

## Output attesi

- matrice diagnostica `milestone-26-response-validity-field-quality-recovery-matrix.yaml`;
- classificazione redatta dei quattro nuovi capture;
- test offline per ogni ramo del validator;
- decisione motivata v0.3.2 oppure v0.4.0;
- candidate e nuova Field Adoption soltanto dopo i gate diagnostici.

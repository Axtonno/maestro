# Milestone 15 — Report finale

Data: 2026-08-28

Stato: **COMPLETATA**

Esito: **`verified_agent_rejected`**

La Milestone 15 qualifica la piattaforma WSL2/Ubuntu 24.04/RTX 5070 e la
modalità read-only `direct/chat`. Non qualifica il verified agent, il baseline
Laravel multi-file B01 o Controlled Mutation. La stop rule è stata applicata
correttamente e non è stata prodotta alcuna release.

## Esito complessivo

| Asse | Esito |
|---|---|
| WSL2 / Ubuntu 24.04 / RTX 5070 | **PASS** |
| Provider Ollama e offload GPU | **PASS** |
| `direct/chat` con `qwen2.5-coder:7b` | **PASS** |
| Streaming e non-streaming | **PASS** |
| Invarianza del workspace nei test direct/chat | **PASS** |
| Verified agent con `qwen3.5:9b` | **FAIL** |
| B01 Laravel multi-file | **NOT_RUN** |
| Controlled Mutation | **NON QUALIFICATA** |
| Milestone 16 mutativa | **CLOSED / NON AUTORIZZATA** |

## Piattaforma e provider

La nuova piattaforma ha superato i controlli ambientali e operativi:

- Windows 11 Pro con WSL2 e Ubuntu 24.04.4 LTS;
- workspace su ext4 sotto `/home`;
- NVIDIA GeForce RTX 5070 con 12.227 MiB di VRAM;
- Ollama 0.33.1 raggiungibile tramite API loopback;
- doctor 10/10, suite normali e development, race detector e `go vet` verdi;
- `100% GPU` osservato con context 4096 e 8192;
- nessun OOM, reset del provider o fallback CPU.

L'evidenza completa della piattaforma, dei modelli e dei digest è conservata
nel [report di Fase 1](milestone-15-phase-1.md).

Questi risultati indicano che il limite prestazionale osservato sul precedente
ThinkPad era principalmente hardware. La mancata progressione del verified
agent, invece, non può essere attribuita principalmente a quella macchina.

## Qualificazione `direct/chat`

Il candidate `qwen2.5-coder:7b`, con context 4096 e thinking disabilitato, ha
superato il protocollo congelato:

| Gate | Risultato |
|---|---|
| C0 senza file | PASS 3/3; contesto insufficiente dichiarato e nessun endpoint inventato |
| C1 single-file | PASS 3/3; `POST /orders` e `OrderController::store`, senza endpoint aggiuntivi |
| C2 streaming/non-streaming | PASS 2/2; ground truth e terminali equivalenti |
| C3 operatività | PASS; `completed`, `stop`, latenze warm 275–3.936 ms |
| C4 sicurezza | PASS; zero tool, retrieval, fallback o mutazioni |

Il digest redatto della fixture è rimasto invariato. Il risultato qualifica
quindi una capacità read-only concreta, ripetibile ed esplicitamente separata
dal percorso agentico. Configurazione, digest e osservazioni sono riportati nel
[report di Fase 2](milestone-15-phase-2.md).

## Verified agent

Il verified agent `qwen3.5:9b`, con context 8192 e thinking predefinito, non ha
superato la prima progressione live del gate sintetico. La read bootstrap
verificata ha coperto la route; nello stesso turno il modello ha poi emesso una
seconda tool call e il runtime ha terminato con `tool_failure`, rappresentato
dalla CLI come `execution_failed`.

| Campo | Valore osservato |
|---|---|
| Terminale | `execution_failed` / `tool_failure` |
| Durata | 31.092 ms |
| Turni modello / tool call | 1 / 2 |
| Token input / output | 3.177 / 225 |
| GPU / context | 100% GPU / 8192 |
| Workspace | invariato |

Il file richiesto dal prompt,
`app/Http/Controllers/OrderController.php`, esisteva nella fixture Laravel
incorporata. Per ogni tentativo la fixture veniva materializzata in una nuova
directory temporanea, impostata come root effettiva del workspace e verificata
tramite digest prima dell'avvio. Il percorso logico era quindi valido e
raggiungibile; il controller mostrava che `store()` invoca
`OrderService::create()`.

Non sono stati osservati OOM, fallback CPU, reset del provider, timeout o
mutazioni del workspace. Il fallimento non dimostra pertanto un problema
hardware e non è spiegato dall'assenza del controller.

### Limite dell'evidenza disponibile

`tool_failure` è la categoria terminale, non la causa concreta. Gli artefatti
pubblicabili applicano la redazione prevista dal piano e non conservano nome,
argomenti ed errore interno della seconda tool call. Non è quindi possibile
stabilire retroattivamente se il failure sia stato prodotto da:

- una chiamata malformata o un percorso alterato dal modello;
- una chiamata valida respinta dal contratto dello strumento;
- un errore del tool runtime, del filesystem o dell'adapter;
- un'interazione difettosa con la choreography/state machine.

Una diagnosi esatta richiederebbe una nuova esecuzione development-only con un
record forense locale e protetto della tool call e dell'errore originario. Tale
diagnosi non modifica il verdetto, non riapre la milestone e non autorizza B01.
Il dettaglio osservabile del run è nel [report di Fase 3](milestone-15-phase-3.md).

## Stop rule e conseguenze

Il gate del verified agent richiedeva due progressioni complete. Dopo il primo
failure il requisito 2/2 non era più raggiungibile senza un retry opportunistico
vietato dal piano. Di conseguenza:

- B01 e la matrice Laravel multi-file sono `NOT_RUN`, come registrato nel
  [report di Fase 4](milestone-15-phase-4.md);
- il baseline read-only multi-file non è qualificato;
- Controlled Mutation resta non supportata;
- la Milestone 16 mutativa resta `CLOSED` e non autorizzata;
- `v0.3.0` non viene costruita, taggata o pubblicata;
- `v0.2.0` conserva esclusivamente il proprio perimetro storico.

La Milestone 15 è quindi conclusa con un esito negativo ammesso dal piano, non
con una release.

## Decisione di prodotto e direzione successiva

La sequenza originaria

```text
nuovo hardware → verified agent → Controlled Mutation → v0.3.0
```

non è autorizzata dall'evidenza raccolta. È però emersa una sequenza distinta e
già qualificata:

```text
nuovo hardware → direct/chat PASS → productization read-only
```

La direzione raccomandata è:

1. mantenere chiusa la Milestone 16 mutativa;
2. mantenere verified agent e Controlled Mutation come capacità sperimentali;
3. aprire una milestone separata per la productization di `direct/chat`;
4. distribuire la modalità diretta come funzionalità esplicita, senza fallback
   agentico implicito;
5. riprendere verified agent e mutazione solo con un nuovo candidate oppure
   dopo avere dimostrato e corretto una causa software specifica.

L'eventuale assegnazione di `direct/chat` a `v0.3.0` deve essere oggetto di una
decisione formale di roadmap, perché cambierebbe il perimetro precedentemente
associato a quella versione.

## Verdetto finale

La nuova macchina ha risolto il vincolo prestazionale, ma non quello agentico.
La Milestone 15 non autorizza Controlled Mutation o la precedente Milestone 16;
qualifica però `direct/chat` come capacità read-only utilizzabile e come base
concreta per la prossima decisione di prodotto.

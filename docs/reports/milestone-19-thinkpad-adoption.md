# Milestone 19 — ThinkPad adoption e lower bound

Data: 2026-08-29

Stato: **COMPLETATA**

Verdetto: **`operationally_impractical`**

## Sintesi

L'esatto asset pubblico v0.3.0 è installabile e funzionale sul ThinkPad
CPU-only. `doctor chat` passa e tutte le risposte ottenute sono corrette
rispetto al singolo file fornito. L'esperienza quotidiana non è però pratica:
soltanto 3/5 domande single-file completano, due raggiungono il deadline di 5
minuti e la mediana delle completion single-file è 91,4 secondi.

Questo risultato è un'osservazione post-release sul lower bound, non una
qualification. Il ThinkPad non diventa reference hardware e il support claim
v0.3.0 resta invariato.

## Installazione e preflight

Archive e checksum sono stati scaricati dalla GitHub Release pubblica in una
directory nuova, non copiati da `dist/` e non ricostruiti dal checkout.

| Controllo | Esito |
|---|---|
| checksum archive | PASS — `6c8f0e883ec8f8c05571fc2e7bc1f4ecac608c2bd7e338395ae0a4253fff1aaf` |
| dimensione archive | PASS — 3.775.317 byte |
| allowlist/path archive | PASS — nessun path assoluto, traversal o symlink |
| manifest | PASS — v0.3.0, release, commit `3f4c7d4…` |
| SHA-256 binario installato | PASS — `378a0533083b9a00be6c0212ca52001cebc5f77b476a20038bc8e08d1fc3d42d` |
| installazione | PASS — prefix isolato sotto `/tmp`, fuori checkout |
| provider/modello | PASS — Ollama loopback e digest qualificato |
| `doctor --mode chat` | PASS 5/5 in 0,57 s |

L'installazione isolata non ha sostituito il precedente comando trovato nel
`PATH`; tutte le prove hanno invocato esplicitamente il binario v0.3.0
verificato.

## Risultati redatti

| Caso | Ambito | Modalità | Terminale | Durata | Token in/out | Qualità |
|---|---|---|---|---:|---:|---|
| `M19-C0` | nessun file | complete | completed | 49,9 s | 394 / 65 | correct |
| `M19-C1` | route, 1.093 byte | complete | completed | 164,9 s | 721 / 288 | correct |
| `M19-C2` | controller | complete | deadline_exceeded | 300,1 s | non pubblicati | unevaluable |
| `M19-C3` | service, 1.173 byte | complete | deadline_exceeded | 300,1 s | non pubblicati | unevaluable |
| `M19-C4` | service, 396 byte | complete | completed | 91,4 s | 577 / 186 | correct |
| `M19-C5` | model, 540 byte | stream | completed | 70,6 s | 599 / 98 | correct |

La matrice single-file ha completion rate 60% (3/5). La qualità è 3/3
`correct` sulle risposte disponibili; non esiste evidenza per classificare le
due run scadute. La mediana delle tre completion single-file è 91,4 secondi;
il caso più lento riuscito richiede 164,9 secondi.

Durante `M19-C5`, stdout e stderr erano ancora entrambi vuoti dopo circa 60
secondi. `--stream` completa correttamente ma, per il contratto di output
atomico, non riduce la latenza percepita prima del terminale.

## Workspace e sicurezza

`project-a` era Git-clean prima delle prove e resta Git-clean dopo. Il digest
aggregato di tutti i file tracciati coincide pre/post:

```text
5ff49977a0ca987035643a28f35f39a77a9738ea0bf4dd3341dca66cc7d96963
```

Le evidenze grezze locali hanno permessi `0600`. Nessun path fisico, remote,
prompt, risposta completa o contenuto del progetto è incluso nel report.
Nessun tool, write, patch, retrieval, agent o accesso Docker è stato concesso.

## Problemi d'uso osservati

1. Errori di configurazione strict differenti vengono ridotti a
   `invalid_request`, senza indicare il campo non valido; il setup di un
   workspace reale richiede quindi troubleshooting manuale.
2. Un precedente binario nel `PATH` può mascherare l'installazione isolata;
   l'identità va verificata e il percorso di invocazione reso esplicito.
3. Il tempo base è alto anche su file da 396–540 byte: 70–91 secondi per una
   risposta breve.
4. File reali più impegnativi possono consumare l'intero timeout di 5 minuti
   senza risultato utilizzabile.
5. Lo streaming atomico protegge da output parziale ma non offre feedback
   applicativo durante le attese lunghe.

## Decisione

Il lower bound osservato non giustifica l'apertura immediata di nuove
capability. Restano fermi verified agent, multi-file, Controlled Mutation,
nuovi provider e altri modelli ufficialmente supportati.

I miglioramenti con maggior valore da valutare in una futura decisione sono
diagnostica di configurazione più precisa, chiarezza dell'identità/PATH,
feedback di progresso non sensibile e riduzione misurabile della latenza sul
profilo già supportato. Nessuno di questi cambi viene implementato o promesso
dalla Milestone 19.

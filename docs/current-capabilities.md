# Capacità correnti di Maestro v0.5.0

Aggiornata: 2026-09-08

Questa è la pagina verità della release pubblica. Se un esempio, un documento
storico o del codice sperimentale sembra promettere di più, prevale il
perimetro descritto qui.

Maestro v0.5.0 è qualificato sull'asset Linux `amd64`, incluso un ciclo di
adozione su Linux nativo CPU-only. La prova nativa M38 è stata eseguita su un
ThinkPad T490s specifico: non equivale a una garanzia universale per ogni PC
Linux.

## Cosa funziona oggi

| Capacità | Perimetro supportato |
| --- | --- |
| Direct Chat | Domanda senza file oppure su un solo file scelto esplicitamente |
| Controlled Mutation | Una sostituzione su un solo intervallo di un file PHP regolare sotto `app/` |
| Autorizzazione | Preview completa e approvazione allow-once obbligatoria in una TTY reale |
| Integrità | Controllo stale, fingerprint della preview e sostituzione atomica |
| Provider | Ollama 0.33.1 locale su `127.0.0.1:11434` |
| Artifact | `maestro-v0.5.0-linux-amd64.tar.gz` |

Controlled Mutation è supportato soltanto nel
[perimetro testato](controlled-mutation-support.md). Non individua il target e
non scrive senza approvazione esplicita dell'utente.

## Modelli qualificati

| Uso | Modello | Digest richiesto |
| --- | --- | --- |
| Chat | `qwen3.5:9b` | `6488c96fa5faab64bb65cbd30d4289e20e6130ef535a93ef9a49f42eda893ea7` |
| Mutation | `qwen2.5-coder:14b` | `9ec8897f747e246e970bc5cfdda85d22f1123dc2e3d34978a010a75968716849` |

Altri tag, digest o provider non fanno parte del support claim v0.5.0. Maestro
non scarica i modelli e si arresta se l'identità non coincide.

## Hardware consigliato

Come punto di partenza pratico, non come minimo universale qualificato:

- Linux `amd64` nativo;
- CPU x86-64 con almeno 4 core / 8 thread;
- 16 GiB di RAM;
- circa 20 GiB liberi per archive, modelli e margine operativo;
- swap disponibile; GPU discreta opzionale.

M38 ha completato i gate su Ubuntu 24.04.4, ThinkPad T490s, Intel i5-8365U,
16,4 GB di RAM e inferenza CPU-only. L'ambiente WSL2 con RTX 5070 resta il
riferimento più veloce della qualifica di release. Latenza e memoria cambiano
sensibilmente con l'hardware.

## Sperimentale, non parte del prodotto v0.5.0

- componenti agent, context e tool presenti nel repository;
- streaming nei profili storici, disabilitato nel profilo v4 distribuito;
- profili di packaging diversi da quello mutation-productization;
- design e milestone future descritti nella roadmap.

La presenza del codice non costituisce una promessa operativa.

## Non supportato oggi

- modifiche multi-file, più intervalli o sequenze autonome;
- scelta autonoma di file, righe o azioni;
- agent autonomi, retrieval multi-file e tool calling come prodotto;
- Windows nativo, macOS e Linux `arm64`;
- provider alternativi, endpoint remoti o modelli diversi dai due digest;
- mutation fuori `app/`, su linguaggi non PHP o senza TTY;
- esecuzione di shell, Git, Docker, Composer o test da parte del modello;
- sandbox, rollback automatico o garanzia di correttezza semantica.

## Difetto documentale dell'archive v0.5.0

L'archive pubblico, già pubblicato e quindi immutabile, contiene un paragrafo
legacy con posizionamento schema v3 e non rende abbastanza evidente il nome
reale del profilo distribuito. Il file operativo corretto è
`configs/maestro.v0.5.0-candidate.yaml`. Il difetto non interessa binario,
manifest o configurazione ed è corretto nei sorgenti per le release
successive; v0.5.0 non viene ripubblicata o sovrascritta.

## Da dove iniziare

Seguire [Installa e prova](install-and-try.md), quindi consultare la
[Compatibility Matrix](compatibility.md), il
[Security Model](security-model.md) e le [evidenze M38](reports/milestone-38-final.md).
Le capacità future sono separate nella [Roadmap](roadmap.md).


# Maestro v0.5.0 Compatibility Matrix

Aggiornata: 2026-09-09

Questa pagina dettaglia il support claim corrente. Funzioni presenti nel
repository, documenti di design e milestone future non ampliano questo
perimetro.

La sintesi autorevole è [Capacità correnti](current-capabilities.md).

## Superficie operativa

| Dimensione | Stato | Confine verificato |
| --- | --- | --- |
| Artifact | Supportato | Release pubblica Linux `amd64` v0.5.0 |
| Provider | Supportato | Ollama 0.33.1 su `http://127.0.0.1:11434` |
| Direct Chat | Supportata | Domanda senza file o su un file esplicito |
| Modello chat | Supportato | `qwen3.5:9b`, digest `6488c96fa5faab64bb65cbd30d4289e20e6130ef535a93ef9a49f42eda893ea7` |
| Controlled Mutation | Supportata, opt-in | Una sostituzione su un intervallo PHP esplicito sotto `app/` |
| Modello mutation | Supportato | `qwen2.5-coder:14b`, digest `9ec8897f747e246e970bc5cfdda85d22f1123dc2e3d34978a010a75968716849` |
| Configurazione | Supportata | Schema strict v4 distribuito |
| Approval | Obbligatoria | Preview completa e allow-once in TTY reale |
| Streaming chat | Compatibilità | Implementato nei profili storici; il profilo v4 distribuito lo disabilita |
| Isolamento | Non fornito | Processo trusted con privilegi dell'utente, nessuna sandbox |

## Piattaforme qualificate e osservate

| Ambiente | Stato | Significato |
| --- | --- | --- |
| Windows → WSL2 → filesystem Linux, RTX 5070 12 GB | Qualifica release M36–M37 | Ambiente esatto di productization e pubblicazione; non equivale a supporto Windows nativo o WSL2 universale |
| Ubuntu 24.04.4 nativo, ThinkPad T490s, i5-8365U CPU-only | Field adoption M38 | Asset pubblico verificato sul campo, inclusi chat, mutation e residency |
| Altri PC Linux `amd64` | Non qualificati individualmente | L'artifact può essere eseguito, ma non esiste una promessa universale di latenza o memoria |
| Linux `arm64`, macOS, Windows nativo | Non supportati | Nessun artifact o gate v0.5.0 qualificato |

CPU-only sul T490s è un ambiente osservato e qualificato per M38, non un
requisito minimo. Ollama ha riportato zero VRAM e un solo modello residente per
volta; le latenze sono sensibilmente dipendenti dall'hardware.

La compatibilità futura viene valutata per dimensioni indipendenti, non con la
sola ricompilazione del binario:

| Dimensione | Valori da qualificare separatamente |
| --- | --- |
| Architettura CPU | `amd64`, `arm64` |
| Sistema operativo | Linux, Windows, macOS |
| Accelerazione | CPU-only, CUDA, Metal, ROCm |
| Risorse | RAM, swap, memoria disponibile e VRAM quando applicabile |
| Provider runtime | Ollama, llama.cpp/GGUF, altri adapter |

M38 qualifica precisamente Linux `amd64` nativo, CPU-only, sul ThinkPad T490s
registrato. Non qualifica automaticamente Linux ARM, Windows nativo, macOS
Intel o Apple Silicon, GPU NVIDIA generiche, AMD/ROCm, llama.cpp oppure WSL2
come target pubblico definitivo. I target futuri e la loro priorità sono
esplicitati nella [Roadmap](roadmap.md#direzione-post-m41-profili-e-piattaforme).

## Direct Chat

Percorso supportato:

```text
domanda + zero o un file esplicito
  -> validazione workspace e limiti
  -> qwen3.5:9b
  -> risposta validata
```

Senza file non viene eseguita ricerca nel workspace. Con `--file` viene
reso disponibile al provider soltanto il file indicato. Direct Chat non abilita tool,
retrieval, agent o mutation come fallback.

## Controlled Mutation

Percorso supportato:

```text
file PHP sotto app/ + righe start:end + istruzione
  -> qwen2.5-coder:14b
  -> proposta host-bound
  -> preview + approval allow-once
  -> stale check + apply atomico
```

Il perimetro normativo completo è in
[Controlled Mutation: perimetro supportato](controlled-mutation-support.md).

## Non supportato oggi

- individuazione autonoma del target;
- retrieval, contesto o modifiche multi-file;
- insert/delete arbitrari e modifiche fuori dallo span selezionato;
- mutation fuori `app/`, non PHP, senza TTY o senza allow-once;
- altri modelli, digest o prompt/schema mutativi;
- llama.cpp ed endpoint Ollama remoti;
- agent autonomi, tool calling, shell, Git, Docker o remote execution;
- sandbox, isolamento di rete, secret manager e rollback automatico;
- plugin o tool di terze parti come superficie di prodotto.

“Non supportato” non significa impossibile nell'architettura: significa che
v0.5.0 non offre una promessa operativa per quel percorso.

## Evidenza

M38 ha verificato l'asset pubblico senza checkout o rebuild su Linux nativo:
doctor 14/14, target e preview 8/8, sei apply esatti, deny e stale senza
scritture, correttezza semantica 11/12 e completion 12/12. L'errore semantico
F02 resta documentato.

CLI e schema restano versionati e possono evolvere durante la serie 0.x. Le
capacità future sono elencate nella [Roadmap](roadmap.md), non in questa
matrice.

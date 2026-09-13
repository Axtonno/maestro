# Piattaforme supportate

Aggiornata: 2026-09-12

| Ambiente | Stato pubblico |
| --- | --- |
| Linux `amd64`, Ollama 0.33.1 | Supportato nel perimetro v0.5.0 |
| Linux `amd64` CPU-only | Qualificato su hardware di riferimento; prestazioni variabili |
| WSL2 su filesystem Linux | Trial post-v0.5 qualificato sull'ambiente di riferimento; non supporto Windows nativo |
| Windows nativo | Non supportato; in valutazione |
| macOS Intel o Apple Silicon | Non supportato; in valutazione |
| Linux `arm64` | Non supportato |

## Confine del supporto

La compatibilità non deriva dalla sola compilazione del binario. Sistema
operativo, architettura, accelerazione, memoria, versione del provider e
identità del modello devono essere qualificati insieme.

Per v0.5.0 il support claim comprende:

- artifact Linux `amd64`;
- Ollama 0.33.1 su `http://127.0.0.1:11434`;
- `qwen3.5:9b` per Direct Chat;
- `qwen2.5-coder:14b` per Controlled Mutation;
- configurazione strict v4 distribuita.

CPU-only è stato verificato su Ubuntu 24.04.4, ThinkPad T490s, Intel
i5-8365U e 16 GiB di RAM. È un ambiente osservato, non una garanzia di latenza
uniforme su ogni macchina Linux.

Il trial post-v0.5 ha verificato il packaging candidate su Windows 11, WSL2 e
Ubuntu 24.04 su filesystem Linux, con Ollama 0.33.1 eseguito dentro la stessa
distro. Clean install, setup idempotente, chat, preview, deny, allow-once,
stale e confini dei path hanno superato il gate. Il candidate non è una
release pubblica: questa evidenza non amplia il support claim v0.5.0 e non vale
per esecuzione o filesystem Windows nativi.

Il perimetro funzionale completo è nella pagina [Capacità correnti](current-capabilities.md).

## Trial llama.cpp / GGUF post-v0.5

M46 ha qualificato su Windows 11 `amd64` il profilo single-model dichiarato in
`configs/maestro.llamacpp-gguf.example.yaml`: `llama-server` build
`b1-0f3a71be1`, GGUF SHA-256 `ac9bc7a…b1ed`, alias
`qwen2.5-coder:14b`, context window 4096 e endpoint loopback. Doctor, chat,
preview, deny, allow-once e stale sono stati provati senza fallback o cambio
modello. È un candidate locale post-release: non modifica la matrice v0.5.0 e
non qualifica altre build, altri GGUF, router mode o endpoint remoti.

# Piattaforme supportate

Aggiornata: 2026-09-09

| Ambiente | Stato pubblico |
| --- | --- |
| Linux `amd64`, Ollama 0.33.1 | Supportato nel perimetro v0.5.0 |
| Linux `amd64` CPU-only | Qualificato su hardware di riferimento; prestazioni variabili |
| WSL2 su filesystem Linux | Ambiente di qualifica, non supporto Windows nativo |
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

Il perimetro funzionale completo è nella pagina [Capacità correnti](current-capabilities.md).

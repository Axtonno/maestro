# Milestone 15 — Fase 1: piattaforma, provider e GPU

Data: 2026-08-28

Stato: **PASS**

## Piattaforma congelata

| Campo | Valore osservato |
|---|---|
| host | Windows 11 Pro, build 26200 |
| topologia | WSL2, Ubuntu 24.04.4 LTS, Linux `amd64` |
| kernel guest | `6.18.33.2-microsoft-standard-WSL2` |
| RAM host | 31,15 GiB osservati |
| RAM / swap guest | 15 GiB / 4 GiB effettivi |
| filesystem workspace | ext4 sotto `/home` |
| GPU | NVIDIA GeForce RTX 5070, 12.227 MiB VRAM |
| driver host / CUDA | 596.36 / 13.2 |
| Ollama | 0.33.1, systemd dentro WSL2, API loopback |
| Go | 1.24.5 linux/amd64, checksum ufficiale verificato |

Il checkout originario non conteneva metadata `.git`; l'identità sorgente
usata per il candidate è stata congelata tramite digest redatto dell'albero.

## Provider e GPU

- `qwen2.5-coder:7b`: ID catalogo `dae161e27b0e`, Q4_K_M, blob principale
  `60e05f2100071479f596b964f89f510f057ce397ea22f2833a0cfe029bfc2463`;
- `qwen3.5:9b`: digest
  `6488c96fa5faab64bb65cbd30d4289e20e6130ef535a93ef9a49f42eda893ea7`,
  Q4_K_M;
- cold load chat osservato in circa 16 secondi;
- `ollama ps` ha riportato `100% GPU` con context 4096 e 8192;
- i log runner hanno confermato CUDA0 e thinking disabilitato per chat;
- nessun OOM, reset provider o fallback CPU osservato.

Doctor è 10/10 dopo l'acquisizione dei due modelli. Suite normale e
development, race detector e `go vet` sono verdi nell'ambiente ext4.


# Milestone 36 — Residency preflight v4

Date: 2026-09-05

The frozen v4 residency profile passed its executable preflight before the formal run.

## Frozen inputs

- profile: `docs/milestone-36-residency-profile-v4.yaml`
- profile SHA-256: `a6f6703716ceead6afeadc9d2fea3400037a2a4abecf488c3f897204106ede64`
- host-bound prompt SHA-256: `594659d52ec6142a5ef79c36dc0db4899e7ef1bb3f99d05017410f68bc1ba732`
- decision schema SHA-256: `bc3432a8f19867eec8e153adaa4434b688974cf34d24b6bd770e887e0dd7557d`
- WSL configuration SHA-256: `f5e1a679dbeca06a712c4f098ca3d17ae3f3c2eb19d20a224904b500e23e4cd6`

## Reference environment

- WSL2 Ubuntu 24.04: `MemTotal=15933088 KiB`, `SwapTotal=0 KiB`, `SwapFree=0 KiB`
- GPU: NVIDIA GeForce RTX 5070, 12227 MiB
- provider: Ollama 0.33.1
- resident models before the run: none
- Direct Chat identity: `qwen3.5:9b` / `6488c96fa5faab64bb65cbd30d4289e20e6130ef535a93ef9a49f42eda893ea7`
- Controlled Mutation identity: `qwen2.5-coder:14b` / `9ec8897f747e246e970bc5cfdda85d22f1123dc2e3d34978a010a75968716849`

The runner validated the model identities, provider version, prompt, schema, GPU identity, zero-swap reference, and exact `.wslconfig` digest without changing model residency.

No M33, M34, or M35 run was repeated.

# Milestone 12 — Final Report

Data: 2026-08-21

Verdetto: **GO — v0.2.0 read-only**

## Risultato

La Milestone 12 consegna una release v0.2.0 installabile e riproducibile per
Linux `amd64`, qualificata sul percorso Ollama/`llama3.1:8b` con reference
agent Laravel read-only. L'archive finale ha SHA-256
`c2d2a6f35178e91ad0c62d3c27f4ff2c33eedb46fd5fb327535890638e963758`
e incorpora il commit `5b05237362370fa79f133e159105a6a99050e81a`.

Il tag annotato `v0.2.0`, il manifest e `maestro version` risolvono allo stesso
commit. La documentazione pubblica era già congelata nel commit antenato
`fac2ae347d9fd6e03e9faef466d11bafa961370c`.

## Fasi

| Fase | Gate | Evidenza principale |
|---|---|---|
| 1 — contratto e baseline | PASS | support claim e stop rule congelati |
| 2 — superficie read-only | PASS | configurazione e documenti coerenti |
| 3 — packaging candidate | PASS | `pc.1` riproducibile e installabile |
| 4 — gate operativi | PASS | deny, EOF, no-TTY, segnali, limiti, anti-leak |
| 5 — qualificazione live e RC | PASS | `pc.5` e `rc.1`, due quick start consecutivi |
| 6 — release e tag | PASS | artifact finale, conferma live, tag verificato |

La serie live ha respinto `pc.1`–`pc.4` senza promozione e ha prodotto
hardening bounded del reference agent. Solo `pc.5` è stato promosso a un RC
distinto; l'artifact finale è stato poi ricostruito separatamente dalla
baseline successiva alla documentazione.

## Confine rilasciato

La release espone nel profilo ufficiale esclusivamente:

- `workspace.list`;
- `workspace.read`;
- `workspace.search`;
- `workspace_mutate: deny`.

Controlled Mutation, `workspace.write`, `workspace.patch`, approval mutativa,
`ibm/granite4.1:8b`, llama.cpp, sandbox, shell, Git, recovery, multi-agent e
tool di terze parti non sono qualificati né supportati dalla v0.2.0.

## Evidenza conclusiva

- doppio packaging finale byte-identico e checksum valido;
- installazione pulita fuori dal checkout;
- doctor 9/9 e modello supportato disponibile;
- quick start finale `completed`, una read reale, risposta corretta e fixture
  invariata;
- suite tripla, race detector e vet verdi;
- audit archive e anti-leak negativi;
- tag, manifest e binario con commit identico.

## Decisione

Tutti i gate previsti sono superati e non restano blocker di release. Il GO è
limitato al support claim read-only sopra descritto e non modifica il verdetto
`mutation_deferred` di ADR-0032.

La milestone è completata. Archive e checksum sono stati prodotti localmente;
push del tag e pubblicazione su un servizio remoto non sono stati eseguiti.

# Milestone 36 — Controlled Mutation Productization

Stato: Aperta — `controlled_mutation_productization_open`.

Data: 2026-09-05. Prerequisito: M35 conclusa con
`mutation_specific_model_qualified`, ADR-0040.

## Obiettivo

Portare il contratto host-bound qualificato nel prodotto con routing esplicito
per capacità: `qwen3.5:9b` per Direct Chat e il profilo M35 congelato di
`qwen2.5-coder:14b` per Controlled Mutation. M36 non ripete le campagne M35 e
non modifica il confine di autorità sul target.

## Fasi

1. Definire configurazione e risoluzione del modello per capacità, inclusi
   digest, disponibilità, acquisizione, residency e diagnostica fail-closed.
2. Collegare il percorso utente `selezione -> richiesta -> preview esatta ->
   approval -> commit atomico -> verifica`, mantenendo path, coordinate e byte
   selezionati sotto autorità host.
3. Integrare terminali, audit e messaggi operativi senza esporre al modello
   ricerca del target o accesso filesystem.
4. Verificare compatibilità Direct Chat, installazione pulita, configurazioni,
   doctor, packaging riproducibile e funzionamento sul reference hardware.
5. Pubblicare matrice di supporto, evidenze di productization e decisione di
   release readiness. Nessun PASS pubblica automaticamente v0.5.0.

## Gate

- Direct Chat mantiene integralmente il claim e i gate v0.4.0;
- il profilo mutativo coincide con modello, digest, prompt, schema e parametri
  congelati in M35;
- selezione singola host-bound, preview e approval sono obbligatorie;
- apply attesi, deny, stale check e rifiuti pre-provider raggiungono il
  terminale previsto nel prodotto installato;
- scritture stale, fuori selezione, errate o non approvate e failure con
  effetti restano a zero;
- configurazione assente, modello errato o digest difforme falliscono chiusi;
- suite, race, vet, packaging e installazione pulita sono verdi su Linux
  `amd64` e reference hardware.

M36 parte senza candidate, package, tag o release v0.5.0 autorizzato.

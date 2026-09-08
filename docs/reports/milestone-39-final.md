# Report finale — Milestone 39

Data: 2026-09-08  
Titolo: **Documentation, Onboarding & Public Trial Readiness**

## Verdetto

M39 si conclude con
`documentation_onboarding_public_trial_ready`. Maestro v0.5.0 dispone ora di
una baseline pubblica coerente, una pagina verità e un percorso riproducibile
che un utente esterno può completare senza checkout, rebuild o conoscenze
dell'architettura interna.

I casi congelati P01–P08 sono tutti PASS. Non sono state aggiunte feature e
il support claim non è stato ampliato oltre M38.

## Baseline documentale

README, installazione, quick start, Controlled Mutation, compatibility,
security, known issues, troubleshooting e release note convergono su cinque
affermazioni:

- v0.5.0 usa l'asset Linux `amd64` ed è stata qualificata anche sul reference
  Linux nativo CPU-only M38;
- Controlled Mutation è supportata soltanto nel perimetro single-file,
  single-range, PHP sotto `app/`;
- ogni scrittura richiede una preview e un'approvazione allow-once in TTY;
- multi-file, agent autonomi, Windows nativo e provider/modelli alternativi
  non sono supportati;
- il difetto documentale legacy dell'archive pubblico è dichiarato, mentre i
  sorgenti sono corretti per le release successive.

La fonte sintetica è [Capacità correnti](../current-capabilities.md). Gli hash
dei documenti sono in
[v0.5.0-public-baseline-freeze.yaml](../v0.5.0-public-baseline-freeze.yaml).

## Prova da utente esterno

È stato scaricato nuovamente
`maestro-v0.5.0-linux-amd64.tar.gz` dalla GitHub Release. Il checksum ufficiale
ha verificato SHA-256
`0afcfe4d648edcde3caf4327c4f995606fb4c3974c05606e13f90dd8cff321d9`.
Manifest e diagnostica dichiarano v0.5.0, stato `release`, commit
`86ed92495c5ce5bd2d0ec8d3c8a11b8c29a316f2` e binario
`e1765f9e8ed919eabe22b200f4145445da0d1b31f90c9b2bf6157a1390e37523`.

I modelli locali coincidevano con i digest del manifest. Non è stato
necessario effettuare un pull; questa è la stessa condizione prevista dalla
guida, che richiede di fermarsi in caso di digest diverso.

La fixture estratta è stata inizializzata come repository Git indipendente e
pulito. Il flusso ha prodotto:

| Passo | Risultato |
| --- | --- |
| Doctor | 14/14 `pass` |
| Direct Chat | `completed`, risposta corretta, 176.390 ms, un tentativo |
| Mutation | preview esatta `201` → `202`, allow-once, `applied`, un tentativo |
| Git | un solo file modificato, 1 inserimento e 1 eliminazione |

La chat ha identificato `customer_id`, `items` e la risposta JSON HTTP 201.
La mutation ha toccato soltanto
`app/Http/Controllers/OrderController.php:22`; l'hash del file è passato da
`4826abe9c6c5d701133817a9dcb565f0b84f760da57e1b518d430b601520b1bd` a
`30097526418b813fef8ba74ece5825637881b83814ba0340b78f5f349c040215`.

## Verifica repository e packaging

Sono verdi `go test ./...`, `go test -race ./...`, `go vet ./...`,
`git diff --check`, parsing YAML, sintassi degli script e controllo dei link
relativi. Un clean snapshot temporaneo ha prodotto due package byte-identici
con il gate completo:

- versione di prova: `v0.5.0-docs.8`;
- profilo: `mutation-productization`;
- SHA-256:
  `d877c0d9311bc2307f93bca39a01ff942272425ddf13baac7ea78cdd4869d1d5`.

Il package era soltanto una verifica locale e non è stato pubblicato. L'asset
v0.5.0 esistente non è stato modificato o sovrascritto.

## Limiti e passaggio di fase

M39 dimostra che il nucleo operativo già qualificato è ora presentabile e
provabile. Non dimostra autonomia, supporto universale Linux o nuove superfici
di mutation. La direzione consigliata successiva è un prototipo VS Code che
esponga il perimetro controllato esistente; M40 resta da aprire con piano e
gate propri.

L'evidenza strutturata completa è in
[milestone-39-public-trial.json](milestone-39-public-trial.json).

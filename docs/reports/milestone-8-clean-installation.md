# Milestone 8 — Clean Installation Validation

Data: 2026-08-15

Stato: Superata su `v0.1.0-pc.5`

---

# Ambiente e identità

La prova è stata eseguita su Linux `amd64`, Intel Core i5-8365U, 15 GiB RAM e
4 GiB swap, in una nuova directory `/tmp` senza checkout. Sono stati copiati
soltanto archive e checksum.

| Campo | Valore |
|---|---|
| Artifact | `maestro-v0.1.0-pc.5-linux-amd64.tar.gz` |
| Versione | `v0.1.0-pc.5` |
| Commit | `2732f26af4550833ad1b2d9cd4ca1caf5d72cd30` |
| SHA-256 | `4eb9abdfab6efbd00dc624b509581ec57666da1c4645d60abadc9316104ffe11` |
| Dimensione | 3598436 byte |
| Provider | Ollama `0.32.5` |
| Modello | `llama3.1:8b` |

`sha256sum -c`, estrazione, `maestro version` e root help sono verdi. Manifest,
nome artifact, versione e commit coincidono.

# Profilo ufficiale

La configurazione estratta:

- usa Ollama e `llama3.1:8b`;
- registra esattamente `workspace.list`, `workspace.read` e
  `workspace.search`;
- non contiene `workspace.write` o `workspace.patch`;
- imposta `workspace_mutate: deny`;
- punta tramite path relativo alla fixture inclusa.

Il controller parte e termina con SHA-256
`4826abe9c6c5d701133817a9dcb565f0b84f760da57e1b518d430b601520b1bd`.
Nessun prompt di approval o effetto mutativo è stato osservato.

# CLI dall'artifact

`doctor` supera tutti i 9 check. `models` elenca il catalogo del solo provider
Ollama configurato. `agents` espone `agent.reference` versione `1.0.0` con
capability planning, run e workspace-aware.

# Quick start consecutivi

I due run usano senza variazioni l'istruzione:

```text
Read app/Http/Controllers/OrderController.php and explain which service its store method calls. Do not modify any file.
```

| Run | Terminale | Turni | Tool call | Token in/out | Durata | Risposta |
|---:|---|---:|---:|---:|---:|---|
| 1 | `completed` | 2 | 1 | 2778 / 74 | 330122 ms | `OrderService::create` |
| 2 | `completed` | 2 | 1 | 2778 / 53 | 52386 ms | `OrderService::create` |

Entrambe le tool call sono letture reali; entrambi i processi terminano con
exit code 0. Il secondo run segue immediatamente il primo sulla stessa
installazione pulita e la fixture resta invariata.

# Verdetto

**Gate di installazione pulita superato.** `pc.5` soddisfa il percorso utente
read-only definito da ADR-0029. La promozione deve produrre un nuovo artifact
con identità e manifest `release-candidate`; `pc.5` non viene rinominato.

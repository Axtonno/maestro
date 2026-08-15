# Milestone 8 — Clean Installation Validation

Data: 2026-08-15

Stato: Superata su `v0.1.0-pc.5`, riconfermata su `pc.6` e `rc.2`

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

# Riconferma dopo `rc.1`

Il distinto `rc.1` ha superato l'installazione ma fallito il run live di
conferma ed è stato escluso. Dopo l'hardening degli argomenti read-only,
`pc.6` ha ripetuto il gate da una nuova directory contenente inizialmente
soltanto archive e checksum. Il candidate, commit
`ab109a5f878b8e1f10d69327736f014ad916a970`, SHA-256
`a148df8ff46d412ba85a39429f02048911d0793d3494db031a79cfa8ea76207b`,
ha completato due run consecutivi in 320075 ms e 66128 ms, ciascuno con una
read reale, risposta corretta ed exit code 0.

La prova conclusiva usa il nuovo artifact, senza rinominare i precedenti:

| Campo | Valore |
|---|---|
| Artifact | `maestro-v0.1.0-rc.2-linux-amd64.tar.gz` |
| Commit | `ab109a5f878b8e1f10d69327736f014ad916a970` |
| SHA-256 | `442090c6e2dac6095aa4532d658def42cd39e04a34baff401b3a92aec1fd9105` |
| Dimensione | 3598576 byte |
| Manifest | `release-candidate` |

Da un'ulteriore directory pulita sono verdi checksum, estrazione, version,
help, configurazione read-only, doctor 9/9, models e agents. Il run di
conferma termina `completed` in 64296 ms, con due turni, una read reale,
2888/94 token, risposta `OrderService::create` ed exit code 0. Il controller
mantiene lo stesso digest prima e dopo tutte le prove.

**La clean installation della release candidate `rc.2` è superata.**

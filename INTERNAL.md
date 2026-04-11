# Guida al Rilascio di una Nuova Versione (Metodo Squash)

Questo documento descrive la procedura standard per rilasciare una nuova versione di `oa-tools` (o di altri progetti correlati), compattando tutti i commit di sviluppo intermedi in un unico "commit di release" pulito. 

Questo approccio si sposa perfettamente con una filosofia di sviluppo essenziale e senza fronzoli: ci permette di sperimentare liberamente in locale, ma di mantenere una cronologia pubblica lineare e comprensibile per chiunque legga il progetto, senza ingombrare il ramo principale con micro-modifiche.

## 1. Identificare l'ultimo tag
Per prima cosa, controlla qual è l'ultimo tag rilasciato, che farà da punto di partenza per il nostro "riavvolgimento":

```bash
# Mostra tutti i tag
git tag --list

# Oppure, per vedere solo l'ultimo in ordine cronologico:
git describe --tags --abbrev=0
```
*Supponiamo che l'ultimo tag sia `v1.0.0` e che vogliamo rilasciare la `v1.1.0`.*

## 2. Il "Trucco": Soft Reset
Usa il comando `reset --soft` puntando all'ultimo tag identificato. 
Questo comando cancella la cronologia dei commit successivi a quel tag, ma **mantiene intatte tutte le modifiche ai file**. In pratica, rimette tutto il tuo nuovo lavoro in area di stage, pronto per essere committato di nuovo.

```bash
git reset --soft v1.0.0
```

## 3. Creare il Commit di Release
Ora hai tutto il lavoro svolto tra la vecchia versione e quella attuale concentrato e pronto. Crea un unico commit solido e descrittivo:

```bash
git commit -m "Release v1.1.0: Aggiunto SSH multiplexing per i pacchetti e ottimizzazioni generali"
```

## 4. Creare il Nuovo Tag
Una volta creato il mega-commit, è il momento di etichettarlo con il numero della nuova release:

```bash
git tag v1.1.0
```

## 5. Allineare il Remote (Push)
Poiché abbiamo riscritto la storia del nostro repository locale, GitHub (o il server remoto) rifiuterà un push standard, avvisandoti che le cronologie non combaciano. 
Dobbiamo imporre la nostra nuova storia pulita sovrascrivendo quella vecchia, usando un push forzato, e poi inviare i nuovi tag.

```bash
# Sovrascrive la storia remota con quella compattata locale
git push --force

# Invia il nuovo tag al server
git push --tags
```

> **⚠️ Avvertenza sul `--force`**
> Il push forzato va usato con consapevolezza. È una pratica sicura se sei l'unico a lavorare sul branch o se il team è allineato sul flusso di rilascio tramite rebase/squash.
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
*Supponiamo che l'ultimo tag sia `v0.6.3` e che vogliamo rilasciare la nuova `v0.6.4`.*

## 2. Il "Trucco": Soft Reset
Usa il comando `reset --soft` puntando all'ultimo tag identificato. 
Questo comando cancella la cronologia dei commit successivi a quel tag, ma **mantiene intatte tutte le modifiche ai file**. In pratica, rimette tutto il tuo nuovo lavoro in area di stage, pronto per essere committato di nuovo.

```bash
git reset --soft v0.6.3
```

## 3. Creare il Commit di Release
Ora hai tutto il lavoro svolto tra la vecchia versione e quella attuale concentrato e pronto. Crea un unico commit solido e descrittivo per la nuova versione:

```bash
git commit -m "Release v0.6.4: first release with clean history"
```

## 4. Creare il Nuovo Tag
**Prima** di compilare, è fondamentale applicare il nuovo tag. In questo modo i tool di compilazione (che leggono la versione da git) useranno la nomenclatura corretta e pulita.

```bash
git tag v0.6.4
```

## 5. Creare ed esportare i pacchetti
Ora che il codice è taggato correttamente, lancia la build per generare i pacchetti nativi (Debian e Arch) che verranno poi caricati nella Release:

```bash
make
coa/coa build
```
On Archlinux `makepkg -ci` on Debian `sudo dpkg -i oa-tools-0.6.4-1`.

## 6. Allineare il Remote (Push)
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

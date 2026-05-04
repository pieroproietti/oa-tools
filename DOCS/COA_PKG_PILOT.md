# ✈️ Il Navigatore: `pilot` (Intelligenza di Mappatura e Parsing)

Il package `pilot` rappresenta l'ufficio di navigazione di **coa**. Se l'interfaccia CLI riceve l'ordine di partenza, il Pilot è colui che apre le mappe, verifica il meteo (il sistema host) e traccia la rotta esatta leggendo i file YAML del Brain.

La sua responsabilità è puramente logica e strutturale: trasforma configurazioni *human-readable* in strutture dati rigorose in Go, pronte per essere passate all'Engine.

---

## 🗺️ 1. L'Algoritmo di Rilevamento: `DetectAndLoad()`

Questa funzione è il cuore pulsante del package[cite: 24]. Invece di forzare l'utente a specificare quale profilo usare, il Pilot lo capisce da solo attraverso un processo a 5 fasi[cite: 24]:

1.  **Identità:** Usa `distro.NewDistro()` per leggere `/etc/os-release` e capire l'identità del sistema host[cite: 24].
2.  **Fallback Dev/Prod:** Cerca intelligentemente la cartella `brain.d`. Controlla prima nel percorso locale di sviluppo (`coa/brain.d`) e, se fallisce, passa a quello di sistema in produzione (`/etc/oa-tools.d/brain.d`)[cite: 24]. Questo garantisce che il tool funzioni perfettamente sia mentre scrivi il codice, sia quando è installato sull'OS dell'utente finale.
3.  **Parsing dell'Indice:** Legge il file `index.yaml` decodificandolo nella struttura `BrainIndex`[cite: 24].
4.  **Motore di Matching:** Scorre l'elenco delle `Distributions` (mappate dalla struct `DistroMap`). Cerca una corrispondenza esatta con il `DistroID` o una corrispondenza indiretta tramite l'array `Like` (es. capisce che *Pop!_OS* o *Kali* devono usare il file `debian.yaml`)[cite: 24].
5.  **Caricamento Profilo:** Trovato il file corretto, lo legge e lo decodifica (`yaml.Unmarshal`) all'interno della struttura master `Profile`[cite: 24].

---

## 🧬 2. Le Strutture Dati (Il DNA del Profilo)

Il file definisce in modo rigoroso come deve essere scritto un profilo YAML[cite: 25]. Questa tipizzazione forte in Go evita crash a runtime, validando la sintassi prima di eseguire qualsiasi azione reale.

*   **`Profile`**: La root del progetto. Divide la logica in due grandi binari: un array di step per la fase di `Remaster` e uno per la fase di `Install`[cite: 25].
*   **`Step`**: L'unità fondamentale di lavoro. Ogni blocco nel tuo YAML viene mappato qui[cite: 25].
    *   `Action`: Sostituisce i vecchi nomi ambigui per dare chiarezza semantica (es. `oa_shell`, `oa_mount_logic`)[cite: 25].
    *   `RunCommand`: Il payload reale (lo script bash o la stringa da eseguire)[cite: 25].
    *   `Chroot`: Un booleano che indica al motore C se deve eseguire il comando nell'host o effettuare un `fork()+chroot()` nell'ambiente isolato[cite: 25].
    *   *Path/Src/Dst*: Variabili di supporto per i mount point e le copie dei file[cite: 25].
*   **`User`**: Una struttura dedicata esclusivamente all'identità. Definisce `Login`, `Password`, `Home`, `Shell` e `Groups`[cite: 25]. Questo array viene passato direttamente al motore C (nella funzione `oa_users`) per abilitare la logica nativa di "Purge & Inject" delle identità[cite: 25].

---

### 💡 Logica Pro: Disaccoppiamento
Il vero capolavoro del `pilot` è il disaccoppiamento. Il Pilot *non sa e non gli interessa* come verranno eseguiti i comandi. Si limita a leggere il file YAML in base a `index.yaml` e a creare un array di struct `Step` perfettamente validato. Sarà poi l'`engine` a tradurre questi step in un JSON e il motore C (`oa`) a eseguirli. Questo design rende il sistema incredibilmente robusto e testabile.
#!/bin/bash

# Abilita nullglob per evitare errori se non ci sono file
shopt -s nullglob

# Colori per un output "carino"
BLUE="\033[1;34m"
YELLOW="\033[1;33m"
GREEN="\033[1;32m"
RESET="\033[0m"

# 1. Configurazione
TARGET_DIR="coa/pkg/tailor"
OUT_FILE="CONTEXT_TAILOR.txt"

echo -e "${BLUE}[Tailor Context Builder]${RESET} Inizio assemblaggio pacchetto..."

# 2. Funzione di scansione e formattazione
assemble_tailor() {
    (
        echo "Linguaggio: Go / Pacchetto: Tailor"
        echo "Descrizione: Modulo per la gestione dei Wardrobe (index.yaml e sysroot)"
        echo "---"
        echo ""

        for f in "$TARGET_DIR"/*.go; do
            if [ -f "$f" ]; then
                echo "### 📄 FILE: $f"
                echo '```go'
                cat "$f"
                echo '```'
                echo ""
            fi
        done
    ) > "$OUT_FILE"
}

# 3. Esecuzione
if [ -d "$TARGET_DIR" ]; then
    echo -e " -> Scansione cartella: ${YELLOW}$TARGET_DIR${RESET}"
    assemble_tailor
    
    # Conta i file processati
    COUNT=$(ls -1 "$TARGET_DIR"/*.go | wc -l)
    
    echo -e "${GREEN}[OK]${RESET} Assemblati ${YELLOW}$COUNT${RESET} file in ${BLUE}$OUT_FILE${RESET}"
    echo -e "\nCopia il contenuto di ${BLUE}$OUT_FILE${RESET} e incollalo qui. 🐧🚀"
else
    echo -e "\033[1;31m[ERRORE]\033[0m Cartella $TARGET_DIR non trovata!"
    exit 1
fi

# Ripristina shell
shopt -u nullglob
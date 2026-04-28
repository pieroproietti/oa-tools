#!/bin/bash

# Abilita nullglob per l'espansione corretta degli array 
# e globstar per la ricerca ricorsiva nelle sottocartelle
shopt -s nullglob
shopt -s globstar

# 1. Configurazione Destinazione
REMOTE_USER="artisan"
REMOTE_HOST="192.168.1.2"
DEST_PATH="/home/artisan/"
TARGET="$REMOTE_USER@$REMOTE_HOST"

# 2. Genera suffisso comune per i file di contesto
RAND_SUFFIX=$(printf "%03d" $((RANDOM % 1000)))
FILE_COA="CONTEXT_COA_${RAND_SUFFIX}.txt"
FILE_OA="CONTEXT_OA_${RAND_SUFFIX}.txt"
FILE_DOCS="CONTEXT_DOCS_${RAND_SUFFIX}.txt"

echo -e "\033[1;34m[Context Builder]\033[0m Session: \033[1m$RAND_SUFFIX\033[0m"

# Funzione per assemblare i file in un unico blocco di testo per l'IA
build_context() {
    local out_file=$1
    shift
    local files=("$@")
    echo -e " -> Assembling \033[1;33m$out_file\033[0m..."
    (
        echo '````'
        for f in "${files[@]}"; do
            if [ -f "$f" ]; then
                echo "### 📄 FILE: $f"
                filename=$(basename "$f")
                ext="${filename##*.}"
                case "$ext" in
                    c|h) lang="c" ;; 
                    go) lang="go" ;; 
                    sh) lang="bash" ;;
                    json) lang="json" ;; 
                    md) lang="markdown" ;; 
                    yaml|yml) lang="yaml" ;;
                    *) lang="text" ;;
                esac
                if [[ "$filename" == "Makefile" || "$filename" == "m" ]]; then lang="make"; fi
                echo '```'"$lang"
                cat "$f"
                echo '```'
                echo ""
            fi
        done
        echo '````'
    ) > "$out_file"
}

# 3. Definizione file (Mind and Body)

# FILES_OA: Il braccio operativo in C
FILES_OA=(
    oa/Makefile 
    oa/README.md 
    oa/include/*.h 
    oa/src/**/*.c
)

FILES_COA=(
    coa/*
    coa/pkg/**/*
    coa/brain.d/*
)

# FILES_DOCS: la documentazione
FILES_DOCS=(
    DOCS/**/*.md
)

# 4. Costruzione locale dei pacchetti di contesto
build_context "$FILE_OA" "${FILES_OA[@]}"
build_context "$FILE_COA" "${FILES_COA[@]}"
build_context "$FILE_DOCS" "${FILES_DOCS[@]}"

# Disabilita le opzioni shell extra
shopt -u nullglob
shopt -u globstar

# 5. Sincronizzazione atomica via SSH
echo -e "\033[1;32m[SYNC]\033[0m Pulizia remota e trasferimento in corso..."

# Usiamo tar per impacchettare, inviare e scompattare in un unico tunnel SSH
tar -cf - "$FILE_OA" "$FILE_COA" "$FILE_DOCS"| ssh "$TARGET" "cd $DEST_PATH && rm -f CONTEXT_OA_*.txt CONTEXT_COA_*.txt CONTEXT_DOCS_*.txt && tar -xf -"

# 6. Pulizia locale dei file temporanei
rm "$FILE_OA" "$FILE_COA" "$FILE_DOCS" 

echo -e "\033[1;32m[OK]\033[0m Sincronizzazione completata! (California dreaming... 🚲)"
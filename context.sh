#!/bin/bash

# Abilita nullglob per l'espansione corretta degli array
shopt -s nullglob

# 1. Configurazione Destinazione
REMOTE_USER="artisan"
REMOTE_HOST="192.168.1.2"
DEST_PATH="/home/artisan/"
TARGET="$REMOTE_USER@$REMOTE_HOST"

# 2. Genera suffisso comune
RAND_SUFFIX=$(printf "%03d" $((RANDOM % 1000)))
FILE_OA="CONTEXT_OA_${RAND_SUFFIX}.txt"
FILE_COA="CONTEXT_COA_${RAND_SUFFIX}.txt"

echo -e "\033[1;34m[Context Builder]\033[0m Session: \033[1m$RAND_SUFFIX\033[0m"

# --- (La funzione build_context rimane identica) ---
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
                    c|h) lang="c" ;; go) lang="go" ;; sh) lang="bash" ;;
                    json) lang="json" ;; md) lang="markdown" ;; yaml|yml) lang="yaml" ;;
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

# 3. Definizione file (OA e COA)
FILES_OA=(oa/CHANGELOG.md oa/Makefile oa/MANIFESTUM.md oa/README.md oa/docs/*.md oa/include/*.h oa/json/*.json oa/src/*.c oa/src/actions/*.c oa/src/vendors/*.c)
FILES_COA=(coa/m coa/go.mod coa/src/*.go coa/conf/*.yaml coa/docs/ROADMAP.md coa/README.md)

# 4. Costruzione locale
build_context "$FILE_OA" "${FILES_OA[@]}"
build_context "$FILE_COA" "${FILES_COA[@]}"
shopt -u nullglob

# 5. IL TRUCCO PER LA PASSWORD SINGOLA
# Usiamo SSH con il "ControlMaster" temporaneo o semplicemente concateniamo
# Se non hai SSH-KEY, il modo più semplice è raggruppare l'azione in un tunnel pipe, 
# ma la soluzione standard è usare sshpass (se installato) o semplicemente 
# accettare che SCP e SSH siano separati. 

# TUTTAVIA, possiamo fare un "reverse": inviamo i file e poi puliamo i VECCHI 
# (escludendo quelli appena mandati) in un colpo solo.
# Ma la cosa più semplice e pulita è:

echo -e "\033[1;32m[SYNC]\033[0m Pulizia e Trasferimento in corso..."

# Usiamo tar via SSH per fare tutto in un'unica connessione (Nessuna doppia password!)
tar -cf - "$FILE_OA" "$FILE_COA" | ssh "$TARGET" "cd $DEST_PATH && rm -f CONTEXT_OA_*.txt CONTEXT_COA_*.txt && tar -xf -"

# 6. Pulizia locale
rm "$FILE_OA" "$FILE_COA"

echo -e "\033[1;32m[OK]\033[0m Sincronizzazione completata con una sola connessione!"

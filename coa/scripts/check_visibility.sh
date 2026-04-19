#!/bin/bash

# check_visibility.sh
# Analizza i pacchetti internal di coa per trovare funzioni esportate (Maiuscole)
# che non vengono utilizzate al di fuori del proprio pacchetto.

echo -e "\033[1;34m[coa-tools]\033[0m Analisi visibilità pacchetti internal...\n"

# Itera su ogni cartella in src/internal (engine, distro, pilot, etc.)
for pkg_dir in src/internal/*; do
    if [ -d "$pkg_dir" ]; then
        pkg_name=$(basename "$pkg_dir")
        echo -e "\033[1;33m--- Pacchetto: $pkg_name ---\033[0m"
        
        # Estrae i nomi delle funzioni che iniziano con una Maiuscola
        funcs=$(grep -rhE "^func [A-Z]" "$pkg_dir" | awk '{print $2}' | cut -d'(' -f1)
        
        for f in $funcs; do
            # Conta le occorrenze della funzione in tutto src/, escludendo la cartella del pacchetto stesso
            count=$(grep -r "$f" src/ --exclude-dir="$pkg_name" | wc -l)
            
            if [ $count -eq 0 ]; then
                echo -e "  \033[0;31m[!]\033[0m $f: Esportata ma usata solo internamente."
            fi
        done
        echo ""
    fi
done

echo -e "\033[1;32m[FINE]\033[0m Analisi completata."

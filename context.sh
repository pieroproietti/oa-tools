# Genera un numero casuale tra 000 e 999
RAND_SUFFIX=$(printf "%03d" $((RANDOM % 1000)))
CONTEXT="CONTEXT_${RAND_SUFFIX}.txt"

(
  echo '````'
  for f in  context.sh \
            CHANGELOG.md \
            Makefile \
            docs\ACTIONS.md \
            docs\ARCHITECTURE.md \
            docs\DISTRO_CHALLENGES.md \
            docs\OA_Context.md \
            docs\ROADMAP.md \
            include/oa.h \
            include/oa-logs.h \
            include/oa-yocto.h \
            json/cleanup.json  \
            json/exclude.json  \
            json/iso.json  \
            json/prepare.json  \
            json/run.json  \
            json/users.json \
            README.md \
            src/actions/action_initrd.c \
            src/actions/action_iso.c \
            src/actions/action_prepare.c \
            src/actions/action_remaster.c \
            src/actions/action_run.c \
            src/actions/action_scan.c \
            src/actions/action_squash.c \
            src/actions/action_users.c \
            src/main.c \
            src/vendors/oa-logger.c \
            src/vendors/oa-yocto.c \
            ; 
    do
    if [ -f "$f" ]; then
      echo "### 📄 FILE: $f"
      # Determina l'estensione per l'evidenziazione del codice
      ext="${f##*.}"
      if [ "$ext" == "c" ] || [ "$ext" == "h" ]; then lang="c"; else lang="markdown"; fi
      echo '```'$lang
      cat "$f"
      echo '```'
      echo ""
    else
      echo "⚠️ ERRORE: $f non trovato"
    fi
  done
  echo '````'
) > $CONTEXT

echo -e "\033[1;32m[oa]\033[0m File \033[1m$CONTEXT\033[0m generato con successo!"
scp $CONTEXT artisan@192.168.1.2:/home/artisan
rm $CONTEXT

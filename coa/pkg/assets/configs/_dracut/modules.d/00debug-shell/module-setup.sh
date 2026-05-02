# ~/penguins-eggs/dracut/modules.d/00debug-shell/module-setup.sh

#!/bin/bash

check() {
    return 0
}

depends() {
    # Non dipende da nulla per essere sicuro che si esegua
    return 0
}

install() {
    # Installa una shell e alcuni comandi utili per il debug
    inst_multiple /bin/sh /bin/ls /bin/cat /bin/mount
    
    # Installa uno script che si eseguirà molto presto durante il boot
    inst_hook cmdline 10 "$moddir/debug-hook.sh"
}
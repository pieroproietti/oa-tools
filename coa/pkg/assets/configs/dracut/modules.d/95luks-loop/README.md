# 95luks-loop

Debian (senza 90block, senza iso-scan) userà 95rootfs-block come base.

Il tuo nuovo modulo 95luks-loop farà la parte di “monta luks.img da ISO”.

Non serve patchare Dracut upstream.

È cross-distro, funziona anche su Fedora (dove block esiste ma non confligge).
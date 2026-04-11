# AlpineLinux

Fare riferimento a [penguins-alpine](https://github.com/pieroproietti/penguins-alpine) per installare penguins-eggs su Alpine e sul funzionamento del sidecar.

## sidecar

Il sidecar consiste in una patch a inittra,s-init che aggiunge le catatteristiche di permettere l'avvio da una ISO prodotta con penguins-eggs.

Il sidecar:

Cerca un dispositivo con la label specificata
Monta il dispositivo live
Monta il filesystem.squashfs come layer read-only
Crea un overlay con tmpfs per le modifiche
Configura il sistema per il boot live
Attaccare il sidecar alla moto

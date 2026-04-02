/*
* oa: eggs in my dialect🥚🥚
* remastering core
*
* Author: Piero Proietti <piero.proietti@gmail.com>
* License: GPL-3.0-or-later
*/
#ifndef OA_H
#define OA_H

// --- Inclusioni di Sistema Standard ---
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <limits.h>
#include <stdbool.h>
#include <stdint.h>
#include <errno.h>
#include <sys/utsname.h>
#include <sys/wait.h>
#include <sys/stat.h>
#include <sys/mount.h>   // <--- FONDAMENTALE PER I MOUNT (MS_BIND, etc.)
#include <ftw.h>

// --- Librerie esterne ---
#include "cJSON.h"
#include "logger.h"
#include "oa-yocto.h"

// --- Costanti Globali --
#define PATH_INPUT PATH_MAX   // 4096 - Per i percorsi che leggiamo
#define PATH_OUT   8192       // 8K - Per i percorsi che costruiamo
#define CMD_MAX    32768      // 32K - Per i comandi system()

// da rimuovere in futuro
#define PATH_SAFE 8192        // Il doppio di PATH_MAX: ora GCC non ha più dubbi

// OA_context
typedef struct {
    cJSON *root;    // Il JSON intero (configurazione globale) 
    cJSON *task;    // Il comando specifico nel plan (configurazione locale) 
} OA_Context;

// --- Inclusioni dei Moduli Actions ---
// Devono stare QUI in fondo, per evitare inclusioni circolari,
// in modo che OA_Context sia già definito quando vengono letti.
#include "action_prepare.h"  // Contiene anche action_cleanup
#include "action_initrd.h"
#include "action_remaster.h"
#include "action_run.h"
#include "action_scan.h"
#include "action_squash.h"
#include "action_iso.h"
#include "action_pause.h"
#include "action_users.h"
#include "action_crypted.h"

#endif
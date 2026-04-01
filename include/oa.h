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
#include "oa-logger.h"
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

// --- Prototipi Aggiornati ---
int action_prepare(OA_Context *ctx);
int action_cleanup(OA_Context *ctx);
int action_initrd(OA_Context *ctx);
int action_remaster(OA_Context *ctx);
int action_run(OA_Context *ctx);
int action_scan(OA_Context *ctx);
int action_squash(OA_Context *ctx);
int action_iso(OA_Context *ctx);
int action_pause(OA_Context *ctx);
int action_users(OA_Context *ctx); 

#endif
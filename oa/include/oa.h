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

#include "helpers.h"

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

// --- Inclusioni dei Moduli ---

// LAY (Remastering)
#include "remaster_cleanup.h"
#include "remaster_crypted.h"
#include "remaster_initrd.h"
#include "remaster_iso.h"
#include "remaster_isolinux.h"
#include "remaster_livestruct.h"
#include "remaster_prepare.h"
#include "remaster_squash.h"
#include "remaster_uefi.h"
#include "remaster_users.h"

// HATCH (Installazione Fisica)
#include "install_format.h"
#include "install_partition.h"
#include "install_unpack.h"
#include "install_fstab.h"
#include "install_users.h"
#include "install_initrd.h"
#include "install_uefi.h"
#include "install_bios.h"
#include "install_prepare.h"

// SYS (Utility Generiche)
#include "sys_run.h"
#include "sys_shell.h"
#include "sys_scan.h"
#include "sys_suspend.h"

#endif
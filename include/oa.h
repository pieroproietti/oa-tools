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

// --- Costanti Globali --
#define PATH_INPUT PATH_MAX   // 4096 - Per i percorsi che leggiamo
#define PATH_OUT   8192       // 8K - Per i percorsi che costruiamo
#define CMD_MAX    32768      // 32K - Per i comandi system()

// da rimuovere in futuro
#define PATH_SAFE 8192        // Il doppio di PATH_MAX: ora GCC non ha più dubbi

// --- Prototipi (Ereditati da helpers e actions) ---
void build_initrd_command(char *dest, const char *template, const char *out, const char *ver);
void append_eggs_exclusion(char *buffer, size_t buf_size, const char *path);

int action_prepare(cJSON *json);
int action_cleanup(cJSON *json);
int action_initrd(cJSON *json);
int action_remaster(cJSON *json);
int action_run(cJSON *json);
int action_scan(cJSON *json);
int action_squash(cJSON *json);
int action_iso(cJSON *json);

#endif

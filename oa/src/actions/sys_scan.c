/*
 * src/actions/install_uefi.c
 * Remastering core: GRUB installation on physical hardware (Krill)
 * oa: eggs in my dialect🥚🥚
 *
 * Author: Piero Proietti <piero.proietti@gmail.com>
 * License: GPL-3.0-or-later
 */
#include "oa.h"

// Variabili globali per il conteggio [cite: 181]
static uint64_t current_total_bytes = 0;
static uint64_t current_file_count = 0;

// Callback: eseguita per ogni file/directory [cite: 182]
static int scan_callback(const char *fpath, const struct stat *sb, int tflag, struct FTW *ftwbuf) {
    (void)fpath;
    (void)ftwbuf;

    if (tflag == FTW_F) {
        current_total_bytes += sb->st_size;
        current_file_count++;
    }
    return 0; 
}

int sys_scan(OA_Context *ctx) {
    // 1. Lookup a cascata: cerca "path" nel task locale, poi nel root [cite: 184]
    cJSON *path_obj = cJSON_GetObjectItemCaseSensitive(ctx->task, "path");
    if (!path_obj) path_obj = cJSON_GetObjectItemCaseSensitive(ctx->root, "path");

    if (!cJSON_IsString(path_obj) || (path_obj->valuestring == NULL)) {
        fprintf(stderr, "{\"error\": \"Path non specificato per lo scan\"}\n");
        return 1;
    }

    current_total_bytes = 0;
    current_file_count = 0;

    // 2. Esecuzione scansione [cite: 187]
    if (nftw(path_obj->valuestring, scan_callback, 64, FTW_PHYS | FTW_MOUNT) == -1) {
        fprintf(stderr, "{\"error\": \"Errore durante la scansione di %s\"}\n", path_obj->valuestring);
        return 1;
    }

    // 3. Risposta JSON [cite: 188]
    cJSON *response = cJSON_CreateObject();
    cJSON_AddStringToObject(response, "status", "ok");
    cJSON_AddStringToObject(response, "path", path_obj->valuestring);
    cJSON_AddNumberToObject(response, "total_bytes", (double)current_total_bytes);
    
    // Correzione MB (divisore 1024*1024) 
    cJSON_AddNumberToObject(response, "total_mb", (double)current_total_bytes / (1024.0 * 1024.0));
    cJSON_AddNumberToObject(response, "file_count", (double)current_file_count);

    char *out = cJSON_PrintUnformatted(response);
    printf("%s\n", out);

    free(out);
    cJSON_Delete(response);
    return 0;
}

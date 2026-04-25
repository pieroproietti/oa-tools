#include "oa.h"
#include <sys/mount.h>
#include <sys/stat.h>
#include <sys/types.h>
#include <errno.h>

/**
 * @brief Crea una directory (mkdir -p)
 */
int oa_mkdir(OA_Context *ctx) {
    cJSON *path = cJSON_GetObjectItemCaseSensitive(ctx->task, "path");
    
    if (!cJSON_IsString(path)) {
        LOG_ERR("oa_mkdir: parametro 'path' mancante o non valido");
        return 1;
    }

    LOG_INFO("Creazione directory: %s", path->valuestring);
    
    char cmd[PATH_SAFE];
    snprintf(cmd, sizeof(cmd), "mkdir -p %s", path->valuestring);
    return (system(cmd) == 0) ? 0 : 1;
}

/**
 * @brief Esegue un bind mount (anche ricorsivo e readonly)
 */
int oa_bind(OA_Context *ctx) {
    const char *src = cJSON_GetObjectItemCaseSensitive(ctx->task, "src")->valuestring;
    const char *dst = cJSON_GetObjectItemCaseSensitive(ctx->task, "dst")->valuestring;
    int ro = cJSON_IsTrue(cJSON_GetObjectItemCaseSensitive(ctx->task, "readonly"));

    LOG_INFO("Bind mount: %s -> %s (RO: %s)", src, dst, ro ? "SI" : "NO");

    // ---> AGGIUNTA FONDAMENTALE <---
    // Assicuriamoci che la directory di destinazione esista prima di montare!
    char mkdir_cmd[PATH_SAFE + 32];
    snprintf(mkdir_cmd, sizeof(mkdir_cmd), "mkdir -p %s", dst);
    system(mkdir_cmd);
    // --------------------------------

    // 1. Bind iniziale (Recursive)
    if (mount(src, dst, NULL, MS_BIND | MS_REC, NULL) != 0) {
        LOG_ERR("Bind fallito: %s -> %s (Errore: %s)", src, dst, strerror(errno));
        return 1;
    }

    // 2. Remount ReadOnly se richiesto
    if (ro) {
        mount(NULL, dst, NULL, MS_BIND | MS_REC | MS_REMOUNT | MS_RDONLY, NULL);
    }

    // 3. Fortificazione (MS_PRIVATE)
    mount(NULL, dst, NULL, MS_PRIVATE | MS_REC, NULL);

    return 0;
}
/**
 * @brief Mount generico (proc, sysfs, tmpfs, overlay)
 */
int oa_mount_generic(OA_Context *ctx) {
    const char *type = cJSON_GetObjectItemCaseSensitive(ctx->task, "type")->valuestring;
    const char *src = cJSON_GetObjectItemCaseSensitive(ctx->task, "src")->valuestring;
    const char *dst = cJSON_GetObjectItemCaseSensitive(ctx->task, "dst")->valuestring;
    
    // opts può essere opzionale (es. per proc o sysfs)
    cJSON *opts_obj = cJSON_GetObjectItemCaseSensitive(ctx->task, "opts");
    const char *opts = cJSON_IsString(opts_obj) ? opts_obj->valuestring : NULL;

    LOG_INFO("Mount %s: %s su %s", type, src, dst);

    // ---> AGGIUNTA FONDAMENTALE <---
    // Il kernel si arrabbia se la destinazione non esiste, quindi la creiamo al volo.
    char mkdir_cmd[PATH_SAFE + 32];
    snprintf(mkdir_cmd, sizeof(mkdir_cmd), "mkdir -p %s", dst);
    system(mkdir_cmd);
    // --------------------------------

    if (mount(src, dst, type, 0, opts) != 0) {
        LOG_ERR("Mount %s fallito su %s: %s", type, dst, strerror(errno));
        return 1;
    }

    return 0;
}

/**
 * @brief Esegue una copia fisica preservando permessi e link (equivalente a cp -a)
 */
int oa_cp(OA_Context *ctx) {
    cJSON *src_obj = cJSON_GetObjectItemCaseSensitive(ctx->task, "src");
    cJSON *dst_obj = cJSON_GetObjectItemCaseSensitive(ctx->task, "dst");

    if (!cJSON_IsString(src_obj) || !cJSON_IsString(dst_obj)) {
        LOG_ERR("oa_cp: parametri 'src' o 'dst' mancanti o non validi");
        return 1;
    }

    const char *src = src_obj->valuestring;
    const char *dst = dst_obj->valuestring;

    LOG_INFO("Copia fisica: %s -> %s", src, dst);

    char cmd[CMD_MAX];
    // Usiamo 'cp -a' per garantire che permessi, symlink e timestamp 
    // vengano mantenuti identici (fondamentale per /etc e /boot)
    snprintf(cmd, sizeof(cmd), "cp -a %s %s", src, dst);
    
    int ret = system(cmd);
    if (ret != 0) {
        LOG_ERR("Fallimento comando di copia (cp restituisce %d)", ret);
        return 1;
    }

    return 0;
}
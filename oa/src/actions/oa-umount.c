#include "oa.h"
#include <sys/mount.h>
#include <mntent.h>
#include <stdio.h>
#include <string.h>
#include <stdlib.h>

typedef struct {
    char path[PATH_SAFE];
    int length;
} MountPoint;

// Ordina dal percorso più lungo al più corto (deepest first)
int compare_mounts(const void *a, const void *b) {
    MountPoint *ma = (MountPoint *)a;
    MountPoint *mb = (MountPoint *)b;
    return mb->length - ma->length;
}

/**
 * @brief Smonta ricorsivamente tutto ciò che appartiene al progetto
 */
int oa_umount(OA_Context *ctx) {
    cJSON *pathLiveFs = cJSON_GetObjectItemCaseSensitive(ctx->task, "pathLiveFs");
    if (!pathLiveFs) pathLiveFs = cJSON_GetObjectItemCaseSensitive(ctx->root, "pathLiveFs");
    
    if (!cJSON_IsString(pathLiveFs)) {
        LOG_ERR("oa_umount: pathLiveFs non specificato");
        return 1;
    }

    const char *base_path = pathLiveFs->valuestring;
    LOG_INFO("Inizio pulizia dei mount in: %s", base_path);

    FILE *fp = setmntent("/proc/mounts", "r");
    if (!fp) {
        LOG_ERR("Impossibile leggere /proc/mounts: %s", strerror(errno));
        return 1;
    }

    MountPoint mounts[512]; // Aumentato a 512 per sicurezza
    int count = 0;
    struct mntent *ent;

    // Rilevamento automatico di tutto ciò che sta sotto la base del progetto
    while ((ent = getmntent(fp)) != NULL) {
        // Se il mount point inizia con il nostro base_path, va smontato
        if (strncmp(ent->mnt_dir, base_path, strlen(base_path)) == 0) {
            strncpy(mounts[count].path, ent->mnt_dir, PATH_SAFE);
            mounts[count].length = strlen(ent->mnt_dir);
            count++;
            if (count >= 512) break;
        }
    }
    endmntent(fp);

    if (count == 0) {
        LOG_INFO("Nessun mount point trovato per %s. Ambiente pulito.", base_path);
        return 0;
    }

    // Ordina: prima smontiamo /home/eggs/liveroot/proc, poi /home/eggs/liveroot
    qsort(mounts, count, sizeof(MountPoint), compare_mounts);

    int errors = 0;
    for (int i = 0; i < count; i++) {
        // MNT_DETACH (Lazy unmount) è fondamentale per non restare appesi
        if (umount2(mounts[i].path, MNT_DETACH) == 0) {
            LOG_INFO("Smontato correttamente: %s", mounts[i].path);
        } else {
            LOG_WARN("Fallimento smontaggio %s: %s", mounts[i].path, strerror(errno));
            errors++;
        }
    }

    if (errors == 0) {
        LOG_INFO("Cleanup completato con successo (%d mount rimossi).", count);
    } else {
        LOG_WARN("Cleanup terminato con %d errori.", errors);
    }

    return 0;
}

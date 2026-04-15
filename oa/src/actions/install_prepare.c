#include "oa.h"

/**
 * @brief Prepara l'ambiente chroot sul disco fisico montando i filesystem virtuali.
 */
int install_prepare(OA_Context *ctx) {
    cJSON *pathLiveFs = cJSON_GetObjectItemCaseSensitive(ctx->task, "pathLiveFs");
    if (!pathLiveFs) pathLiveFs = cJSON_GetObjectItemCaseSensitive(ctx->root, "pathLiveFs");

    if (!cJSON_IsString(pathLiveFs)) {
        LOG_ERR("oa_install_prepare: pathLiveFs mancante.");
        return 1;
    }

    char target_root[PATH_SAFE];
    // Puntiamo alla liveroot montata sulla partizione fisica
    snprintf(target_root, sizeof(target_root), "%s/liveroot", pathLiveFs->valuestring);

    LOG_INFO("Preparing physical chroot environment at %s", target_root);
    printf("\033[1;34m[oa HATCH]\033[0m Mounting virtual filesystems for physical target...\n");

    char cmd[CMD_MAX];
    // Usiamo mount --bind e mount -t per popolare /dev, /proc, /sys e /run
    snprintf(cmd, sizeof(cmd),
             "mount --bind /dev %s/dev && "
             "mount -t proc /proc %s/proc && "
             "mount -t sysfs /sys %s/sys && "
             "mount --bind /run %s/run",
             target_root, target_root, target_root, target_root);

    if (system(cmd) != 0) {
        LOG_ERR("Failed to mount virtual filesystems on physical disk");
        return 1;
    }

    printf("\033[1;32m[SUCCESS]\033[0m Physical environment ready for chroot actions.\n");
    return 0;
}

/*
 * src/actions/install_uefi.c
 * Remastering core: GRUB installation on physical hardware (Krill)
 * oa: eggs in my dialect🥚🥚
 *
 * Author: Piero Proietti <piero.proietti@gmail.com>
 * License: GPL-3.0-or-later
 */
#include "oa.h"
#include <mntent.h>

typedef struct {
    char path[PATH_SAFE];
    int length;
} MountPoint;

int compare_mounts(const void *a, const void *b) {
    MountPoint *ma = (MountPoint *)a;
    MountPoint *mb = (MountPoint *)b;
    return mb->length - ma->length;
}

int remaster_cleanup(OA_Context *ctx) {
    cJSON *pathLiveFs = cJSON_GetObjectItemCaseSensitive(ctx->task, "pathLiveFs");
    if (!pathLiveFs) pathLiveFs = cJSON_GetObjectItemCaseSensitive(ctx->root, "pathLiveFs");
    if (!cJSON_IsString(pathLiveFs)) return 1;

    char liveroot_dir[PATH_SAFE], overlay_dir[PATH_SAFE];
    snprintf(liveroot_dir, PATH_SAFE, "%s/liveroot", pathLiveFs->valuestring);
    snprintf(overlay_dir, PATH_SAFE, "%s/.overlay", pathLiveFs->valuestring);

    int liveroot_len = strlen(liveroot_dir);
    int overlay_len = strlen(overlay_dir);

    printf("\033[1;34m[oa CLEANUP]\033[0m Smart unmounting filesystem projections...\n");

    FILE *fp = setmntent("/proc/mounts", "r");
    if (!fp) {
        LOG_ERR("Failed to open /proc/mounts");
        return 1;
    }

    MountPoint mounts[256];
    int count = 0;
    struct mntent *ent;

    while ((ent = getmntent(fp)) != NULL) {
        if (strncmp(ent->mnt_dir, liveroot_dir, liveroot_len) == 0 ||
            strncmp(ent->mnt_dir, overlay_dir, overlay_len) == 0) {
            strncpy(mounts[count].path, ent->mnt_dir, PATH_SAFE);
            mounts[count].length = strlen(ent->mnt_dir);
            count++;
            if (count >= 256) break;
        }
    }
    endmntent(fp);

    if (count == 0) {
        printf("\033[1;32m[oa CLEANUP]\033[0m No mounts found to clean up.\n");
        return 0;
    }

    qsort(mounts, count, sizeof(MountPoint), compare_mounts);

    for (int i = 0; i < count; i++) {
        if (umount2(mounts[i].path, MNT_DETACH) == 0) {
            LOG_INFO("Unmounted %s", mounts[i].path);
        } else {
            LOG_WARN("Failed to unmount %s", mounts[i].path);
        }
    }

    printf("\033[1;32m[oa CLEANUP]\033[0m Cleanup completed successfully.\n");
    return 0;
}
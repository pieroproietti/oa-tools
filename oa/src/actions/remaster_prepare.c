/*
 * oa: eggs in my dialect🥚🥚
 * remastering core - Environment Preparation
 */
#include "oa.h"
#include <errno.h>
#include <sys/mount.h>
#include <sys/stat.h>
#include <stdio.h>
#include <string.h>
#include <unistd.h>

static int make_full_dir(const char *base, const char *rel) {
    char full_path[PATH_SAFE];
    snprintf(full_path, sizeof(full_path), "%s/%s", base, rel);
    if (mkdir(full_path, 0755) != 0 && errno != EEXIST)
        return -1;
    return 0;
}

static void clone_symlink(const char *src_path, const char *liveroot_base, const char *name) {
    char link_data[PATH_SAFE];
    ssize_t len = readlink(src_path, link_data, sizeof(link_data) - 1);
    if (len != -1) {
        link_data[len] = '\0';
        char dst_path[PATH_SAFE];
        snprintf(dst_path, sizeof(dst_path), "%s/%s", liveroot_base, name);
        unlink(dst_path); 
        symlink(link_data, dst_path);
    }
}

static int fortified_bind_mount(const char *src, const char *tgt, unsigned long flags) {
    if (mount(src, tgt, NULL, flags, NULL) != 0)
        return -1;
    mount(NULL, tgt, NULL, MS_PRIVATE | MS_REC, NULL);
    return 0;
}

int remaster_prepare(OA_Context *ctx) {
    cJSON *path_item = cJSON_GetObjectItemCaseSensitive(ctx->task, "pathLiveFs");
    if (!path_item) path_item = cJSON_GetObjectItemCaseSensitive(ctx->root, "pathLiveFs");
    cJSON *mode_item = cJSON_GetObjectItemCaseSensitive(ctx->task, "mode");
    if (!mode_item) mode_item = cJSON_GetObjectItemCaseSensitive(ctx->root, "mode");

    if (!cJSON_IsString(path_item)) return 1;

    const char *base = path_item->valuestring;
    const char *mode = cJSON_IsString(mode_item) ? mode_item->valuestring : "standard";

    char liveroot_path[PATH_SAFE], overlay_path[PATH_SAFE];
    snprintf(liveroot_path, sizeof(liveroot_path), "%s/liveroot", base);
    snprintf(overlay_path, sizeof(overlay_path), "%s/.overlay", base);

    // 1. Setup Struttura
    mkdir(base, 0755);
    make_full_dir(base, "liveroot");
    make_full_dir(base, ".overlay");
    make_full_dir(base, ".overlay/lowerdir");
    make_full_dir(base, ".overlay/upperdir");
    make_full_dir(base, ".overlay/workdir");

    // 2. COPIE FISICHE (/etc e ora anche /boot per Arch compatibilità)
    char cp_cmd[CMD_MAX];
    printf("\033[1;34m[oa PREPARE]\033[0m Physical copy of /etc and /boot...\n");
    snprintf(cp_cmd, sizeof(cp_cmd), "cp -a /etc %s/ && cp -a /boot %s/", liveroot_path, liveroot_path);
    system(cp_cmd);

    // 3. BIND MOUNTS (RDONLY)
    // Nota: 'boot' rimosso perché copiato fisicamente
    const char *root_entries[] = {
        "bin", "sbin", "lib", "lib64", "opt",
        "root", "srv", "vmlinuz", "initrd.img"
    };
    for (size_t i = 0; i < sizeof(root_entries)/sizeof(char*); i++) {
        char src_path[PATH_SAFE], dst_path[PATH_SAFE];
        snprintf(src_path, sizeof(src_path), "/%s", root_entries[i]);
        snprintf(dst_path, sizeof(dst_path), "%s/%s", liveroot_path, root_entries[i]);

        struct stat st;
        if (lstat(src_path, &st) != 0) continue;

        if (S_ISLNK(st.st_mode)) {
            clone_symlink(src_path, liveroot_path, root_entries[i]);
        } else if (S_ISDIR(st.st_mode)) {
            mkdir(dst_path, 0755);
            if (fortified_bind_mount(src_path, dst_path, MS_BIND | MS_REC) == 0) {
                mount(NULL, dst_path, NULL, MS_BIND | MS_REC | MS_REMOUNT | MS_RDONLY, NULL);
            }
        }
    }

    // 4. OVERLAY PER USR E VAR (Questi sono solitamente nello stesso FS del root)
    const char *ovl_dirs[] = {"usr", "var"};
    for (int i = 0; i < 2; i++) {
        char lower[PATH_SAFE], upper[PATH_SAFE], work[PATH_SAFE], merged[PATH_SAFE], src[PATH_SAFE];
        snprintf(lower, sizeof(lower), "%s/lowerdir/%s", overlay_path, ovl_dirs[i]);
        snprintf(upper, sizeof(upper), "%s/upperdir/%s", overlay_path, ovl_dirs[i]);
        snprintf(work, sizeof(work), "%s/workdir/%s", overlay_path, ovl_dirs[i]);
        snprintf(merged, sizeof(merged), "%s/%s", liveroot_path, ovl_dirs[i]);
        snprintf(src, sizeof(src), "/%s", ovl_dirs[i]);

        mkdir(lower, 0755); mkdir(upper, 0755); mkdir(work, 0755); mkdir(merged, 0755);

        if (fortified_bind_mount(src, lower, MS_BIND | MS_REC) == 0) {
            mount(NULL, lower, NULL, MS_BIND | MS_REC | MS_REMOUNT | MS_RDONLY, NULL);
        }

        char opts[CMD_MAX];
        snprintf(opts, sizeof(opts), "lowerdir=%s,upperdir=%s,workdir=%s", lower, upper, work);

        if (mount("overlay", merged, "overlay", 0, opts) != 0) {
            fprintf(stderr, "\033[1;31m[oa ERROR]\033[0m Overlay mount failed for /%s: %s\n", ovl_dirs[i], strerror(errno));
            return 1;
        }
    }

    // 5. API FILESYSTEMS
    const char *api_fs[] = {"proc", "sys", "run", "dev", "tmp"};
    char p[PATH_SAFE];
    for(int i=0; i<5; i++) {
        snprintf(p, sizeof(p), "%s/%s", liveroot_path, api_fs[i]);
        mkdir(p, 0755);
        if (strcmp(api_fs[i], "tmp") == 0) mount("tmpfs", p, "tmpfs", 0, NULL);
    }
    
    snprintf(p, sizeof(p), "%s/proc", liveroot_path); mount("proc", p, "proc", 0, NULL);
    snprintf(p, sizeof(p), "%s/sys", liveroot_path);  mount("sysfs", p, "sysfs", 0, NULL);
    snprintf(p, sizeof(p), "%s/dev", liveroot_path);  fortified_bind_mount("/dev", p, MS_BIND | MS_REC);
    snprintf(p, sizeof(p), "%s/run", liveroot_path);  fortified_bind_mount("/run", p, MS_BIND | MS_REC);

    printf("\033[1;32m[oa PREPARE]\033[0m Full environment mirrored at %s\n", liveroot_path);
    return 0;
}
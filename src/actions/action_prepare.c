/*
* oa: eggs in my dialect🥚🥚
* remastering core
*
* Author: Piero Proietti <piero.proietti@gmail.com>
* License: GPL-3.0-or-later
*/
#include "oa.h"

// Helper per creare directory con buffer sicuro
static int make_full_dir(const char *base, const char *rel) {
    char full_path[PATH_SAFE];
    snprintf(full_path, sizeof(full_path), "%s/%s", base, rel);
    if (mkdir(full_path, 0755) != 0 && errno != EEXIST)
        return -1;
    return 0;
}

// Helper per clonare un link simbolico dall'host al liveroot
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

// Bind mount fortificato (aggiunge MS_PRIVATE per evitare propagazione inversa)
static int fortified_bind_mount(const char *src, const char *tgt, unsigned long flags) {
    if (mount(src, tgt, NULL, flags, NULL) != 0)
        return -1;
    mount(NULL, tgt, NULL, MS_PRIVATE | MS_REC, NULL);
    return 0;
}

int action_prepare(cJSON *json) {
    cJSON *pathLiveFs = cJSON_GetObjectItemCaseSensitive(json, "pathLiveFs");
    if (!cJSON_IsString(pathLiveFs)) return 1;

    const char *base = pathLiveFs->valuestring;
    char liveroot_path[PATH_SAFE], overlay_path[PATH_SAFE];
    
    snprintf(liveroot_path, sizeof(liveroot_path), "%s/liveroot", base);
    snprintf(overlay_path, sizeof(overlay_path), "%s/.overlay", base);

    printf("{\"status\": \"starting\", \"action\": \"prepare_mirror\", \"path\": \"%s\"}\n", base);

    // 1. Setup Struttura Base
    mkdir(base, 0755);
    make_full_dir(base, "liveroot");
    make_full_dir(base, ".overlay");
    make_full_dir(base, ".overlay/lowerdir");
    make_full_dir(base, ".overlay/upperdir");
    make_full_dir(base, ".overlay/workdir");

    // 2. COPIA DI /ETC (L'unica fisica) - Usiamo CMD_MAX
    char cp_cmd[CMD_MAX];
    snprintf(cp_cmd, sizeof(cp_cmd), "cp -a /etc %s/", liveroot_path);
    system(cp_cmd);

    // 3. ANALISI E CLONAZIONE DELLA ROOT
    const char *root_entries[] = {
        "bin", "sbin", "lib", "lib64", "boot", "opt",
        "root", "srv", "vmlinuz", "vmlinuz.old", "initrd.img", "initrd.img.old"
    };
    int entries_count = sizeof(root_entries) / sizeof(root_entries[0]);

    for (int i = 0; i < entries_count; i++) {
        char src_path[PATH_SAFE];
        snprintf(src_path, sizeof(src_path), "/%s", root_entries[i]);

        struct stat st;
        if (lstat(src_path, &st) != 0) continue;

        if (S_ISLNK(st.st_mode)) {
            clone_symlink(src_path, liveroot_path, root_entries[i]);
        } else if (S_ISDIR(st.st_mode)) {
            char dst_path[PATH_SAFE];
            snprintf(dst_path, sizeof(dst_path), "%s/%s", liveroot_path, root_entries[i]);
            mkdir(dst_path, 0755);
            if (fortified_bind_mount(src_path, dst_path, MS_BIND | MS_REC) == 0) {
                mount(NULL, dst_path, NULL, MS_BIND | MS_REC | MS_REMOUNT | MS_RDONLY, NULL);
            }
        }
    }

    // 4. OVERLAY PER USR E VAR
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

        char opts[CMD_MAX]; // Overlayfs vuole molto spazio per le opzioni
        snprintf(opts, sizeof(opts), "lowerdir=%s,upperdir=%s,workdir=%s,index=off,metacopy=off,xino=off",
                 lower, upper, work);

        if (mount("overlay", merged, "overlay", 0, opts) == 0) {
            mount(NULL, merged, NULL, MS_PRIVATE | MS_REC, NULL);
        }
    }

    // 5. DIRECTORY VUOTE E API FILESYSTEMS
    const char *empty_dirs[] = {"home", "media", "mnt", "run"};
    for (int i = 0; i < 4; i++) make_full_dir(liveroot_path, empty_dirs[i]);

    char p[PATH_SAFE];
    snprintf(p, sizeof(p), "%s/tmp", liveroot_path);
    mkdir(p, 01777); chmod(p, 01777);

    snprintf(p, sizeof(p), "%s/proc", liveroot_path);
    mkdir(p, 0755); mount("proc", p, "proc", 0, NULL);

    snprintf(p, sizeof(p), "%s/sys", liveroot_path);
    mkdir(p, 0755); mount("sysfs", p, "sysfs", 0, NULL);

    snprintf(p, sizeof(p), "%s/dev", liveroot_path);
    mkdir(p, 0755); fortified_bind_mount("/dev", p, MS_BIND | MS_REC);

    snprintf(p, sizeof(p), "%s/dev/pts", liveroot_path);
    mkdir(p, 0755); mount("devpts", p, "devpts", MS_NOSUID | MS_NOEXEC, "ptmxmode=0666");

    printf("{\"status\": \"ok\", \"action\": \"prepare_complete\", \"liveroot\": \"%s\"}\n", liveroot_path);
    return 0;
}

int action_cleanup(cJSON *json) {
    cJSON *pathLiveFs = cJSON_GetObjectItemCaseSensitive(json, "pathLiveFs");
    if (!cJSON_IsString(pathLiveFs)) return 1;

    const char *base = pathLiveFs->valuestring;
    char liveroot_path[PATH_SAFE], overlay_path[PATH_SAFE];
    snprintf(liveroot_path, sizeof(liveroot_path), "%s/liveroot", base);
    snprintf(overlay_path, sizeof(overlay_path), "%s/.overlay", base);

    printf("{\"status\": \"starting\", \"action\": \"cleanup\", \"path\": \"%s\"}\n", base);

    char p[PATH_SAFE];
    snprintf(p, sizeof(p), "%s/dev/pts", liveroot_path); umount2(p, MNT_DETACH);
    snprintf(p, sizeof(p), "%s/dev", liveroot_path);     umount2(p, MNT_DETACH);
    snprintf(p, sizeof(p), "%s/sys", liveroot_path);     umount2(p, MNT_DETACH);
    snprintf(p, sizeof(p), "%s/proc", liveroot_path);    umount2(p, MNT_DETACH);

    const char *ovl_dirs[] = {"var", "usr"};
    for (int i = 0; i < 2; i++) {
        char merged[PATH_SAFE], lower[PATH_SAFE];
        snprintf(merged, sizeof(merged), "%s/%s", liveroot_path, ovl_dirs[i]);
        snprintf(lower, sizeof(lower), "%s/lowerdir/%s", overlay_path, ovl_dirs[i]);
        umount2(merged, MNT_DETACH);
        umount2(lower, MNT_DETACH);
    }

    const char *selective[] = {"boot", "opt", "root", "srv"};
    for (int i = 0; i < 4; i++) {
        char full_tgt[PATH_SAFE];
        snprintf(full_tgt, sizeof(full_tgt), "%s/%s", liveroot_path, selective[i]);
        struct stat st;
        if (lstat(full_tgt, &st) == 0 && S_ISDIR(st.st_mode)) {
            umount2(full_tgt, MNT_DETACH);
        }
    }

    printf("{\"status\": \"ok\", \"action\": \"cleanup_complete\"}\n");
    return 0;
}

/*
 * oa: eggs in my dialect🥚🥚
 * remastering core - Environment Preparation
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

// Bind mount fortificato con MS_PRIVATE per evitare propagazione inversa
static int fortified_bind_mount(const char *src, const char *tgt, unsigned long flags) {
    if (mount(src, tgt, NULL, flags, NULL) != 0)
        return -1;
    mount(NULL, tgt, NULL, MS_PRIVATE | MS_REC, NULL);
    return 0;
}

int action_prepare(OA_Context *ctx) {
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

    printf("{\"status\": \"starting\", \"action\": \"prepare_mirror\", \"path\": \"%s\", \"mode\": \"%s\"}\n", base, mode);

    // 1. Setup Struttura Base 
    mkdir(base, 0755);
    make_full_dir(base, "liveroot");
    make_full_dir(base, ".overlay");
    make_full_dir(base, ".overlay/lowerdir");
    make_full_dir(base, ".overlay/upperdir");
    make_full_dir(base, ".overlay/workdir");

    // 2. COPIA DI /ETC (L'unica fisica)
    char cp_cmd[CMD_MAX];
    snprintf(cp_cmd, sizeof(cp_cmd), "cp -a /etc %s/", liveroot_path);
    system(cp_cmd);

    // 3. ANALISI E CLONAZIONE DELLA ROOT (Bind mounts sola lettura)
    const char *root_entries[] = {
        "bin", "sbin", "lib", "lib64", "boot", "opt",
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

    // 3.5 GESTIONE /HOME IN BASE AL MODE
    char home_dst[PATH_SAFE];
    snprintf(home_dst, sizeof(home_dst), "%s/home", liveroot_path);
    mkdir(home_dst, 0755); // Creata vuota per 'standard'

    if (strcmp(mode, "clone") == 0 || strcmp(mode, "crypted") == 0) {
        char home_src[] = "/home";
        struct stat st;
        if (lstat(home_src, &st) == 0 && S_ISDIR(st.st_mode)) {
            if (fortified_bind_mount(home_src, home_dst, MS_BIND | MS_REC) == 0) {
                mount(NULL, home_dst, NULL, MS_BIND | MS_REC | MS_REMOUNT | MS_RDONLY, NULL);
            }
        }
        printf("\033[1;34m[oa PREPARE]\033[0m Mode '%s': /home has been bind-mounted read-only.\n", mode);
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

        char opts[CMD_MAX];
        snprintf(opts, sizeof(opts), "lowerdir=%s,upperdir=%s,workdir=%s,index=off,metacopy=off,xino=off",
                 lower, upper, work);

        mount("overlay", merged, "overlay", 0, opts);
    }

    // 5. API FILESYSTEMS
    char p[PATH_SAFE];
    const char *api_fs[] = {"proc", "sys", "run", "dev"};
    for(int i=0; i<4; i++) {
        snprintf(p, sizeof(p), "%s/%s", liveroot_path, api_fs[i]);
        mkdir(p, 0755);
    }
    
    snprintf(p, sizeof(p), "%s/proc", liveroot_path); mount("proc", p, "proc", 0, NULL);
    snprintf(p, sizeof(p), "%s/sys", liveroot_path);  mount("sysfs", p, "sysfs", 0, NULL);
    snprintf(p, sizeof(p), "%s/dev", liveroot_path);  fortified_bind_mount("/dev", p, MS_BIND | MS_REC);
    snprintf(p, sizeof(p), "%s/run", liveroot_path);  fortified_bind_mount("/run", p, MS_BIND | MS_REC);


    // 6. --- PREVENZIONE RICORSIONE GLOBALE (Anti-Inception) ---
    // Invece di mascherare le singole cartelle (che potrebbero non esistere ancora),
    // mascheriamo l'intero percorso di lavoro all'interno della liveroot.
    char nested_base[PATH_SAFE];
    snprintf(nested_base, sizeof(nested_base), "%s%s", liveroot_path, base);

    struct stat nst;
    if (lstat(nested_base, &nst) == 0 && S_ISDIR(nst.st_mode)) {
        // Montiamo un tmpfs vuoto sopra la cartella di lavoro riflessa
        mount("tmpfs", nested_base, "tmpfs", 0, "size=1k");
        printf("\033[1;32m[oa PREPARE]\033[0m Anti-recursion mask applied on: %s\n", nested_base);
    }

    printf("\033[1;32m[oa PREPARE]\033[0m Full environment mirrored at %s\n", liveroot_path);
    return 0;
}

int action_cleanup(OA_Context *ctx) {
    cJSON *path_item = cJSON_GetObjectItemCaseSensitive(ctx->task, "pathLiveFs");
    if (!path_item) path_item = cJSON_GetObjectItemCaseSensitive(ctx->root, "pathLiveFs");
    if (!cJSON_IsString(path_item)) return 1;

    const char *base = path_item->valuestring;
    char liveroot_path[PATH_SAFE];
    snprintf(liveroot_path, sizeof(liveroot_path), "%s/liveroot", base);

    // 1. Smontiamo la maschera Anti-Ricorsione Globale se presente
    char nested_base[PATH_SAFE];
    snprintf(nested_base, sizeof(nested_base), "%s%s", liveroot_path, base);
    umount2(nested_base, MNT_DETACH);

    // 2. Smontaggio principale in ordine inverso, includendo /home
    char p[PATH_SAFE];
    const char *to_umount[] = {"dev/pts", "dev", "run", "sys", "proc", "usr", "var", "home"};
    for(int i=0; i<8; i++) {
        snprintf(p, sizeof(p), "%s/%s", liveroot_path, to_umount[i]);
        umount2(p, MNT_DETACH); // Fallisce silenziosamente se non era montato
    }

    printf("\033[1;32m[oa CLEANUP]\033[0m Cleanup completed.\n");
    return 0;
}

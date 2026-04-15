#include "oa.h"
#include <sys/utsname.h>

int remaster_livestruct(OA_Context *ctx) {
    cJSON *path_obj = cJSON_GetObjectItemCaseSensitive(ctx->task, "pathLiveFs");
    if (!path_obj) path_obj = cJSON_GetObjectItemCaseSensitive(ctx->root, "pathLiveFs");
    if (!cJSON_IsString(path_obj)) return 1;

    const char *work_base = path_obj->valuestring;
    char liveroot_boot[PATH_SAFE];
    char iso_live_dir[PATH_SAFE];

    snprintf(liveroot_boot, sizeof(liveroot_boot), "%s/liveroot/boot", work_base);
    snprintf(iso_live_dir, sizeof(iso_live_dir), "%s/iso/live", work_base);

    char cmd[CMD_MAX];
    snprintf(cmd, sizeof(cmd), "mkdir -p %s", iso_live_dir);
    system(cmd);

    struct utsname buffer;
    uname(&buffer);
    char *kver = buffer.release;

    printf("\033[1;34m[oa LIVESTRUCT]\033[0m Gathering boot files (Universal Search)...\n");

    // -------------------------------------------------------------------------
    // 1. RICERCA INITRD / INITRAMFS
    // -------------------------------------------------------------------------
    bool found_initrd = false;
    const char *initrd_patterns[] = {
        "initrd.img",                       // Custom (generato da noi)
        "initrd.img-%s",                    // Debian / Ubuntu
        "initramfs-%s.img",                 // Fedora / RHEL / Arch (versionato)
        "initramfs-linux.img",              // Arch Standard
        "initramfs-linux-lts.img"           // Arch LTS
    };

    for (size_t i = 0; i < sizeof(initrd_patterns)/sizeof(char*); i++) {
        char src[PATH_SAFE];
        snprintf(src, sizeof(src), initrd_patterns[i], kver); // Applica kver se c'è %s
        
        snprintf(cmd, sizeof(cmd), "cp %s/%s %s/initrd.img 2>/dev/null", liveroot_boot, src, iso_live_dir);
        if (system(cmd) == 0) {
            LOG_INFO("Initrd found using pattern: %s", src);
            found_initrd = true;
            break;
        }
    }

    if (!found_initrd) {
        LOG_ERR("Could not find any initrd/initramfs in %s", liveroot_boot);
        return 1;
    }

    // -------------------------------------------------------------------------
    // 2. RICERCA KERNEL (VMLINUZ)
    // -------------------------------------------------------------------------
    bool found_vmlinuz = false;
    const char *vmlinuz_patterns[] = {
        "vmlinuz-%s",       // Debian / Fedora / RHEL (versionato)
        "vmlinuz-linux",    // Arch Standard
        "vmlinuz-linux-lts" // Arch LTS
    };

    for (size_t i = 0; i < sizeof(vmlinuz_patterns)/sizeof(char*); i++) {
        char src[PATH_SAFE];
        snprintf(src, sizeof(src), vmlinuz_patterns[i], kver);
        
        snprintf(cmd, sizeof(cmd), "cp %s/%s %s/vmlinuz 2>/dev/null", liveroot_boot, src, iso_live_dir);
        if (system(cmd) == 0) {
            LOG_INFO("Kernel found using pattern: %s", src);
            found_vmlinuz = true;
            break;
        }
    }

    if (!found_vmlinuz) {
        LOG_ERR("Could not find any vmlinuz kernel in %s", liveroot_boot);
        return 1;
    }

    // -------------------------------------------------------------------------
    // 3. FIX PERMESSI
    // -------------------------------------------------------------------------
    snprintf(cmd, sizeof(cmd), "chmod 644 %s/vmlinuz %s/initrd.img", iso_live_dir, iso_live_dir);
    system(cmd);

    printf("\033[1;32m[oa LIVESTRUCT]\033[0m ISO boot folder populated for %s kernel.\n", kver);
    return 0;
}

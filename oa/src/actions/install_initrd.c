#include "oa.h"
#include <sys/mount.h>

int install_initrd(OA_Context *ctx) {
    cJSON *pathLiveFs = cJSON_GetObjectItemCaseSensitive(ctx->task, "pathLiveFs");
    if (!pathLiveFs) pathLiveFs = cJSON_GetObjectItemCaseSensitive(ctx->root, "pathLiveFs");

    cJSON *initrd_cmd_tpl = cJSON_GetObjectItemCaseSensitive(ctx->task, "initrd_cmd");
    if (!initrd_cmd_tpl) initrd_cmd_tpl = cJSON_GetObjectItemCaseSensitive(ctx->root, "initrd_cmd");

    if (!cJSON_IsString(pathLiveFs) || !cJSON_IsString(initrd_cmd_tpl)) {
        LOG_ERR("oa_sysinstall_initrd: initrd_cmd or pathLiveFs missing");
        return 1;
    }

    char liveroot_dir[PATH_SAFE];
    snprintf(liveroot_dir, PATH_SAFE, "%s/liveroot", pathLiveFs->valuestring);

    // --- PREPARAZIONE MOUNT NATIIVI (Cruciale per Arch) ---
    char p_proc[PATH_SAFE], p_sys[PATH_SAFE], p_dev[PATH_SAFE];
    snprintf(p_proc, PATH_SAFE, "%s/proc", liveroot_dir);
    snprintf(p_sys, PATH_SAFE, "%s/sys", liveroot_dir);
    snprintf(p_dev, PATH_SAFE, "%s/dev", liveroot_dir);

    printf("\033[1;34m[oa SYSINSTALL]\033[0m Preparing virtual filesystems for disk initrd...\n");
    mount("proc", p_proc, "proc", 0, NULL);
    mount("sysfs", p_sys, "sysfs", 0, NULL);
    mount("/dev", p_dev, NULL, MS_BIND, NULL);

    // Rilevamento kernel
    struct utsname buffer;
    uname(&buffer);
    char *kversion = buffer.release;

    // Costruzione comando (scriviamo temporaneamente in /initrd-coa.img)
    char final_cmd[4096], chroot_cmd[4096];
    build_initrd_command(final_cmd, initrd_cmd_tpl->valuestring, "/initrd-coa.img", kversion);
    snprintf(chroot_cmd, 4096, "chroot %s /bin/bash -c \"%s\"", liveroot_dir, final_cmd);

    printf("\033[1;34m[oa]\033[0m Executing inside chroot: %s\n", final_cmd);
    int res = system(chroot_cmd);

    // --- PULIZIA IMMEDIATA ---
    umount(p_proc);
    umount(p_sys);
    umount(p_dev);

    if (res != 0) {
        LOG_ERR("Initrd generation failed on disk.");
        return 1;
    }

    // Spostiamo il file nella destinazione corretta (/boot) del sistema installato
    char mv_cmd[CMD_MAX];
    snprintf(mv_cmd, sizeof(mv_cmd), "mv %s/initrd-coa.img %s/boot/initramfs-linux.img 2>/dev/null || mv %s/initrd-coa.img %s/boot/initrd.img", 
             liveroot_dir, liveroot_dir, liveroot_dir, liveroot_dir);
    system(mv_cmd);

    printf("\033[1;32m[SUCCESS]\033[0m System initrd generated and placed in /boot\n");
    return 0;
}
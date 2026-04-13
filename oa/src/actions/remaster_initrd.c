/*
* oa: eggs in my dialect🥚🥚
* remastering core
*
* Author: Piero Proietti <piero.proietti@gmail.com>
* License: GPL-3.0-or-later
*/

#include "oa.h"

/**
 * @brief Genera l'initrd (Initial RAM Disk) usando il contesto OA
 */
int remaster_initrd(OA_Context *ctx) {
    // 1. Lookup a cascata (Locale > Globale)
    cJSON *pathLiveFs = cJSON_GetObjectItemCaseSensitive(ctx->task, "pathLiveFs");
    if (!pathLiveFs) pathLiveFs = cJSON_GetObjectItemCaseSensitive(ctx->root, "pathLiveFs");

    cJSON *initrd_cmd_tpl = cJSON_GetObjectItemCaseSensitive(ctx->task, "initrd_cmd");
    if (!initrd_cmd_tpl) initrd_cmd_tpl = cJSON_GetObjectItemCaseSensitive(ctx->root, "initrd_cmd");

    if (!cJSON_IsString(pathLiveFs) || !cJSON_IsString(initrd_cmd_tpl)) {
        fprintf(stderr, "\033[1;31m[ERROR]\033[0m initrd_cmd or pathLiveFs missing\n");
        return 1;
    }

    char live_dir[PATH_SAFE], liveroot_dir[PATH_SAFE];
    snprintf(live_dir, PATH_SAFE, "%s/iso/live", pathLiveFs->valuestring);
    snprintf(liveroot_dir, PATH_SAFE, "%s/liveroot", pathLiveFs->valuestring);

    // 2. Rilevamento kernel host (Visto che è un remastering live, host = chroot)
    struct utsname buffer;
    if (uname(&buffer) != 0) {
        fprintf(stderr, "\033[1;31m[ERROR]\033[0m Failed to execute uname\n");
        return 1;
    }
    char *kversion = buffer.release;
    printf("\033[1;34m[oa]\033[0m Host kernel detected: %s\n", kversion);

    // 3. Costruzione comando. 
    char final_cmd[4096], chroot_cmd[4096];
    
    // FONDAMENTALE: Scriviamo nella radice del chroot per aggirare il blocco su /boot
    char chroot_out[] = "/initrd-coa.img"; 

    build_initrd_command(final_cmd, initrd_cmd_tpl->valuestring, chroot_out, kversion);

    // Wrappiamo il comando finale in un chroot per l'isolamento
    snprintf(chroot_cmd, 4096, "chroot %s /bin/bash -c \"%s\"", liveroot_dir, final_cmd);

    printf("\033[1;34m[oa]\033[0m Executing inside chroot: %s\n", final_cmd);
    // /tmp DEVE ESISTERE ed essere SCRIVIVBILE in chroot
    char fix_tmp_cmd[PATH_SAFE];
    snprintf(fix_tmp_cmd, PATH_SAFE, "mkdir -p %s/tmp && chmod 1777 %s/tmp", liveroot_dir, liveroot_dir);
    system(fix_tmp_cmd);

    
    // Assicuriamoci che la cartella di destinazione su host esista
    char mkdir_cmd[PATH_SAFE];
    snprintf(mkdir_cmd, PATH_SAFE, "mkdir -p %s", live_dir);
    system(mkdir_cmd);

    // 4. Esecuzione generatore Initrd
    if (system(chroot_cmd) != 0) {
        fprintf(stderr, "\033[1;31m[ERROR]\033[0m initrd generation failed\n");
        return 1;
    }

    // 5. Estrazione: spostiamo l'initrd appena generato dal chroot alla cartella ISO
    char extract_cmd[PATH_SAFE];
    snprintf(extract_cmd, PATH_SAFE, "mv %s/initrd-coa.img %s/initrd.img", liveroot_dir, live_dir);
    if (system(extract_cmd) != 0) {
        fprintf(stderr, "\033[1;31m[ERROR]\033[0m Failed to extract generated initrd.img to iso/live/\n");
        return 1;
    }

    // Copia del vmlinuz per avere il kernel allineato nella cartella live (preso dall'host!)
    char cp_kernel_cmd[PATH_SAFE];
    snprintf(cp_kernel_cmd, PATH_SAFE, "cp /boot/vmlinuz-%s %s/vmlinuz >/dev/null 2>&1", kversion, live_dir);
    system(cp_kernel_cmd);

    printf("\033[1;32m[SUCCESS]\033[0m initrd and kernel ready in %s\n", live_dir);

    return 0;
}

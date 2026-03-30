/*
* oa: eggs in my dialect🥚🥚
* remastering core
*
* Author: Piero Proietti <piero.proietti@gmail.com>
* License: GPL-3.0-or-later
*/
#include "oa.h"

/**
 * @brief Genera l'initrd (Initial RAM Disk) usando il template fornito
 */
int action_initrd(cJSON *json) {
    cJSON *pathLiveFs = cJSON_GetObjectItemCaseSensitive(json, "pathLiveFs");
    cJSON *initrd_cmd_tpl = cJSON_GetObjectItemCaseSensitive(json, "initrd_cmd");

    if (!cJSON_IsString(pathLiveFs) || !cJSON_IsString(initrd_cmd_tpl)) {
        fprintf(stderr, "{\"error\": \"initrd_cmd or pathLiveFs missing\"}\n");
        return 1;
    }

    char live_dir[PATH_SAFE], final_cmd[4096], initrd_out[PATH_SAFE];
    snprintf(live_dir, PATH_SAFE, "%s/iso/live", pathLiveFs->valuestring);
    snprintf(initrd_out, PATH_SAFE, "%s/initrd.img", live_dir);

    // Rilevamento kernel host
    struct utsname buffer;
    uname(&buffer);
    char *kversion = buffer.release;

    // Usiamo l'helper build_initrd_command che abbiamo scritto prima
    build_initrd_command(final_cmd, initrd_cmd_tpl->valuestring, initrd_out, kversion);

    printf("\033[1;34m[oa]\033[0m Generating initrd: %s\n", final_cmd);
    
    // Assicuriamoci che la cartella esista prima di lanciare il comando
    char mkdir_cmd[PATH_SAFE];
    snprintf(mkdir_cmd, PATH_SAFE, "mkdir -p %s", live_dir);
    system(mkdir_cmd);

    if (system(final_cmd) != 0) {
        fprintf(stderr, "{\"status\": \"error\", \"msg\": \"initrd generation failed\"}\n");
        return 1;
    }

    return 0;
}
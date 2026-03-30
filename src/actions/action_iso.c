/*
* oa: eggs in my dialect🥚🥚
* remastering core
*
* Author: Piero Proietti <piero.proietti@gmail.com>
* License: GPL-3.0-or-later
*/
#include "oa.h"

/**
 * @brief Finalizza la ISO avviabile
 */
#include "oa.h"

/**
 * @brief Finalizza la ISO avviabile
 */
int action_iso(cJSON *json) {
    // Cerchiamo le chiavi ESATTE che sono nel tuo plan.json
    cJSON *pathLiveFs = cJSON_GetObjectItemCaseSensitive(json, "pathLiveFs");
    cJSON *volid = cJSON_GetObjectItemCaseSensitive(json, "volid");      // Corretto da volume_id
    cJSON *output_iso = cJSON_GetObjectItemCaseSensitive(json, "output_iso"); // Corretto da filename

    if (!cJSON_IsString(pathLiveFs)) {
        fprintf(stderr, "{\"error\": \"pathLiveFs mancante nell'azione ISO\"}\n");
        return 1;
    }

    char iso_root[PATH_SAFE], output_iso_path[PATH_SAFE];
    
    // Definiamo la sorgente (cartella iso)
    snprintf(iso_root, sizeof(iso_root), "%s/iso", pathLiveFs->valuestring);
    
    // Definiamo la destinazione (file .iso)
    snprintf(output_iso_path, sizeof(output_iso_path), "%s/%s", 
             pathLiveFs->valuestring, 
             cJSON_IsString(output_iso) ? output_iso->valuestring : "live-system.iso");

    // Usiamo CMD_MAX (32K) per il comando xorriso
    char cmd[CMD_MAX];
    snprintf(cmd, sizeof(cmd),
             "xorriso -as mkisofs -J -joliet-long -r -l -iso-level 3 "
             "-isohybrid-mbr /usr/lib/ISOLINUX/isohdpfx.bin "
             "-partition_offset 16 -V '%s' "
             "-b isolinux/isolinux.bin -c isolinux/boot.cat "
             "-no-emul-boot -boot-load-size 4 -boot-info-table "
             "-o %s %s/",
             cJSON_IsString(volid) ? volid->valuestring : "OA_LIVE",
             output_iso_path, 
             iso_root);

    printf("\n\033[1;32m[oa ISO Mode]\033[0m Finalizing ISO: %s\n", 
           cJSON_IsString(output_iso) ? output_iso->valuestring : "live-system.iso");
           
    return system(cmd);
}

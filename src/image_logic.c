#include "image_logic.h"
#include "cJSON.h"
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

// --- HELPERS ---

static void append_eggs_exclusion(char *buffer, size_t buf_size, const char *path) {
    const char *p = (path[0] == '/') ? path + 1 : path;
    strncat(buffer, " '", buf_size - strlen(buffer) - 1);
    strncat(buffer, p, buf_size - strlen(buffer) - 1);
    strncat(buffer, "'", buf_size - strlen(buffer) - 1);
}

// --- AZIONI ---

/**
 * @brief Prepara lo scheletro della ISO (Bootloader e Kernel)
 * Ispirato alla "via di Adrian" (mx-snapshot): resiliente e standalone.
 */
#include <sys/utsname.h>

int action_skeleton(cJSON *json) {
    cJSON *pathLiveFs = cJSON_GetObjectItemCaseSensitive(json, "pathLiveFs");
    if (!cJSON_IsString(pathLiveFs)) return 1;

    char iso_dir[1024], live_dir[1024], isolinux_dir[1024];
    snprintf(iso_dir, 1024, "%s/iso", pathLiveFs->valuestring);
    snprintf(live_dir, 1024, "%s/iso/live", pathLiveFs->valuestring);
    snprintf(isolinux_dir, 1024, "%s/isolinux", iso_dir);

    // 1. Setup Struttura
    char cmd[4096];
    snprintf(cmd, sizeof(cmd), "mkdir -p %s %s %s/boot/grub", live_dir, isolinux_dir, iso_dir);
    system(cmd);

    // 2. Rilevamento Kernel
    struct utsname buffer;
    if (uname(&buffer) != 0) return 1;
    char *kversion = buffer.release;

    printf("{\"status\": \"imaging\", \"step\": \"skeleton\", \"kernel\": \"%s\", \"distro_hint\": \"debian_way\"}\n", kversion);

    // 3. Copia Kernel (usando il percorso diretto di /boot per evitare symlink rotti)
    printf("\033[1;34m[Vitellus]\033[0m Copying kernel vmlinuz-%s...\n", kversion);
    snprintf(cmd, sizeof(cmd), "cp /boot/vmlinuz-%s %s/vmlinuz", kversion, live_dir);
    if (system(cmd) != 0) {
        // Fallback estremo al symlink in root
        system("cp -L /vmlinuz %s/vmlinuz");
    }

    // 4. Generazione Initrd (Debian-Only for now)
    printf("\033[1;34m[Vitellus]\033[0m Generating initrd via mkinitramfs...\n");
    if (access("/usr/sbin/mkinitramfs", X_OK) == 0) {
        snprintf(cmd, sizeof(cmd), "mkinitramfs -o %s/initrd.img %s", live_dir, kversion);
        int res = system(cmd);
        if (res != 0) {
            fprintf(stderr, "{\"status\": \"error\", \"msg\": \"mkinitramfs failed\"}\n");
            return 1;
        }
    } else {
        printf("\033[1;31m[Warning]\033[0m mkinitramfs not found. Falling back to copy (risky)...\n");
        snprintf(cmd, sizeof(cmd), "cp /boot/initrd.img-%s %s/initrd.img || cp -L /initrd.img %s/initrd.img", 
                 kversion, live_dir, live_dir);
        system(cmd);
    }

    // 5. Bootloader BIOS (Isolinux)
    printf("\033[1;34m[Vitellus]\033[0m Populating Isolinux binaries...\n");
    // Copiamo i file critici. Nota: usiamo il wildcard per i moduli .c32
    snprintf(cmd, sizeof(cmd), "cp /usr/lib/ISOLINUX/isolinux.bin %s/ && "
                               "cp /usr/lib/syslinux/modules/bios/*.c32 %s/", 
             isolinux_dir, isolinux_dir);
    system(cmd);

    // 6. Configurazione Isolinux (Se manca, la creiamo)
    char cfg_path[1024];
    snprintf(cfg_path, 1024, "%s/isolinux.cfg", isolinux_dir);
    if (access(cfg_path, F_OK) != 0) {
        FILE *f = fopen(cfg_path, "w");
        if (f) {
            fprintf(f, "UI vesamenu.c32\n"
                       "PROMPT 0\n"
                       "TIMEOUT 50\n"
                       "DEFAULT live\n"
                       "MENU TITLE Vitellus Boot Menu\n"
                       "LABEL live\n"
                       "  MENU LABEL Vitellus Live (Debian Core)\n"
                       "  KERNEL /live/vmlinuz\n"
                       "  APPEND initrd=/live/initrd.img boot=live components quiet splash\n");
            fclose(f);
            printf("\033[1;32m[Vitellus]\033[0m isolinux.cfg generated.\n");
        }
    }

    return 0;
}

/**
 * @brief Crea il filesystem compresso SquashFS (Turbo Mode)
 */
int action_squash(cJSON *json) {
    cJSON *pathLiveFs = cJSON_GetObjectItemCaseSensitive(json, "pathLiveFs");
    cJSON *comp = cJSON_GetObjectItemCaseSensitive(json, "compression");
    cJSON *comp_lvl = cJSON_GetObjectItemCaseSensitive(json, "compression_level");
    cJSON *exclude_file = cJSON_GetObjectItemCaseSensitive(json, "exclude_list");
    cJSON *include_root_home = cJSON_GetObjectItemCaseSensitive(json, "include_root_home");

    if (!cJSON_IsString(pathLiveFs)) return 1;

    long nprocs = sysconf(_SC_NPROCESSORS_ONLN);
    int level = cJSON_IsNumber(comp_lvl) ? comp_lvl->valueint : 3;
    const char *comp_str = cJSON_IsString(comp) ? comp->valuestring : "zstd";

    char liveroot[1024], squash_out[1024];
    snprintf(liveroot, 1024, "%s/liveroot", pathLiveFs->valuestring);
    snprintf(squash_out, 1024, "%s/iso/live/filesystem.squashfs", pathLiveFs->valuestring);

    char session_excludes[4096] = "";
    const char *fexcludes[] = {
        "boot/efi/EFI", "boot/loader/entries/", "etc/fstab", "var/lib/docker/",
        "proc/*", "sys/*", "dev/*", "run/*", "tmp/*"
    };
    for (size_t i = 0; i < 9; i++) append_eggs_exclusion(session_excludes, 4096, fexcludes[i]);

    if (!cJSON_IsBool(include_root_home) || !include_root_home->valueint) {
        append_eggs_exclusion(session_excludes, 4096, "root/*");
    }

    char cmd[8192], comp_opts[256] = "";
    if (strcmp(comp_str, "zstd") == 0) snprintf(comp_opts, 256, "-Xcompression-level %d", level);

    snprintf(cmd, sizeof(cmd), "mksquashfs %s %s -comp %s %s -processors %ld -b 1M -noappend -wildcards", 
             liveroot, squash_out, comp_str, comp_opts, nprocs);

    if (cJSON_IsString(exclude_file)) snprintf(cmd + strlen(cmd), 8192 - strlen(cmd), " -ef %s", exclude_file->valuestring);
    if (strlen(session_excludes) > 0) snprintf(cmd + strlen(cmd), 8192 - strlen(cmd), " -e%s", session_excludes);

    printf("\n\033[1;34m[Vitellus Turbo Squash]\033[0m Cores: %ld | Lvl: %d\n", nprocs, level);
    return system(cmd);
}

/**
 * @brief Genera la ISO avviabile con xorriso
 */
int action_iso(cJSON *json) {
    cJSON *pathLiveFs = cJSON_GetObjectItemCaseSensitive(json, "pathLiveFs");
    cJSON *volid = cJSON_GetObjectItemCaseSensitive(json, "volume_id");
    cJSON *iso_name = cJSON_GetObjectItemCaseSensitive(json, "filename");

    if (!cJSON_IsString(pathLiveFs)) return 1;

    char iso_root[1024], output_iso[1024];
    snprintf(iso_root, 1024, "%s/iso", pathLiveFs->valuestring);
    snprintf(output_iso, 1024, "%s/%s", pathLiveFs->valuestring, 
             cJSON_IsString(iso_name) ? iso_name->valuestring : "live-system.iso");

    char cmd[8192];
    snprintf(cmd, sizeof(cmd),
             "xorriso -as mkisofs -J -joliet-long -r -l -iso-level 3 "
             "-isohybrid-mbr /usr/lib/ISOLINUX/isohdpfx.bin "
             "-partition_offset 16 -V '%s' "
             "-b isolinux/isolinux.bin -c isolinux/boot.cat "
             "-no-emul-boot -boot-load-size 4 -boot-info-table "
             "-o %s %s/",
             cJSON_IsString(volid) ? volid->valuestring : "VITELIUS_LIVE",
             output_iso, iso_root);

    printf("\n\033[1;34m[Vitellus ISO Mode]\033[0m Finalizing ISO...\n");
    return system(cmd);
}

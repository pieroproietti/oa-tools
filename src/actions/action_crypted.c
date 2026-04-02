/*
 * src/actions/action_crypted.c
 * Remastering core: LUKS Encryption Wrapper
 * oa: eggs in my dialect🥚🥚
 *
 * Author: Piero Proietti <piero.proietti@gmail.com>
 * License: GPL-3.0-or-later
 */
#include "oa.h"

int action_crypted(OA_Context *ctx) {
    // 1. Lookup a cascata (Locale > Globale)
    cJSON *pathLiveFs = cJSON_GetObjectItemCaseSensitive(ctx->task, "pathLiveFs");
    if (!pathLiveFs) pathLiveFs = cJSON_GetObjectItemCaseSensitive(ctx->root, "pathLiveFs");

    cJSON *pass_obj = cJSON_GetObjectItemCaseSensitive(ctx->task, "crypted_password");
    if (!pass_obj) pass_obj = cJSON_GetObjectItemCaseSensitive(ctx->root, "crypted_password");

    if (!cJSON_IsString(pathLiveFs)) {
        LOG_ERR("action_crypted: pathLiveFs missing");
        return 1;
    }

    // Se non c'è una password nel JSON, usiamo "evolution" come default
    const char *luks_pass = cJSON_IsString(pass_obj) ? pass_obj->valuestring : "evolution";

    char squash_path[PATH_SAFE], root_img_path[PATH_SAFE];
    snprintf(squash_path, sizeof(squash_path), "%s/iso/live/filesystem.squashfs", pathLiveFs->valuestring);
    snprintf(root_img_path, sizeof(root_img_path), "%s/iso/live/root.img", pathLiveFs->valuestring);

    // 2. Verifica esistenza e calcolo dimensioni di filesystem.squashfs
    struct stat st;
    if (stat(squash_path, &st) != 0) {
        LOG_ERR("action_crypted: %s non trovato. Lo step action_squash ha fallito?", squash_path);
        return 1;
    }

    // Aggiungiamo ~500MB di tolleranza (ext4 overhead, journal e LUKS header)
    uint64_t needed_mb = (st.st_size / (1024 * 1024)) + 500;
    
    printf("\033[1;35m[oa CRYPT]\033[0m Encapsulating filesystem.squashfs into LUKS container...\n");
    printf("\033[1;35m[oa CRYPT]\033[0m Allocated size for root.img: %lu MB\n", needed_mb);

    char cmd[CMD_MAX];
    int res;

    // 3. Creazione file vuoto root.img (fallocate è istantaneo, fallback su dd)
    snprintf(cmd, sizeof(cmd), "fallocate -l %luM %s || dd if=/dev/zero of=%s bs=1M count=%lu status=none", 
             needed_mb, root_img_path, root_img_path, needed_mb);
    if (system(cmd) != 0) {
        LOG_ERR("action_crypted: Creazione root.img fallita");
        return 1;
    }

    // 4. Formattazione LUKS (via stdin per evitare prompt interattivi)
    printf("\033[1;35m[oa CRYPT]\033[0m Formatting LUKS container...\n");
    snprintf(cmd, sizeof(cmd), "echo -n '%s' | cryptsetup luksFormat --type luks2 %s -", luks_pass, root_img_path);
    if (system(cmd) != 0) {
        LOG_ERR("action_crypted: luksFormat fallito");
        return 1;
    }

    // 5. Apertura LUKS
    const char *mapper_name = "oa-crypt-build";
    snprintf(cmd, sizeof(cmd), "echo -n '%s' | cryptsetup luksOpen %s %s -", luks_pass, root_img_path, mapper_name);
    if (system(cmd) != 0) {
        LOG_ERR("action_crypted: luksOpen fallito");
        return 1;
    }

    // 6. Formattazione EXT4 del volume mappato
    printf("\033[1;35m[oa CRYPT]\033[0m Creating ext4 filesystem inside LUKS...\n");
    snprintf(cmd, sizeof(cmd), "mkfs.ext4 -q /dev/mapper/%s", mapper_name);
    res = system(cmd);
    if (res != 0) {
        LOG_ERR("action_crypted: mkfs.ext4 fallito. Eseguo cleanup...");
        system("cryptsetup luksClose oa-crypt-build");
        return 1;
    }

    // 7. Montaggio e incapsulamento
    char mnt_dir[] = "/tmp/oa_crypt_mnt";
    mkdir(mnt_dir, 0755);
    
    snprintf(cmd, sizeof(cmd), "mount /dev/mapper/%s %s", mapper_name, mnt_dir);
    system(cmd);

    printf("\033[1;35m[oa CRYPT]\033[0m Moving filesystem.squashfs into LUKS container...\n");
    snprintf(cmd, sizeof(cmd), "mv %s %s/filesystem.squashfs", squash_path, mnt_dir);
    system(cmd);

    // 8. Smontaggio e chiusura
    printf("\033[1;35m[oa CRYPT]\033[0m Closing and finalizing container...\n");
    snprintf(cmd, sizeof(cmd), "umount %s && cryptsetup luksClose %s", mnt_dir, mapper_name);
    system(cmd);
    
    rmdir(mnt_dir);

    printf("\033[1;32m[oa CRYPT]\033[0m Encryption process completed successfully. root.img generated.\n");
    return 0;
}
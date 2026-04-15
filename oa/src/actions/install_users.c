/*
 * src/actions/install_users.c
 * Remastering core: User Identity injection for installation (Krill)
 * oa: eggs in my dialect🥚🥚
 *
 * Author: Piero Proietti <piero.proietti@gmail.com>
 * License: GPL-3.0-or-later
 */
#include "oa.h"

int install_users(OA_Context *ctx) {
    // 1. Lookup a cascata: nell'installazione i dati sono spesso nel root JSON
    cJSON *pathLiveFs = cJSON_GetObjectItemCaseSensitive(ctx->task, "pathLiveFs");
    if (!pathLiveFs) pathLiveFs = cJSON_GetObjectItemCaseSensitive(ctx->root, "pathLiveFs");
    
    cJSON *users = cJSON_GetObjectItemCaseSensitive(ctx->task, "users");
    if (!users) users = cJSON_GetObjectItemCaseSensitive(ctx->root, "users");

    if (!cJSON_IsString(pathLiveFs)) {
        LOG_ERR("pathLiveFs is missing in install_users");
        return 1;
    }

    // 2. Puntiamo DIRETTAMENTE alla radice montata sul disco fisico (es. /mnt/krill-target)
    char target_root[PATH_SAFE], p_path[PATH_SAFE], s_path[PATH_SAFE], g_path[PATH_SAFE];
    snprintf(target_root, sizeof(target_root), "%s/liveroot", pathLiveFs->valuestring);
    snprintf(p_path, sizeof(p_path), "%s/etc/passwd", target_root);
    snprintf(s_path, sizeof(s_path), "%s/etc/shadow", target_root);
    snprintf(g_path, sizeof(g_path), "%s/etc/group", target_root);

    printf("\033[1;34m[oa HATCH]\033[0m Injecting new machine owner identity...\n");

    // 3. Iniezione Yocto-style diretta
    if (cJSON_IsArray(users)) {
        FILE *fp = fopen(p_path, "a");
        FILE *fs = fopen(s_path, "a");

        if (!fp || !fs) {
            if(fp) fclose(fp);
            if(fs) fclose(fs);
            LOG_ERR("Failed to open passwd or shadow on physical disk");
            return 1;
        }

        cJSON *u;
        cJSON_ArrayForEach(u, users) {
            cJSON *login_obj  = cJSON_GetObjectItemCaseSensitive(u, "login");
            cJSON *pass_obj   = cJSON_GetObjectItemCaseSensitive(u, "password");
            cJSON *home_obj   = cJSON_GetObjectItemCaseSensitive(u, "home");
            cJSON *shell_obj  = cJSON_GetObjectItemCaseSensitive(u, "shell");
            cJSON *gecos_obj  = cJSON_GetObjectItemCaseSensitive(u, "gecos");
            cJSON *groups_obj = cJSON_GetObjectItemCaseSensitive(u, "groups");

            if (!login_obj || !pass_obj || !home_obj) {
                LOG_WARN("Skipping user entry: missing mandatory fields");
                continue;
            }

            const char *login = login_obj->valuestring;
            const char *pass  = pass_obj->valuestring;
            const char *home  = home_obj->valuestring;
            const char *shell = shell_obj ? shell_obj->valuestring : "/bin/bash";
            const char *gecos = gecos_obj ? gecos_obj->valuestring : ",,,";

            // In una nuova installazione Krill, il padrone di casa prende l'UID 1000
            cJSON *uid_obj   = cJSON_GetObjectItemCaseSensitive(u, "uid");
            cJSON *gid_obj   = cJSON_GetObjectItemCaseSensitive(u, "gid");
            int val_uid = uid_obj ? uid_obj->valueint : OE_UID_HUMAN_MIN;
            int val_gid = gid_obj ? gid_obj->valueint : OE_UID_HUMAN_MIN;

            printf("\033[1;32m[oa HATCH]\033[0m Handcrafting identity: %s on physical disk\n", login);
            LOG_INFO("Injecting user %s (UID:%d) into target %s", login, val_uid, target_root);

            // Scriviamo nativamente le stringhe C senza chroot o useradd
            yocto_write_passwd(fp, login, val_uid, val_gid, gecos, home, shell);
            yocto_write_shadow(fs, login, pass);

            if (cJSON_IsArray(groups_obj)) {
                yocto_add_user_to_groups(g_path, login, groups_obj);
            }

            // Copia dello scheletro e setup permessi sulla partizione fisica
            char home_cmd[CMD_MAX];
            snprintf(home_cmd, sizeof(home_cmd),
                     "mkdir -p %s%s && cp -a %s/etc/skel/. %s%s/ 2>/dev/null || true && chown -R %d:%d %s%s",
                     target_root, home, target_root, target_root, home, val_uid, val_gid, target_root, home);
            system(home_cmd);
        }
        fclose(fp);
        fclose(fs);
    } else {
        LOG_WARN("No 'users' array found in JSON plan. System might be unbootable.");
    }

    return 0;
}

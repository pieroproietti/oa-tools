#include "oa.h"
#include "oa-users.h"
#include "oa-yocto.h"
#include <shadow.h>
#include <crypt.h>
#include <pwd.h>

int oa_users(OA_Context *ctx) {
    // 1. Lookup dei percorsi
    cJSON *pathLiveFs = cJSON_GetObjectItemCaseSensitive(ctx->task, "pathLiveFs");
    if (!pathLiveFs) pathLiveFs = cJSON_GetObjectItemCaseSensitive(ctx->root, "pathLiveFs");
    
    cJSON *users = cJSON_GetObjectItemCaseSensitive(ctx->task, "users");
    if (!users) users = cJSON_GetObjectItemCaseSensitive(ctx->root, "users");
    
    cJSON *mode_item = cJSON_GetObjectItemCaseSensitive(ctx->task, "mode");
    const char *mode = cJSON_IsString(mode_item) ? mode_item->valuestring : "standard";

    if (!cJSON_IsString(pathLiveFs)) {
        LOG_ERR("oa_users: pathLiveFs mancante");
        return 1;
    }

    char liveroot[PATH_SAFE], p_path[PATH_SAFE], s_path[PATH_SAFE], g_path[PATH_SAFE];
    snprintf(liveroot, sizeof(liveroot), "%s/liveroot", pathLiveFs->valuestring);
    snprintf(p_path, sizeof(p_path), "%s/etc/passwd", liveroot);
    snprintf(s_path, sizeof(s_path), "%s/etc/shadow", liveroot);
    snprintf(g_path, sizeof(g_path), "%s/etc/group", liveroot);

    LOG_INFO("Inizio gestione utenti in modalità: %s", mode);

    // 2. PULIZIA
    if (strcmp(mode, "clone") != 0 && strcmp(mode, "crypted") != 0) {
        LOG_INFO("Esecuzione sanitize identità host (modalità standard)...");
        yocto_sanitize_file(p_path, OE_UID_HUMAN_MIN, OE_UID_HUMAN_MAX);
        yocto_sanitize_shadow(s_path, p_path);
        yocto_sanitize_file(g_path, OE_UID_HUMAN_MIN, OE_UID_HUMAN_MAX);
    }

    // 3. SCRITTURA
    if (cJSON_IsArray(users)) {
        FILE *fp = fopen(p_path, "a");
        FILE *fs = fopen(s_path, "a");
        
        if (!fp || !fs) { 
            LOG_ERR("Errore fatale: impossibile aprire i database utenti in %s/etc/", liveroot);
            if(fp) fclose(fp); if(fs) fclose(fs); 
            return 1;
        }

        cJSON *u;
        cJSON_ArrayForEach(u, users) {
            const char *login = cJSON_GetObjectItemCaseSensitive(u, "login")->valuestring;
            const char *pass  = cJSON_GetObjectItemCaseSensitive(u, "password")->valuestring;
            const char *home  = cJSON_GetObjectItemCaseSensitive(u, "home")->valuestring;
            
            // Logghiamo l'utente che stiamo creando
            LOG_INFO("Creazione identità nativa: user='%s' home='%s'", login, home);

            // Scrittura (qui scriviamo la password, ma non la logghiamo!)
            yocto_write_passwd(fp, login, OE_UID_HUMAN_MIN, OE_UID_HUMAN_MIN, "live,,,", home, "/bin/bash");
            yocto_write_shadow(fs, login, pass);

            // Gruppi secondari
            cJSON *groups_obj = cJSON_GetObjectItemCaseSensitive(u, "groups");
            if (cJSON_IsArray(groups_obj)) {
                LOG_INFO("Aggiunta utente '%s' ai gruppi secondari...", login);
                yocto_add_user_to_groups(g_path, login, groups_obj);
            }

            // Home e Skel
            char full_home[PATH_SAFE];
            snprintf(full_home, sizeof(full_home), "%s%s", liveroot, home);
            
            LOG_INFO("Popolamento home directory: %s", full_home);
            if (mkdir(full_home, 0755) != 0 && errno != EEXIST) {
                LOG_WARN("Attenzione: mkdir fallita per %s (errno: %d)", full_home, errno);
            }

            char home_cmd[CMD_MAX];
            snprintf(home_cmd, sizeof(home_cmd), 
                     "cp -a %s/etc/skel/. %s/ 2>/dev/null || true && chown -R %d:%d %s", 
                     liveroot, full_home, OE_UID_HUMAN_MIN, OE_UID_HUMAN_MIN, full_home);
            
            if (system(home_cmd) == 0) {
                LOG_INFO("Home di '%s' configurata con successo (skel+chown)", login);
            } else {
                LOG_ERR("Errore durante il comando di setup home per '%s'", login);
            }
        }
        fclose(fp);
        fclose(fs);
    }
    
    LOG_INFO("Gestione utenti completata.");
    return 0;
}
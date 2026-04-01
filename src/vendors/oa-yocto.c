/*
 * oa: eggs in my dialect🥚🥚
 *
 * src/vendors/oe-yocto.c
 * Logica di classificazione utenti basata su OpenEmbedded-Core
 * e sulla filosofia di penguins-eggs.
 */
#include "oa.h"

/**
 * @brief Scrive una riga in formato passwd (USER:x:UID:GID:GECOS:HOME:SHELL)
 */
void yocto_write_passwd(FILE *f, const char *user, int uid, int gid, const char *gecos, const char *home, const char *shell) {
    if (f) fprintf(f, "%s:x:%d:%d:%s:%s:%s\n", user, uid, gid, gecos, home, shell);
}

/**
 * @brief Scrive una riga in formato shadow (USER:PASS:LAST:MIN:MAX:WARN:INACT:EXP:RES)
 */
void yocto_write_shadow(FILE *f, const char *user, const char *enc_pass) {
    // 19750 è un valore di last_change approssimativo per il 2024+
    if (f) fprintf(f, "%s:%s:19750:0:99999:7:::\n", user, enc_pass);
}

/**
 * @brief Scrive una riga in formato group (GROUP:x:GID:USERS)
 */
void yocto_write_group(FILE *f, const char *group, int gid, const char *users) {
    if (f) fprintf(f, "%s:x:%d:%s\n", group, gid, users ? users : "");
}

/**
 * @brief Filtra un file di testo (passwd/group) rimuovendo gli UID/GID umani
 */
int yocto_sanitize_file(const char *src_path, int min_id, int max_id) {
    char tmp_path[PATH_SAFE];
    snprintf(tmp_path, sizeof(tmp_path), "%s.tmp", src_path);

    FILE *src = fopen(src_path, "r");
    FILE *dst = fopen(tmp_path, "w");
    if (!src || !dst) return -1;

    char line[PATH_SAFE];
    while (fgets(line, sizeof(line), src)) {

        char line_copy[PATH_SAFE];
        strcpy(line_copy, line);

        strtok(line_copy, ":");           // Salta il nome (era 'name')
        strtok(NULL, ":");                // Salta la password (era 'pass')
        char *id_str = strtok(NULL, ":"); // Questo ci serve per l'ID

        if (id_str) {
            int id = atoi(id_str);
            // Se l'ID è fuori dal range umano (OE-Core), preserviamo la riga
            if (id < min_id || id > max_id) {
                fputs(line, dst);
            }
        }
    }

    fclose(src);
    fclose(dst);
    return rename(tmp_path, src_path);
}

/**
 * @brief Controlla se un utente esiste nel file passwd
 */
static bool user_exists_in_passwd(const char *passwd_path, const char *username) {
    FILE *f = fopen(passwd_path, "r");
    if (!f) return false;
    
    char line[PATH_SAFE];
    while (fgets(line, sizeof(line), f)) {
        char line_copy[PATH_SAFE];
        strcpy(line_copy, line);
        char *user = strtok(line_copy, ":");
        if (user && strcmp(user, username) == 0) {
            fclose(f);
            return true;
        }
    }
    fclose(f);
    return false;
}

/**
 * @brief Filtra il file shadow mantenendo solo gli utenti presenti in passwd
 */
int yocto_sanitize_shadow(const char *shadow_path, const char *passwd_path) {
    char tmp_path[PATH_SAFE];
    snprintf(tmp_path, sizeof(tmp_path), "%s.tmp", shadow_path);

    FILE *src = fopen(shadow_path, "r");
    FILE *dst = fopen(tmp_path, "w");
    if (!src || !dst) {
        if (src) fclose(src);
        if (dst) fclose(dst);
        return -1;
    }

    char line[PATH_SAFE];
    int removed = 0, kept = 0;
    
    while (fgets(line, sizeof(line), src)) {
        char line_copy[PATH_SAFE];
        strcpy(line_copy, line);
        char *user = strtok(line_copy, ":");

        if (user) {
            // Se l'utente è sopravvissuto alla pulizia di passwd, lo teniamo
            if (user_exists_in_passwd(passwd_path, user)) {
                fputs(line, dst);
                kept++;
            } else {
                LOG_INFO("Shadow sanitize: purging removed user '%s'", user);
                removed++;
            }
        }
    }

    fclose(src);
    fclose(dst);

    // Ripristiniamo i permessi restrittivi richiesti da PAM per shadow
    chmod(tmp_path, 0640); 

    LOG_INFO("Sanitized shadow: kept %d, removed %d entries", kept, removed);
    return rename(tmp_path, shadow_path);
}

/**
 * @brief Verifica se il percorso della home è in una whitelist di sistema.
 */
static bool is_path_allowed(const char *home) {
    if (home == NULL || strlen(home) < 2) {
        return false;
    }

    const char *whitelist[] = {"home", "opt", "srv", "usr", "var", NULL};
    char path_tmp[PATH_SAFE];
    
    strncpy(path_tmp, home, sizeof(path_tmp));
    path_tmp[sizeof(path_tmp) - 1] = '\0'; // Sicurezza extra per il terminatore

    char *fLevel = strtok(path_tmp, "/");
    if (fLevel == NULL) return false;

    for (int i = 0; whitelist[i] != NULL; i++) {
        if (strcmp(fLevel, whitelist[i]) == 0) {
            return true;
        }
    }

    return false;
}

/**
 * @brief yocto_is_human_user
 * Decide se un utente dell'host deve essere processato.
 */
bool yocto_is_human_user(uint32_t uid, const char *home) {
    // 1. Filtro UID basato su OE-Core (1000-59999)
    if (uid < OE_UID_HUMAN_MIN || uid > OE_UID_HUMAN_MAX) {
        return false;
    }

    // 2. Controllo Whitelist dei percorsi
    if (!is_path_allowed(home)) {
        return false;
    }

    // 3. Verifica fisica
    struct stat st;
    if (stat(home, &st) != 0 || !S_ISDIR(st.st_mode)) {
        return false;
    }

    // 4. Analisi sottocartelle vietate
    if (strstr(home, "/cache") || strstr(home, "/run") || strstr(home, "/spool")) {
        return false;
    }

    return true;
}

/**
 * @brief Aggiunge nativamente un utente a un elenco di gruppi secondari in /etc/group
 */
void yocto_add_user_to_groups(const char *group_file, const char *username, cJSON *groups_array) {
    if (!cJSON_IsArray(groups_array)) return;

    char tmp_file[PATH_SAFE];
    snprintf(tmp_file, sizeof(tmp_file), "%s.tmp", group_file);

    FILE *src = fopen(group_file, "r");
    FILE *dst = fopen(tmp_file, "w");
    if (!src || !dst) {
        if (src) fclose(src);
        if (dst) fclose(dst);
        return;
    }

    char line[PATH_SAFE];
    while (fgets(line, sizeof(line), src)) {
        line[strcspn(line, "\n")] = 0; // Rimuove il newline

        char line_copy[PATH_SAFE];
        strcpy(line_copy, line);
        char *gname = strtok(line_copy, ":");

        bool match = false;
        cJSON *g;
        cJSON_ArrayForEach(g, groups_array) {
            if (cJSON_IsString(g) && strcmp(gname, g->valuestring) == 0) {
                match = true;
                break;
            }
        }

        if (match) {
            // Se la riga finisce con ':', non ci sono altri utenti. Altrimenti aggiungiamo una virgola.
            if (line[strlen(line) - 1] == ':') {
                fprintf(dst, "%s%s\n", line, username);
            } else {
                fprintf(dst, "%s,%s\n", line, username);
            }
        } else {
            fprintf(dst, "%s\n", line);
        }
    }

    fclose(src);
    fclose(dst);
    rename(tmp_file, group_file);
}
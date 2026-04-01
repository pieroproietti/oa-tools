/* yocto_ids ispirati a OpenEmbedded-Core  */
#define OE_UID_ROOT           0
#define OE_UID_SYSTEM_MAX     999
#define OE_UID_HUMAN_MIN      1000
#define OE_UID_HUMAN_MAX      59999

// Prototipi per la gestione utenti
bool yocto_is_human_user(uint32_t uid, const char *home);
void yocto_write_passwd(FILE *f, const char *user, int uid, int gid, const char *gecos, const char *home, const char *shell);
void yocto_write_shadow(FILE *f, const char *user, const char *enc_pass);
void yocto_write_group(FILE *f, const char *group, int gid, const char *users);
int yocto_sanitize_file(const char *src_path, int min_id, int max_id);
int yocto_sanitize_shadow(const char *shadow_path, const char *passwd_path);
void yocto_add_user_to_groups(const char *group_file, const char *username, cJSON *groups_array);

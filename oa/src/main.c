#include "oa.h"
int oa_mkdir(OA_Context *ctx);
int oa_bind(OA_Context *ctx);
int oa_cp(OA_Context *ctx);
int oa_mount_generic(OA_Context *ctx);
int oa_umount(OA_Context *ctx);

int execute_verb(cJSON *root, cJSON *task) {
    cJSON *command = cJSON_GetObjectItemCaseSensitive(task, "command");
    if (!cJSON_IsString(command)) return 1;

    const char *cmd_name = command->valuestring;
    OA_Context ctx = { .root = root, .task = task };

    // Log dell'azione
    cJSON *info = cJSON_GetObjectItemCaseSensitive(task, "info");
    if (cJSON_IsString(info)) printf("%s\n", info->valuestring);

    LOG_INFO(">>> Esecuzione verbo: %s", cmd_name);
    // Dispatcher

    if (strcmp(cmd_name, "oa_umount") == 0) 
        return oa_umount(&ctx);
    else if (strcmp(cmd_name, "oa_shell") == 0) {
        return oa_shell(&ctx);
    } else if (strcmp(cmd_name, "oa_users") == 0) {
        return oa_users(&ctx);
    } else if (strcmp(cmd_name, "oa_mkdir") == 0) {          // <--- AGGIUNGI QUESTO
        return oa_mkdir(&ctx);
    } else if (strcmp(cmd_name, "oa_bind") == 0) {           // <--- AGGIUNGI QUESTO
        return oa_bind(&ctx);
    } else if (strcmp(cmd_name, "oa_cp") == 0) {             // <--- AGGIUNGI QUESTO
        return oa_cp(&ctx);
    } else if (strcmp(cmd_name, "oa_mkdir") == 0) {             // <--- AGGIUNGI QUESTO
        return oa_mkdir(&ctx);
    } else if (strcmp(cmd_name, "oa_mount_generic") == 0) {  // <--- AGGIUNGI QUESTO
        return oa_mount_generic(&ctx);
    } else if (strcmp(cmd_name, "oa_umount") == 0) {         // <--- AGGIUNGI QUESTO
        return oa_umount(&ctx);
    } else {
        LOG_ERR("Unknown command requested: %s", cmd_name);
        return 1;
    }

    LOG_ERR("Comando sconosciuto: %s", cmd_name);
    return 1;
}

char *read_file(const char *filename) {
    FILE *f = fopen(filename, "rb");
    if (!f) return NULL;
    fseek(f, 0, SEEK_END);
    long len = ftell(f);
    fseek(f, 0, SEEK_SET);
    char *data = malloc(len + 1);
    if (data) {
        LOG_ERR("IO_ERROR: Impossibile aprire il piano di volo '%s' (errno: %d)", filename, errno);
        fread(data, 1, len, f);
        data[len] = '\0';
    }
    fclose(f);
    return data;
}

int main(int argc, char *argv[]) {
    if (argc < 2) {
        printf("oa engine v%s\nUsage: %s <plan.json>\n", OA_VERSION, argv[0]);
        return 1;
    }

    oa_init_log("/var/log/oa-tools.log");
    char *json_data = read_file(argv[1]);
    if (!json_data) return 1;

    cJSON *json = cJSON_Parse(json_data);
    if (!json) return 1;

    cJSON *plan = cJSON_GetObjectItemCaseSensitive(json, "plan");
    int status = 0;

    if (cJSON_IsArray(plan)) {
        cJSON *task;
        cJSON_ArrayForEach(task, plan) {
            if ((status = execute_verb(json, task)) != 0) break;
        }
    } else {
        status = execute_verb(json, json);
    }

    cJSON_Delete(json);
    free(json_data);
    oa_close_log();
    return status;
}
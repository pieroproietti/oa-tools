/*
 * oa: eggs in my dialect🥚🥚
 * remastering core
 *
 * Author: Piero Proietti <piero.proietti@gmail.com>
 * License: GPL-3.0-or-later
 */
#include "oa.h"

// Helper per leggere il file JSON
char *read_file(const char *filename) {
    FILE *f = fopen(filename, "rb");
    if (!f) return NULL;
    fseek(f, 0, SEEK_END);
    long len = ftell(f);
    fseek(f, 0, SEEK_SET);
    char *data = malloc(len + 1);
    if (data) {
        fread(data, 1, len, f);
        data[len] = '\0';
    }
    fclose(f);
    return data;
}

// Il "Vigile Urbano": smista i verbi ai vari moduli tramite OA_Context
/* oa/src/main.c */
int execute_verb(cJSON *root, cJSON *task) {
    cJSON *command = cJSON_GetObjectItemCaseSensitive(task, "command");
    if (!cJSON_IsString(command) || (command->valuestring == NULL)) {
        LOG_ERR("Task without a valid 'command' field found.");
        return 1;
    }

    const char *cmd_name = command->valuestring;
    OA_Context ctx = { .root = root, .task = task };

    LOG_INFO(">>> dispatching to: %s", cmd_name);
    printf("\033[1;34m[oa]\033[0m Executing action '%s'...\n", cmd_name);
    int status = 1;

    // --- FASE 1: REMASTER (Ex LAY) ---
    if (strcmp(cmd_name, "oa_remaster_prepare") == 0)          status = remaster_prepare(&ctx);
    else if (strcmp(cmd_name, "oa_remaster_cleanup") == 0)     status = remaster_cleanup(&ctx);
    else if (strcmp(cmd_name, "oa_remaster_crypted") == 0)     status = remaster_crypted(&ctx);
    else if (strcmp(cmd_name, "oa_remaster_initrd") == 0)      status = remaster_initrd(&ctx);
    else if (strcmp(cmd_name, "oa_remaster_iso") == 0)         status = remaster_iso(&ctx);
    else if (strcmp(cmd_name, "oa_remaster_isolinux") == 0)    status = remaster_isolinux(&ctx);
    else if (strcmp(cmd_name, "oa_remaster_livestruct") == 0)  status = remaster_livestruct(&ctx);
    else if (strcmp(cmd_name, "oa_remaster_squash") == 0)      status = remaster_squash(&ctx);
    else if (strcmp(cmd_name, "oa_remaster_uefi") == 0)        status = remaster_uefi(&ctx);
    else if (strcmp(cmd_name, "oa_remaster_users") == 0)       status = remaster_users(&ctx);

    // --- FASE 2: INSTALL (Ex HATCH) ---
    else if (strcmp(cmd_name, "oa_install_partition") == 0)    status = install_partition(&ctx);
    else if (strcmp(cmd_name, "oa_install_format") == 0)       status = install_format(&ctx);
    else if (strcmp(cmd_name, "oa_install_unpack") == 0)       status = install_unpack(&ctx);
    else if (strcmp(cmd_name, "oa_install_prepare") == 0)      status = install_prepare(&ctx);
    else if (strcmp(cmd_name, "oa_install_fstab") == 0)        status = install_fstab(&ctx);
    else if (strcmp(cmd_name, "oa_install_initrd") == 0)       status = install_initrd(&ctx);
    else if (strcmp(cmd_name, "oa_install_users") == 0)        status = install_users(&ctx);
    else if (strcmp(cmd_name, "oa_install_uefi") == 0)         status = install_uefi(&ctx);
    else if (strcmp(cmd_name, "oa_install_bios") == 0)         status = install_bios(&ctx);
    else if (strcmp(cmd_name, "oa_install_cleanup") == 0)      status = remaster_cleanup(&ctx);


    // --- FASE 3: SYS (Utility) ---
    else if (strcmp(cmd_name, "oa_sys_shell") == 0)            status = sys_shell(&ctx); 
    else if (strcmp(cmd_name, "oa_sys_run") == 0)              status = sys_run(&ctx);
    else if (strcmp(cmd_name, "oa_sys_scan") == 0)             status = sys_scan(&ctx);
    else if (strcmp(cmd_name, "oa_sys_suspend") == 0)          status = sys_suspend(&ctx);

    else {
        LOG_ERR("Unknown command requested: %s", cmd_name);
        return 1;
    }

    return status;
}

int main(int argc, char *argv[]) {
    if (argc < 2) {
        printf("oa engine v%s\nUsage: %s <plan.json>\n",OA_VERSION, argv[0]);
        return 1;
    }

    // Inizializziamo il logger subito (es. oa.log per chiarezza)
    oa_init_log("oa.log");
    LOG_INFO("=== STARTING OA ENGINE ===");
    LOG_INFO("Input plan: %s", argv[1]);

    char *json_data = read_file(argv[1]);
    if (!json_data) {
        LOG_ERR("Could not read file: %s", argv[1]);
        fprintf(stderr, "Error: Could not read file %s\n", argv[1]);
        oa_close_log();
        return 1;
    }

    cJSON *json = cJSON_Parse(json_data);
    if (!json) {
        LOG_ERR("JSON parsing failed for %s", argv[1]);
        fprintf(stderr, "Error: Invalid JSON format\n");
        free(json_data);
        oa_close_log();
        return 1;
    }

    cJSON *plan = cJSON_GetObjectItemCaseSensitive(json, "plan");
    int final_status = 0;

    // --- LOGICA DEL PIANO DI VOLO ---
    if (cJSON_IsArray(plan)) {
        LOG_INFO("Plan detected: processing %d tasks", cJSON_GetArraySize(plan));
        cJSON *task;
        int step = 0;
        cJSON_ArrayForEach(task, plan) {
            step++;
            LOG_INFO("--- Task %d ---", step);
            if (execute_verb(json, task) != 0) {
                LOG_ERR("Plan halted at step %d due to previous error", step);
                fprintf(stderr, "Error: Plan halted at step %d\n", step);
                final_status = 1;
                break;
            }
        }
    } else {
        LOG_INFO("No plan array found. Executing root as a single task.");
        final_status = execute_verb(json, json);
    }

    if (final_status == 0) {
        LOG_INFO("=== PLAN COMPLETED SUCCESSFULLY ===");
    } else {
        LOG_ERR("=== PLAN FAILED ===");
    }

    cJSON_Delete(json);
    free(json_data);
    oa_close_log();
    return final_status;
}

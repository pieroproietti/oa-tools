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

// Il "Vigile Urbano": smista i verbi ai vari moduli
int execute_verb(cJSON *task) {
    cJSON *command = cJSON_GetObjectItemCaseSensitive(task, "command");
    if (!cJSON_IsString(command) || (command->valuestring == NULL))
        return 1;

    // Branding aggiornato a [oa]
    printf("\033[1;34m[oa]\033[0m Executing action '%s'...\n", command->valuestring);

    // Mappatura comandi
    if (strcmp(command->valuestring, "action_prepare") == 0)
        return action_prepare(task);
    if (strcmp(command->valuestring, "action_initrd") == 0)
        return action_initrd(task);
    if (strcmp(command->valuestring, "action_remaster") == 0) 
        return action_remaster(task); 
    if (strcmp(command->valuestring, "action_cleanup") == 0)
        return action_cleanup(task);
    if (strcmp(command->valuestring, "action_run") == 0)
        return action_run(task);
    if (strcmp(command->valuestring, "action_squash") == 0)
        return action_squash(task);
    if (strcmp(command->valuestring, "action_iso") == 0)
        return action_iso(task);

    fprintf(stderr, "{\"error\": \"Unknown command '%s'\"}\n", command->valuestring);
    return 1;
}

int main(int argc, char *argv[]) {
    if (argc < 2) {
        printf("oa engine v0.2\nUsage: %s <plan.json>\n", argv[0]);
        return 1;
    }

    char *json_data = read_file(argv[1]);
    if (!json_data) {
        fprintf(stderr, "Error: Could not read file %s\n", argv[1]);
        return 1;
    }

    cJSON *json = cJSON_Parse(json_data);
    if (!json) {
        fprintf(stderr, "Error: Invalid JSON format\n");
        free(json_data);
        return 1;
    }

    cJSON *plan = cJSON_GetObjectItemCaseSensitive(json, "plan");
    
    // Parametri Globali per l'ereditarietà
    cJSON *global_path = cJSON_GetObjectItemCaseSensitive(json, "pathLiveFs");
    cJSON *global_mode = cJSON_GetObjectItemCaseSensitive(json, "mode");
    cJSON *global_initrd = cJSON_GetObjectItemCaseSensitive(json, "initrd_cmd");

    int final_status = 0;

    // --- LOGICA DEL PIANO DI VOLO ---
    if (cJSON_IsArray(plan)) {
        cJSON *task;
        cJSON_ArrayForEach(task, plan) {
            
            // 1. Ereditarietà pathLiveFs
            if (global_path && !cJSON_GetObjectItemCaseSensitive(task, "pathLiveFs")) {
                cJSON_AddItemToObject(task, "pathLiveFs", cJSON_Duplicate(global_path, 1));
            }

            // 2. Ereditarietà mode (fondamentale per remaster e squash)
            if (global_mode && !cJSON_GetObjectItemCaseSensitive(task, "mode")) {
                cJSON_AddItemToObject(task, "mode", cJSON_Duplicate(global_mode, 1));
            }

            // 3. Ereditarietà initrd_cmd
            if (global_initrd && !cJSON_GetObjectItemCaseSensitive(task, "initrd_cmd")) {
                cJSON_AddItemToObject(task, "initrd_cmd", cJSON_Duplicate(global_initrd, 1));
            }

            if (execute_verb(task) != 0) {
                fprintf(stderr, "{\"status\": \"halted\", \"error\": \"Plan failed at task\"}\n");
                final_status = 1;
                break; 
            }
        }
    } else {
        // Fallback per comando singolo
        final_status = execute_verb(json);
    }

    cJSON_Delete(json);
    free(json_data);
    return final_status;
}

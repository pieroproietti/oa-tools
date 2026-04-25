#include "oa.h"
#include <sys/mount.h> // Necessario per MNT_DETACH e umount2

// Prototipi che potrebbero mancare nell'inclusione sopra
int oa_mkdir(OA_Context *ctx);
int oa_bind(OA_Context *ctx);
int oa_cp(OA_Context *ctx);
int oa_mount_generic(OA_Context *ctx);
int oa_umount(OA_Context *ctx);
int oa_shell(OA_Context *ctx);
int oa_users(OA_Context *ctx);

int execute_verb(cJSON *root, cJSON *task) {
    cJSON *command = cJSON_GetObjectItemCaseSensitive(task, "command");
    if (!cJSON_IsString(command)) return 1;

    const char *cmd_name = command->valuestring;
    OA_Context ctx = { .root = root, .task = task };

    // Log dell'azione (se presente "info" nel JSON)
    cJSON *info = cJSON_GetObjectItemCaseSensitive(task, "info");
    if (cJSON_IsString(info)) {
        LOG_INFO("INFO TASK: %s", info->valuestring);
    }

    LOG_INFO(">>> Esecuzione verbo: %s", cmd_name);
    
    // Dispatcher pulito senza duplicati
    if (strcmp(cmd_name, "oa_umount") == 0) {
        return oa_umount(&ctx);
    } else if (strcmp(cmd_name, "oa_shell") == 0) {
        return oa_shell(&ctx);
    } else if (strcmp(cmd_name, "oa_users") == 0) {
        return oa_users(&ctx);
    } else if (strcmp(cmd_name, "oa_mkdir") == 0) {
        return oa_mkdir(&ctx);
    } else if (strcmp(cmd_name, "oa_bind") == 0) {
        return oa_bind(&ctx);
    } else if (strcmp(cmd_name, "oa_cp") == 0) {
        return oa_cp(&ctx);
    } else if (strcmp(cmd_name, "oa_mount_generic") == 0) {
        return oa_mount_generic(&ctx);
    } else {
        LOG_ERR("Comando sconosciuto o non supportato: %s", cmd_name);
        return 1;
    }
}

char *read_file(const char *filename) {
    FILE *f = fopen(filename, "rb");
    if (!f) {
        // QUI va il log di errore reale (quando fopen fallisce)
        LOG_ERR("IO_ERROR: Impossibile aprire il piano di volo '%s' (errno: %d - %s)", 
                filename, errno, strerror(errno));
        return NULL;
    }

    fseek(f, 0, SEEK_END);
    long len = ftell(f);
    fseek(f, 0, SEEK_SET);

    char *data = malloc(len + 1);
    if (data) {
        fread(data, 1, len, f);
        data[len] = '\0';
    } else {
        // Errore di memoria allocata
        LOG_ERR("MEM_ERROR: Impossibile allocare %ld bytes per il file '%s'", len, filename);
    }
    
    fclose(f);
    return data;
}

void print_help(const char *prog_name) {
    printf("oa engine v%s - Il motore operativo di coa\n\n", OA_VERSION);
    printf("USO:\n");
    printf("  %s <percorso_piano.json>  Esegue il piano di volo specificato\n", prog_name);
    printf("  %s cleanup                Forza lo smontaggio di emergenza (MNT_DETACH)\n", prog_name);
    printf("  %s -h, --help             Mostra questo messaggio di aiuto\n\n", prog_name);
}

int main(int argc, char *argv[]) {
    // 1. Se lanciato senza argomenti, mostra l'help ed esce con errore (1)
    if (argc < 2) {
        print_help(argv[0]);
        return 1;
    }

    // 2. Se l'utente chiede esplicitamente aiuto, mostra l'help ed esce pulito (0)
    if (strcmp(argv[1], "-h") == 0 || strcmp(argv[1], "--help") == 0) {
        print_help(argv[0]);
        return 0;
    }

    oa_init_log("/var/log/oa-tools.log");
    
    // --- INTERCETTAZIONE COMANDO DI CLEANUP ---
    if (strcmp(argv[1], "cleanup") == 0) {
        LOG_INFO("Comando diretto ricevuto: Esecuzione cleanup (Emergency Unmount)...");
        
        // Smontaggio d'emergenza forzato (Lazy unmount)
        umount2("/home/eggs/liveroot/dev/pts", MNT_DETACH);
        umount2("/home/eggs/liveroot/dev", MNT_DETACH);
        umount2("/home/eggs/liveroot/proc", MNT_DETACH);
        umount2("/home/eggs/liveroot/sys", MNT_DETACH);
        umount2("/home/eggs/liveroot/run", MNT_DETACH);
        umount2("/home/eggs/liveroot", MNT_DETACH); // Smonta l'OverlayFS principale
        
        LOG_INFO("Smontaggio di emergenza completato.");
        oa_close_log();
        return 0; // Esce con successo senza cercare file JSON
    }
    // ------------------------------------------

    char *json_data = read_file(argv[1]);
    if (!json_data) {
        // Se read_file fallisce, esce in modo pulito
        oa_close_log();
        return 1; 
    }

    cJSON *json = cJSON_Parse(json_data);
    if (!json) {
        LOG_ERR("JSON_ERROR: Impossibile fare il parsing di '%s'. SINTASSI ERRATA.", argv[1]);
        free(json_data);
        oa_close_log();
        return 1;
    }

    cJSON *plan = cJSON_GetObjectItemCaseSensitive(json, "plan");
    int status = 0;

    if (cJSON_IsArray(plan)) {
        cJSON *task;
        cJSON_ArrayForEach(task, plan) {
            if ((status = execute_verb(json, task)) != 0) {
                LOG_ERR("Esecuzione interrotta. Task fallito.");
                break;
            }
        }
    } else {
        status = execute_verb(json, json);
    }

    cJSON_Delete(json);
    free(json_data);
    
    LOG_INFO("Esecuzione completata con status: %d", status);
    oa_close_log();
    return status;
}


#include "oa.h"
#include <sys/wait.h>
#include <unistd.h>

/**
 * @brief Esegue un comando tramite shell con tracciamento nel log.
 */
int sys_shell(OA_Context *ctx) {
    cJSON *cmd_obj = cJSON_GetObjectItemCaseSensitive(ctx->task, "run_command");
    cJSON *chroot_obj = cJSON_GetObjectItemCaseSensitive(ctx->task, "chroot");
    cJSON *path_obj = cJSON_GetObjectItemCaseSensitive(ctx->root, "pathLiveFs");

    if (!cJSON_IsString(cmd_obj)) {
        LOG_ERR("oa_shell: 'run_command' mancante o non valido.");
        return 1;
    }

    const char *command = cmd_obj->valuestring;
    bool use_chroot = cJSON_IsTrue(chroot_obj);

    // Logging dell'intento
    LOG_INFO("Shell Exec: %s (chroot: %s)", command, use_chroot ? "YES" : "NO");

    if (use_chroot) {
        if (!cJSON_IsString(path_obj)) {
            LOG_ERR("oa_shell: pathLiveFs richiesto per il chroot.");
            return 1;
        }

        char target_root[PATH_SAFE];
        snprintf(target_root, sizeof(target_root), "%s/liveroot", path_obj->valuestring);

        pid_t pid = fork();
        if (pid == 0) { // Processo FIGLIO
            if (chroot(target_root) != 0 || chdir("/") != 0) {
                perror("Errore ingresso chroot");
                _exit(1);
            }
            
            execl("/bin/sh", "sh", "-c", command, (char *)NULL);
            _exit(1); 
        }

        // Processo PADRE: attende la fine del comando
        int status;
        waitpid(pid, &status, 0);

        int exit_code = WIFEXITED(status) ? WEXITSTATUS(status) : 1;

        if (exit_code != 0) {
            LOG_ERR("Command failed with exit code %d: %s", exit_code, command);
        } else {
            LOG_INFO("Command completed successfully.");
        }

        return exit_code;
    } else {
        // Esecuzione standard sull'host
        int res = system(command);
        if (res != 0) {
            LOG_ERR("Host command failed (code %d): %s", res, command);
        }
        return res;
    }
}

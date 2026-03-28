// src/exec_logic.c (VERSIONE ACTION_RUN - MIRROR READY)

#define _GNU_SOURCE
#include "exec_logic.h"
#include "cJSON.h"
#include <errno.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/wait.h>
#include <unistd.h>

static int execute_command(const char *chroot_path, const char *command,
                           char *const argv[]) {
  pid_t pid = fork();

  if (pid == -1) {
    perror("fork");
    return -1;
  }

  if (pid == 0) { // --- FIGLIO ---
    // 1. Entriamo nella liveroot
    if (chdir(chroot_path) != 0) {
      perror("chdir liveroot");
      exit(EXIT_FAILURE);
    }

    // 2. Syscall CHROOT nativa
    if (chroot(chroot_path) != 0) {
      perror("chroot failed");
      exit(EXIT_FAILURE);
    }

    // Torniamo alla radice della nuova chroot
    chdir("/");

    // 3. Setup minimo di ambiente (fondamentale per apt/binari)
    // Possiamo aggiungere PATH se necessario, ma per ora usiamo quello
    // dell'host

    // 4. Esecuzione reale
    execvp(command, argv);

    // Se arriviamo qui, execvp ha fallito
    perror("execvp failed");
    exit(EXIT_FAILURE);
  } else { // --- PADRE ---
    int status;
    if (waitpid(pid, &status, 0) == -1) {
      perror("waitpid");
      return -1;
    }
    return WIFEXITED(status) ? WEXITSTATUS(status) : -1;
  }
}

int action_run(cJSON *json) {
  cJSON *pathLiveFs = cJSON_GetObjectItemCaseSensitive(json, "pathLiveFs");
  cJSON *cmd_obj = cJSON_GetObjectItemCaseSensitive(
      json, "run_command"); // Nome più esplicito
  cJSON *args_obj = cJSON_GetObjectItemCaseSensitive(json, "args");

  if (!cJSON_IsString(pathLiveFs) || !cJSON_IsString(cmd_obj)) {
    fprintf(stderr, "{\"error\": \"Missing pathLiveFs or run_command\"}\n");
    return 1;
  }

  // Costruiamo il percorso verso la liveroot
  char liveroot_path[1024];
  snprintf(liveroot_path, sizeof(liveroot_path), "%s/liveroot",
           pathLiveFs->valuestring);

  // Preparazione degli argomenti (Comando + Args + NULL)
  int args_size = cJSON_IsArray(args_obj) ? cJSON_GetArraySize(args_obj) : 0;
  char **argv = malloc(sizeof(char *) * (args_size + 2));

  argv[0] = strdup(cmd_obj->valuestring);
  for (int i = 0; i < args_size; i++) {
    cJSON *item = cJSON_GetArrayItem(args_obj, i);
    argv[i + 1] = strdup(cJSON_IsString(item) ? item->valuestring : "");
  }
  argv[args_size + 1] = NULL;

  printf("{\"status\": \"starting_chroot_exec\", \"command\": \"%s\"}\n",
         cmd_obj->valuestring);

  int exit_code = execute_command(liveroot_path, cmd_obj->valuestring, argv);

  // Pulizia memoria
  for (int i = 0; i <= args_size; i++)
    free(argv[i]);
  free(argv);

  if (exit_code == 0) {
    printf("{\"status\": \"ok\", \"exit_code\": 0}\n");
    return 0;
  } else {
    fprintf(stderr, "{\"status\": \"error\", \"exit_code\": %d}\n", exit_code);
    return exit_code;
  }
}

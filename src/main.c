// src/main.c (VERSIONE DEFINITIVA E PULITA - Solo
// action_prepare/action_cleanup)

#include "cJSON.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

// Forward declarations omogeneizzate (sarebbe meglio metterle in main.h o nei
// rispettivi .h)
int action_prepare(cJSON *json);
int action_cleanup(cJSON *json);
int action_run(cJSON *json);

// Funzione per leggere un file in memoria
char *read_file(const char *filename) {
  FILE *f = fopen(filename, "rb");
  if (!f)
    return NULL;
  fseek(f, 0, SEEK_END);
  long len = ftell(f);
  fseek(f, 0, SEEK_SET);
  char *data = malloc(len + 1);
  fread(data, 1, len, f);
  fclose(f);
  data[len] = '\0';
  return data;
}

int main(int argc, char *argv[]) {
  if (argc < 2) {
    printf("Vitellus Core v0.1\nUsage: %s <task.json>\n", argv[0]);
    return 1;
  }

  char *json_data = read_file(argv[1]);
  if (!json_data) {
    fprintf(stderr, "Error: Could not read file %s\n", argv[1]);
    return 1;
  }

  // Parsiamo il JSON e lo salviamo nella variabile 'json'
  cJSON *json = cJSON_Parse(json_data);
  if (!json) {
    fprintf(stderr, "Error: Invalid JSON format\n");
    free(json_data);
    return 1;
  }

  // Estraiamo il comando principale
  cJSON *command = cJSON_GetObjectItemCaseSensitive(json, "command");
  if (cJSON_IsString(command) && (command->valuestring != NULL)) {
    printf("Vitellus: Executing action '%s'...\n", command->valuestring);

    // --- DISPATCHING COMANDO OMOGENEO DEFINITIVO ---

    // Cerchiamo ESATTAMENTE 'action_prepare'
    if (strcmp(command->valuestring, "action_prepare") == 0) {
      return action_prepare(json);
    }
    // Cerchiamo ESATTAMENTE 'action_cleanup'
    else if (strcmp(command->valuestring, "action_cleanup") == 0) {
      return action_cleanup(json);
    }
    // Se usi chroot_run
    else if (strcmp(command->valuestring, "action_run") == 0) {
      return action_run(json);
    }
    // Vecchi comandi obsoleti (PULIZIA TOTALE)
    else if (strcmp(command->valuestring, "prepare") == 0 ||
             strcmp(command->valuestring, "cleanup") == 0 ||
             strcmp(command->valuestring, "cmd_scan") == 0 ||
             strcmp(command->valuestring, "action_run") == 0) {
      fprintf(stderr,
              "{\"error\": \"Command '%s' is obsolete. Use 'action_prepare' or "
              "'action_cleanup'.\"}\n",
              command->valuestring);
      return 1;
    } else {
      fprintf(stderr, "{\"error\": \"Unknown command '%s'\"}\n",
              command->valuestring);
    }
  }

  // Pulizia finale
  cJSON_Delete(json);
  free(json_data);
  return 0;
}
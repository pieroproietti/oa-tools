// src/image_logic.c (Imaging Engine per Vitellus)

#include "image_logic.h"
#include "cJSON.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int action_squash(cJSON *json) {
  cJSON *pathLiveFs = cJSON_GetObjectItemCaseSensitive(json, "pathLiveFs");
  cJSON *comp = cJSON_GetObjectItemCaseSensitive(json, "compression");

  if (!cJSON_IsString(pathLiveFs))
    return 1;

  char liveroot_path[1024], squash_path[1024], live_dir[1024];
  snprintf(liveroot_path, 1024, "%s/liveroot", pathLiveFs->valuestring);
  snprintf(live_dir, 1024, "%s/iso/live", pathLiveFs->valuestring);
  snprintf(squash_path, 1024, "%s/filesystem.squashfs", live_dir);

  // Assicuriamoci che la directory di output esista
  char mkdir_cmd[1100];
  snprintf(mkdir_cmd, 1100, "mkdir -p %s", live_dir);
  system(mkdir_cmd);

  printf(
      "{\"status\": \"imaging\", \"step\": \"squashfs\", \"output\": \"%s\"}\n",
      squash_path);

  char cmd[4096];
  // Parametri ottimizzati: zstd, 1MB block size, esclusioni per non comprimere
  // il 'nido'
  snprintf(
      cmd, 4096,
      "mksquashfs %s %s -comp %s -b 1M -noappend -wildcards -e 'home/eggs/*'",
      liveroot_path, squash_path,
      cJSON_IsString(comp) ? comp->valuestring : "zstd");

  return system(cmd);
}

// Qui domani metteremo action_iso...
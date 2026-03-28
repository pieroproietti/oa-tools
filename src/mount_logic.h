// src/mount_logic.h

#ifndef MOUNT_LOGIC_H
#define MOUNT_LOGIC_H

#include "cJSON.h"

// Nuovi nomi consolidati e politici (Agnostici)
int action_prepare(cJSON *json);
int action_cleanup(cJSON *json);

#endif
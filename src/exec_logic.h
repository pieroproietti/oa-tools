#ifndef EXEC_LOGIC_H
#define EXEC_LOGIC_H

#include "cJSON.h"

// Esegue comandi all'interno della chroot definita nel JSON
int cmd_chroot_run(cJSON *json);

#endif

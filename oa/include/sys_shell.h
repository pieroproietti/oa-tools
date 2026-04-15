#ifndef SYS_SHELL_H
#define SYS_SHELL_H

#include "oa.h"

/**
 * @brief Esegue un comando arbitrario tramite shell (/bin/sh -c)
 * Parametri JSON:
 * - "run_command": string (il comando da eseguire)
 * - "chroot": boolean (se true, esegue dentro liveroot)
 */
int sys_shell(OA_Context *ctx);

#endif

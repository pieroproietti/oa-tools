/*
* oa: eggs in my dialect🥚🥚
* remastering core
*
* Author: Piero Proietti <piero.proietti@gmail.com>
* License: GPL-3.0-or-later
*/
#ifndef HELPERS_H
#define HELPERS_H

#include <stddef.h>

void build_initrd_command(char *dest, const char *template, const char *out, const char *ver);
void append_eggs_exclusion(char *buffer, size_t buf_size, const char *path);

#endif

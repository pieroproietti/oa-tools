/*
* oa: eggs in my dialect🥚🥚
* remastering core
*
* Author: Piero Proietti <piero.proietti@gmail.com>
* License: GPL-3.0-or-later
*/
#include "oa.h"

/**
 * @brief Aggiunge un pattern di esclusione formattato per mksquashfs
 */
void append_eggs_exclusion(char *buffer, size_t buf_size, const char *path) {
    const char *p = (path[0] == '/') ? path + 1 : path;
    strncat(buffer, " '", buf_size - strlen(buffer) - 1);
    strncat(buffer, p, buf_size - strlen(buffer) - 1);
    strncat(buffer, "'", buf_size - strlen(buffer) - 1);
}

/**
 * @brief Rimpiazza {{out}} e {{ver}} nel template del comando initrd
 */
void build_initrd_command(char *dest, const char *template, const char *out, const char *ver) {
    char buffer[4096];
    strncpy(buffer, template, sizeof(buffer));
    
    char *res1 = strstr(buffer, "{{out}}");
    if (res1) {
        char temp[4096];
        size_t len = res1 - buffer;
        strncpy(temp, buffer, len);
        temp[len] = '\0';
        sprintf(temp + len, "%s%s", out, res1 + 7);
        strcpy(buffer, temp);
    }
    
    char *res2 = strstr(buffer, "{{ver}}");
    if (res2) {
        char temp[4096];
        size_t len = res2 - buffer;
        strncpy(temp, buffer, len);
        temp[len] = '\0';
        sprintf(temp + len, "%s%s", ver, res2 + 7);
        strcpy(buffer, temp);
    }
    strcpy(dest, buffer);
}

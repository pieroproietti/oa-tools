/*
 * oa: remastering core
 * src/vendors/logger.c
 */
#include "oa.h"
#include <stdarg.h>
#include <time.h>
#include <stdio.h>
#include <string.h>

static FILE *log_file = NULL;

// La funzione deve essere definita prima di essere utilizzata
void oa_log(const char *level, const char *file, int line, const char *fmt, ...) {
    if (!log_file) return;

    time_t now = time(NULL);
    struct tm *t = localtime(&now);
    char time_str[20];
    strftime(time_str, sizeof(time_str), "%Y-%m-%d %H:%M:%S", t);

    const char *short_file = strrchr(file, '/');
    short_file = short_file ? short_file + 1 : file;

    fprintf(log_file, "[%s] [%-5s] [%s:%d] ", time_str, level, short_file, line);

    va_list args;
    va_start(args, fmt);
    vfprintf(log_file, fmt, args);
    va_end(args);

    fprintf(log_file, "\n");
    fflush(log_file);
}

void oa_init_log(const char *filename) {
    log_file = fopen(filename, "w");
    if (!log_file) {
        perror("\033[1;31m[oa ERROR]\033[0m Impossibile creare il file di log");
    } else {
        oa_log("INFO", __FILE__, __LINE__, "=== Inizio sessione oa ===");
    }
}

void oa_close_log(void) {
    if (log_file) {
        oa_log("INFO", __FILE__, __LINE__, "=== Fine sessione oa ===");
        fclose(log_file);
        log_file = NULL;
    }
}
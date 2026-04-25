#ifndef OA_H
#define OA_H

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <sys/stat.h>
#include <sys/types.h>
#include <sys/mount.h>
#include <errno.h>
#include <stdbool.h>
#include <cJSON.h>

// Definizioni di limiti e contesti che già usi
#define PATH_SAFE 4096
#define CMD_MAX 8192

typedef struct {
    cJSON *root;    // Il JSON intero (configurazione globale) 
    cJSON *task;    // Il comando specifico nel plan (configurazione locale) 
} OA_Context;


// Cambia il prototipo
void oa_log(const char *level, const char *file, int line, const char *fmt, ...);

// Aggiorna le macro per passare automaticamente file e riga
#define LOG_INFO(fmt, ...) oa_log("INFO", __FILE__, __LINE__, fmt, ##__VA_ARGS__)
#define LOG_ERR(fmt, ...)  oa_log("ERROR", __FILE__, __LINE__, fmt, ##__VA_ARGS__)
#define LOG_WARN(fmt, ...) oa_log("WARN", __FILE__, __LINE__, fmt, ##__VA_ARGS__)

// Prototipi delle azioni
int oa_mkdir(OA_Context *ctx);
int oa_mount_generic(OA_Context *ctx);
int oa_bind(OA_Context *ctx);
int oa_shell(OA_Context *ctx);
int oa_umount(OA_Context *ctx);
int oa_users(OA_Context *ctx);

// Utils e Log
void oa_init_log(const char *logfile);
void oa_close_log();
char *read_file(const char *filename);

// Costanti Colore ANSI
#define CLR_CYAN    "\033[1;36m"
#define CLR_RED     "\033[1;31m"
#define CLR_GREEN   "\033[1;32m"
#define CLR_YELLOW  "\033[1;33m"
#define CLR_RESET   "\033[0m"

// Tag standardizzati
#define TAG_OA      CLR_CYAN "[oa]" CLR_RESET
#define TAG_ERR     CLR_RED "[ERRORE]" CLR_RESET

#endif

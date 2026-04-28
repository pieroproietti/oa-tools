#include "oa.h"
#include <stdarg.h>

// Puntatore globale al file di log
FILE *oa_log_file = NULL;

void oa_init_log(const char *logfile) {
    oa_log_file = fopen(logfile, "a");
    if (!oa_log_file) {
        perror("Impossibile aprire il file di log");
    }
}

void oa_close_log() {
    if (oa_log_file) {
        fclose(oa_log_file);
        oa_log_file = NULL;
    }
}

void oa_log(const char *level, const char *file, int line, const char *fmt, ...) {
    va_list args;
    va_start(args, fmt);

    // Prepariamo il messaggio formattato in un buffer
    char buffer[CMD_MAX];
    vsnprintf(buffer, sizeof(buffer), fmt, args);
    va_end(args);

    // Definiamo i colori in base al livello di gravità
    const char *color = CLR_RESET;
    if (strcmp(level, "INFO") == 0) color = CLR_CYAN;
    else if (strcmp(level, "ERROR") == 0) color = CLR_RED;
    else if (strcmp(level, "WARN") == 0) color = CLR_YELLOW;

    // Estraiamo solo il nome del file (rimuove il path completo per pulizia visiva)
    const char *short_file = strrchr(file, '/');
    short_file = short_file ? short_file + 1 : file;

    // 1. STAMPA SU CONSOLE (CON COLORI E CON [ao])
    fprintf(stdout, "%s[oa] [%s] %s:%d%s - %s\n", 
            color, level, short_file, line, CLR_RESET, buffer);

    // 2. STAMPA SU FILE (SENZA COLORI, TESTO PULITO)
    if (oa_log_file) {
        // Nel log scriviamo testo normale per non corrompere la lettura
        fprintf(oa_log_file, "[ao] [%s] %s:%d - %s\n", 
                level, short_file, line, buffer);
        fflush(oa_log_file); // Forza la scrittura immediata su disco
    }
}

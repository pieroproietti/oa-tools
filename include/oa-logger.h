// --- Sistema di Logging ---
void oa_init_log(const char *logfile);
void oa_close_log(void);
void oa_log(const char *level, const char *file, int line, const char *fmt, ...);

#ifndef LOG_INFO
#define LOG_INFO(fmt, ...)  oa_log("INFO",  __FILE__, __LINE__, fmt, ##__VA_ARGS__)
#define LOG_WARN(fmt, ...)  oa_log("WARN",  __FILE__, __LINE__, fmt, ##__VA_ARGS__)
#define LOG_ERR(fmt, ...)   oa_log("ERROR", __FILE__, __LINE__, fmt, ##__VA_ARGS__)
#endif

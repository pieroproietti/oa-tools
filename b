gcc src/main.c \
    src/cJSON.c \
    src/mount_logic.c \
    src/scan_logic.c \
    src/exec_logic.c \
    src/image_logic.c \
    -o vitellus \
    -lm
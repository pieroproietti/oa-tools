# artisan/oa/Makefile

# Fallback se non viene passata dal Makefile principale o da coa
VERSION ?= 0.0.0-dev

CC = gcc
# Aggiunto -DOA_VERSION con il valore della variabile VERSION
CFLAGS = -Wall -Wextra -Iinclude -D_GNU_SOURCE -Wno-format-truncation -DOA_VERSION=\"$(VERSION)\"

# --- MODIFICA QUI: Aggiunto -lcrypt ---
LDFLAGS = -lm -lcrypt

# Struttura delle cartelle
SRC_DIR = src
OBJ_DIR = obj
INC_DIR = include

# Trova tutti i file .c ricorsivamente in src/
SOURCES = $(shell find $(SRC_DIR) -name '*.c')

# Mappa i sorgenti agli oggetti
OBJECTS = $(SOURCES:$(SRC_DIR)/%.c=$(OBJ_DIR)/%.o)

# Nome del binario finale
TARGET = oa

# Target principale
all: $(TARGET)

# Linker
$(TARGET): $(OBJECTS)
	@echo "  LD    $@"
	@$(CC) $(OBJECTS) -o $@ $(LDFLAGS)

# Compilatore
$(OBJ_DIR)/%.o: $(SRC_DIR)/%.c
	@mkdir -p $(dir $@)
	@echo "  CC    $<"
	@$(CC) $(CFLAGS) -c $< -o $@

# Pulizia
clean:
	@echo "  Cleaning oa..."
	@rm -rf $(OBJ_DIR) $(TARGET)
	@rm -f /tmp/remaster.json /tmp/sysinstall.json

.PHONY: all clean

# artisan/Makefile

# 1. THE SINGLE SOURCE OF TRUTH
# Legge la versione dinamicamente dal file di testo dell'orchestratore Go.
# Se il file è assente, usa la versione di sviluppo in modo sicuro.
VERSION := $(shell cat coa/src/VERSION 2>/dev/null || echo "0.0.0-dev")

# Directories
OA_DIR = oa
COA_DIR = coa

# Binaries
OA_BIN = $(OA_DIR)/oa
COA_BIN = $(COA_DIR)/coa

all: build_oa build_coa
	@echo "--------------------------------------"
	@echo "Hatching completed successfully! 🐣"
	@echo "Version:          $(VERSION)"
	@echo "coa Brain (Go):   ./$(COA_BIN)"
	@echo "oa Workhorse (C): ./$(OA_BIN)"
	@echo "--------------------------------------"

build_oa:
	@echo "  MAKING oa..."
	@$(MAKE) -C $(OA_DIR) VERSION="$(VERSION)"

build_coa:
	@echo "  MAKING coa..."
	@cd $(COA_DIR) && go build -o coa ./src

clean:
	@echo "  Pulizia in corso..."
	@$(MAKE) -C $(OA_DIR) clean
	@rm -f $(COA_BIN)
	@rm -f $(COA_DIR)/plan_coa_tmp.json

.PHONY: all build_oa build_coa clean
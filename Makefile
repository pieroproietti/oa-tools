# 1. THE SINGLE SOURCE OF TRUTH
VERSION := $(shell git describe --tags --always 2>/dev/null || echo "0.0.0-dev")

# Directories
OA_DIR = oa
COA_DIR = coa

# Binaries
OA_BIN = $(OA_DIR)/oa
COA_BIN = $(COA_DIR)/coa

# Patterns per i pacchetti nativi
PACKAGES = *.deb *.rpm *.pkg.tar.zst PKGBUILD

# Target principale
all: build_oa build_coa
	@echo "--------------------------------------"
	@echo "Hatching completed successfully! 🐣"
	@echo "Version:           $(VERSION)"
	@echo "coa Brain (Go):    ./$(COA_BIN)"
	@echo "oa Workhorse (C):  ./$(OA_BIN)"
	@echo "--------------------------------------"

build_oa:
	@echo "  MAKING oa..."
	# Passiamo LIBS="-lcrypt" al Makefile interno di oa
	@$(MAKE) -C $(OA_DIR) VERSION="$(VERSION)" LIBS="-lcrypt"

build_coa:
	@echo "  MAKING coa (Version: $(VERSION))..."
	@cd $(COA_DIR) && go build -ldflags "-X 'coa/pkg/cmd.AppVersion=$(VERSION)'" -o coa main.go
	@echo "  GENERATING DOCUMENTATION..."
	# Il '-' iniziale impedisce a make di fallire se il comando _gen_docs non esiste ancora
	@-./$(COA_BIN) _gen_docs --target ./$(COA_DIR)/docs 2>/dev/null || true

clean:
	@echo "  Pulizia binari e piani di volo..."
	@$(MAKE) -C $(OA_DIR) clean || true
	@rm -f $(COA_BIN)
	@rm -f /tmp/oa-remaster.json /tmp/sysinstall.json /tmp/coa/finalize-plan.json
	@echo "  Rimozione pacchetti nativi ($(PACKAGES))..."
	@rm -f $(PACKAGES)
	@echo "  Pulizia documentazione e completamenti..."
	@rm -rf $(COA_DIR)/docs/man/*
	@rm -rf $(COA_DIR)/docs/completion/*
	@rm -rf $(COA_DIR)/docs/md/*

.PHONY: all build_oa build_coa clean
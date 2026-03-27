APP_NAME := Pruebas-Ollama
SSH_USER := root
SSH_HOST := 138.197.101.64
SSH_PORT := 22
REMOTE_DIR := /opt/$(APP_NAME)

SSH := ssh -p $(SSH_PORT)
RSYNC := rsync -az --delete

RSYNC_EXCLUDES := \
	--exclude '.git' \
	--exclude '.github' \
	--exclude '.vscode' \
	--exclude '.idea' \
	--exclude 'bin' \
	--exclude 'dist' \
	--exclude '*.zip' \
	--exclude '*.tar.gz' \
	--exclude '.DS_Store'

.PHONY: help provision sync remote-prepare remote-deps remote-build remote-clean

help:
	@echo "Targets disponibles:"
	@echo "  make provision      -> envia archivos al servidor, descarga dependencias y compila"
	@echo "  make sync           -> envia archivos al servidor"
	@echo "  make remote-prepare -> crea el directorio remoto"
	@echo "  make remote-deps    -> ejecuta go mod tidy y go mod download en remoto"
	@echo "  make remote-build   -> compila en remoto con go build ./..."
	@echo "  make remote-clean   -> borra el directorio remoto"

provision: remote-prepare sync remote-deps remote-build

remote-prepare:
	$(SSH) $(SSH_USER)@$(SSH_HOST) "mkdir -p $(REMOTE_DIR)"

sync:
	$(RSYNC) $(RSYNC_EXCLUDES) ./ $(SSH_USER)@$(SSH_HOST):$(REMOTE_DIR)/

remote-deps:
	$(SSH) $(SSH_USER)@$(SSH_HOST) '\
		set -e; \
		cd $(REMOTE_DIR); \
		if ! command -v go >/dev/null 2>&1; then \
			echo "ERROR: Go no esta instalado en el servidor remoto."; \
			exit 1; \
		fi; \
		go mod tidy; \
		go mod download \
	'

remote-build:
	$(SSH) $(SSH_USER)@$(SSH_HOST) '\
		set -e; \
		cd $(REMOTE_DIR); \
		go build ./... \
	'

remote-clean:
	$(SSH) $(SSH_USER)@$(SSH_HOST) "rm -rf $(REMOTE_DIR)"
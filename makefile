SSH_HOST := 138.197.101.64

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
	@echo "  make provision      -> envia archivos, instala deps y compila binarios"
	@echo "  make sync           -> envia archivos al servidor"
	@echo "  make remote-build   -> compila binarios en remoto"
	@echo "  make remote-clean   -> borra el directorio remoto"

provision: remote-prepare sync remote-deps remote-build

remote-prepare:
	ssh root@$(SSH_HOST) "mkdir -p /opt/pruebas-ollama /opt/pruebas-ollama/bin"

sync:
	$(RSYNC) $(RSYNC_EXCLUDES) ./ root@$(SSH_HOST):/opt/pruebas-ollama/

remote-deps:
	ssh root@$(SSH_HOST) '\
		set -e; \
		cd /opt/pruebas-ollama; \
		if ! command -v go >/dev/null 2>&1; then \
			echo "ERROR: Go no esta instalado en el servidor remoto."; \
			exit 1; \
		fi; \
		go mod tidy; \
		go mod download \
	'

remote-build:
	ssh root@$(SSH_HOST) '\
		set -e; \
		cd /opt/pruebas-ollama; \
		mkdir -p /opt/pruebas-ollama/bin; \
		for dir in cmd/*; do \
			if [ -d "$$dir" ]; then \
				name=$$(basename $$dir); \
				echo "Compilando $$name..."; \
				go build -o /opt/pruebas-ollama/bin/$$name ./cmd/$$name; \
			fi; \
		done \
	'

remote-clean:
	ssh root@$(SSH_HOST) "rm -rf /opt/pruebas-ollama"
SSH_HOST := 138.197.101.64

.PHONY: help provision remote-prepare github-sync hf-sync local-deps local-build upload-bin remote-verify remote-tree remote-clean clean-bin

help:
	@echo "Targets disponibles:"
	@echo "  make provision    -> prepara remoto, sincroniza GitHub y Hugging Face, compila local y sube binarios"
	@echo "  make github-sync  -> clona o actualiza el repo de GitHub en el servidor"
	@echo "  make hf-sync      -> clona o actualiza el repo de Hugging Face en el servidor"
	@echo "  make local-build  -> compila binarios Linux amd64 en ./bin"
	@echo "  make upload-bin   -> sube ./bin al servidor"
	@echo "  make remote-tree  -> muestra el tree remoto"
	@echo "  make remote-clean -> borra /opt/pruebas-ollama del servidor"
	@echo "  make clean-bin    -> borra ./bin local"

provision: remote-prepare github-sync hf-sync local-deps local-build upload-bin remote-verify

remote-prepare:
	ssh root@$(SSH_HOST) 'bash -lc '\''
		set -Eeuo pipefail
		mkdir -p /opt
		mkdir -p /opt/pruebas-ollama
		mkdir -p /opt/pruebas-ollama/bin
		mkdir -p /opt/pruebas-ollama/vendor
		mkdir -p /opt/pruebas-ollama/vendor/huggingface
		if ! command -v git >/dev/null 2>&1; then
			echo "ERROR: git no está instalado en el servidor remoto."
			exit 1
		fi
	'\'''

github-sync:
	ssh root@$(SSH_HOST) 'bash -lc '\''
		set -Eeuo pipefail
		if [ ! -d /opt/pruebas-ollama/.git ]; then
			rm -rf /opt/pruebas-ollama
			git clone https://github.com/galvarez0/Pruebas-Ollama.git /opt/pruebas-ollama
		else
			cd /opt/pruebas-ollama
			git fetch --all --prune
			git checkout main
			git pull --ff-only origin main
		fi
		mkdir -p /opt/pruebas-ollama/bin
		mkdir -p /opt/pruebas-ollama/vendor/huggingface
	'\'''

hf-sync:
	ssh root@$(SSH_HOST) 'bash -lc '\''
		set -Eeuo pipefail
		if [ ! -d /opt/pruebas-ollama/vendor/huggingface/Qwen2.5-0.5B-Instruct/.git ]; then
			git clone https://huggingface.co/Qwen/Qwen2.5-0.5B-Instruct /opt/pruebas-ollama/vendor/huggingface/Qwen2.5-0.5B-Instruct
		else
			cd /opt/pruebas-ollama/vendor/huggingface/Qwen2.5-0.5B-Instruct
			git fetch --all --prune
			git pull --ff-only
		fi

		if command -v git-xet >/dev/null 2>&1; then
			echo "git-xet detectado en remoto."
		else
			echo "WARNING: git-xet no está instalado en el servidor."
			echo "WARNING: si este repo de Hugging Face usa archivos grandes administrados por Xet, instala git-xet."
		fi

		if command -v git-lfs >/dev/null 2>&1; then
			cd /opt/pruebas-ollama/vendor/huggingface/Qwen2.5-0.5B-Instruct
			git lfs pull || true
		fi
	'\'''

local-deps:
	go mod tidy
	go mod download

local-build:
	bash -lc '\
		set -Eeuo pipefail; \
		rm -rf ./bin; \
		mkdir -p ./bin; \
		for dir in ./cmd/*; do \
			if [ -d "$$dir" ]; then \
				name=$$(basename "$$dir"); \
				echo "Compilando $$name..."; \
				CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o ./bin/$$name ./cmd/$$name; \
			fi; \
		done \
	'

upload-bin:
	rsync -az --delete ./bin/ root@$(SSH_HOST):/opt/pruebas-ollama/bin/

remote-verify:
	ssh root@$(SSH_HOST) 'bash -lc '\''
		set -Eeuo pipefail
		test -d /opt/pruebas-ollama/cmd
		test -d /opt/pruebas-ollama/internal
		test -f /opt/pruebas-ollama/go.mod
		test -d /opt/pruebas-ollama/bin
		echo "Contenido remoto:"
		find /opt/pruebas-ollama -maxdepth 3 | sort
	'\'''

remote-tree:
	ssh root@$(SSH_HOST) 'bash -lc '\''
		set -Eeuo pipefail
		find /opt/pruebas-ollama -maxdepth 4 | sort
	'\'''

remote-clean:
	ssh root@$(SSH_HOST) 'bash -lc '\''
		set -Eeuo pipefail
		rm -rf /opt/pruebas-ollama
	'\'''

clean-bin:
	rm -rf ./bin
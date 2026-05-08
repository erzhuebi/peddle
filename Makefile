PEDDLEC_BIN := build/peddlec
ASM        := 64tass
VICE       ?= $(shell command -v x64sc 2>/dev/null || command -v x64 2>/dev/null)

EXAMPLE ?= hello
OPT     ?= speed
MEM_FLAGS ?=

SRC     := examples/$(EXAMPLE).ped
ASM_OUT := build/$(EXAMPLE).asm
PRG_OUT := build/$(EXAMPLE).prg

.PHONY: help check check-run build run example hello clean test examples

# default target
all: help

test: check
	go test ./...

help:
	@echo "Available targets:"
	@echo ""
	@echo "  make build                         - build peddlec compiler"
	@echo "  make run EXAMPLE=hello             - compile examples/hello.ped, assemble PRG, run in VICE"
	@echo "  make run EXAMPLE=hello OPT=size    - same, but compile with --opt=size"
	@echo "  make example EXAMPLE=x             - compile examples/x.ped and assemble PRG without running"
	@echo "  make hello                         - same as make run EXAMPLE=hello"
	@echo "  make examples                      - list available examples"
	@echo "  make check                         - check compiler toolchain"
	@echo "  make clean                         - remove build artifacts"
	@echo ""
	@echo "Optional variables:"
	@echo "  OPT=speed|size"
	@echo "  MEM_FLAGS='--mem-report --mem-limit=32768'"
	@echo ""
	@echo "Toolchain:"
	@echo "  macOS: brew install go 64tass vice"
	@echo "  Linux: sudo apt install golang 64tass vice"

check:
	@command -v go >/dev/null 2>&1 || { \
		echo "missing: go"; \
		echo "macOS: install with: brew install go"; \
		echo "Linux: install with: sudo apt install golang"; \
		exit 1; \
	}
	@echo "compiler toolchain ok"

check-run: check
	@command -v $(ASM) >/dev/null 2>&1 || { \
		echo "missing: $(ASM)"; \
		echo "macOS: install with: brew install 64tass"; \
		echo "Linux: install with: sudo apt install 64tass"; \
		exit 1; \
	}
	@if [ -z "$(VICE)" ]; then \
		echo "missing: x64sc or x64"; \
		echo "macOS: install with: brew install vice"; \
		echo "Linux: install with: sudo apt install vice"; \
		exit 1; \
	fi
	@echo "demo toolchain ok: $(VICE)"

build: check
	@mkdir -p build
	go build -o $(PEDDLEC_BIN) ./cmd/peddlec
	@echo "wrote $(PEDDLEC_BIN)"

examples:
	@find examples -name '*.ped' -maxdepth 1 -type f | sed 's|examples/||; s|\.ped$$||' | sort

example: check-run build
	@if [ ! -f "$(SRC)" ]; then \
		echo "missing example: $(SRC)"; \
		echo ""; \
		echo "Available examples:"; \
		$(MAKE) --no-print-directory examples; \
		exit 1; \
	fi
	$(PEDDLEC_BIN) --opt=$(OPT) $(MEM_FLAGS) -o $(ASM_OUT) $(SRC)
	$(ASM) $(ASM_OUT) -o $(PRG_OUT)
	@echo "wrote $(PRG_OUT)"

run: example
	$(VICE) -autostart $(PRG_OUT)

hello: run

clean:
	rm -rf build
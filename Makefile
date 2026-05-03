PEDDLEC_BIN := build/peddlec
ASM        := 64tass
VICE       ?= $(shell command -v x64sc 2>/dev/null || command -v x64 2>/dev/null)

HELLO_SRC := examples/hello.ped
HELLO_ASM := build/hello.asm
HELLO_PRG := build/hello.prg

.PHONY: help check check-run build hello clean test

test: check
	go test ./...

help:
	@echo "Available targets:"
	@echo ""
	@echo "  make build   - build peddlec compiler"
	@echo "  make hello   - build compiler, compile hello example, assemble PRG, run in VICE"
	@echo "  make check   - check compiler toolchain"
	@echo "  make clean   - remove build artifacts"
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

hello: check-run build
	$(PEDDLEC_BIN) -o $(HELLO_ASM) $(HELLO_SRC)
	$(ASM) $(HELLO_ASM) -o $(HELLO_PRG)
	$(VICE) -autostart $(HELLO_PRG)

clean:
	rm -rf build
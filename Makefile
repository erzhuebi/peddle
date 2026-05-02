PEDDLEC := go run ./cmd/peddlec
ASM     := 64tass
VICE    := x64

DEMO_SRC := examples/hello.ped
ASM_OUT  := build/out.asm
PRG_OUT  := build/out.prg

.PHONY: all check check-run build asm prg run demo clean

all: prg

check:
	@command -v go >/dev/null 2>&1 || { echo "missing: go"; exit 1; }
	@command -v $(ASM) >/dev/null 2>&1 || { echo "missing: $(ASM)"; exit 1; }
	@echo "toolchain ok"

check-run: check
	@command -v $(VICE) >/dev/null 2>&1 || { echo "missing: $(VICE)"; exit 1; }
	@echo "emulator ok"

build:
	@mkdir -p build

asm: check build
	$(PEDDLEC) -o $(ASM_OUT) $(DEMO_SRC)

prg: asm
	$(ASM) $(ASM_OUT) -o $(PRG_OUT)
	@echo "wrote $(PRG_OUT)"

run: check-run prg
	$(VICE) $(PRG_OUT)

demo: run

clean:
	rm -rf build
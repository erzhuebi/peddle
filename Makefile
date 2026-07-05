PEDDLEC_BIN := build/peddlec
ASM        := 64tass
VICE       ?= $(shell command -v x64sc 2>/dev/null || command -v x64 2>/dev/null)

GO ?= go

VERSION_FILE := .version
RELEASE_NOTES_FILE ?= release-notes.txt
DEVTOOL_CMD ?= ./cmd/devtool

BASE_VERSION := $(shell [ -f $(VERSION_FILE) ] && cat $(VERSION_FILE) || echo "0.0.0")
VERSION := $(BASE_VERSION)-dev

DEFAULT_EXAMPLE := smoke/hello
PRIMARY_GOAL := $(firstword $(MAKECMDGOALS))
EXAMPLE_GOAL := $(word 2,$(MAKECMDGOALS))
KNOWN_TARGETS := help all check check-run build run example clean test examples version release-notes bump-patch bump-minor bump-major _bump_version _write_release_notes
EXAMPLE_ARG :=

ifneq ($(filter run example,$(PRIMARY_GOAL)),)
ifneq ($(EXAMPLE_GOAL),)
ifeq ($(filter $(EXAMPLE_GOAL),$(KNOWN_TARGETS)),)
EXAMPLE_ARG := $(EXAMPLE_GOAL)
endif
endif
endif

ifeq ($(origin EXAMPLE), undefined)
ifneq ($(EXAMPLE_ARG),)
EXAMPLE := $(EXAMPLE_ARG)
else
EXAMPLE := $(DEFAULT_EXAMPLE)
endif
endif

OPT     ?= speed
MEM_FLAGS ?=

SRC     := examples/$(EXAMPLE).ped
ASM_OUT := build/$(EXAMPLE).asm
PRG_OUT := build/$(EXAMPLE).prg

.PHONY: help all check check-run build run example clean test examples version release-notes bump-patch bump-minor bump-major _bump_version _write_release_notes

ifneq ($(EXAMPLE_ARG),)
$(EXAMPLE_ARG):
	@:
endif

# default target
help:
	@echo "Available targets:"
	@echo ""
	@echo "  make all                           - run tests and build peddlec compiler"
	@echo "  make build                         - build peddlec compiler"
	@echo "  make run smoke/hello               - compile examples/smoke/hello.ped, assemble PRG, run in VICE"
	@echo "  make run smoke/hello OPT=size      - same, but compile with --opt=size"
	@echo "  make example demos/pong            - compile examples/demos/pong.ped and assemble PRG without running"
	@echo "  make examples                      - list available examples"
	@echo "  make check                         - check compiler toolchain"
	@echo "  make test                          - run Go tests"
	@echo "  make version                       - print current development version"
	@echo "  make release-notes                 - update release-notes.txt from git history"
	@echo "  make bump-patch                    - tag current version and bump patch"
	@echo "  make bump-minor                    - tag current version and bump minor"
	@echo "  make bump-major                    - tag current version and bump major"
	@echo "  make clean                         - remove build artifacts"
	@echo ""
	@echo "Optional variables:"
	@echo "  OPT=speed|size"
	@echo "  EXAMPLE=smoke/hello                - explicit form, useful for scripts"
	@echo "  MEM_FLAGS='--mem-report --mem-limit=32768'"
	@echo ""
	@echo "Version:"
	@echo "  base: $(BASE_VERSION)"
	@echo "  dev : $(VERSION)"
	@echo ""
	@echo "Version workflow:"
	@echo "  .version stores only the base version, for example 0.17.0"
	@echo "  normal builds append -dev, so work-in-progress builds are 0.17.0-dev"
	@echo "  make bump-patch/minor/major tags the current state as v0.17.0"
	@echo "  then it bumps .version to the next base version for the next -dev cycle"
	@echo "  no -dev branch is created; create/switch sprint branches manually if needed"
	@echo ""
	@echo "Toolchain:"
	@echo "  macOS: brew install go 64tass vice"
	@echo "  Linux: sudo apt install golang 64tass vice"

all: test build

test: check
	$(GO) test ./...

check:
	@command -v $(GO) >/dev/null 2>&1 || { \
		echo "missing: $(GO)"; \
		echo "macOS: install with: brew install go"; \
		echo "Linux: install with: sudo apt install golang"; \
		exit 1; \
	}
	@if [ ! -f "$(VERSION_FILE)" ]; then \
		echo "missing: $(VERSION_FILE)"; \
		echo "create it with an initial version, for example:"; \
		echo "  echo 0.1.0 > $(VERSION_FILE)"; \
		exit 1; \
	fi
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
	$(GO) build -ldflags "-X main.Version=$(VERSION)" -o $(PEDDLEC_BIN) ./cmd/peddlec
	@echo "wrote $(PEDDLEC_BIN)"
	@echo "version $(VERSION)"

version: check
	@echo "$(VERSION)"

release-notes: check
	@$(MAKE) --no-print-directory _write_release_notes RELEASE_VERSION="$(BASE_VERSION)"
	@echo "updated $(RELEASE_NOTES_FILE)"

bump-patch:
	@$(MAKE) --no-print-directory _bump_version BUMP=patch

bump-minor:
	@$(MAKE) --no-print-directory _bump_version BUMP=minor

bump-major:
	@$(MAKE) --no-print-directory _bump_version BUMP=major

_bump_version: test
	@set -e; \
	if [ -z "$(BUMP)" ]; then \
		echo "missing BUMP=patch|minor|major"; \
		exit 1; \
	fi; \
	if [ ! -d .git ]; then \
		echo "not a git repository"; \
		exit 1; \
	fi; \
	if [ -n "$$(git status --porcelain)" ]; then \
		echo "Git working tree is dirty. Commit or stash changes first."; \
		echo ""; \
		git status --short; \
		exit 1; \
	fi; \
	base=$$(cat "$(VERSION_FILE)"); \
	case "$$base" in \
		*.*.*) ;; \
		*) echo "invalid version in $(VERSION_FILE): $$base"; exit 1 ;; \
	esac; \
	echo "Releasing v$$base"; \
	$(MAKE) --no-print-directory _write_release_notes RELEASE_VERSION="$$base"; \
	git add "$(RELEASE_NOTES_FILE)" "$(VERSION_FILE)"; \
	git commit -m "release: v$$base"; \
	git tag -a "v$$base" -m "v$$base"; \
	new=$$($(GO) run "$(DEVTOOL_CMD)" -quiet "$(VERSION_FILE)" "$(BUMP)"); \
	echo "Next development version: $$new-dev"; \
	git add "$(VERSION_FILE)"; \
	git commit -m "chore: start $$new-dev"; \
	echo ""; \
	echo "Released v$$base"; \
	echo "Current development version: $$new-dev"; \
	echo ""; \
	echo "Push with:"; \
	echo "  git push"; \
	echo "  git push --tags"

_write_release_notes:
	@set -e; \
	if [ -z "$(RELEASE_VERSION)" ]; then \
		echo "missing RELEASE_VERSION"; \
		exit 1; \
	fi; \
	if [ ! -d .git ]; then \
		echo "not a git repository"; \
		exit 1; \
	fi; \
	last_tag=$$(git describe --tags --abbrev=0 2>/dev/null || true); \
	tmp_notes=$$(mktemp); \
	{ \
		echo "# Release notes"; \
		echo ""; \
		echo "## v$(RELEASE_VERSION) - $$(date +%Y-%m-%d)"; \
		echo ""; \
		if [ -n "$$last_tag" ]; then \
			commits=$$(git log --pretty=format:'- %s' "$$last_tag"..HEAD); \
		else \
			commits=$$(git log --pretty=format:'- %s'); \
		fi; \
		if [ -n "$$commits" ]; then \
			echo "$$commits"; \
		else \
			echo "- No changes recorded."; \
		fi; \
		echo ""; \
		if [ -f "$(RELEASE_NOTES_FILE)" ]; then \
			awk 'NR==1 && $$0 == "# Release notes" {drop_blank=1; next} drop_blank && $$0 == "" {drop_blank=0; next} {drop_blank=0; print}' "$(RELEASE_NOTES_FILE)"; \
		fi; \
	} > "$$tmp_notes"; \
	mv "$$tmp_notes" "$(RELEASE_NOTES_FILE)"

examples:
	@find examples -type f -name '*.ped' ! -path '*/lib/*' | sed 's|examples/||; s|\.ped$$||' | sort

example: check-run build
	@if [ ! -f "$(SRC)" ]; then \
		echo "missing example: $(SRC)"; \
		echo ""; \
		echo "Available examples:"; \
		$(MAKE) --no-print-directory examples; \
		exit 1; \
	fi
	@mkdir -p "$(dir $(ASM_OUT))" "$(dir $(PRG_OUT))"
	$(PEDDLEC_BIN) --opt=$(OPT) $(MEM_FLAGS) -o $(ASM_OUT) $(SRC)
	$(ASM) $(ASM_OUT) -o $(PRG_OUT)
	@echo "wrote $(PRG_OUT)"

run: example
	$(VICE) -autostart $(PRG_OUT)

clean:
	rm -rf build

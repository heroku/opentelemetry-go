TOP_LEVEL=$(shell git rev-parse --show-toplevel)
ALL_PKGS := $(shell go list ./...)
GO_GENERATE_FILES=$(shell find . -name "*.go" -exec grep -lw {} -e '//go:generate' \;)

EXAMPLES := \
	./example/basic \
	./example/http/client \
	./example/http/server \
	./experimental/streaming/example/basic

# All source code and documents. Used in spell check.
ALL_DOCS := $(shell find . -name '*.md' -type f | sort)
# All directories with go.mod files. Used in go mod tidy.
ALL_GO_MOD_DIRS := $(shell find . -type f -name 'go.mod' -exec dirname {} \; | sort)

GOTEST=go test
GOTEST_OPT?=-v -race -timeout 30s
GOTEST_OPT_WITH_COVERAGE = $(GOTEST_OPT) -coverprofile=coverage.txt -covermode=atomic

.DEFAULT_GOAL := precommit

TOOLS_DIR=$(TOP_LEVEL)/.tools

$(TOOLS_DIR):
go.mod go.sum tools.go:
$(GO_GENERATE_FILES):

$(TOOLS_DIR)/golangci-lint: go.mod go.sum tools.go
	go build -o $(TOOLS_DIR)/golangci-lint github.com/golangci/golangci-lint/cmd/golangci-lint

$(TOOLS_DIR)/misspell: go.mod go.sum tools.go
	go build -o $(TOOLS_DIR)/misspell github.com/client9/misspell/cmd/misspell

$(TOOLS_DIR)/stringer: go.mod go.sum tools.go
	go build -o $(TOOLS_DIR)/stringer golang.org/x/tools/cmd/stringer

.PHONY: lint
# TODO: Fix this on windows.
lint: $(TOOLS_DIR)/golangci-lint
	@for dir in $(ALL_GO_MOD_DIRS); do \
		echo "'golangci-lint run --fix' in $${dir}"; \
		(cd $${dir} && $(TOOLS_DIR)/golangci-lint run --fix); \
	done

.PHONY: generate
generate: $(GO_GENERATE_FILES) $(TOOLS_DIR)/stringer 
	PATH="$(abspath $(TOOLS_DIR)):$${PATH}" go generate ./...

.PHONY: misspell
misspell: $(TOOLS_DIR)/misspell
	$(TOOLS_DIR)/misspell -w $(ALL_DOCS)

.PHONY: precommit
precommit: generate misspell lint verify tidy 

.PHONY: test-with-coverage
test-with-coverage:
	$(GOTEST) $(GOTEST_OPT_WITH_COVERAGE) $(ALL_PKGS)
	go tool cover -html=coverage.txt -o coverage.html

.PHONY: circle-ci
circle-ci: precommit test-clean-work-tree test-with-coverage test-386 examples

.PHONY: test-clean-work-tree
test-clean-work-tree:
	@if ! git diff --quiet; then \
	  echo; \
	  echo "Working tree is not clean"; \
	  echo; \
	  git status; \
	  exit 1; \
	fi

.PHONY: test
test: examples
	$(GOTEST) $(GOTEST_OPT) $(ALL_PKGS)

.PHONY: test-386
test-386:
	GOARCH=386 $(GOTEST) -v -timeout 30s $(ALL_PKGS)

.PHONY: examples
examples:
	@for ex in $(EXAMPLES); do \
	  echo "Building $${ex}"; \
	  (cd "$${ex}" && go build .); \
	done

.PHONY: all-pkgs
all-pkgs:
	@echo $(ALL_PKGS) | tr ' ' '\n' | sort

.PHONY: all-docs
all-docs:
	@echo $(ALL_DOCS) | tr ' ' '\n' | sort

.PHONY: tidy
tidy:
	@for dir in $(ALL_GO_MOD_DIRS); do \
	  echo "'go mod tidy' in $${dir}"; \
	  (cd "$${dir}" && go mod tidy); \
	done

.PHONY: verify
verify:
	@for dir in $(ALL_GO_MOD_DIRS); do \
	  echo "'go mod verify' in $${dir}"; \
	  (cd "$${dir}" && go mod verify); \
	done

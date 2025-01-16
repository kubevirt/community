.PHONY: generate validate-sigs test coverage lint

export GOLANGCI_LINT_VERSION := v1.62.2
ifndef GOPATH
    GOPATH=$(shell go env GOPATH)
    export GOPATH
endif
WORKDIR := /tmp
LOCAL_BIN := $(WORKDIR)/local_bin
PATH := $(LOCAL_BIN):${PATH}
ifndef ARTIFACTS
	export ARTIFACTS := $(WORKDIR)/artifacts
endif
ifndef COVERAGE_OUTPUT_PATH
	COVERAGE_OUTPUT_PATH := ${ARTIFACTS}/coverage.html
endif

work-dirs:
	mkdir -p ${ARTIFACTS}
	mkdir -p ${LOCAL_BIN}

generate:
	go run ./validators/cmd/sigs --dry-run=false
	go run ./generators/cmd/sigs
	go run ./generators/cmd/alumni

validate-sigs:
	go run ./validators/cmd/sigs

test:
	go test ./...

coverage: work-dirs
	if ! command -V covreport; then GOBIN=$(LOCAL_BIN) go install github.com/cancue/covreport@latest; fi
	go test ./... -coverprofile=/tmp/coverage.out
	covreport -i /tmp/coverage.out -o $(COVERAGE_OUTPUT_PATH)

lint: work-dirs
	if ! command -V golangci-lint; then curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "${LOCAL_BIN}" ${GOLANGCI_LINT_VERSION} ; fi
	golangci-lint run --verbose

.PHONY: make-artifacts-dir generate validate-sigs test coverage install-metrics-binaries lint

export GOLANGCI_LINT_VERSION := v1.62.2
ifndef GOPATH
    GOPATH=$(shell go env GOPATH)
    export GOPATH
endif
ifndef ARTIFACTS
	ARTIFACTS=/tmp/artifacts
    export ARTIFACTS
endif
ifndef COVERAGE_OUTPUT_PATH
	COVERAGE_OUTPUT_PATH=${ARTIFACTS}/coverage.html
    export COVERAGE_OUTPUT_PATH
endif

make-artifacts-dir:
	mkdir -p ${ARTIFACTS}

generate:
	go run ./validators/cmd/sigs --dry-run=false
	go run ./generators/cmd/sigs
	go run ./generators/cmd/alumni

validate-sigs:
	go run ./validators/cmd/sigs

test:
	go test ./...

coverage: make-artifacts-dir
	if ! command -V covreport; then go install github.com/cancue/covreport@latest; fi
	go test ./... -coverprofile=/tmp/coverage.out
	covreport  -i /tmp/coverage.out -o ${COVERAGE_OUTPUT_PATH}

install-metrics-binaries:
	if ! command -V golangci-lint; then curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ${GOPATH}/bin ${GOLANGCI_LINT_VERSION} ; fi

lint: install-metrics-binaries
	golangci-lint run --verbose

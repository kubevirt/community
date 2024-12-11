.PHONY: generate validate-sigs install-metrics-binaries lint

export GOLANGCI_LINT_VERSION := v1.62.2
ifndef $(GOPATH)
    GOPATH=$(shell go env GOPATH)
    export GOPATH
endif

generate:
	go run ./validators/cmd/sigs --dry-run=false
	go run ./generators/cmd/sigs
	go run ./generators/cmd/alumni

validate-sigs:
	go run ./validators/cmd/sigs

install-metrics-binaries:
	if ! command -V golangci-lint; then curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ${GOPATH}/bin ${GOLANGCI_LINT_VERSION} ; fi

lint: install-metrics-binaries
	golangci-lint run --verbose

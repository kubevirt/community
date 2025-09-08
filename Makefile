.PHONY: generate validate-sigs test coverage lint

ifndef GOPATH
    GOPATH=$(shell go env GOPATH)
    export GOPATH
endif
WORKDIR := /tmp
ifndef LOCAL_BIN
    LOCAL_BIN=$(WORKDIR)/local_bin
    export LOCAL_BIN
endif
PATH := $(LOCAL_BIN):${PATH}
ifndef ARTIFACTS
	export ARTIFACTS := $(WORKDIR)/artifacts
endif
ifndef COVERAGE_OUTPUT_PATH
	COVERAGE_OUTPUT_PATH=${ARTIFACTS}/coverage.html
	export COVERAGE_OUTPUT_PATH
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
	./hack/coverage.sh

lint: work-dirs
	./hack/lint.sh

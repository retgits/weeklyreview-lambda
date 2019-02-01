# Description: Makefile
# Author: retgits
# Last Updated: 2019-01-31

#--- Variables ---
## The name of the user for Docker
DOCKERUSER=retgits
## Get the name of the project
PROJECT=weekly-review
## Set a default test directory
TESTDIR=$(CURDIR)/test
## Create a list of all packages in this repository
PACKAGES=$(shell go list ./... | grep -v "vendor")

#--- Help ---
.PHONY: help
help: ## Displays the help for each target (this message)
	@echo 
	@echo Makefile targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' Makefile | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
	@echo 

#--- Linting targets ---
fmt: ## Fmt runs the commands 'gofmt -l -w' and 'gofmt -s -w' and prints the names of the files that are modified.
	env GO111MODULE=on go fmt ./...
	env GO111MODULE=on gofmt -s -w .

vet: ## Vet examines Go source code and reports suspicious constructs.
	env GO111MODULE=on go vet ./...

lint: ## Lint examines Go source code and prints style mistakes for all packages.
	env GO111MODULE=on golint -set_exit_status $(ALL_PACKAGES)

#--- Setup targets ---
setup: ## Make preparations to be able to run tests.
	mkdir -p ${TESTDIR}
	mkdir -p $(GOPATH)/bin
	go get -u golang.org/x/lint/golint
	go get -u github.com/gojp/goreportcard/cmd/goreportcard-cli

deps: ## Get all the Go dependencies.
	go get -u ./...

#--- Test targets ---
.PHONY: test
test: ## Run all testcases.
	env TESTDIR=${TESTDIR} go test -race ./...

test-cover-html: ## Run all test cases and generate a coverage report.
	@echo "mode: count" > coverage-all.out

	$(foreach pkg, $(PACKAGES),\
	env TESTDIR=${TESTDIR} go test -coverprofile=coverage.out -covermode=count $(pkg);\
	tail -n +2 coverage.out >> coverage-all.out;)
	go tool cover -html=coverage-all.out -o out/coverage.html

score: ## Get a score based on GoReportcard.
	goreportcard-cli -v

#--- Build targets ---
compile-lin: ## Compiles and creates a Linux executable in the 'out' folder.
	mkdir -p out/
	env GO111MODULE=on GOOS=linux CGO_ENABLED=0 go build -v -a -installsuffix cgo -o out/${PROJECT}-linux *.go

compile-win: ## Compiles and creates a Windows executable in the 'out' folder.
	mkdir -p out/
	env GO111MODULE=on GOOS=windows CGO_ENABLED=0 go build -v -a -installsuffix cgo -o out/${PROJECT}.exe *.go

compile-mac: ## Compiles and creates a MacOS executable in the 'out' folder.
	mkdir -p out/
	env GO111MODULE=on GOOS=darwin CGO_ENABLED=0 go build -v -a -installsuffix cgo -o out/${PROJECT}-macos *.go

compile: compile-lin compile-win compile-mac ## Compiles and creates all executables in the 'out' folder.

install: ## Compiles and installs the packages named by the import paths.
	go install
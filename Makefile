# -----------------------------------------------------------------------------
# Description: Makefile
# Author(s): retgits <https://github.com/retgits/>
# Last updated: 2019-09-10
# 
# This software may be modified and distributed under the terms of the
# MIT license. See the LICENSE file for details.
# -----------------------------------------------------------------------------

## The stage to deploy to
stage         := dev

## The name of the user in GitHub
github_user   := retgits

## The name of the project, defaults to the name of the current directory
project_name  := $(notdir $(CURDIR))

## The version of the project, either uses the current commit hash, or will default to "dev"
version       := $(strip $(if $(shell git describe --tags --always --dirty="-dev"),$(shell git describe --tags --always --dirty="-dev"),dev))

## The current date in UTC
date          := $(shell date -u '+%Y-%m-%d-%H:%M UTC')

## The Amazon S3 bucket to upload files to
aws_bucket    ?= $$AWSBUCKET

## Version flags for Go builds
version_flags := -ldflags='-X "github.com/$(github_user)/$(project_name)/main.Version=$(version)" -X "github.com/$(github_user)/$(project_name)/main.BuildTime=$(date)"'

# Suppress checking files and all Make output
.PHONY: help deps test build clean local deploy stage
.SILENT: help deps test build clean local deploy stage

# Targets
help: ## Displays the help for each target (this message).
	echo
	echo Usage: make [TARGET]
	echo
	echo Makefile targets
	grep -E '^[a-zA-Z_-]+:.*?## .*$$' Makefile | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
	echo

deps: ## Get the Go modules from the GOPROXY
	echo
	echo Getting Go modules from: $(shell go env GOPROXY)
	go get ./...
	echo

test: ## Run all unit tests and print coverage
	echo
	go test -cover ./...
	echo

build: ## Build the executable for Lambda
	echo
	GOOS=linux GOARCH=amd64 go build -o bin/$(project_name) $(if $V,-v) $(version_flags)
	echo

clean: ## Remove all generated files
	echo
	-rm -rf bin
	-rm temp-template.yaml
	echo

local: ## Run SAM to test the Lambda function using Docker
	echo
	sam local invoke WeeklyReviewFunction -e ./test/event.json
	echo

preparetemplate:
	rm -f temp-template.yaml
	cp template.yaml temp-template.yaml
	sed -i 's/version: xxx/version: $(version)/g' temp-template.yaml

deploy: clean build preparetemplate ## Deploy the app to AWS Lambda
	echo
	sam package --output-template-file packaged.yaml --s3-bucket $(aws_bucket)
	sam deploy --template-file packaged.yaml --stack-name $(project_name)-$(stage) --capabilities CAPABILITY_IAM
	aws cloudformation describe-stacks --stack-name $(project_name)-$(stage) --query 'Stacks[].Outputs'
	echo
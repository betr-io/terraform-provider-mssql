SHELL := /bin/bash

OPERATING_SYSTEM=Linux
TEST?=$$(go list ./... | grep -v 'vendor')
HOSTNAME=betr.io
NAMESPACE=betr
NAME=mssql
BINARY=terraform-provider-${NAME}
VERSION=0.2.6
OS_ARCH=linux_amd64
TERRAFORM=terraform

default: install

build:
	go build -o ${BINARY}

release:
	# Runs goreleaser locally (testrun)
	goreleaser release --rm-dist --skip-sign --skip-publish

install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

test:
	go test -i $(TEST) || exit 1
	echo $(TEST) | xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4

testacc:
	if [ -f .local.env ]; then source .local.env; fi && TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m

testacc-local:
	if [ -f .local.env ]; then source .local.env; fi && TF_ACC_LOCAL=1 go test $(TEST) -v $(TESTARGS) -timeout 120m

docker-start:
	cd test-fixtures/local && ${TERRAFORM} init && ${TERRAFORM} apply -auto-approve -var="operating_system=${OPERATING_SYSTEM}"

docker-stop:
	cd test-fixtures/local && ${TERRAFORM} destroy -auto-approve -var="operating_system=${OPERATING_SYSTEM}"

azure-create:
	cd test-fixtures/all && ${TERRAFORM} init && ${TERRAFORM} apply -auto-approve -var="operating_system=${OPERATING_SYSTEM}"

azure-destroy:
	cd test-fixtures/all && ${TERRAFORM} destroy -auto-approve -var="operating_system=${OPERATING_SYSTEM}"

PROJECTNAME=$(shell basename "$(PWD)")
CURRENT=$(shell echo $(PWD))

.PHONY: help

help: Makefile
	@echo
	@echo " Choose a command run in "$(PROJECTNAME)":"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo

## install: Install missing dependencies.
install:
	go mod download

## build: Builds the project.
build: main.go
	go build -ldflags "-X 'main.gitCommit=$$(git rev-parse --short HEAD)' -X 'main.date=$$(date --utc +%F_%T)'" -o build/cniep

## build-all: Build all linux plattforms
build-all: main.go
	for arch in amd64; do \
		for os in linux darwin; do \
			CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch go build -o "build/cniep_"$$os"_$$arch" $(LDFLAGS) -ldflags "-X 'main.gitCommit=$$(git rev-parse --short HEAD)' -X 'main.date=$$(date --utc +%F_%T)'"; \
		done; \
	done;
	/bin/chmod +x build/*

## docker-build: Build docker image
docker-build:
	docker build -t cniep .

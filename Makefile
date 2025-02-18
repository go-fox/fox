USER	:=	$(shell whoami)
REV 	:= 	$(shell git rev-parse --short HEAD)
OS		:=	$(shell uname -s)
VERSION_PROD := 0.0.0
VERSION_GIT_COMMIT=$(shell git describe --tags --always)
CUR_PATH := $(shell pwd)
MOD_FILES := $(shell find $(CUR_PATH) -name "*.mod")
MOD_PATHS := $(dir $(MOD_FILES))
FOX_UPDATE_PATHS := $(filter-out $(CUR_PATH)/,$(MOD_PATHS))
ENV := $(shell echo ${ENV})
# GOBIN > GOPATH > INSTALLDIR
# Mac OS X
ifeq ($(shell uname),Darwin)
GOBIN	:=	$(shell echo ${GOBIN} | cut -d':' -f1)
GOPATH	:=	$(shell echo $(GOPATH) | cut -d':' -f1)
endif

# Linux
ifeq ($(os),Linux)
GOBIN	:=	$(shell echo ${GOBIN} | cut -d':' -f1)
GOPATH	:=	$(shell echo $(GOPATH) | cut -d':' -f1)
endif

ifeq ($(ENV),prod)
VERSION := ${VERSION_PROD}
else
VERSION := ${VERSION_GIT_COMMIT}
endif

# Windows
ifeq ($(os),MINGW)
GOBIN	:=	$(subst \,/,$(GOBIN))
GOPATH	:=	$(subst \,/,$(GOPATH))
GOBIN :=/$(shell echo "$(GOBIN)" | cut -d';' -f1 | sed 's/://g')
GOPATH :=/$(shell echo "$(GOPATH)" | cut -d';' -f1 | sed 's/://g')
endif

.PHONY: env
env:
	@echo ${ENV}

.PHONY: revive
revive:
	revive -config revive.toml -exclude ./vendor/... ./...

.PHONY: proto
proto:
	protoc --proto_path=./third_party --go_out=paths=source_relative:./ --go-grpc_out=paths=source_relative:./ api/annotations/annotations.proto
	protoc --proto_path=./third_party --go_out=paths=source_relative:./ --go-grpc_out=paths=source_relative:./ api/protocol/protocol.proto
	protoc --proto_path=./third_party --go_out=paths=source_relative:./ --go-grpc_out=paths=source_relative:./ api/pagination/pagination.proto
	protoc --proto_path=./ --go_out=paths=source_relative:./ --go-grpc_out=paths=source_relative:./ errors/errors.proto

.PHONY: mod-tidy
mod-tidy:
	@for dir in ${MOD_PATHS}; do cd $$dir && go mod tidy ||exit; done
	@echo "go mod tidy done"

.PHONY: mod-update
mod-update:
	@for dir in ${MOD_PATHS}; do cd $$dir && go get -u && go mod tidy ||exit; done
	@echo "go mod update done"

.PHONY: mod-fox
mod-fox:
	@for dir in ${FOX_UPDATE_PATHS}; do cd $$dir && go get github.com/go-fox/fox@${VERSION_GIT_COMMIT}  ||exit; done
	@echo "go mod fox ${VERSION_GIT_COMMIT} done"

.PHONY: mod
mod: mod-fox mod-tidy

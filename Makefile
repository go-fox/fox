USER	:=	$(shell whoami)
REV 	:= 	$(shell git rev-parse --short HEAD)
OS		:=	$(shell uname -s)

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

# Windows
ifeq ($(os),MINGW)
GOBIN	:=	$(subst \,/,$(GOBIN))
GOPATH	:=	$(subst \,/,$(GOPATH))
GOBIN :=/$(shell echo "$(GOBIN)" | cut -d';' -f1 | sed 's/://g')
GOPATH :=/$(shell echo "$(GOPATH)" | cut -d';' -f1 | sed 's/://g')
endif

.PHONY: revive
revive:
	revive -config revive.toml -exclude ./vendor/... ./...

.PHONY: proto
proto:
	protoc --proto_path=./third_party --go_out=paths=source_relative:./ --go-grpc_out=paths=source_relative:./ api/annotations/annotations.proto
	protoc --proto_path=./third_party --go_out=paths=source_relative:./ --go-grpc_out=paths=source_relative:./ api/protocol/protocol.proto
	protoc --proto_path=./ --go_out=paths=source_relative:./ --go-grpc_out=paths=source_relative:./ errors/errors.proto

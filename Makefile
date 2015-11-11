
VERSION := 0.10.0
GOPATH  := ${PWD}
export GOPATH

SOURCES := $(shell find . -name '*.go')

BINARIES := \
  bin/windows_386/qpm.exe \
  bin/windows_amd64/qpm.exe \
  bin/linux_386/qpm \
  bin/linux_amd64/qpm \
  bin/darwin_386/qpm \
  bin/darwin_amd64/qpm

default: $(SOURCES)
	go install qpm.io/qpm

src/qpm.io/common/messages/qpm.pb.go: src/qpm.io/common/messages/qpm.proto bin/protoc-gen-go
	cd src/qpm.io/common/messages; \
	protoc --plugin=$$GOPATH/bin/protoc-gen-go --go_out=plugins=grpc:. *.proto

bin/protoc-gen-go:
	go install github.com/golang/protobuf/protoc-gen-go

.all: $(BINARIES)
	echo test

$(BINARIES): $(SOURCES)
	GOOS=$(firstword $(subst _, , $(word 2, $(subst /, ,$@)))) \
	GOARCH=$(word 2, $(subst _, , $(word 2, $(subst /, ,$@)))) \
	go install qpm.io/qpm

clean:
	@rm -rf $(BINARIES)


.PHONY: default clean .all


VERSION := 0.10.0
GOPATH  := ${PWD}
TS      := $(shell /bin/date "+%Y-%m-%d---%H-%M-%S")
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

.protobuf: src/qpm.io/common/messages/qpm.proto bin/protoc-gen-go
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
	@rm -rf staging/
	@rm -rf repository/

# Targets for building the Qt Maintence Tool Repository

bin/packager: $(SOURCES)
	go install qpm.io/tools/packager

staging/packages: $(BINARIES) bin/packager
	bin/packager staging

repository: staging/packages
	repogen -p staging/packages -r repository
	gsutil -m cp -r gs://www.qpm.io/repository gs://www.qpm.io/repository_$(TS)
	gsutil -m rsync -r repository gs://www.qpm.io/repository

downloads: $(BINARIES)
	gsutil -m cp -r gs://www.qpm.io/download gs://www.qpm.io/download_$(TS)
	gsutil -m rsync -x 'qpm|packager' -r bin gs://www.qpm.io/download/v$(VERSION)

.PHONY: default clean .all .protobuf

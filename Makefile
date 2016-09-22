
BINARY := qpm
VERSION := 0.11.0
BUILD := $(git rev-parse head)
TS      := $(shell /bin/date "+%Y-%m-%d---%H-%M-%S")
SOURCES := $(shell find . -name '*.go')
LDFLAGS := -ldflags "-X main.Version=${VERSION} -X main.Build=${BUILD}"

BINARIES := \
  ${GOPATH}/bin/windows_386/qpm.exe \
  ${GOPATH}/bin/windows_amd64/qpm.exe \
  ${GOPATH}/bin/linux_386/qpm \
  ${GOPATH}/bin/linux_amd64/qpm \
  ${GOPATH}/bin/darwin_386/qpm \
  ${GOPATH}/bin/darwin_amd64/qpm

default: $(SOURCES)
	go install ${LDFLAGS} qpm.io/qpm

.protobuf: common/messages/qpm.proto ${GOPATH}/bin/protoc-gen-go
	cd common/messages; \
	protoc --plugin=$$GOPATH/bin/protoc-gen-go --go_out=plugins=grpc:. *.proto

${GOPATH}/bin/protoc-gen-go:
	go get -u github.com/golang/protobuf/protoc-gen-go
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

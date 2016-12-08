
BINARY := qpm
VERSION := 0.11.0
BUILD := $(shell git rev-parse head | cut -c1-8)
TS      := $(shell /bin/date "+%Y-%m-%d---%H-%M-%S")
SOURCES := $(shell find . -name '*.go')
LDFLAGS := -ldflags "-X qpm.io/qpm/core.Version=${VERSION} -X qpm.io/qpm/core.Build=${BUILD}"
go_build = GOOS=$(1) GOARCH=$(2) go build ${LDFLAGS} -o ${GOPATH}/bin/$(1)_$(2)/$(3) qpm.io/qpm

default: $(SOURCES)
	go install ${LDFLAGS} qpm.io/qpm

## Supported Platforms ##

.windows: .windows_386 .windows_amd64

.linux: .linux_386 .linux_amd64

.darwin: .darwin_386 .darwin_amd64

.freebsd: .freebsd_386 .freebsd_amd64

.all: .windows .linux .darwin .freebsd

## Platform targets ##

.windows_386: ${GOPATH}/bin/windows_386/qpm.exe

.windows_amd64: ${GOPATH}/bin/windows_amd64/qpm.exe

.linux_386: ${GOPATH}/bin/linux_386/qpm

.linux_amd64: ${GOPATH}/bin/linux_amd64/qpm

.darwin_386: ${GOPATH}/bin/darwin_386/qpm

.darwin_amd64: ${GOPATH}/bin/darwin_amd64/qpm

.freebsd_386: ${GOPATH}/bin/freebsd_386/qpm

.freebsd_amd64: ${GOPATH}/bin/freebsd_amd64/qpm

## Target build steps ##

${GOPATH}/bin/windows_386/qpm.exe: $(SOURCES)
	$(call go_build,windows,386,qpm.exe)

${GOPATH}/bin/windows_amd64/qpm.exe: $(SOURCES)
	$(call go_build,windows,amd64,qpm.exe)
	
${GOPATH}/bin/linux_386/qpm: $(SOURCES)
	$(call go_build,linux,386,qpm)
	
${GOPATH}/bin/linux_amd64/qpm: $(SOURCES)
	$(call go_build,linux,amd64,qpm)
	
${GOPATH}/bin/darwin_386/qpm: $(SOURCES)
	$(call go_build,darwin,386,qpm)
	
${GOPATH}/bin/darwin_amd64/qpm: $(SOURCES)
	$(call go_build,darwin,amd64,qpm)
	
${GOPATH}/bin/freebsd_386/qpm: $(SOURCES)
	$(call go_build,freebsd,386,qpm)
	
${GOPATH}/bin/freebsd_amd64/qpm: $(SOURCES)
	$(call go_build,freebsd,amd64,qpm)

## Protobuf generation ##

.protobuf: common/messages/qpm.proto ${GOPATH}/bin/protoc-gen-go
	cd common/messages; \
	protoc --plugin=$$GOPATH/bin/protoc-gen-go --go_out=plugins=grpc:. *.proto

${GOPATH}/bin/protoc-gen-go:
	go get -u github.com/golang/protobuf/protoc-gen-go

clean:
	@rm -rf $(BINARIES)
	@rm -rf staging/
	@rm -rf repository/

## Targets for building the Qt Maintence Tool Repository ##

${GOPATH}/bin/packager: $(SOURCES)
	go install ${LDFLAGS} qpm.io/tools/packager

.staging: .all ${GOPATH}/bin/packager
	${GOPATH}/bin/packager staging

.repository: .staging
	repogen -p staging/packages -r repository
	gsutil -m cp -r gs://dev.qpm.io/repository gs://dev.qpm.io/repository_$(TS)
	gsutil -m rsync -r repository gs://dev.qpm.io/repository

.downloads: .all
	gsutil -m cp -r gs://www.qpm.io/download gs://www.qpm.io/download_$(TS)
	gsutil -m rsync -x 'qpm|packager' -r bin gs://www.qpm.io/download/v$(VERSION)

.PHONY: default clean .protobuf .all  \
	.downloads .repository .staging/packages \
	.windows .windows_386 .windows_amd64 \
	.linux .linux_386 .linux_amd64 \
	.darwin .darwin_386 .darwin_amd64 \
	.freebsd .freebsd_386 .freebsd_amd64

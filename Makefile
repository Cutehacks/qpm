install:
	GOPATH=$(TRAVIS_BUILD_DIR):$(PWD):$(GOPATH) go install qpm.io/qpm
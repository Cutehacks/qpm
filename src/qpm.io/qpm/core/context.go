package core

import (
	"google.golang.org/grpc"
	"log"
	"os"
	msg "qpm.io/common/messages"
)

const (
	Version     = "0.0.1"
	PackageFile = "qpm.json"
	Vendor      = "vendor"
	Address     = "pkg.qpm.io:7000"
	GitHub      = "https://api.github.com/repos"
	Tarball     = "tarball"
	TarSuffix   = ".tar.gz"
	License     = "LICENSE"
)

type Context struct {
	Log    *log.Logger
	Client msg.QpmClient
}

func NewContext() *Context {
	log := log.New(os.Stderr, "QPM: ", log.LstdFlags)

	conn, err := grpc.Dial(Address)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	return &Context{
		Log:    log,
		Client: msg.NewQpmClient(conn),
	}
}

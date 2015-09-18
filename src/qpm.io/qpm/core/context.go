package core

import (
	"google.golang.org/grpc"
	"log"
	"os"
	msg "qpm.io/common/messages"
	"google.golang.org/grpc/credentials"
)

const (
	Version       = "0.0.1"
	PackageFile   = "qpm.json"
	SignatureFile = "qpm.asc"
	Vendor        = "vendor"
	Address       = "pkg.qpm.io:7000"
	LicenseFile   = "LICENSE"
)

type Context struct {
	Log    *log.Logger
	Client msg.QpmClient
}

func NewContext() *Context {
	log := log.New(os.Stderr, "QPM: ", log.LstdFlags)

	creds := credentials.NewClientTLSFromCert(nil, "")
	address := os.Getenv("SERVER")
	if address == "" {
		address = Address
	}
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	return &Context{
		Log:    log,
		Client: msg.NewQpmClient(conn),
	}
}

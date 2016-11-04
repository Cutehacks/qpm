package core

import (
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"log"
	"os"
	msg "qpm.io/common/messages"
	"runtime"
)

var (
	Version = "0.10.0"
	Build   = ""
)

const (
	PackageFile   = "qpm.json"
	SignatureFile = "qpm.asc"
	Vendor        = "vendor"
	Address       = "pkg.qpm.io:7000"
	LicenseFile   = "LICENSE"
)

var UA = fmt.Sprintf("qpm/%v (%s; %s)", Version, runtime.GOOS, runtime.GOARCH)

type Context struct {
	Log    *log.Logger
	Client msg.QpmClient
}

func NewContext() *Context {
	log := log.New(os.Stderr, "QPM: ", log.LstdFlags)

	address := os.Getenv("SERVER")
	if address == "" {
		address = Address
	}

	noTls := os.Getenv("NO_TLS") == "1"

	var tlsOption grpc.DialOption
	if noTls {
		tlsOption = grpc.WithInsecure()
	} else {
		tlsOption = grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, ""))
	}

	conn, err := grpc.Dial(address, tlsOption, grpc.WithUserAgent(UA))

	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	return &Context{
		Log:    log,
		Client: msg.NewQpmClient(conn),
	}
}

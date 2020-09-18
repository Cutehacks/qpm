package core

import (
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"runtime"
	"time"

	"github.com/mwitkow/go-http-dialer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	msg "qpm.io/common/messages"
)

var (
	Version = "0.X.x"
	Build   = "master"
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

	// noTls := os.Getenv("NO_TLS") == "1"

	// var tlsOption grpc.DialOption
	// if noTls {
	// 	tlsOption = grpc.WithInsecure()
	// } else {
	// 	tlsOption = grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, ""))
	// }

	// conn, err := grpc.Dial(address, tlsOption, grpc.WithUserAgent(UA))

	opts := make([]grpc.DialOption, 0)
	opts = append(opts, grpc.WithUserAgent(UA))

	noTLS := os.Getenv("NO_TLS") == "1"
	if noTLS {
		opts = append(opts, grpc.WithInsecure())
	} else {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, "")))
	}

	httpProxy := os.Getenv("HTTP_PROXY")
	if httpProxy != "" {
		log.Println("env: ", httpProxy)
		httpProxyURL, err := url.Parse(httpProxy)
		if err != nil {
			log.Fatalf("did not get http proxy: %v", err)
		} else {
			proxyDialer := http_dialer.New(httpProxyURL)
			opts = append(opts, grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) { return proxyDialer.Dial("tcp", addr) }))
		}
	}

	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	return &Context{
		Log:    log,
		Client: msg.NewQpmClient(conn),
	}
}

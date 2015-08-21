
[![Build Status](https://travis-ci.org/Cutehacks/qpm.svg?branch=master)](https://travis-ci.org/Cutehacks/qpm)

# Introduction

> qpm is a command line tool for accessing the qpm.io Qt package registry and installing dependencies.

qpm and the corresponding qpm.io service provide a way for Qt developers to search, install and
publish source code (QML, JS, C++) components that can be compiled into Qt applications or libraries.

## Goals

qpm aims to:

* Help application and library developers.
* Provide a central registry of re-usable Qt components.
* Produce consistent dependencies for applications.

qpm is **not**:

* A binary package manager (like apt-get, yum, or brew).
* A build system (like qmake or qbs).
* A version control system (like Git or SVN.

## How it works

Packages are published to the qpm.io registry by a package author. Only the meta data is stored in
the registry, the actual source code resides in a public repository under the author's control
(eg: Github).

An application developer installs a package that he or she wishes to use and qpm will fetch the source
code for that package and also fetch any *nested* dependencies that the package has. The application
developer can then use those package dependencies in their application.

A package in qpm terms consists of a `package.json` file which contains meta data about the package
such as the name, maintainer, version, dependencies, etc. The name of the package must be unique as
it is used later to avoid naming collisions.

When an application developer installs a package for use in a Qt app, qpm will automatically create a
package.json file to track the apps dependencies. Even though apps contain package.json files, they
should not be published to the registry. The package.json file should be checked into your version
control system so that the dependencies can be re-created later. Upon installing a package dependency, 
the qpm tool will create a directory called `vendor` which contains the source code of the package.
This is included in the application (see below) and can used as normal.

# Installing

There are currently 3 ways to install the qpm client:

## Download a binary

The pre-compiled binaries can be downloaded from http://www.qpm.io.

## Use Qt's Maintenance Tool

qpm is a available via a custom repository which can be added to Qt's Maintence Tool
which is part of the SDK.

* Open maintenancetool.
* Select "Add or remove components" (do not click Continue yet).
* Click "Settings" in the bottom left.
* Go to the "Repositories" tab.
* Scroll to the bottom and select "User defined repositories".
* Click the "Add" button underneath.
* Enter the following URL: `https://storage.googleapis.com/www.qpm.io/repository`.[1]
* Leave username and password empty.
* Clicked "Test" to ensure everything is working.
* Click "Ok" to exit Settings.
* Click "Continue".
* Expand the `qpm` item and choose your platform.

NOTE[1]: You can (and should) use HTTPS for the URL if it works, but at least on Mac OS X
Yosemite, there is a problem with Qt's handing of Google's SSL certificate that prevents
it from working.

## Compile from source

* Ensure you have [Go](http://golang.org/) installed (tested with 1.4.2 and 1.5)
* Clone this repository

The qpm tool has it's dependencies stored in the repo as Git submodules so to initialize
those you need to navigate to the root of the project and run:

```
git submodule init
git submodule update
```

Go will search GOPATH for dependencies so set that accordingly. From the root of the repository:

```
export GOPATH=`pwd`
export PATH=$GOPATH/bin:$PATH
```

Compiling the command line tool is as simple as:

```
go build qpm.io/qpm
```

Although you should probably install it as well with (installing runs the build step):

```
go install qpm.io/qpm
```

If you are using Go 1.5, then cross compiling is simply a matter of setting ´GOARCH´
and ´GOOS´, for example to generate a 32-bit Windows build:

```
GOOS=windows GOARCH=386 go install qpm.io/qpm
```

The compiled binaries are placed in `$GOPATH/bin`.

# Usage for App Developers

Installing package dependencies with qpm requires no login or registration. You can search for packages
using the following command:

```
qpm search <package name>
```

This will list a package and version number as well as the author. To get more information about the
package, you can run:

```
qpm info <package name>
```

If the package sounds useful, you can install it:

```
qpm install <package name>
```

By default qpm will install the latest version of the component, however you can request a specific
version using the following syntax:

```
qpm install package@1.0.1
```

Installing a package for the first time will create a new file called `package.json`. Subsequent
installs will update this file with the new package. If you want to install all of the packages
listed in your `package.json` file, then you use:

```
qpm install
```

With no arguments, this will install your dependent packages. Upon installing a new package, there
will be a directory called `vendor` which contains the code for each package in it's own
subdirectory. The vendor directory will also contain a file called `vendor.pri` which should be
included in your applications .pro file like so:

```
include(vendor/vendor.pri)
```

The vendor.pri takes care of including each package's .pri file which will expose the contents of the
package to your project's build. Package .pri typically add files to `SOURCES`, `HEADERS` and
`RESOURCES` so that they can be accessible to your app.

If the package contains QML or Javascript code, then you need to register it with the QML engine, like
so in your main.cpp (or wherever):

```
    QQmlApplicationEngine engine;
	QPM_INIT(engine)
    engine.load(QUrl(QStringLiteral("qrc:/main.qml")));
```

It is important that you call `QPM_INIT` before calling `engine.load()` because otherwise you will
errors about missing components.

For packages that contain C++ code, these need to be manually exposed to QML for the time being, but
perhaps we can find something clever here in the future.

Uninstalling package can be done with:

```
qpm uninstall <package name>
```

# Usage for Package Authors

If you have an idea for a Qt component that you would like to share, you can publish it on qpm.io.
qpm can help you get started by running the following interactive command:

```
qpm init
```

This asks some basic questions to get you going and to generate a package.json file containing your
package's meta data. If you are starting from scratch, you can let the `init` command generate some
boilerplate code for you. This generates some basic files that you can extend as you create your
module. The boilerplate currently generated is:

* **qmldir**: QML module definition file
* **package.pri**: Pri file for inclusion (indirectly) by apps
* **package.qrc**: A Qt resource file for listing embedded source such as QML, JS, etc.

To simplify deployment of applications that use your package, we recommend package authors to add
as much as possible (QML, JS, PNG, etc.) to the resource file so everything gets compiled into
the application binary.

If you don't use the init command to generate the boilerplate, then you should use the following
command to ensure that everything is in the right place with the right format:

```
qpm check
```

This command checks for common mistakes that will likely result in yor package not working correctly.

Finally, when your package is ready to go, you can run the following command:

```
qpm publish
```

This command will prompt you to login or register and is required every time you publish a package.
This is to prevent other people from publishing your package. In the future, it will be possible to
have several contributors that can publish the same package.

## A note on versioning

qpm is somewhat strict on versioning. This is because one of the goals of qpm is to have consistent
dependencies. If two applications use the same version of a package, they will contain the same code.

A version consists of a label (eg: 0.0.1) and a revision (eg: Git SHA1 or tag). For package authors,
this means that once a version has been published, it cannot be republished with a different revision.
Publishing a new revision entails publishing a new version. Multiple labels can point to the same
revision though. This is to handle the case where, for example, a release candidate (1.0.0-rc1) is
identical to the final release (1.0.0).

Publishing a new version of a package, is simply a matter of modifying the respective fields in the
package.json file and running `qpm publish`. The meta data for a package is not versioned, so
changing the description or author will affect all versions of a package. Nested dependencies for a
package **are** versioned so publishing a new version with new dependencies will not affect previous
versions.

# Contributing

qpm is open source and we encourage other developers to contribute if they see an opportunity. The tool
is written in Go so following the steps above to compile from source is a good place to start.

qpm uses Google's Protocol Buffers to communicate between the command line tool and the server. More specifically,
it uses the new "proto3" version which has support for Go and also gRPC which is also used.  

If you make change to the .proto files you need to follow the steps below.

## Prerequisites

Ensure the following component is installed and the `protoc` command is in your path somewhere.
Note that repository is independent of Go so it should be cloned and installed globally.

* https://github.com/google/protobuf (brew install --devel protobuf) or (port install protobuf3-cpp)

## Generating Protocol Buffers

The generated code for the protocol buffers is checked in to the repo so the following
steps are only needed if you modify the .proto file(s).

In order to use the source generator you need to first build the `protoc-gen-go` tool:
 
```
cd src/github.com/golang/protobuf/protoc-gen-go
go build
go install
```

You can verify that the binary was built correctly in `$GOPATH/bin`. The above is only
done once.

Every time you make a modification to a .proto file, you need to regenerate some code
by doing the following:

```
cd src/qpm.io/common/messages
protoc --go_out=plugins=grpc:. *.proto
```

This will generate files of the format \*.pb.go in the same directory. You should
check in the generated files for now.


# FAQ

## How stable is it?

This project is very young so your mileage may vary. At this point we are at the
"Developer Preview" stage where feedback is the most important thing. We took
a few shortcuts to get this out quickly, but we hope it scratches an itch for a
few people at least.

## Why did you make qpm?

Having now worked on frameworks outside of Qt, we saw that there was a big hole in
Qt's developer offering. iOS has CocoaPods, Android has Maven/Gradle, Node has npm,
Ruby has gem, Python has pip, Go has.. well Go. We wanted something that integrated
well with Qt so we made something. We took inspiration from all of the above tools
in some way or another.

## Why Go?

The decision came down to Node.js, Go and Qt. Ideally we wanted to use a language
that allowed us to share code between the client and server. Node and Go have
stronger server side support than Qt/C++ so that eliminated Qt. There are plenty of
cloud providers that support Go and Node but very few (any?) that support C++.

The beautiful thing about Go is that it compiles to a single static binary which is
exactly what we wanted for a command line tool. It also makes server deploys easy :)

## Why is the binary so big?

This seems to be a result of using Go which presumably has a rather large runtime
associated with it.

## How do I add a new FAQ?

Submit a pull request to this file with your question and create an issue for us
to answer it!

#!/usr/bin/env bash
mkdir ./dstroot
install -m 755 ../bin/qpm ./dstroot/
pkgbuild --identifier io.qpm --root ./dstroot/ --install-location /usr/local/bin qpm.pkg
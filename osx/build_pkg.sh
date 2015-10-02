mkdir ./dstroot
install -m 755 ../bin/qpm ./dstroot/
pkgbuild --identifier io.qpm --root ./dstroot/ --install-location /opt/local/bin qpm.pkg
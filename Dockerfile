FROM kjhman21/dev:go1.11.2-solc0.4.24
MAINTAINER Jesse Lee jesse.lee@groundx.xyz

ENV PKG_DIR /klaytn-docker-pkg
ENV SRC_DIR /go/src/github.com/ground-x/klaytn

RUN mkdir -p $PKG_DIR/bin
RUN mkdir -p $PKG_DIR/conf

ADD . $SRC_DIR
RUN cd $SRC_DIR && make klay kcn kpn ken kscn

RUN cp $SRC_DIR/build/bin/klay /usr/bin/
RUN cp $SRC_DIR/build/bin/kcn /usr/bin/
RUN cp $SRC_DIR/build/bin/kpn /usr/bin/
RUN cp $SRC_DIR/build/bin/ken /usr/bin/
RUN cp $SRC_DIR/build/bin/kscn /usr/bin/

# packaging
RUN cp $SRC_DIR/build/bin/kcn $PKG_DIR/bin/
RUN cp $SRC_DIR/build/bin/kpn $PKG_DIR/bin/
RUN cp $SRC_DIR/build/bin/ken $PKG_DIR/bin/
RUN cp $SRC_DIR/build/bin/kscn $PKG_DIR/bin/

RUN cp $SRC_DIR/build/packaging/linux/bin/kcnd $PKG_DIR/bin/
RUN cp $SRC_DIR/build/packaging/linux/bin/kpnd $PKG_DIR/bin/
RUN cp $SRC_DIR/build/packaging/linux/bin/kend $PKG_DIR/bin/

RUN cp $SRC_DIR/build/packaging/linux/conf/kcnd.conf $PKG_DIR/conf/
RUN cp $SRC_DIR/build/packaging/linux/conf/kpnd.conf $PKG_DIR/conf/
RUN cp $SRC_DIR/build/packaging/linux/conf/kend.conf $PKG_DIR/conf/

EXPOSE 8551 8552 32323 61001 32323/udp

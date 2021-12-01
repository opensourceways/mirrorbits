FROM openeuler/openeuler:21.03

LABEL maintainer="etix@l0cal.com"

RUN mkdir -p /go/bin && mkdir -p /go/src/mirrorbits
RUN yum -y update
RUN yum install -y sudo
RUN sudo yum install -y go
ENV GOROOT=/usr/lib/golang
ENV PATH=$PATH:/usr/lib/golang/bin
ENV GOPATH=/go
ENV PATH=$GOPATH/bin/:$PATH

RUN sudo yum install -y gcc make && sudo yum install -y pkg-config git zlib autoconf automake libtool curl g++ unzip protobuf-compiler rsync python3 python3-pip python3-devel
RUN pip install pyyaml prettytable && pip install git-extras && sudo yum install -y redis
# install geoipupdate binary, NOTE: default configuration file located at /usr/local/etc/GeoIP.conf
# and geoip folder is /usr/share/GeoIP env GO111MODULE=on go get github.com/maxmind/geoipupdate/v4/cmd/geoipupdate
ENV GO111MODULE=on
RUN go get github.com/maxmind/geoipupdate/v4/cmd/geoipupdate && \
    mkdir -p /usr/share/GeoIP
COPY . /go/src/mirrorbits
RUN mkdir -p /srv/repo /var/log/mirrorbits && \
    cd /go/src/mirrorbits && make && \
    make install PREFIX=/usr
RUN cp /go/src/mirrorbits/contrib/docker/mirrorbits.conf /etc/mirrorbits.conf
COPY scripts /
RUN cd / && git clone https://github.com/protocolbuffers/protobuf.git
RUN cd protobuf && git submodule update --init --recursive
RUN ./autogen.sh && ./configure --prefix=/usr/local/protobuf && make && make check && make install
ENV PATH=$PATH:/usr/local/protobuf/bin
ENV LD_LIBRARY_PATH=$LD_LIBRARY_PATH:/usr/local/protobuf/lib
ENV LIBRARY_PATH=$LIBRARY_PATH:/usr/local/protobuf/lib
RUN cd / && git clone https://github.com/tj/git-extras.git
ENV PATH=$PATH:/git-extras/bin
ENTRYPOINT /usr/bin/mirrorbits daemon -config /etc/mirrorbits.conf

EXPOSE 8080

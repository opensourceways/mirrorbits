FROM openeuler/openeuler:22.03-lts-sp1

LABEL maintainer="etix@l0cal.com"

RUN mkdir -p /go/bin && mkdir -p /go/src/mirrorbits
RUN yum -y update && yum install -y sudo && sudo yum install -y go
ENV GOROOT=/usr/lib/golang
ENV PATH=$PATH:/usr/lib/golang/bin
ENV GOPATH=/go
ENV PATH=$GOPATH/bin/:$PATH
ENV GO111MODULE=on
RUN go env -w GOPROXY=https://goproxy.cn,direct

RUN sudo yum install -y gcc make && sudo yum install -y pkg-config git zlib zlib-devel autoconf automake libtool curl g++ unzip protobuf-compiler rsync python3 python3-pip python3-devel bazel
RUN pip install pyyaml prettytable && sudo yum install -y redis
# install geoipupdate binary, NOTE: default configuration file located at /usr/local/etc/GeoIP.conf
# and geoip folder is /usr/share/GeoIP env GO111MODULE=on go get github.com/maxmind/geoipupdate/v4/cmd/geoipupdate
RUN go get github.com/maxmind/geoipupdate/v4/cmd/geoipupdate && mkdir -p /usr/share/GeoIP/
COPY . /go/src/mirrorbits
COPY GeoIP /usr/share/GeoIP/
COPY scripts /
RUN chmod u+x /usr/share/GeoIP

RUN mkdir -p /srv/repo /var/log/mirrorbits && \
    cd /go/src/mirrorbits && make && \
    make install PREFIX=/usr
RUN cp /go/src/mirrorbits/contrib/docker/mirrorbits.conf /etc/mirrorbits.conf

RUN cd / && git clone https://github.com/tj/git-extras.git
RUN /git-extras/bin/git-extras update
ENTRYPOINT /usr/bin/mirrorbits daemon -config /etc/mirrorbits.conf
EXPOSE 8080

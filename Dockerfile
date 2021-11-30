FROM openeuler/openeuler:21.03

LABEL maintainer="etix@l0cal.com"

RUN yum install -y update
RUN yum install -y sudo
RUN sudo yum install -y go
ENV GOROOT=/usr/lib/golang
ENV PATH=$PATH:/usr/lib/golang/bin
ENV GOPATH=/go
RUN mkdir -p /go/bin && mkdir -p /go/src
ENV PATH=$GOPATH/bin
ADD . /go/src/mirrorbits

RUN yum install -y gcc && yum install -y pkg-config zlib1g-dev protobuf-compiler libprotoc-dev rsync python3 python3-pip git-extras python3-devel
RUN pip install pyyaml prettytable && yum install -y redis
# install geoipupdate binary, NOTE: default configuration file located at /usr/local/etc/GeoIP.conf
# and geoip folder is /usr/share/GeoIP
RUN GO111MODULE=on && go get github.com/maxmind/geoipupdate/v4/cmd/geoipupdate && \
    mkdir /usr/share/GeoIP
RUN mkdir /srv/repo /var/log/mirrorbits && \
    cd /go/src/mirrorbits && make && \
    make install PREFIX=/usr
RUN cp /go/src/mirrorbits/contrib/docker/mirrorbits.conf /etc/mirrorbits.conf
COPY scripts /

ENTRYPOINT /usr/bin/mirrorbits daemon -config /etc/mirrorbits.conf

EXPOSE 8080

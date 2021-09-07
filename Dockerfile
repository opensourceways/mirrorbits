FROM golang:latest

LABEL maintainer="etix@l0cal.com"

ADD . /go/mirrorbits

RUN apt-get update -y && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y pkg-config zlib1g-dev protobuf-compiler libprotoc-dev rsync python3 python3-pip git-extras && \
    apt-get clean
RUN pip3 install pyyaml prettytable && apt install -y redis
# install geoipupdate binary, NOTE: default configuration file located at /usr/local/etc/GeoIP.conf
# and geoip folder is /usr/share/GeoIP
RUN GO111MODULE=on && go get github.com/maxmind/geoipupdate/v4/cmd/geoipupdate && \
    mkdir /usr/share/GeoIP
RUN mkdir /srv/repo /var/log/mirrorbits && \
    cd /go/mirrorbits && make && \
    make install PREFIX=/usr
RUN cp /go/mirrorbits/contrib/docker/mirrorbits.conf /etc/mirrorbits.conf
COPY scripts /

ENTRYPOINT /usr/bin/mirrorbits daemon -config /etc/mirrorbits.conf

EXPOSE 8080

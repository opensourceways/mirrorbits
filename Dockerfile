FROM openeuler/openeuler:24.03-lts

RUN dnf -y update && \
    dnf install -y golang && \
    go env -w GO111MODULE=on && \
    dnf install -y gcc make pkg-config git zlib zlib-devel autoconf automake libtool curl g++ unzip protobuf-compiler rsync python3 python3-pip python3-devel bazel && \
    groupadd -g 1000 mirrorbits && useradd -u 1000 -g 1000 -s /sbin/nologin -m mirrorbits && \
    mkdir -p /go/bin && mkdir -p /go/src/mirrorbits && \
    go install github.com/maxmind/geoipupdate/v7/cmd/geoipupdate@latest && cp ~/go/bin/geoipupdate /usr/local/bin/ && \
    mkdir -p /opt/mirrorbits/GeoIP

WORKDIR /go/src/mirrorbits
COPY . .

RUN mkdir -p /var/log/mirrorbits && chown 1000:1000 /var/log/mirrorbits && \
    cd /go/src/mirrorbits && mkdir -p bin && make build && cp bin/mirrorbits /usr/local/bin/ && \
    cd /usr/local/src && git clone https://github.com/tj/git-extras.git && \
    cd /usr/local/src/git-extras && git checkout tags/7.2.0 -b tag-7.2.0 && cp bin/* /usr/local/bin && \
    mkdir -p /opt/mirrorbits/templates /opt/mirrorbits/python-scripts && \
    cp /go/src/mirrorbits/templates/* /opt/mirrorbits/templates && cp /go/src/mirrorbits/scripts/* /opt/mirrorbits/python-scripts && \
    chown -R 1000:1000 /opt/mirrorbits

USER mirrorbits
WORKDIR /opt/mirrorbits

RUN cd /opt/mirrorbits/python-scripts && python3 -m venv venv && ./venv/bin/pip install pyyaml prettytable

ENTRYPOINT geoipupdate -f /vault/secrets/geoip.conf -d /opt/mirrorbits/GeoIP && rm -f /vault/secrets/geoip.conf && mirrorbits daemon -config /vault/secrets/mirrorbits.conf
EXPOSE 8080

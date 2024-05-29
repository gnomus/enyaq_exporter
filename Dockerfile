FROM golang

RUN git clone https://github.com/chr4/enyaq_exporter
WORKDIR /go/enyaq_exporter
RUN git checkout v0.1.2
RUN go build .


COPY entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]

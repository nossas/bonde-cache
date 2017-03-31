FROM golang

ENV CACHE_INTERVAL 600

VOLUME ["/go/src/github.com/bonde-cache"]
WORKDIR /go/src/github.com/bonde-cache
COPY . .

RUN go get github.com/tools/godep
RUN godep restore
RUN go install github.com/bonde-cache

ENTRYPOINT /go/bin/bonde-cache

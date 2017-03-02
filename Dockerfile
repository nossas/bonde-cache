FROM golang:onbuild
VOLUME ["/go/src/app"]
RUN go get github.com/codegangsta/gin
ENV CACHE_INTERVAL 600

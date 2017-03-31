FROM golang

ENV CACHE_INTERVAL 600

VOLUME ["/go/src/app"]
WORKDIR /go/src/app
COPY . .

ENTRYPOINT ./bonde-cache

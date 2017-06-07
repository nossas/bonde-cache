FROM golang:onbuild
VOLUME ["/go/src/app"]
WORKDIR /go/src/app
ENV CACHE_INTERVAL 60
EXPOSE 80 443

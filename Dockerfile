FROM golang:onbuild
VOLUME ["/go/src/app"]
ENV CACHE_INTERVAL 600
EXPOSE 80 443

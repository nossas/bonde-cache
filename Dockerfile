FROM golang:onbuild
RUN go get github.com/codegangsta/gin
ENV CACHE_INTERVAL 300

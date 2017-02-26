FROM golang:onbuild
RUN go get github.com/codegangsta/gin
ENV PORT 5000

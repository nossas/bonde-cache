FROM golang

ENV CACHE_INTERVAL 600

VOLUME ["/go/src/app"]
WORKDIR /go/src/app
COPY . .

CMD ["bonde-cache"]
EXPOSE 443 80

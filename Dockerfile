FROM golang
WORKDIR /go/src/app

ENV CACHE_INTERVAL 600
COPY . .
CMD ["bonde-cache"]
EXPOSE 443 80

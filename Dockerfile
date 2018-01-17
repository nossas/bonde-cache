FROM golang

WORKDIR /go/src/app
ARG TIMEZONE="America/Sao_Paulo"
RUN set -x \
	&& apt-get update \
	&& apt-get upgrade -y \
	&& echo "=> Needed packages:" \
    && apt-get install -y --no-install-recommends apt-utils curl ca-certificates tar openssl xz-utils s3cmd \
    && echo "=> Configuring and installing timezone (${TIMEZONE}):" \
    && echo ${TIMEZONE} > /etc/timezone \
    && dpkg-reconfigure -f noninteractive tzdata \
    && rm -rf /var/lib/apt/lists/* \
    && apt-get purge -y --auto-remove apt-utils \
	&& mkdir -p ./data/db ./data/certificates \
	&& touch ./data/db/bonde-cache.db \
	&& chmod -R 777 data

ENV CACHE_INTERVAL 30
EXPOSE 80 443
RUN go get -u github.com/golang/dep/cmd/dep
COPY . .
RUN dep ensure && go build && go install
CMD ["app"]

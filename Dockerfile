FROM golang

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
    && apt-get purge -y --auto-remove apt-utils

ENV CACHE_INTERVAL 30
EXPOSE 80 443
VOLUME ["/go/src/app"]
WORKDIR /go/src/app
RUN mkdir -p data/db && chmod -R 777 data && go get -u github.com/golang/dep/cmd/dep
COPY . .
RUN dep ensure && go build && go install
CMD ["app"]
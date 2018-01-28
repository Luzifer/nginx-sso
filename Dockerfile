FROM golang:alpine

LABEL maintainer "Knut Ahlers <knut@ahlers.me>"

ADD . /go/src/github.com/Luzifer/nginx-sso
WORKDIR /go/src/github.com/Luzifer/nginx-sso

RUN set -ex \
 && apk --no-cache add git ca-certificates bash curl \
 && go install -ldflags "-X main.version=$(git describe --tags || git rev-parse --short HEAD || echo dev)" \
 && curl -sSfLo /usr/local/bin/dumb-init https://github.com/Yelp/dumb-init/releases/download/v1.2.1/dumb-init_1.2.1_amd64 \
 && chmod +x /usr/local/bin/dumb-init \
 && apk --no-cache del --purge git curl

EXPOSE 8082

VOLUME ["/data"]

ADD docker-start.sh /usr/local/bin

ENTRYPOINT ["/usr/local/bin/docker-start.sh"]
CMD ["--"]

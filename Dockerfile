FROM golang:1.25-alpine@sha256:26111811bc967321e7b6f852e914d14bede324cd1accb7f81811929a6a57fea9 as builder

ADD . /go/src/github.com/Luzifer/nginx-sso
WORKDIR /go/src/github.com/Luzifer/nginx-sso

ENV CGO_ENABLED=0

RUN set -ex \
 && apk add --no-cache \
      git \
 && go install \
      -ldflags "-X main.version=$(git describe --tags || git rev-parse --short HEAD || echo dev)" \
      -mod=readonly


FROM alpine:3.23@sha256:51183f2cfa6320055da30872f211093f9ff1d3cf06f39a0bdb212314c5dc7375

LABEL maintainer "Knut Ahlers <knut@ahlers.me>"

RUN set -ex \
 && apk --no-cache add \
      bash \
      ca-certificates \
      dumb-init \
 && mkdir /data \
 && chown -R 1000:1000 /data

COPY --from=builder /go/bin/nginx-sso                                     /usr/local/bin/
COPY --from=builder /go/src/github.com/Luzifer/nginx-sso/config.yaml      /usr/local/share/nginx-sso/
COPY --from=builder /go/src/github.com/Luzifer/nginx-sso/docker-start.sh  /usr/local/bin/
COPY --from=builder /go/src/github.com/Luzifer/nginx-sso/frontend/*       /usr/local/share/nginx-sso/frontend/

EXPOSE 8082
VOLUME ["/data"]

USER 1000:1000

ENTRYPOINT ["/usr/local/bin/docker-start.sh"]
CMD ["--"]

# vim: set ft=Dockerfile:

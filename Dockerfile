FROM golang:1.26-alpine@sha256:c2a1f7b2095d046ae14b286b18413a05bb82c9bca9b25fe7ff5efef0f0826166 as builder

ADD . /go/src/github.com/Luzifer/nginx-sso
WORKDIR /go/src/github.com/Luzifer/nginx-sso

ENV CGO_ENABLED=0

RUN set -ex \
 && apk add --no-cache \
      git \
 && go install \
      -ldflags "-X main.version=$(git describe --tags || git rev-parse --short HEAD || echo dev)" \
      -mod=readonly


FROM alpine:3.23@sha256:c69a6ff7c24d1ffa913798501d0e7104e0e9764e28eb44a930939f91ef829e64

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

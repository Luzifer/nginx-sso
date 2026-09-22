FROM golang:1.27.1-alpine@sha256:4cb7ac979db5fcc41cae44b2227ba5ab8a51e8807f40d9ba4dee20a0ad960b5b as builder

ADD . /go/src/github.com/Luzifer/nginx-sso
WORKDIR /go/src/github.com/Luzifer/nginx-sso

ENV CGO_ENABLED=0

RUN set -ex \
 && apk add --no-cache \
      git \
 && go install \
      -ldflags "-X main.version=$(git describe --tags || git rev-parse --short HEAD || echo dev)" \
      -mod=readonly


FROM alpine:3.24.2@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6

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

FROM golang:1.26-alpine@sha256:27f829349da645e287cb195a9921c106fc224eeebbdc33aeb0f4fca2382befa6 as builder

ADD . /go/src/github.com/Luzifer/nginx-sso
WORKDIR /go/src/github.com/Luzifer/nginx-sso

ENV CGO_ENABLED=0

RUN set -ex \
 && apk add --no-cache \
      git \
 && go install \
      -ldflags "-X main.version=$(git describe --tags || git rev-parse --short HEAD || echo dev)" \
      -mod=readonly


FROM alpine:3.23@sha256:5b10f432ef3da1b8d4c7eb6c487f2f5a8f096bc91145e68878dd4a5019afde11

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

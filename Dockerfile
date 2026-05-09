FROM golang:1.26.3-alpine@sha256:91eda9776261207ea25fd06b5b7fed8d397dd2c0a283e77f2ab6e91bfa71079d as builder

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

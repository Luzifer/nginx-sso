#!/usr/local/bin/dumb-init /bin/bash
set -euo pipefail

# Copy frontend if not available
[ -d /data/frontend ] || cp -r /go/src/github.com/Luzifer/nginx-sso/frontend /data/frontend

[ -e /data/config.yaml ] || {
  cp /go/src/github.com/Luzifer/nginx-sso/config.yaml /data/config.yaml
  echo "An example configuration was copied to /data/config.yaml - You want to edit that one!"
  exit 1
}

echo "Starting nginx-sso"
exec /go/bin/nginx-sso \
  --config /data/config.yaml \
  --frontend-dir /data/frontend \
  $@

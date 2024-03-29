#!/usr/bin/dumb-init /bin/bash
set -euo pipefail

# Copy frontend if not available
[ -f /data/frontend/index.html ] || cp -r /usr/local/share/nginx-sso/frontend /data/

[ -e /data/config.yaml ] || {
  cp /usr/local/share/nginx-sso/config.yaml /data/config.yaml
  echo "An example configuration was copied to /data/config.yaml - You want to edit that one!"
  exit 1
}

echo "Starting nginx-sso"
exec /usr/local/bin/nginx-sso \
  --config /data/config.yaml \
  --frontend-dir /data/frontend \
  $@

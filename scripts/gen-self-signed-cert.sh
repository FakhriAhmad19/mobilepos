#!/usr/bin/env bash
# Generate a self-signed TLS certificate for staging / local HTTPS testing.
# For real production, use Let's Encrypt (certbot) instead — see docs/DEPLOYMENT.md.
#
# Usage: ./scripts/gen-self-signed-cert.sh [common_name]
#        common_name defaults to "localhost".
set -euo pipefail

cd "$(dirname "$0")/.."
CN="${1:-localhost}"
CERT_DIR="docker/nginx/certs"

mkdir -p "$CERT_DIR"

if [ -f "$CERT_DIR/fullchain.pem" ] && [ -f "$CERT_DIR/privkey.pem" ]; then
  echo "• Certificate already exists in $CERT_DIR — leaving it. Delete to regenerate."
  exit 0
fi

openssl req -x509 -nodes -newkey rsa:2048 \
  -days 365 \
  -keyout "$CERT_DIR/privkey.pem" \
  -out "$CERT_DIR/fullchain.pem" \
  -subj "/CN=${CN}" \
  -addext "subjectAltName=DNS:${CN}"

chmod 600 "$CERT_DIR/privkey.pem"
echo "✓ Self-signed cert for '${CN}' written to $CERT_DIR (valid 365 days)."
echo "  Browsers will warn on the self-signed cert — expected for staging."

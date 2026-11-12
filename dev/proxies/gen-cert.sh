#!/bin/sh
# Dev TLS PKI for the reverse-proxy stack (localhost only), written to the
# bind-mounted dev/proxies/certs/ (gitignored):
#   ca.crt                  CA to import in the browser once
#   tls.crt, tls.key        server cert for localhost / vcv.localhost
#   haproxy.pem             same cert + key in the single file HAProxy wants
# Idempotent: keeps existing files so an imported CA keeps working.
set -eu
dir=/certs
[ -f "$dir/tls.crt" ] && exit 0
apk add --no-cache openssl >/dev/null
cd "$dir"

openssl req -x509 -newkey rsa:2048 -nodes -days 825 -subj "/CN=VCV dev CA" \
  -addext "basicConstraints=critical,CA:true" -addext "keyUsage=critical,keyCertSign,cRLSign" \
  -keyout ca.key -out ca.crt 2>/dev/null

openssl req -newkey rsa:2048 -nodes -subj "/CN=localhost" -keyout tls.key -out tls.csr 2>/dev/null
printf 'extendedKeyUsage=serverAuth\nsubjectAltName=DNS:localhost,DNS:vcv.localhost,IP:127.0.0.1\n' > tls.ext
openssl x509 -req -in tls.csr -CA ca.crt -CAkey ca.key -CAcreateserial -days 825 \
  -extfile tls.ext -out tls.crt 2>/dev/null

cat tls.crt tls.key > haproxy.pem

rm -f tls.csr tls.ext ca.srl
chmod 644 ./*

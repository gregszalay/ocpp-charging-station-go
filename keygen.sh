openssl req -newkey rsa:2048 \
  -new -nodes -x509 \
  -days 365 \
  -out client_cert.pem \
  -keyout key.pem \
  -subj "/C=HU/ST=Budapest/L=Budapest/O=chargerevolution/OU=CSO/CN=localhost"
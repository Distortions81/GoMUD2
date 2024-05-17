#!/bin/bash

openssl ecparam -genkey -name prime256v1 -out data/privkey.pem
openssl req -new -x509 -key data/privkey.pem -out data/cert.pem -days 3650
echo "Generated."
echo "If you have a domain name, use letsEncrypt instead!"

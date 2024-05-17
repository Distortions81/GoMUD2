#!/bin/bash

openssl ecparam -genkey -name prime256v1 -out data/key.pem
openssl req -new -x509 -key data/key.pem -out data/cert.pem -days 3650
echo "Generated."
echo "If you have a domain name, use letsEncrypt instead!"

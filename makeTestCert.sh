#!/bin/bash

openssl ecparam -genkey -name prime256v1 -out privkey.pem
openssl req -new -x509 -key privkey.pem -out fullchain.pem -days 3650
echo "Generated."
echo "If you have a domain name, use letsEncrypt instead!"

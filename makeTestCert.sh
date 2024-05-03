#!/bin/bash

openssl ecparam -genkey -name prime256v1 -out data/server.pem
openssl req -new -x509 -key fullchain.pem -out data/server.pem -days 3650
echo "Generated."
echo "If you have a domain name, use letsEncrypt instead!"
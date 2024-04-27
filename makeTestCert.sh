#!/bin/bash

openssl ecparam -genkey -name prime256v1 -out data/server.key
openssl req -new -x509 -key server.key -out data/server.pem -days 3650
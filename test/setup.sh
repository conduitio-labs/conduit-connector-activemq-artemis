#!/bin/bash

# Clean up previous state
rm -rf certs
mkdir certs

cd certs

keytool -genkey -alias broker -keyalg RSA -keystore broker.ks -storepass password -keypass password -dname "CN=localhost, OU=ID, O=Example, L=City, S=State, C=US" -validity 360 -noprompt
keytool -export -alias broker -keystore broker.ks -file broker_cert -storepass password -noprompt
keytool -genkey -alias client -keyalg RSA -keystore client.ks -storepass password -keypass password -dname "CN=client, OU=ID, O=Example, L=City, S=State, C=US" -validity 360 -noprompt
keytool -import -alias broker -keystore client.ts -file broker_cert -storepass password -noprompt


keytool -importkeystore -srckeystore client.ks -destkeystore client.p12 -srcstoretype JKS -deststoretype PKCS12 -srcstorepass password -deststorepass password -noprompt
openssl pkcs12 -in client.p12 -out client_cert.pem -clcerts -nokeys -password pass:password
openssl pkcs12 -in client.p12 -out client_key.pem -nocerts -nodes -password pass:password
keytool -importkeystore -srckeystore broker.ks -destkeystore broker.p12 -srcstoretype JKS -deststoretype PKCS12 -srcstorepass password -deststorepass password -noprompt
openssl pkcs12 -in broker.p12 -out broker.pem -clcerts -nokeys -password pass:password

chmod 644 *

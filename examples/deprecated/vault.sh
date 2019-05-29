#! /bin/bash


if [[ $# -ne 1 ]]; then
    echo "Usage: rgp_creds namespace"
    exit 1
fi

(kubectl --insecure-skip-tls-verify=true -n $1 port-forward svc/vault 8200:8200 > portforward.log)&

sleep 3
vault login -tls-skip-verify  $(kubectl --insecure-skip-tls-verify=true -n $1 get secret vault-init-secret -o go-template='{{.data.root_token}}' | base64 --decode) > /dev/null
vault kv get -tls-skip-verify --format=json secret/auth/private/admin | jq .data.data

kill $!

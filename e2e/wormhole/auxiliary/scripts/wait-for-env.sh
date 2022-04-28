#!/usr/bin/env bash

set -e

retries=10
for i in `seq $retries`; do
    l1_xdomain_messenger=$(curl -s http://localhost:8080/addresses.json | jq -r .OVM_L1CrossDomainMessenger)
    if [ -z "$l1_xdomain_messenger" ]; then
        echo "OVM_L1CrossDomainMessenger not found yet, waiting on deployer..."
        echo "Sleeping 30 sec ($i/$retries)."
        sleep 30
    else
        echo "OVM_L1CrossDomainMessenger found. Good to go."
        exit
    fi
done

echo "Timed out... :/"
echo "Try 'docker logs deployer' -- never able to get OVM_L1CrossDomainMessenger value."

exit 1

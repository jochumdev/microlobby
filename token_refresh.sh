#!/bin/bash
function main() {
    local HEALTH=$(curl -s -H "Content-Type: application/json" $MICROLOBBY/health | jq -r '.message')
    echo "Health: $HEALTH"

    local JSON=$(curl -s -X POST -H "Content-Type: application/json" -d '{"refreshToken": "'$REFRESH_TOKEN'"}' $MICROLOBBY/auth/v1/refresh)
    echo "JSON: $JSON"

    export ACCESS_TOKEN=$(echo $JSON | jq -r '.accessToken')
    export REFRESH_TOKEN=$(echo $JSON | jq -r '.refreshToken')

    echo -e "Access Token: $ACCESS_TOKEN"
}

main
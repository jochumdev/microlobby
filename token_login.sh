#!/bin/bash
function main() {
    export MICROLOBBY=$1
    local HEALTH=$(curl -s -H "Content-Type: application/json" ${MICROLOBBY}/health | jq -r '.message')
    echo "${MICROLOBBY} Health: $HEALTH"

    local JSON=$(curl -s -X POST -H "Content-Type: application/json" -d '{"username": "'$2'", "password": "'$3'"}' ${MICROLOBBY}/auth/v1/login)
    echo "JSON: $JSON"

    export ACCESS_TOKEN=$(echo $JSON | jq -r '.accessToken')
    export REFRESH_TOKEN=$(echo $JSON | jq -r '.refreshToken')

    echo -e "Access Token: $ACCESS_TOKEN"
}

main $1 $2 $3;
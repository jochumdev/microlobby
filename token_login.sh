#!/bin/bash
function main() {
    export MICROLOBBY=$1

    local JSON=$(curl -s -X POST -H "Content-Type: application/json" -d '{"username": "'$2'", "password": "'$3'"}' ${MICROLOBBY}/api/auth/v1/login)
    echo "JSON: $JSON"

    export ACCESS_TOKEN=$(echo $JSON | jq -r '.accessToken')
    export REFRESH_TOKEN=$(echo $JSON | jq -r '.refreshToken')
}

main $1 $2 $3;
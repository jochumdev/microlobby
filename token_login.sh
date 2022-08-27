#!/bin/bash
MICROLOBBY=$1
HEALTH=$(curl -s -H "Content-Type: application/json" ${MICROLOBBY}/health | jq -r '.message')
echo "${MICROLOBBY} Health: $HEALTH"

JSON=$(curl -s -X POST -H "Content-Type: application/json" -d '{"username": "'$2'", "password": "'$3'"}' http://localhost:8080/auth/v1/login)
echo "JSON: $JSON"

ACCESS_TOKEN=$(echo $JSON | jq -r '.accessToken')
REFRESH_TOKEN=$(echo $JSON | jq -r '.refreshToken')

echo -e "Access Token: $ACCESS_TOKEN"
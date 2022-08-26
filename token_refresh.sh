#!/bin/bash
HEALTH=$(curl -s -H "Content-Type: application/json" http://localhost:8080/health | jq -r '.message')
echo "Health: $HEALTH"

JSON=$(curl -s -X POST -H "Content-Type: application/json" -d '{"refreshToken": "'$REFRESH_TOKEN'"}' http://localhost:8080/auth/v1/refresh)
echo "JSON: $JSON"

ACCESS_TOKEN=$(echo $JSON | jq -r '.accessToken')
REFRESH_TOKEN=$(echo $JSON | jq -r '.refreshToken')

echo -e "Access Token: $ACCESS_TOKEN"
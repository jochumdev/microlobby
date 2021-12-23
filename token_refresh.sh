#!/bin/bash
JSON=$(curl -s -X POST -H "Content-Type: application/json" -d '{"refreshToken": "'$REFRESH_TOKEN'"}' http://localhost:8080/auth/v1/refresh)

ACCESS_TOKEN=$(echo $JSON | jq -r '.accessToken')
REFRESH_TOKEN=$(echo $JSON | jq -r '.refreshToken')
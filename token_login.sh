#!/bin/bash
JSON=$(curl -s -X POST -H "Content-Type: application/json" -d '{"username": "'$1'", "password": "'$2'"}' http://localhost:8080/auth/v1/login)

ACCESS_TOKEN=$(echo $JSON | jq -r '.accessToken')
REFRESH_TOKEN=$(echo $JSON | jq -r '.refreshToken')
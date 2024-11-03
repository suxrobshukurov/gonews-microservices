#!/bin/bash

declare -A services=(
  ["APIGateway"]="cd ./APIGateway/cmd/server && go build -o server.exe && ./server.exe"     
  ["Gonews"]="cd ./Gonews/cmd/server && go build -o server.exe && ./server.exe"                 
  ["Comments"]="cd ./Comments/cmd/server && go build -o server.exe && ./server.exe"           
  ["Cenzor"]="cd ./Cenzor/cmd/server && go build -o server.exe && ./server.exe"                 
)

for service in "${!services[@]}"; do
  eval "${services[$service]}" &
done

wait
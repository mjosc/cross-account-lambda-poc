#!/bin/bash

set -e

GOARCH=amd64 GOOS=linux go build -o gurgler-lambda main.go
zip gurgler-lambda.zip gurgler-lambda
aws lambda update-function-code --function-name gurgler --zip-file fileb://gurgler-lambda.zip

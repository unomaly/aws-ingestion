#!/usr/bin/env bash

cd cloudwatch
GOOS=linux GOARCH=amd64 go build -o cloudwatch main.go
zip cloudwatch.zip cloudwatch

cd -

# Here is how to publish the lambda to AWS for testing
#aws lambda update-function-code --region eu-west-1 --function-name alex-lambda-go --zip-file fileb://cloudwatch/cloudwatch.zip

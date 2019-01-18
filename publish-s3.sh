#!/usr/bin/env bash

# Build first
./build.sh

# Push to our public release bucket
aws s3 cp --acl public-read cloudwatch/cloudwatch.zip s3://unomaly/releases/aws/lambdas/


# Push the templates
aws s3 cp --acl public-read cloudwatch/sam-cloudwatch-logs.yml s3://unomaly/releases/aws/sam/

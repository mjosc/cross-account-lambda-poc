#!/bin/bash

aws lambda invoke \
--function-name gurgler \
--invocation-type "RequestResponse" \
--payload '{ "parameterName": "/lambda/gurgler/test", "parameterValue": "abcabc" }' \
response.txt
head response.txt

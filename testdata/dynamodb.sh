#!/bin/bash

set -e

aws dynamodb delete-table \
    --table-name ProductCatalog \
    --no-paginate \
    --endpoint-url http://localhost:18000 || true

aws dynamodb create-table \
    --table-name ProductCatalog \
    --attribute-definitions AttributeName=Id,AttributeType=N \
    --key-schema AttributeName=Id,KeyType=HASH \
    --provisioned-throughput ReadCapacityUnits=1,WriteCapacityUnits=1 \
    --no-paginate \
    --endpoint-url http://localhost:18000

aws dynamodb delete-table \
    --table-name Forum \
    --no-paginate \
    --endpoint-url http://localhost:18000 || true

aws dynamodb create-table \
    --table-name Forum \
    --attribute-definitions AttributeName=Name,AttributeType=S \
    --key-schema AttributeName=Name,KeyType=HASH \
    --provisioned-throughput ReadCapacityUnits=1,WriteCapacityUnits=1 \
    --no-paginate \
    --endpoint-url http://localhost:18000

aws dynamodb delete-table \
    --table-name Thread \
    --no-paginate \
    --endpoint-url http://localhost:18000 || true

aws dynamodb create-table \
    --table-name Thread \
    --attribute-definitions AttributeName=ForumName,AttributeType=S AttributeName=Subject,AttributeType=S \
    --key-schema AttributeName=ForumName,KeyType=HASH AttributeName=Subject,KeyType=RANGE \
    --provisioned-throughput ReadCapacityUnits=1,WriteCapacityUnits=1 \
    --no-paginate \
    --endpoint-url http://localhost:18000

aws dynamodb delete-table \
    --table-name Reply \
    --no-paginate \
    --endpoint-url http://localhost:18000 || true

aws dynamodb create-table \
    --table-name Reply \
    --attribute-definitions AttributeName=Id,AttributeType=S AttributeName=ReplyDateTime,AttributeType=S AttributeName=PostedBy,AttributeType=S \
    --key-schema AttributeName=Id,KeyType=HASH AttributeName=ReplyDateTime,KeyType=RANGE \
    --provisioned-throughput ReadCapacityUnits=1,WriteCapacityUnits=1 \
    --local-secondary-indexes "IndexName=PostedBy-index,KeySchema=[{AttributeName=Id,KeyType=HASH},{AttributeName=PostedBy,KeyType=RANGE}],Projection={ProjectionType=KEYS_ONLY}" \
    --no-paginate \
    --endpoint-url http://localhost:18000

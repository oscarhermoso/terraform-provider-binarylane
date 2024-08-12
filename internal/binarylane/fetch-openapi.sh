#!/bin/bash

OPENAPI_FILE=$(dirname "$0")/openapi.json

# Fetch the latest OpenAPI spec
curl https://api.binarylane.com.au/reference/openapi.json --output $OPENAPI_FILE
# Remove the /v2 prefix from the paths
cat <<<$(jq '.paths |= with_entries(.key |= sub("/v2/"; "/"))' $OPENAPI_FILE) >$OPENAPI_FILE
# Set the base URL with a /v2 suffix
cat <<<$(jq '.servers[0].url = "https://api.binarylane.com.au/v2"' $OPENAPI_FILE) >$OPENAPI_FILE
# Terraform can't hanlde oneOf types, so we need to remove them and add a type: "string"
cat <<<$(jq '.components.schemas.CreateServerRequest.properties.image |= del(.oneOf) + {type:"string"}' $OPENAPI_FILE) >$OPENAPI_FILE
# Remove the "/paths/{image_id}" path because its duplicated by "/images/{image_id_or_slug}"
cat <<<$(jq 'del(.paths."/images/{image_id}")' $OPENAPI_FILE) >$OPENAPI_FILE

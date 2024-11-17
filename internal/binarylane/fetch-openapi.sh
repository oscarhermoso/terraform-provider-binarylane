#!/bin/bash

OPENAPI_FILE=$(dirname "$0")/openapi.json

# Fetch the latest OpenAPI spec
curl https://api.binarylane.com.au/reference/openapi.json --output $OPENAPI_FILE

# Move the /v2 prefix to the base URL and remove it from the paths
cat <<<$(jq '.servers[0].url = "https://api.binarylane.com.au/v2"' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.paths |= with_entries(.key |= sub("/v2/"; "/"))' $OPENAPI_FILE) >$OPENAPI_FILE

# Terraform can't handle oneOf types, so we need to replace them with basic types
cat <<<$(jq '.components.schemas.CreateServerRequest.properties.image |= del(.oneOf) + {type:"string"}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.CreateServerRequest.properties.ssh_keys.items |= del(.oneOf) + {type:"integer"}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.Rebuild.properties.image |= del(.oneOf) + {type:"string"}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.ImageOptions.properties.ssh_keys.items |= del(.oneOf) + {type:"integer"}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.ChangeImage.properties.image |= del(.oneOf) + {type:"string"}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.paths["/account/keys/{key_id}"].delete.parameters[0].schema |= del(.oneOf) + {type:"integer"}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.paths["/account/keys/{key_id}"].put.parameters[0].schema |= del(.oneOf) + {type:"integer"}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.paths["/account/keys/{key_id}"].get.parameters[0].schema |= del(.oneOf) + {type:"integer"}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.paths["/account/keys/{key_id}"].delete.parameters[0].schema |= del(.oneOf) + {type:"integer"}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.paths["/account/keys/{key_id}"].delete.parameters[0].schema |= del(.oneOf) + {type:"integer"}' $OPENAPI_FILE) >$OPENAPI_FILE

# Remove the "/paths/{image_id}" path because its duplicated by "/images/{image_id_or_slug}"
cat <<<$(jq 'del(.paths."/images/{image_id}")' $OPENAPI_FILE) >$OPENAPI_FILE

# Add x-oapi-codegen-extra-tags so structs can be reflected

## RouteEntryRequest
cat <<<$(jq '.components.schemas.RouteEntryRequest.properties.destination += {"x-oapi-codegen-extra-tags": {"tfsdk": "destination"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.RouteEntryRequest.properties.description += {"x-oapi-codegen-extra-tags": {"tfsdk": "description"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.RouteEntryRequest.properties.router += {"x-oapi-codegen-extra-tags": {"tfsdk": "router"}}' $OPENAPI_FILE) >$OPENAPI_FILE

## AdvancedFirewallRule
cat <<<$(jq '.components.schemas.AdvancedFirewallRule.properties.description += {"x-oapi-codegen-extra-tags": {"tfsdk": "description"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.AdvancedFirewallRule.properties.protocol += {"x-oapi-codegen-extra-tags": {"tfsdk": "protocol"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.AdvancedFirewallRule.properties.action += {"x-oapi-codegen-extra-tags": {"tfsdk": "action"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.AdvancedFirewallRule.properties.destination_addresses += {"x-oapi-codegen-extra-tags": {"tfsdk": "destination_addresses"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.AdvancedFirewallRule.properties.destination_ports += {"x-oapi-codegen-extra-tags": {"tfsdk": "destination_ports"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.AdvancedFirewallRule.properties.source_addresses += {"x-oapi-codegen-extra-tags": {"tfsdk": "source_addresses"}}' $OPENAPI_FILE) >$OPENAPI_FILE

## Load Balancer
cat <<<$(jq '.components.schemas.CreateLoadBalancerRequest.properties.forwarding_rules += {"x-oapi-codegen-extra-tags": {"tfsdk": "forwarding_rules"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.ForwardingRule.properties.entry_protocol += {"x-oapi-codegen-extra-tags": {"tfsdk": "entry_protocol"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.LoadBalancer.properties.health_check += {"x-oapi-codegen-extra-tags": {"tfsdk": "health_check"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.HealthCheckProtocol |= del(.enum)' $OPENAPI_FILE) >$OPENAPI_FILE

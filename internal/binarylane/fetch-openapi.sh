#!/bin/bash

OPENAPI_FILE=$(dirname "$0")/openapi.json

# Fetch the latest OpenAPI spec
curl https://api.binarylane.com.au/reference/openapi.json --output $OPENAPI_FILE

# Move the /v2 prefix to the base URL and remove it from the paths
cat <<<$(jq '.servers[0].url = "https://api.binarylane.com.au/v2"' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.paths |= with_entries(.key |= sub("/v2/"; "/"))' $OPENAPI_FILE) >$OPENAPI_FILE

# Terraform can't handle oneOf/allOf types, so we need to replace them with basic types
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
cat <<<$(jq '.paths["/images"].get.parameters[0].schema |= del(.allOf) + {type:"string"}' $OPENAPI_FILE) >$OPENAPI_FILE

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

## Images
cat <<<$(jq '.components.schemas.Image.properties.backup_info += {"x-oapi-codegen-extra-tags": {"tfsdk": "backup_info"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.Image.properties.created_at += {"x-oapi-codegen-extra-tags": {"tfsdk": "created_at"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.Image.properties.description += {"x-oapi-codegen-extra-tags": {"tfsdk": "description"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.Image.properties.distribution += {"x-oapi-codegen-extra-tags": {"tfsdk": "distribution"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.Image.properties.distribution_info += {"x-oapi-codegen-extra-tags": {"tfsdk": "distribution_info"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.Image.properties.distribution_surcharges += {"x-oapi-codegen-extra-tags": {"tfsdk": "distribution_surcharges"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.Image.properties.error_message += {"x-oapi-codegen-extra-tags": {"tfsdk": "error_message"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.Image.properties.full_name += {"x-oapi-codegen-extra-tags": {"tfsdk": "full_name"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.Image.properties.id += {"x-oapi-codegen-extra-tags": {"tfsdk": "id"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.Image.properties.min_disk_size += {"x-oapi-codegen-extra-tags": {"tfsdk": "min_disk_size"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.Image.properties.min_memory_megabytes += {"x-oapi-codegen-extra-tags": {"tfsdk": "min_memory_megabytes"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.Image.properties.name += {"x-oapi-codegen-extra-tags": {"tfsdk": "name"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.Image.properties.public += {"x-oapi-codegen-extra-tags": {"tfsdk": "public"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.Image.properties.regions += {"x-oapi-codegen-extra-tags": {"tfsdk": "regions"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.Image.properties.size_gigabytes += {"x-oapi-codegen-extra-tags": {"tfsdk": "size_gigabytes"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.Image.properties.slug += {"x-oapi-codegen-extra-tags": {"tfsdk": "slug"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.Image.properties.status += {"x-oapi-codegen-extra-tags": {"tfsdk": "status"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.Image.properties.type += {"x-oapi-codegen-extra-tags": {"tfsdk": "type"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.DistributionInfo.properties.features += {"x-oapi-codegen-extra-tags": {"tfsdk": "features"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.DistributionInfo.properties.remote_access_user += {"x-oapi-codegen-extra-tags": {"tfsdk": "remote_access_user"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.DistributionInfo.properties.password_recovery += {"x-oapi-codegen-extra-tags": {"tfsdk": "password_recovery"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.DistributionInfo.properties.image_id += {"x-oapi-codegen-extra-tags": {"tfsdk": "image_id"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.DistributionSurcharges.properties.surcharge_base_cost += {"x-oapi-codegen-extra-tags": {"tfsdk": "surcharge_base_cost"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.DistributionSurcharges.properties.surcharge_per_memory_megabyte += {"x-oapi-codegen-extra-tags": {"tfsdk": "surcharge_per_memory_megabyte"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.DistributionSurcharges.properties.surcharge_per_memory_max_megabytes += {"x-oapi-codegen-extra-tags": {"tfsdk": "surcharge_per_memory_max_megabytes"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.DistributionSurcharges.properties.surcharge_per_vcpu += {"x-oapi-codegen-extra-tags": {"tfsdk": "surcharge_per_vcpu"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.DistributionSurcharges.properties.surcharge_min_vcpu += {"x-oapi-codegen-extra-tags": {"tfsdk": "surcharge_min_vcpu"}}' $OPENAPI_FILE) >$OPENAPI_FILE

# Edit description here because it's hard to override nested schema properties
cat <<<$(jq '.components.schemas.ForwardingRule.properties.entry_protocol.description = "The protocol that traffic must match for the load balancer to forward it. Valid values are \"http\" and \"https\"."' $OPENAPI_FILE) >$OPENAPI_FILE

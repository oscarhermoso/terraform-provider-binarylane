#!/bin/bash

OPENAPI_FILE=$(dirname "$0")/openapi.json

# Fetch the latest OpenAPI spec
curl https://api.binarylane.com.au/reference/openapi.json --output $OPENAPI_FILE

# Move the /v2 prefix to the base URL and remove it from the paths
cat <<<$(jq '.servers[0].url = "https://api.binarylane.com.au/v2"' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.paths |= with_entries(.key |= sub("/v2/"; "/"))' $OPENAPI_FILE) >$OPENAPI_FILE

# Remove all unhealthy responses, as they are not useful for Terraform
cat <<<$(jq 'walk(if type == "object" and has("responses") then .responses |= with_entries(select(.key | tonumber <= 299)) else . end)' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq 'del(.components.schemas.ProblemDetails)' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq 'del(.components.schemas.ValidationProblemDetails)' $OPENAPI_FILE) >$OPENAPI_FILE

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
cat <<<$(jq '.paths["/sizes"].get.parameters[1].schema |= del(.oneOf) + {type:"string"}' $OPENAPI_FILE) >$OPENAPI_FILE

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

## Regions
cat <<<$(jq '.components.schemas.Region.properties.available += {"x-oapi-codegen-extra-tags": {"tfsdk": "available"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.Region.properties.features += {"x-oapi-codegen-extra-tags": {"tfsdk": "features"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.Region.properties.name += {"x-oapi-codegen-extra-tags": {"tfsdk": "name"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.Region.properties.name_servers += {"x-oapi-codegen-extra-tags": {"tfsdk": "name_servers"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.Region.properties.sizes += {"x-oapi-codegen-extra-tags": {"tfsdk": "sizes"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.Region.properties.slug += {"x-oapi-codegen-extra-tags": {"tfsdk": "slug"}}' $OPENAPI_FILE) >$OPENAPI_FILE

## Sizes
cat <<<$(jq '.components.schemas.Size.properties.slug += {"x-oapi-codegen-extra-tags": {"tfsdk": "slug"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.Size.properties.description += {"x-oapi-codegen-extra-tags": {"tfsdk": "description"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.Size.properties.cpu_description += {"x-oapi-codegen-extra-tags": {"tfsdk": "cpu_description"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.Size.properties.storage_description += {"x-oapi-codegen-extra-tags": {"tfsdk": "storage_description"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.Size.properties.size_type += {"x-oapi-codegen-extra-tags": {"tfsdk": "size_type"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.Size.properties.available += {"x-oapi-codegen-extra-tags": {"tfsdk": "available"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.Size.properties.regions += {"x-oapi-codegen-extra-tags": {"tfsdk": "regions"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.Size.properties.regions_out_of_stock += {"x-oapi-codegen-extra-tags": {"tfsdk": "regions_out_of_stock"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.Size.properties.price_monthly += {"x-oapi-codegen-extra-tags": {"tfsdk": "price_monthly"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.Size.properties.price_hourly += {"x-oapi-codegen-extra-tags": {"tfsdk": "price_hourly"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.Size.properties.disk += {"x-oapi-codegen-extra-tags": {"tfsdk": "disk"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.Size.properties.memory += {"x-oapi-codegen-extra-tags": {"tfsdk": "memory"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.Size.properties.transfer += {"x-oapi-codegen-extra-tags": {"tfsdk": "transfer"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.Size.properties.excess_transfer_cost_per_gigabyte += {"x-oapi-codegen-extra-tags": {"tfsdk": "excess_transfer_cost_per_gigabyte"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.Size.properties.vcpus += {"x-oapi-codegen-extra-tags": {"tfsdk": "vcpus"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.Size.properties.vcpu_units += {"x-oapi-codegen-extra-tags": {"tfsdk": "vcpu_units"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.Size.properties.options += {"x-oapi-codegen-extra-tags": {"tfsdk": "options"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.SizeOptions.properties.disk_min += {"x-oapi-codegen-extra-tags": {"tfsdk": "disk_min"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.SizeOptions.properties.disk_max += {"x-oapi-codegen-extra-tags": {"tfsdk": "disk_max"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.SizeOptions.properties.disk_cost_per_additional_gigabyte += {"x-oapi-codegen-extra-tags": {"tfsdk": "disk_cost_per_additional_gigabyte"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.SizeOptions.properties.restricted_disk_values += {"x-oapi-codegen-extra-tags": {"tfsdk": "restricted_disk_values"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.SizeOptions.properties.memory_max += {"x-oapi-codegen-extra-tags": {"tfsdk": "memory_max"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.SizeOptions.properties.memory_cost_per_additional_megabyte += {"x-oapi-codegen-extra-tags": {"tfsdk": "memory_cost_per_additional_megabyte"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.SizeOptions.properties.transfer_max += {"x-oapi-codegen-extra-tags": {"tfsdk": "transfer_max"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.SizeOptions.properties.transfer_cost_per_additional_gigabyte += {"x-oapi-codegen-extra-tags": {"tfsdk": "transfer_cost_per_additional_gigabyte"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.SizeOptions.properties.ipv4_addresses_max += {"x-oapi-codegen-extra-tags": {"tfsdk": "ipv4_addresses_max"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.SizeOptions.properties.ipv4_addresses_cost_per_address += {"x-oapi-codegen-extra-tags": {"tfsdk": "ipv4_addresses_cost_per_address"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.SizeOptions.properties.discount_for_no_public_ipv4 += {"x-oapi-codegen-extra-tags": {"tfsdk": "discount_for_no_public_ipv4"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.SizeOptions.properties.daily_backups += {"x-oapi-codegen-extra-tags": {"tfsdk": "daily_backups"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.SizeOptions.properties.weekly_backups += {"x-oapi-codegen-extra-tags": {"tfsdk": "weekly_backups"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.SizeOptions.properties.monthly_backups += {"x-oapi-codegen-extra-tags": {"tfsdk": "monthly_backups"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.SizeOptions.properties.backups_cost_per_backup_per_gigabyte += {"x-oapi-codegen-extra-tags": {"tfsdk": "backups_cost_per_backup_per_gigabyte"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.SizeOptions.properties.offsite_backups_cost_per_gigabyte += {"x-oapi-codegen-extra-tags": {"tfsdk": "offsite_backups_cost_per_gigabyte"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.SizeOptions.properties.offsite_backup_frequency_cost += {"x-oapi-codegen-extra-tags": {"tfsdk": "offsite_backup_frequency_cost"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.OffsiteBackupFrequencyCost.properties.daily_per_gigabyte += {"x-oapi-codegen-extra-tags": {"tfsdk": "daily_per_gigabyte"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.OffsiteBackupFrequencyCost.properties.weekly_per_gigabyte += {"x-oapi-codegen-extra-tags": {"tfsdk": "weekly_per_gigabyte"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.OffsiteBackupFrequencyCost.properties.monthly_per_gigabyte += {"x-oapi-codegen-extra-tags": {"tfsdk": "monthly_per_gigabyte"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.SizeType.properties.slug += {"x-oapi-codegen-extra-tags": {"tfsdk": "slug"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.SizeType.properties.name += {"x-oapi-codegen-extra-tags": {"tfsdk": "name"}}' $OPENAPI_FILE) >$OPENAPI_FILE
cat <<<$(jq '.components.schemas.SizeType.properties.description += {"x-oapi-codegen-extra-tags": {"tfsdk": "description"}}' $OPENAPI_FILE) >$OPENAPI_FILE

# Edit description here because it's hard to override nested schema properties
cat <<<$(jq '.components.schemas.ForwardingRule.properties.entry_protocol.description = "The protocol that traffic must match for the load balancer to forward it. Valid values are \"http\" and \"https\"."' $OPENAPI_FILE) >$OPENAPI_FILE

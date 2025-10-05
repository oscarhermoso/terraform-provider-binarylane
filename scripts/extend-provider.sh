#!/bin/bash

GENERATOR_CONFIG=$(dirname "$0")/../provider_code_spec.json

# Add JSON schema
cat <<<$(jq --tab '. + {"$schema": "https://raw.githubusercontent.com/hashicorp/terraform-plugin-codegen-spec/main/spec/v0.1/schema.json"}' $GENERATOR_CONFIG) >$GENERATOR_CONFIG

# Server configuration
ADVANCED_FEATURES_CONFIG=$(dirname "$0")/data/server_advanced_features.json
cat <<<$(jq --tab --slurpfile adv_feat_cfg $ADVANCED_FEATURES_CONFIG '.resources[1].schema.attributes |= . + $adv_feat_cfg' $GENERATOR_CONFIG) >$GENERATOR_CONFIG

# Use set instead of list for server_ids in load_balancer resource
jq --tab '
  .resources |= map(
    if .name == "load_balancer" then
      .schema.attributes |= map(
        if .name == "server_ids" then
          with_entries(.key |= gsub("list"; "set"))
        else .
        end
      )
    else .
    end
  )
' "$GENERATOR_CONFIG" >"$GENERATOR_CONFIG.tmp" && mv "$GENERATOR_CONFIG.tmp" "$GENERATOR_CONFIG"

#!/bin/bash

GENERATOR_CONFIG=$(dirname "$0")/../provider_code_spec.json

# Server configuration
ADVANCED_FEATURES_CONFIG=$(dirname "$0")/data/server_advanced_features.json
cat <<<$(jq --tab --slurpfile adv_feat_cfg $ADVANCED_FEATURES_CONFIG '.resources[1].schema.attributes |= . + $adv_feat_cfg' $GENERATOR_CONFIG) >$GENERATOR_CONFIG

DISKS_CONFIG=$(dirname "$0")/data/server_disks.json
cat <<<$(jq --tab --slurpfile disks_cfg $DISKS_CONFIG '.resources[1].schema.attributes |= . + $disks_cfg' $GENERATOR_CONFIG) >$GENERATOR_CONFIG

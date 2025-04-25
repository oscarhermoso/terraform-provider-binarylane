#!/bin/bash

GENERATOR_CONFIG=$(dirname "$0")/../provider_code_spec.json

# Advanced features configuration
ADVANCED_FEATURES_CONFIG=$(dirname "$0")/data/advanced_features.json
cat <<<$(jq --tab --slurpfile adv_feat_cfg $ADVANCED_FEATURES_CONFIG '.resources[1].schema.attributes |= . + $adv_feat_cfg' $GENERATOR_CONFIG) >$GENERATOR_CONFIG

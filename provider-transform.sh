#!/bin/bash

# The current provider generator will generate duplicate `OptionsType`s, so we
# need to fix that by replacing `"name": "options"` with `"name": "options1"`,
# `"name": "options2"`, etc. in the provider_code_spec.json file.

sed -i '0,/"name": "options"/ s/"name": "options"/"name": "options1"/' provider_code_spec.json
sed -i '0,/"name": "options"/ s/"name": "options"/"name": "options2"/' provider_code_spec.json
sed -i '0,/"name": "options"/ s/"name": "options"/"name": "options3"/' provider_code_spec.json

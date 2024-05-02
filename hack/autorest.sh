#!/bin/bash

# This script is to be run via go generate.
# It is run in the context of the folder getting generated
# Please ensure you downloaded the MSI data plane swagger file and added it to directory /hack/swagger
# MSI data plane swagger may be found here: https://msazure.visualstudio.com/One/_git/ManagedIdentity-MIRP?path=/src/Product/MSI/swagger/CredentialsDataPlane&version=GBmaster

API_VERSION="2024-01-01"

die() {
  echo "$1"
  exit 1
}

node --version || die "node must be installed"
autorest --version || die "autorest must be installed"

script_dir=$(dirname "$(readlink -f "$0")")
swagger_dir="$script_dir/swagger"

json_file="$swagger_dir/msi-credentials-data-plane-$API_VERSION.json"

if [ ! -f "$json_file" ]; then
  die "File $json_file not found"
fi

cat autorest.md
rm zz_generated_*
autorest autorest.md --input-file="$json_file" --output-folder=.

echo "\trunning go fmt..."
go fmt .
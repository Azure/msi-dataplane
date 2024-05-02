#!/bin/bash

# This script is used to generate mock files for the given interfaces.
# It uses the mockgen tool from the uber-go/mock package to generate the mocks.
# The generated mocks are placed in the mocks directory.

die() {
  echo "$1"
  exit 1
}

mockgen --version || die "mockgen must installed: go install go.uber.org/mock/mockgen@latest"
mockgen -destination="$1" -package=mock -source="$2"
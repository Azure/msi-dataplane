#!/bin/bash

# This script is used to run the tests for the project.
# Script should be executed in the root of the project.

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
VCR_ENV_FILE="$SCRIPT_DIR/vcr/vcr-env"

# Tags for various tests
INTEGRATION_TAG="integration"
UNIT_TAG="unit"

print_usage() {
  echo "Usage: $0 [-u] [-i] [-r]"
  echo "  -u  Run unit tests"
  echo "  -i  Run integration tests"
  echo "  -r  Run integration tests in record mode"
}

run_integration_tests() {
  echo "Running integration tests..."
  go test ./... --tags "$INTEGRATION_TAG"
}

run_integration_tests_record() {
  echo "Running integration tests in record mode..."
  if [ ! -f "$VCR_ENV_FILE" ]; then
    echo "VCR environment file not found. Please create the VCR environment first."
    exit 1
  fi
  env $(cat $VCR_ENV_FILE | xargs) go test ./... --tags "$INTEGRATION_TAG"
}

run_unit_tests() {
  echo "Running unit tests..."
  go test ./... --tags "$UNIT_TAG"
}

# Begin script execution
while getopts ":auir" opt; do
  case ${opt} in
    a )
      run_unit_tests
      run_integration_tests
      ;;
    u )
      run_unit_tests
      ;;
    i )
      run_integration_tests
      ;;
    r )
      run_integration_tests_record
      ;;
    \? )
      print_usage
      exit 1
      ;;
  esac
done
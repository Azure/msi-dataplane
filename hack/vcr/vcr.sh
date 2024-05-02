#!/bin/bash

set -euo pipefail

ENV_FILE="vcr-env"
KV_PREFIX="vcr-kv"
RESOURCE_GROUP="vcr-rg"
SP_NAME="vcr-sp"

azure_login() {
    local tenant_id=$1

    if ! az account show > /dev/null; then
        az login --tenant "$tenant_id"
    fi
}

check_params() {
    if [ -z "$1" ]; then
        echo "Tenant ID must be provided"
        print_usage
        exit 1
    fi
}

create_kv() {
    az group create --name $RESOURCE_GROUP --location eastus

    random_string=$(openssl rand -base64 6 | tr -dc 'a-zA-Z0-9' | fold -w 6 | head -n 1)
    kv_name="$KV_PREFIX-$random_string"
    az keyvault create --name $kv_name --resource-group $RESOURCE_GROUP --location eastus --enable-rbac-authorization
    
    client_id=$(az ad sp list --display-name $SP_NAME | jq -r '.[0].appId')

    # Get the Key Vault's resource ID
    kv_id=$(az keyvault show --name $kv_name --query id --output tsv)

    # Assign the 'Key Vault Secrets Officer' role to the service principal
    az role assignment create --assignee $client_id --role "Key Vault Secrets Officer" --scope $kv_id

    # Export the key vault URL to the environment file
    kv_url=$(az keyvault show --name $kv_name --query properties.vaultUri --output tsv)
    echo "KEYVAULT_URL=$kv_url" >> $ENV_FILE
}

create_sp() {
    sp_json=$(az ad sp create-for-rbac --name $SP_NAME)

    # Export the service principal details to the environment file
    client_id=$(echo $sp_json | jq -r '.appId')
    echo "AZURE_CLIENT_ID=$client_id" >> $ENV_FILE

    client_secret=$(echo $sp_json | jq -r '.password')
    echo "AZURE_CLIENT_SECRET=$client_secret" >> $ENV_FILE
}

create_vcr_env() {
    echo "Creating VCR environment file..."

    # Delete the environment file if it already exists
    rm -f $ENV_FILE

    # Export record mode to the environment file
    echo "RECORD_MODE=record" >> $ENV_FILE
    
    # Export the tenant ID to the environment file
    tenant_id=$1
    echo "AZURE_TENANT_ID=$tenant_id" >> $ENV_FILE

    azure_login $tenant_id
    create_sp
    create_kv
}

delete_rg() {
    az group delete --name $RESOURCE_GROUP --yes
    echo "Resource group '$RESOURCE_GROUP' deleted"
}

delete_sp() {
    client_id=$(az ad sp list --display-name $SP_NAME | jq -r '.[0].appId')
    az ad sp delete --id $client_id
    az ad app delete --id $client_id

    echo "Service principal '$SP_NAME' deleted"
}

delete_vcr_env() {
    echo "Deleting VCR environment file and cleaning up resources..."

    # Delete the service principal and key vault
    tenant_id=$1
    azure_login $tenant_id

    delete_rg
    delete_sp

    # Delete the environment file
    rm -f $ENV_FILE
    echo "Deleted VCR environment file"
}

print_usage() {
  echo "Usage: $0 [-c tenant_id] [-d tenant_id]"
  echo "  -c tenant_id  Create VCR environment with the provided tenant ID"
  echo "  -d tenant_id  Delete VCR environment with the provided tenant ID"
}

# Begin script execution
while getopts "c:d:" opt; do
  case ${opt} in
    c )
      check_params $OPTARG
      create_vcr_env $OPTARG
      ;;
    d )
      check_params $OPTARG
      delete_vcr_env $OPTARG
      ;;
    \? )
      print_usage
      exit 1
      ;;
  esac
done
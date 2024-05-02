This directory contains scripts for Go-VCR testing.

# How to Use

In this directory, run `vcr.sh` with flag `-c *tenant-id*`.
This script will output a file `vcr-env`, which contains the environment variables 
needed to run VCR tests in record mode.

When you're done with VCR recording, run `vcr.sh` with flag `-d *tenant-id*` to cleanup 
resources created for VCR recording.

# FAQ

## Any restrictions on which tenant I can use?
Don't use the MSIT tenant - there are policies that restrict service principal client ID/secret
usage. 
#!/usr/bin/env bash

set +e

aws configure set aws_access_key_id $bot_aws_access_id
aws configure set aws_secret_access_key $bot_aws_secret_access_key
aws configure set region $Region

set -e
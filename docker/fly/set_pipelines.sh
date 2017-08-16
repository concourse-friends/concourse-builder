#!/usr/bin/env bash

set -ex

fly --target trgt login --insecure --concourse-url $CONCOURSE_URL --username $CONCOURSE_USER --password $CONCOURSE_PASSWORD

cd $PIPELINES

for yml in *
do
    name=$(echo $yml | cut -f 1 -d '.')
    fly -t trgt set-pipeline --non-interactive --pipeline=$name --config=$yml
done
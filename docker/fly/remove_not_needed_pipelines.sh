#!/usr/bin/env bash

set -ex

fly --target trgt login --insecure --concourse-url $CONCOURSE_URL --username $CONCOURSE_USER --password $CONCOURSE_PASSWORD


branches=$(cat $BRANCHES_DIR/branches)

cd $PIPELINES

function destroy_pipeline {
    fly -t trgt destroy-pipeline --non-interactive --pipeline=$1
}

for yml in *
do
    name=$(echo $yml | cut -f 1 -d '.')

    if [ -z "$PIPELINE_REGEX"  ]; then
        for branch in $branches; do
            if [[ $branch == *"$name"* ]]; then
                destroy_pipeline $name
            fi
        done
    else
        if [[ $name =~ $PIPELINE_REGEX ]]; then
            destroy_pipeline $name
        fi
    fi
done

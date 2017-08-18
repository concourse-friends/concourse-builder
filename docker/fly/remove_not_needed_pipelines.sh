#!/usr/bin/env bash

set -ex

fly --target trgt login --insecure --concourse-url $CONCOURSE_URL --username $CONCOURSE_USER --password $CONCOURSE_PASSWORD


#BRANCHES=$(cat $BRANCHES_DIR/branches)

cd $PIPELINES

EXIST_PIPELINES=$(fly --target trgt pipelines | awk '{ print $1 }' | sort)
PIPELINE_FILES=$(for yml in *; do name=$(echo $yml | cut -f 1 -d '.');done)

if [ ! -z "$PIPELINE_REGEX" ]; then
    for pipeline in $EXIST_PIPELINES
    do
        if [[ $pipeline =~ $PIPELINE_REGEX ]]; then
            if ! echo "$PIPELINE_FILES" | grep "$pipeline"; then
                fly -t trgt destroy-pipeline --pipeline=$pipeline --non-interactive
            fi
        fi
    done
fi
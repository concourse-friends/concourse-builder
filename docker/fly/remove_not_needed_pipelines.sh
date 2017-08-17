#!/usr/bin/env bash

set -ex

fly --target trgt login --insecure --concourse-url $CONCOURSE_URL --username $CONCOURSE_USER --password $CONCOURSE_PASSWORD


branches=$(cat $BRANCHES_DIR/branches)

cd $PIPELINES

for yml in *
do
    name=$(echo $yml | cut -f 1 -d '.')

    if [ -z "$PIPELINE_REGEX"  ]; then
        name="$PIPELINE_REGEX"
    fi

    if echo $branches | grep -w name; then
        fly -t trgt destroy-pipeline --non-interactive --pipeline=$name
    fi
done

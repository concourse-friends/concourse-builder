#!/usr/bin/env bash

set -ex

. /bin/fly/authenticate.sh

cd $PIPELINES

for yml in *
do
    name=$(echo $yml | cut -f 1 -d '.')
    fly -t trgt set-pipeline --non-interactive --pipeline=$name --config=$yml
done
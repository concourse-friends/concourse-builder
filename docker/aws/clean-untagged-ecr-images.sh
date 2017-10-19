#!/usr/bin/env bash

set -ex

. /bin/aws/configure.sh

repositories=$(aws ecr describe-repositories --output text | awk '{print $5}')
while IFS= read -r line
do
IMAGES=$(aws ecr list-images --repository-name $line --filter tagStatus=UNTAGGED --query 'imageIds[?type(imageTag)!=`string`].[imageDigest]' --output text)
echo "Deleting all images with no tag from $line..."
for DIGEST in ${IMAGES[*]}
do
aws ecr batch-delete-image  --repository-name $line --image-ids imageDigest=$DIGEST
done
done <<< "$repositories"
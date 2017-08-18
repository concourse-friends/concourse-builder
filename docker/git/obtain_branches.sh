#!/usr/bin/env bash
BUILD_DIR=`pwd`

set -ex

if [ -z "$GIT_REPO_DIR"  ]
then
  echo "Please specify GIT_REPO_DIR env variable"
  exit 1
fi

if [ -z "$OUTPUT_DIR"  ]
then
  echo "Please specify OUTPUT_DIR env variable"
  exit 1
fi

mkdir -p $BUILD_DIR/$OUTPUT_DIR

cd $BUILD_DIR/$GIT_REPO_DIR

{
  git branch -r | \
  grep -v ">" | \
  while read rbranch
     do echo $rbranch | rev | cut -d/ -f1 | rev
  done
 } >  $BUILD_DIR/$OUTPUT_DIR/branches

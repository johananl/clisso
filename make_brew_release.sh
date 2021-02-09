#!/usr/bin/env bash
# expects the following variables
# $1 BINARY_NAME
# $2 VERSION


# See https://gist.github.com/maelvalais/068af21911c7debc4655cdaa41bbf092 for a rough guide on CI/CD for Brew.
# This script runs the following process:
# 1. Tap the repo in case it's not yet tapped.
# 2. Disregard all changes and switch to master.
# 3. Create a new formula from the template with the correct version.
# 4. Get the correct SHA256 sum for the version and update the formula.
# 5. Build the bottle for the current operating system.
# 6. Create a final formula with the correct version, SHA256 and bottle info.
# 7. Commit to a new branch and push.

# setup error handling
set -e -o pipefail

# set some variables used below
BINARY_NAME=$1
VERSION=$2
BINARY_DIR=$(pwd)/$3


function cleanup() {
  rm -f ${BINARY_NAME}.rb.bottle*
}

trap cleanup EXIT

# add tap in case it's missing
brew tap allcloud-io/tools
TAP_DIR=$(brew --repo allcloud-io/tools)
# change to tap directory
cd $TAP_DIR
# stash all changes so we have a clean working directory
git clean -d -x -f && git reset --hard && git checkout master

# set the correct version
sed "s:%VERSION%:${VERSION}:" ${BINARY_NAME}.rb.template | sed "s:%BOTTLE%::" > ${BINARY_NAME}.rb
# and calc sha256
SHA256=$(brew fetch ${BINARY_NAME} --build-from-source 2>/dev/null | grep SHA256 | cut -d" " -f2 || true)

# replace version and sha256 placeholder in template
sed "s:%VERSION%:${VERSION}:" ${BINARY_NAME}.rb.template | \
sed "s:%SOURCE_SHA%:${SHA256}:" > ${BINARY_NAME}.rb.bottle

# generate parts to be assembled later
grep -B100 '%BOTTLE%' ${BINARY_NAME}.rb.bottle | grep -v '%BOTTLE%' > ${BINARY_NAME}.rb.bottle.head
grep -A100 '%BOTTLE%' ${BINARY_NAME}.rb.bottle | grep -v '%BOTTLE%' > ${BINARY_NAME}.rb.bottle.tail

# skip the bottle placeholder for now
cat ${BINARY_NAME}.rb.bottle.head ${BINARY_NAME}.rb.bottle.tail > ${BINARY_NAME}.rb

# change back to original workdir
cd $BINARY_DIR
# build the bottle
brew test-bot allcloud-io/tools/${BINARY_NAME}

# create a tempfile
TEMPFILE=$(mktemp)

for json in `ls -1 *bottle.json`; do
  # extract the mac version the bottle was build for
  MAC_VERSION=$(echo $json | cut -d. -f4);
  # extract the sha256 of the bottle
  SHA=$(cat $json | jq ".\"allcloud-io/tools/${BINARY_NAME}\".bottle.tags.$MAC_VERSION.sha256")
  # get the local file name
  LOCAL=$(cat $json | jq -r ".\"allcloud-io/tools/${BINARY_NAME}\".bottle.tags.$MAC_VERSION.local_filename")
  # get the remote filename
  REMOTE=$(cat $json | jq -r ".\"allcloud-io/tools/${BINARY_NAME}\".bottle.tags.$MAC_VERSION.filename")
  # rename to the correct name
  mv $LOCAL $REMOTE
  # append to tempfile
  echo "    sha256 $SHA => :$MAC_VERSION" >> ${TEMPFILE}
  rm $json
done

cd $TAP_DIR

# add all the bottles
cat ${BINARY_NAME}.rb.bottle.head ${TEMPFILE} ${BINARY_NAME}.rb.bottle.tail > ${BINARY_NAME}.rb

# commit to git and push to origin
BRANCHNAME=auto/${BINARY_NAME}-${VERSION}
git checkout -b $BRANCHNAME || git checkout $BRANCHNAME
git add ${BINARY_NAME}.rb && git commit -m "Automatic commit of bottle build for version $VERSION of $BINARY_NAME."
git push origin $BRANCHNAME

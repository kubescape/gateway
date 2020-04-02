#!/usr/bin/env bash
set -ex

export IMAGE_NAME="canotificationserver"
export IMAGE_TAG="test"

./build.sh

passed=$(python3 component_test.py --image $IMAGE_NAME:$IMAGE_TAG| grep "TEST PASSED")

if [ ! -z "$passed" ]; then
    echo "<--------------- COMPENENT Test Tests PASSED ---------------------->"
else
    echo "<--------------- COMPENENT Test Tests FAILED ---------------------->"
fi
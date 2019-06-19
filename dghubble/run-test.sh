#!/bin/bash -ex

for line in $(lpass show Personal\\api_keys/twitter --notes --sync=no); do
    export $line
done

ginkgo $@

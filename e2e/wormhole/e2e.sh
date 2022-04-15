#!/usr/bin/env bash

# Simply put, this script looks for any scripts prefixed by 't_' in the tests/
# folder and runs them.

set -e

gitroot=$(git rev-parse --show-toplevel)
cd $gitroot
wd=$gitroot/e2e/wormhole

for t in $(ls $wd/tests/t_*.sh); do
    echo $(basename "$t")
    source $t 
done

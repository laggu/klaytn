#!/bin/sh

set -e

if [ ! -f "build/env.sh" ]; then
    echo "$0 must be run from the root of the repository."
    exit 2
fi

# Create fake Go workspace if it doesn't exist yet.
workspace="$PWD/build/_workspace"
root="$PWD"
gxpdir="$workspace/src/github.com/ground-x"
if [ ! -L "$gxpdir/klaytn" ]; then
    mkdir -p "$gxpdir"
    cd "$gxpdir"
    ln -s ../../../../../. klaytn
    cd "$root"
fi

# Set up the environment to use the workspace.
GOPATH="$workspace"
export GOPATH

# Run the command inside the workspace.
cd "$gxpdir/klaytn"
PWD="$gxpdir/klaytn"

# Launch the arguments with the configured environment.
exec "$@"

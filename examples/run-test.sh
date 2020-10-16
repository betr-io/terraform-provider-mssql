#!/usr/bin/env bash

set -e

if [ -n "${BASH_VERSION}" ]; then
    SOURCE="${BASH_SOURCE[0]}"
elif [ -n "${ZSH_VERSION}" ]; then
    SOURCE="$0"
else
    exit_with_message "Unknown shell!"
fi
CURRENT="$(cd "$(dirname "$SOURCE")" >/dev/null && pwd)"
while [ -h "$SOURCE" ]; do # resolve $SOURCE until the file is no longer a symlink
    DIR="$(cd -P "$(dirname "$SOURCE")" >/dev/null 2>&1 && pwd)"
    SOURCE="$(readlink "$SOURCE")"
    [[ $SOURCE != /* ]] && SOURCE="$DIR/$SOURCE" # if $SOURCE was a relative symlink, we need to resolve it relative to the path where the symlink file was located
done
DIR="$(cd -P "$(dirname "$SOURCE")" >/dev/null 2>&1 && pwd)"

# Parse arguments
while (( "$#" )); do
    case "$1" in
        -h|--help)
            echo -e "\nUsage: $0 [--skip-build] [terraform cmd/options]\n"
            exit 1
            ;;
        --skip-build)
            skip_build=1
            shift
            ;;
        #-*|--*=)    # Unsupported flags
        #    echo "Error: Unsupported flag $1" >&2
        #    exit 1
        #    ;;
        *)          # Preserve positional arguments
            PARAMS="$PARAMS $1"
            shift
            ;;
    esac
done

eval set -- "$PARAMS"

if [ $# == 0 ]; then
    eval set -- "apply -auto-approve"
fi

if [ -e "$CURRENT/.local.env" ]; then
    source "$CURRENT/.local.env"
fi

if [ ! $skip_build ]; then
    (cd $(dirname $DIR) && make install)
    if [[ $? > 0 ]]; then
        exit 1
    fi

    (cd $CURRENT && terraform init)
fi

(cd $CURRENT && rm -f *.log && terraform "$@")

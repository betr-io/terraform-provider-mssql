#!/usr/bin/env bash

set -Eeuo pipefail

function exit_with_message() {
    echo -e "\n$@\n" >&2
    exit 1
}

if [ -n "${BASH_VERSION}" ]; then
    SOURCE="${BASH_SOURCE[0]}"
elif [ -n "${ZSH_VERSION}" ]; then
    SOURCE="$0"
else
    exit_with_message "Unknown shell!"
fi
while [ -h "$SOURCE" ]; do # resolve $SOURCE until the file is no longer a symlink
    DIR="$( cd -P "$( dirname "$SOURCE" )" >/dev/null 2>&1 && pwd )"
    SOURCE="$(readlink "$SOURCE")"
    [[ $SOURCE != /* ]] && SOURCE="$DIR/$SOURCE" # if $SOURCE was a relative symlink, we need to resolve it relative to the path where the symlink file was located
done
DIR="$( cd -P "$( dirname $( dirname "$SOURCE" ) )" >/dev/null 2>&1 && pwd )"

PARAMS=""
HELP=0
LIST=0
CURRENT=1
KEEP=0
FORCE=0

# Parse arguments
while (( "$#" )); do
    case "$1" in
        -h|--help)
            HELP=1
            shift
            ;;
        -l|--list)
            LIST=1
            shift
            ;;
        -c|--current)
            CURRENT=1
            shift
            ;;
        -n|--no-current)
            CURRENT=0
            shift
            ;;
        -k|--keep)
            KEEP=1
            shift
            ;;
        -f|--force)
            FORCE=1
            shift
            ;;
        -*|--*=)    # Unsupported flags
            exit_with_message "Error: Unsupported flag $1"
            ;;
        *)          # Preserve positional arguments
            PARAMS="$PARAMS $1"
            shift
            ;;
    esac
done

eval set -- "$PARAMS"

[ -e "${DIR}/scripts/version.sh" ] && source "${DIR}/scripts/versions.sh"
if [[ $(id -u) == 0 ]]; then
    DEFAULT_TERRAFORM_DIR=/usr/local/bin
else
    DEFAULT_TERRAFORM_DIR=$HOME/bin
fi
TERRAFORM_DIR=${TERRAFORM_DIR:-$DEFAULT_TERRAFORM_DIR}

echoerr() { cat <<< "$@" 1>&2; }
if [[ $HELP == 1 || $# > 1 ]]; then
    echoerr "
    Usage: $0 [options] [version]

    Options:
        -l --list        List installed terraform versions.
        -d --dir         Install directory (default: $TERRAFORM_DIR)
        -c --current     Set this version as current (default).
        -n --no-current  Do not set this version as current.
        -k --keep        Keep other patch versions.
        -f --force       Force download.
        -h --help        Display this message.
"
    exit 1
fi

if [[ $LIST == 1 || ${#} == 0 ]]; then
    echo -e "\nInstalled terraform versions:\n"
    for f in $(ls ${TERRAFORM_DIR} | grep 'terraform-' | sort -r); do
        if [ $(readlink "${TERRAFORM_DIR}/terraform") == $(basename "$f") ]; then CUR='*'; else CUR=' '; fi
        echo "    $CUR $(echo $f | cut -d- -f 2)"
    done
    echo
    exit
fi


if [[ $# == 1 ]]; then
    TERRAFORM_VERSION=$1
fi

DOWNLOAD=0
if [[ ! -e "${TERRAFORM_DIR}/terraform-${TERRAFORM_VERSION}" ]]; then
    echo -e "\nDownloading terraform v${TERRAFORM_VERSION}"
    DOWNLOAD=1
elif [[ $FORCE == 1 ]]; then
    echo -e "\nForce downloading terraform v${TERRAFORM_VERSION}"
    DOWNLOAD=1
fi

if [[ $DOWNLOAD == 1 ]]; then
    TERRAFORM_ZIP="${TERRAFORM_DIR}/terraform.zip"
    wget -q https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip -O "${TERRAFORM_ZIP}" || (rm "${TERRAFORM_ZIP}" && exit_with_message "Failed to download terraform-${TERRAFORM_VERSION}")
    echo -e "\nInstalling terraform v${TERRAFORM_VERSION}"
    unzip -p ${TERRAFORM_ZIP} terraform > "${TERRAFORM_DIR}/terraform-${TERRAFORM_VERSION}"
    chmod +x "${TERRAFORM_DIR}/terraform-${TERRAFORM_VERSION}"
    rm "${TERRAFORM_ZIP}"
fi

if [[ $CURRENT == 1 ]]; then
    if [[ "$(readlink "${TERRAFORM_DIR}/terraform")" != "terraform-${TERRAFORM_VERSION}" ]]; then
        echo -e "\nSetting current version to v${TERRAFORM_VERSION}"
        ln -sfnr "${TERRAFORM_DIR}/terraform-${TERRAFORM_VERSION}" "${TERRAFORM_DIR}/terraform"
    else
        echo -e "\nCurrent version is already v${TERRAFORM_VERSION}"
    fi
fi

if [[ $KEEP != 1 ]]; then
    VERSION_PREFIX=$(echo $TERRAFORM_VERSION | sed 's|\(.*\)\..*|\1|')
    readarray -t FILES <<< $(echo ${TERRAFORM_DIR}/terraform-${VERSION_PREFIX}.* | xargs basename -a | sort -V)
    # Always keep latest version in patch range
    unset -v FILES[-1]
    # Always keep current version
    for i in "${!FILES[@]}"; do
        if [[ "${FILES[$i]}" = "terraform-${TERRAFORM_VERSION}" ]]; then
            unset -v FILES[$i]
        fi
    done
    if [[ ${#FILES[@]} > 0 ]]; then
        echo -e "\nRemoving old version(s): ${FILES[@]}"
        for f in "${FILES[@]}"; do
            rm -f "${TERRAFORM_DIR}/$f"
        done
    fi
fi


if [[ ":$PATH:" == *":$TERRAFORM_DIR:"* ]]; then
    echo -e "\nYour path is correctly set\n"
else
    echo -e "\nSet yout path to: \n    PATH=$PATH:$TERRAFORM_DIR\n"
    PATH=$PATH:$TERRAFORM_DIR export PATH
fi

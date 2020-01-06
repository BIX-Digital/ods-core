#!/usr/bin/env bash
set -eu

function usage {
   printf "usage: %s [options]\n", $0
   printf "\t-h|--help\tPrints the usage\n"
   printf "\t-v|--verbose\tVerbose output\n"
   printf "\t-b|--ods-ref\tReference to be created in the mocked git repo.\n"

}

urlencode() {
    # urlencode <string>
    
    local length="${#1}"
    for (( i = 0; i < length; i++ )); do
        local c="${1:i:1}"
        case $c in
            [a-zA-Z0-9.~_-]) printf "$c" ;;
            *) printf '%%%02X' "'$c" ;;
        esac
    done
    
   
}

REF=""
NAMESPACE=""


URL=$(oc config view --minify -o jsonpath='{.clusters[*].cluster.server}')
if [ ${URL} != "https://127.0.0.1:8443" ]; then
    echo "You are not in a local cluster. Stopping now!!!"
fi

while [[ "$#" -gt 0 ]]; do case $1 in

   -v|--verbose) set -x;;
   
   -h|--help) usage; exit 0;;

   -b=*|--ods-ref=*) REF="${1#*=}";;
   -b|--ods-ref) REF="$2"; shift;;

   *) echo "Unknown parameter passed: $1"; usage; exit 1;;
 esac; shift; done

if git remote -v | grep mockbucket; then
    git remote remove mockbucket
fi


if [ -z "${REF}" ]; then 
    echo "Reference --ods-ref must be provided"
    exit 1
fi 

source ${BASH_SOURCE%/*}/../../ods-config/ods-core.env

# git checkout -b "${REF}"
HEAD=$(git rev-parse --abbrev-ref HEAD)
if [ "${HEAD}" = "HEAD" ]; then
    HEAD="ci/cd"
    git checkout -b ${HEAD}
fi
git remote add mockbucket http://$(urlencode ${CD_USER_ID}):$(urlencode ${CD_USER_PWD})@${BITBUCKET_HOST}/scm/opendevstack/ods-core.git
git -c http.sslVerify=false push mockbucket --set-upstream "${HEAD}:${REF}"
git remote remove mockbucket

mkdir -p "${HOME}/ods-configuration"
git -C "${HOME}/ods-configuration" init
cp ${BASH_SOURCE%/*}/../../ods-config/ods-core.env ${HOME}/ods-configuration
git -C "${HOME}/ods-configuration" add ods-core.env
git -C "${HOME}/ods-configuration" commit -m "Initial Commit"
git -C "${HOME}/ods-configuration" remote add mockbucket http://$(urlencode ${CD_USER_ID}):$(urlencode ${CD_USER_PWD})@${BITBUCKET_HOST}/scm/opendevstack/ods-configuration.git
git -C "${HOME}/ods-configuration" -c http.sslVerify=false push mockbucket --set-upstream "$(git -C "${HOME}/ods-configuration" rev-parse --abbrev-ref HEAD):${REF}"
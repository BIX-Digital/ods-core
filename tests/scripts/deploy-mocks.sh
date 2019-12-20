#!/usr/bin/env bash

NAMESPACE=mocks

RECREATE="false"
while [[ "$#" -gt 0 ]]; do case $1 in

   -v|--verbose) set -x;;

   -h|--help) usage; exit 0;;

   -n=*|--namespace=*) NAMESPACE="${1#*=}";;
   -n|--namespace) NAMESPACE="$2"; shift;;


   *) echo "Unknown parameter passed: $1"; usage; exit 1;;
 esac; shift; done

URL=$(oc config view --minify -o jsonpath='{.clusters[*].cluster.server}')
if [ ${URL} != "https://127.0.0.1:8443" ]; then
    echo "You are not in a local cluster. Stopping now!!!"
fi

if ! oc project ${NAMESPACE}; then
    echo "Project '${NAMESPACE}' does not exists."    
fi

source ${BASH_SOURCE%/*}/../../ods-config/ods-core.env
docker run -d -p "8080:8080" \
           --env="BASIC_USERNAME=${CD_USER_ID}" \
           --env="BASIC_PASSWORD=${CD_USER_PWD}" \
           --env="REPOS=opendevstack/ods-core.git" \
           --name mockbucket \
           hugowschneider/mockbucket:latest 
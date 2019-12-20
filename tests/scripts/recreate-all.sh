#!/usr/bin/env bash
set -uxe


URL=$(oc config view --minify -o jsonpath='{.clusters[*].cluster.server}')

if [ ${URL} != "https://127.0.0.1:8443" ]; then
    echo "You are not in a local cluster. Stopping now!!!"
fi

if [ ! -d ${BASH_SOURCE%/*}/../../ods-config ]; then
    mkdir -p ${BASH_SOURCE%/*}/../../ods-config
fi
${BASH_SOURCE%/*}/create-env-from-local-cluster.sh --output ${BASH_SOURCE%/*}/../../ods-config/ods-core.env

NAMSPACE="cd"
REF="ci/cd"

if ! oc whoami; then
    echo "You must be logged in to the OC Cluster"
fi

if oc project cd; then
    oc delete project ${NAMSPACE}
fi

if docker ps -a --format "{{.Names}}" | grep mockbucket; then
    docker rm mockbucket --force
fi

if git remote -v | grep mockbucket; then
    git remote remove mockbucket
fi

${BASH_SOURCE%/*}/../../ods-setup/setup-ods-project.sh --verbose --force --namespace ${NAMSPACE}

${BASH_SOURCE%/*}/deploy-mocks.sh --namespace ${NAMSPACE} --verbose
sleep 10 # Waiting for service to boot up
${BASH_SOURCE%/*}/setup-mocked-ods-repo.sh --namespace ${NAMSPACE} --ods-ref ${REF} --verbose

${BASH_SOURCE%/*}/../../ods-setup/setup-ods-infrastructure.sh --force --verbose --ods-ref ${REF}
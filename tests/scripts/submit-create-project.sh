#!/usr/bin/env bash
set -uxe 


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

source ${BASH_SOURCE%/*}/../../ods-config/ods-core.env

REF="ci/cd"
WEBHOOK_PROXY_HOST=$(oc get services -n prov-cd | grep webhook-proxy | awk '{print $3}')
REPO_URL="ods-core"
PROJECT_ID="blaht"
CD_USER_TYPE=general
CD_USER_ID_B64=${CD_USER_ID_B64} 
TRIGGER_SECRET=${PIPELINE_TRIGGER_SECRET}
PIPELINE_TRIGGER_SECRET=${PIPELINE_TRIGGER_SECRET_B64}

JSON="{\"branch\" :  \"${REF}\", \"repository\" :  \"${REPO_URL}\",\"project\" :  \"opendevstack\", \"env\" : [{\"name\":\"PROJECT_ID\",\"value\": \"${PROJECT_ID}\"},{\"name\":\"CD_USER_TYPE\",\"value\":\"general\" }, { \"name\" : \"CD_USER_ID_B64\", \"value\" :  \"${CD_USER_ID_B64}\" }, { \"name\" : \"PIPELINE_TRIGGER_SECRET\", \"value\" :  \"${PIPELINE_TRIGGER_SECRET}\"}, { \"name\" : \"ODS_GIT_REF\", \"value\" :  \"ci/cd\"}] }"

echo "${JSON}"
echo "${#JSON}"

curl -X POST --verbose \
    -H "Content-Type: application/json" \
    -d "$JSON" \
    "http://${WEBHOOK_PROXY_HOST}/build?trigger_secret=${TRIGGER_SECRET}&jenkinsfile_path=create-projects/Jenkinsfile&component=ods-corejob-create-project-${PROJECT_ID}"


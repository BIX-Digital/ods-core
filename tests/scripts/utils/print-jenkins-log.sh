#!/usr/bin/env bash

JENKINS_IP=$(oc get services | grep jenkins | grep 80/TCP | awk '{print $3}')
JENKINS_JNLP_IP=$(oc get services | grep jenkins | grep 50000/TCP | awk '{print $3}')
LOG_URL=$(oc get build ods-corejob-create-project-blaht-ci-cd-1 -o json | jq '.metadata.annotations."openshift.io/jenkins-log-url"' | sed -e 's/"//g' | sed -e 's@https://jenkins-prov-cd.172.30.0.1.nip.io@http://'"${JENKINS_IP}"'@g')
curl ${LOG_URL} --header "Authorization: Bearer $(oc get sa/builder --template='{{range .secrets}}{{ .name }} {{end}}' | xargs -n 1 oc get secret --template='{{ if .data.token }}{{ .data.token }}{{end}}' | head -n 1 | base64 -d -)" -k

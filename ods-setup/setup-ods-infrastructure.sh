#!/usr/bin/env bash

#!/usr/bin/env bash
set -xe

function usage {
   printf "usage: %s [options]\n", $0
   printf "\t--force\tIgnores warnings and error with tailor --force\n"
   printf "\t-h|--help\tPrints the usage\n"
   printf "\t-v|--verbose\tVerbose output\n"
   printf "\t-t|--tailor\tChanges the executable of tailor. Default: tailor\n"
   printf "\t-n|--namespace\tChanges the default namespace. Default: cd\n"

}
TAILOR="tailor"
NAMESPACE="cd"

while [[ "$#" -gt 0 ]]; do case $1 in

   -v|--verbose) set -x;;

   --force) FORCE="--force"; ;;

   -h|--help) usage; exit 0;;

   -t=*|--tailor=*) TAILOR="${1#*=}";;
   -t|--tailor) TAILOR="$2"; shift;;

   -n=*|--namespace=*) NAMESPACE="${1#*=}";;
   -n|--namespace) NAMESPACE="$2"; shift;;

   *) echo "Unknown parameter passed: $1"; usage; exit 1;;
 esac; shift; done

if ! oc whoami; then
  echo "You should be logged to run the script"
  exit 1
fi

if ! oc project ${NAMESPACE}; then
  echo "The project '${NAMESPACE}' does not exists. Please setup the project using 'setup-ods-project.sh'"
  exit 1
fi

echo "Applying Tailorfile to project '${NAMESPACE}'"
cd ${BASE_DIR}/ods-core/jenkins/ocp-config
${TAILOR} update ${FORCE} --context-dir=${BASH_SOURCE%/*}/../ods-core/jenkins/ocp-config --non-interactive

echo "Start Jenkins Builds"
oc start-build -n cd jenkins-master --follow
oc start-build -n cd jenkins-slave-base --follow
oc start-build -n cd jenkins-webhook-proxy --follow




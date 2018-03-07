#!/bin/bash

### Duplicate function from pre-install.sh
if [ -z $NAMESPACE ] ; then
   echo "EXITING: Need to specify the NAMESPACE env var."
   exit 1
fi

function setclient() {
        env | grep K
        if [ -z "$KUBECTL" ]; then
                echo "KUBECTL NOT SET... [ ${KUBECTL} ] setting it to '$OC'"
                export KUBECTL="$OC"
        fi
        if [[ $KUBECTL == "" ]]; then
                echo "[install.sh] EXITING: The Kubectl/OC client isn't set.  run this script with either $OC or $KUBECTL pointing to your openshift, kube client."
                exit 1
        fi
}

function setup() {
  # Skip pre-installation is a debugging feature.
  if [[ `env | grep -q "SKIP_PREINSTALL"` ]] ; then
    echo "skipping preinstall"
  else
    ./pre-install.sh
  fi

  # Reasonable defaults for simple installations in a hub namespace.
  # Note that concurrent scans is limited by hub bandwidth.
  # Could be set via introspection at some point.
  export CONCURRENT_SCAN=2
  export DEF_HUBPORT=443
  export DEF_HUBUSER="sysadmin"
  export DEF_HUB_HOST="nginx-webapp-logstash"

  if [[ $AUTO_INSTALL == "true" ]]; then
    echo "Attempting auto install."
  else
    clear
    echo "============================================"
    echo "Blackduck Hub Configuration Information"
    echo "============================================"
    echo "Interactive"
    echo "============================================"
    read -p "Hub server host (e.g. hub.mydomain.com): " hubHost
    read -p "Hub server port [$DEF_HUBPORT]: " hubPort
    read -p "Hub user name [$DEF_HUBUSER]: " hubUser
    read -sp "Hub user password : " hubPassword
    echo " "
    read -p "Maximum concurrent scans [$CONCURRENT_SCAN]: " noOfConcurrentScan
    echo "============================================"
  fi
}

function setParams() {

  hubPort="${hubPort:-$DEF_HUBHOST}"
  hubPassword="${hubPassword:-$HUB_PASSWORD}"
  hubPort="${hubPort:-$DEF_HUBPORT}"
  hubUser="${hubUser:-$DEF_HUBUSER}"
  namespace="${namespace:-$NAMESPACE}"
  noOfConcurrentScan="${noOfConcurrentScan:-$CONCURRENT_SCAN}"

  DOCKER_PASSWORD=$($OC sa get-token perceptor-scanner-sa)

  if echo $KUBECTL | grep -q oc ; then
      export Openshift="true"
  else
      export Openshift="false"
  fi
}

function create_config() {
  cat << EOF > config.yml
  apiVersion: v1
  kind: List
  metadata:
    name: viper-input
  items:
  - apiVersion: v1
    kind: ConfigMap
    metadata:
      name: viper-input
    data:
      protoform.yaml: |
        DockerPasswordOrToken: "$DOCKER_PASSWORD"
        HubHost: "$hubHost"
        HubPort: "$hubPort"
        HubUser: "$hubUser"
        # TODO, inject as secret.
        HubUserPassword: "$hubPassword"
        ConcurrentScanLimit: "$noOfConcurrentScan"
        DockerUsername: "admin"
        Namespace: "$namespace"
        Openshift: "$Openshift"
        Registry: "$REGISTRY"
        ImagePath: "$IMAGE_PATH"
        PerceptorContainerVersion: "$PERCEPTOR_VERSION"
        ScannerContainerVersion: "$PERCEPTOR_SCAN_VERSION"
        PerceiverContainerVersion: "$PERCEIVER_VERSION"
        ImageFacadeContainerVersion: "$PERCEPTOR_IF_VERSION"
EOF
}

function create_protoform() {
  cat << EOF > protoform.yml
  apiVersion: v1
  kind: Pod
  metadata:
    name: protoform
  spec:
    volumes:
    - name: viper-input
      configMap:
        name: viper-input
    containers:
    - name: protoform
      image: $REGISTRY/$IMAGE_PATH/perceptor-protoform:$VERSION
      imagePullPolicy: Always
      command: [ ./protoform ]
      ports:
      - containerPort: 3001
        protocol: TCP
      volumeMounts:
      - name: viper-input
        mountPath: /etc/protoform/
    restartPolicy: Never
    serviceAccountName: protoform
    serviceAccount: protoform
EOF
}

for opt in "$@"
do
case $opt in
  --registry)
    REGISTRY="$2"
    shift
    shift
    ;;
  --registry=*)
    REGISTRY="${opt#*=}"
    shift
    ;;
  --imagepath)
    IMAGE_PATH="$2"
    shift
    shift
    ;;
  --imagepath=*)
    IMAGE_PATH="${opt#*=}"
    shift
    ;;
  --version)
    VERSION="$2"
    shift
    shift
    ;;
  --version=*)
    VERSION="${opt#*=}"
    shift
    ;;
  --perceptor-version)
    PERCEPTOR_VERSION="$2"
    shift
    shift
    ;;
  --perceptor-version=*)
    PERCEPTOR_VERSION="${opt#*=}"
    shift
    ;;
  --perceiver-version)
    PERCEIVER_VERSION="$2"
    shift
    shift
    ;;
  --perceiver-version=*)
    PERCEIVER_VERSION="${opt#*=}"
    shift
    ;;
  --perceptor-scan-version)
    PERCEPTOR_SCAN_VERSION="$2"
    shift
    shift
    ;;
  --perceptor-scan-version=*)
    PERCEPTOR_SCAN_VERSION="${opt#*=}"
    shift
    ;;
  --perceptor-imagefacade)
    PERCEPTOR_IF_VERSION="$2"
    shift
    shift
    ;;
  --perceptor-imagefacade=*)
    PERCEPTOR_IF_VERSION="${opt#*=}"
    shift
    ;;
esac
done

REGISTRY="${REGISTRY:-gcr.io}"
IMAGE_PATH="${IMAGE_PATH:-gke-verification/blackducksoftware}"
# Note : Someday, we may want to use latest.  For now, for dev purposes, master is a easy default since
# it maps directly to the master branches for repos.
VERSION="${VERSION:-master}"

setclient
setup
setParams

create_config
$KUBECTL create -f config.yml -n $NAMESPACE

if [[ "$AUTO_INSTALL" != "true" ]]; then
  echo "Your configuration is at config.yml, click ENTER to proceed installing, or edit it before continuing."
  read x
fi

create_protoform

echo " *********************** installing protoform now **************************** "
$KUBECTL get ns
$KUBECTL create -f protoform.yml -n $NAMESPACE

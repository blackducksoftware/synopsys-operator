# TODO this logic will be go-ified via short-quinesston
#!/bin/bash

set +x
NS="$NAMESPACE"
if [[ -z "${NS}" ]]; then
	echo "Namespace env required !!!"
	exit 10
fi

function setclient() {
	if [ -z "$KUBECTL" ]; then
		KUBECTL="$OC"
	fi
	if [[ $KUBECTL == "" ]]; then 
		echo "EXITING: The Kubectl/OC client isn't set.  run this script with either $OC or $KUBECTL pointing to your openshift, kube client."
		exit 1
	fi
}

function install-rbac() {
	SCC="add-scc-to-user"
	ROLE="add-role-to-user"
	CLUSTER="add-cluster-role-to-user"
	SYSTEM_SA="system:serviceaccount"

	PERCEPTOR_SC="perceptor-scanner"
	NS_SA="${SYSTEM_SA}:${NS}"
	SCANNER_SA="${NS_SA}:${PERCEPTOR_SCANNER}"

	OS_PERCEIVER="openshift-perceiver"
	OS_PERCEIVER_SA="${NS_SA}:${OS_PERCEIVER}"

	KUBE_PERCEIVER="kube-generic-perceiver"
	KUBE_PERCEIVER_SA="${NS_SA}:${KUBE_PERCEIVER}"

	if [ "$KUBECTL" == "kubectl" ]; then
		echo "Detected Kubernetes... setting up"

		kubectl create ns $NS
		kubectl create sa perceptor-scanner-sa -n $NS
		kubectl create sa kube-generic-perceiver -n $NS
        else
		echo "Detected openshift... setting up "
		# Create the namespace to install all containers
		oc new-project $NS

		# Create the openshift-perceiver service account
		oc create serviceaccount openshift-perceiver -n $NS

		# following allows us to write cluster level metadata for imagestreams
		oc adm policy $CLUSTER cluster-admin system:serviceaccount:$NS:openshift-perceiver

		# Create the serviceaccount for perceptor-scanner to talk with Docker
		oc create sa perceptor-scanner-sa -n $NS

		# allows launching of privileged containers for Docker machine access
		oc adm policy $SCC privileged system:serviceaccount:$NS:perceptor-scanner-sa

		# following allows us to write cluster level metadata for imagestreams
		oc adm policy $CLUSTER cluster-admin system:serviceaccount:$NS:perceptor-scanner-sa

		# To pull or view all images
		oc policy $ROLE view system:serviceaccount::perceptor-scanner-sa
	fi
}

function install-contrib() {
	# Deploy a small, local prometheus.  It is only used for scraping perceptor.  Doesnt need fancy ACLs for
	# cluster discovery etc.
	$KUBECTL create -f prometheus-deployment.yaml --namespace=$NS
}

install-rbac
install-contrib

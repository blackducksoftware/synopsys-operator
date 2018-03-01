# TODO this logic will be go-ified via short-quinesston
#!/bin/bash

set +x
NS=bds-perceptor
KUBECTL="kubectl"

function is_openshift {
	if `which oc` ; then
		oc version
		return 0
	else
		return 1
	fi
	return 1
}

cleanup() {
	is_openshift
	if ! $(exit $?); then
		echo "assuming kube"
		KUBECTL="kubectl"
	else
		KUBECTL="oc"
	fi
	$KUBECTL delete ns $NS
	while $KUBECTL get ns | grep -q $NS ; do
	  echo "Waiting for deletion...`$KUBECTL get ns | grep $NS` "
	  sleep 1
	done
}

install-rbac() {
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
		set -e

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

install-contrib() {
	set -e
	# Deploy a small, local prometheus.  It is only used for scraping perceptor.  Doesnt need fancy ACLs for
	# cluster discovery etc.
	$KUBECTL create -f prometheus-deployment.yaml --namespace=$NS
}

cleanup
install-rbac
install-contrib

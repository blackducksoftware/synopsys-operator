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
                export KUBECTL="$OC"
        fi
        if [[ $KUBECTL == "" ]]; then
                echo "[pre-install.sh] EXITING: The Kubectl/OC client isn't set.  run this script with either $OC or $KUBECTL pointing to your openshift, kube client."
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

        if echo "$KUBECTL" | grep -q "kubectl" ; then
                echo "Detected Kubernetes... setting up"
                # Kubeadm, or other more secure rbac implementations on kube clusters will need to set
                # these service accounts up.  This enables protoform to create replication controllers.
cat << EOF > /tmp/protoform-rbac.yaml
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: protoform
subjects:
- kind: ServiceAccount
  name: protoform
  namespace: ${NS}
roleRef:
  kind: ClusterRole
  name: cluster-admin
  apiGroup: ""
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: protoform
  namespace: ${NS}
EOF

                $KUBECTL create ns $NS
                $KUBECTL create -f  /tmp/protoform-rbac.yaml -n ${NS}
                $KUBECTL create sa perceptor-scanner-sa -n $NS
                $KUBECTL create sa perceiver -n $NS
        else
                echo "Detected openshift... setting up using $KUBECTL as oc client. "
                # Create the namespace to install all containers
                $KUBECTL new-project $NS

                # Create the openshift-perceiver service account
                $KUBECTL create serviceaccount perceiver -n $NS

                # Protoform has its own SA.
                $KUBECTL create serviceaccount protoform -n $NS
                $KUBECTL adm policy $CLUSTER cluster-admin system:serviceaccount:$NS:protoform

                # following allows us to write cluster level metadata for imagestreams
                $KUBECTL adm policy $CLUSTER cluster-admin system:serviceaccount:$NS:perceiver

                # Create the serviceaccount for perceptor-scanner to talk with Docker
                $KUBECTL create sa perceptor-scanner-sa -n $NS

                # allows launching of privileged containers for Docker machine access
                $KUBECTL adm policy $SCC privileged system:serviceaccount:$NS:perceptor-scanner-sa

                # following allows us to write cluster level metadata for imagestreams
                $KUBECTL adm policy $CLUSTER cluster-admin system:serviceaccount:$NS:perceptor-scanner-sa

                # To pull or view all images
                $KUBECTL policy $ROLE view system:serviceaccount::perceptor-scanner-sa
        fi
}

function install-contrib() {
        # Deploy a small, local prometheus.  It is only used for scraping perceptor.  Doesnt need fancy ACLs for
        # cluster discovery etc.
        $KUBECTL create -f prometheus-deployment.yaml --namespace=$NS
}

setclient
install-rbac
install-contrib

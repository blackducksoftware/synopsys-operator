NS=$1

# echo "ns" $NS

oc delete ns $NS

oc delete crd alerts.synopsys.com
oc delete crd blackducks.synopsys.com
oc delete crd opssights.synopsys.com


oc delete clusterrolebinding synopsys-operator-admin
#oc delete clusterrolebinding protoform-admin
#oc delete clusterrolebinding synopsys-operator-cluster-admin

oc delete clusterrole skyfire
oc delete clusterrole pod-perceiver

#oc delete clusterrolebinding skyfire
#oc delete clusterrolebinding perceptor-scanner
#oc delete clusterrolebinding pod-perceiver

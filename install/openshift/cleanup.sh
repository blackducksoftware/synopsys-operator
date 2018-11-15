NS=$1

# echo "ns" $NS

$OC delete ns $NS

$OC delete crd alerts.synopsys.com
$OC delete crd hubs.synopsys.com
$OC delete crd opssights.synopsys.com


$OC delete clusterrolebinding synopsys-operator-admin
#$OC delete clusterrolebinding protoform-admin
#$OC delete clusterrolebinding synopsys-operator-cluster-admin

$OC delete clusterrole skyfire
$OC delete clusterrole pod-perceiver

#$OC delete clusterrolebinding skyfire
#$OC delete clusterrolebinding perceptor-scanner
#$OC delete clusterrolebinding pod-perceiver

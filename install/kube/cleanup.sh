NS=$1

# echo "ns" $NS

kubectl delete ns $NS

kubectl delete crd alerts.synopsys.com
kubectl delete crd blackducks.synopsys.com
kubectl delete crd opssights.synopsys.com

kubectl delete clusterrolebinding synopsys-operator-admin
#kubectl delete clusterrolebinding protoform-admin
#kubectl delete clusterrolebinding synopsys-operator-cluster-admin

kubectl delete clusterrole skyfire
kubectl delete clusterrole pod-perceiver

#kubectl delete clusterrolebinding skyfire
#kubectl delete clusterrolebinding perceptor-scanner
#kubectl delete clusterrolebinding pod-perceiver

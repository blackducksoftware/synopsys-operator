#!/bin/bash

for i in reporting polaris authserver polarisdb;do
	kubectl delete -f config/samples/synopsys_v1_$i.yaml
	sleep 3
done


for i in $(cat delete-list.txt); do
	kubectl delete $i
done 

#!/bin/bash

oc delete rc perceptor-scanner
oc delete rc image-perceiver
oc delete rc pod-perceiver

cat << EOF > perceivers.yml
apiVersion: v1
kind: List
metadata:
  name: "Openshift Perceiver"
items:
- apiVersion: v1
  kind: ReplicationController
  metadata:
    name: image-perceiver
    labels:
      app: image-perceiver
  spec:
    replicas: 1
    selector:
      name: image-perceiver
    template:
      metadata:
        labels:
          name: image-perceiver
        name: image-perceiver
      spec:
        containers:
          - name: image-perceiver
            image: gcr.io/gke-verification/blackducksoftware/image-perceiver:latest
            imagePullPolicy: Always
            ports:
              - containerPort: 4000
            resources:
              requests:
                memory: 1Gi # TODO may not even need this much since it's stateless
                cpu: 50m # TODO same here -- maybe reduce this number
              limits:
                cpu: 500m
            volumeMounts:
              - name: openshift-perceiver-config
                mountPath: /etc/perceiver
            terminationMessagePath: /dev/termination-log
        volumes:
          - name: openshift-perceiver-config
            configMap:
              name: openshift-perceiver-config
        restartPolicy: Always
        terminationGracePeriodSeconds: 30
        dnsPolicy: ClusterFirst
        serviceAccountName: openshift-perceiver
        serviceAccount: openshift-perceiver
- apiVersion: v1
  kind: ReplicationController
  metadata:
    name: pod-perceiver
    labels:
      app: pod-perceiver
  spec:
    replicas: 1
    selector:
      name: pod-perceiver
    template:
      metadata:
        labels:
          name: pod-perceiver
        name: pod-perceiver
      spec:
        containers:
          - name: pod-perceiver
            image: gcr.io/gke-verification/blackducksoftware/pod-perceiver:latest
            imagePullPolicy: Always
            ports:
              - containerPort: 4000
            resources:
              requests:
                memory: 1Gi # TODO may not even need this much since it's stateless
                cpu: 50m # TODO same here -- maybe reduce this number
              limits:
                cpu: 500m
            volumeMounts:
              - name: kube-generic-perceiver-config
                mountPath: /etc/perceiver
            terminationMessagePath: /dev/termination-log
        volumes:
          - name: kube-generic-perceiver-config
            configMap:
              name: kube-generic-perceiver-config
        restartPolicy: Always
        terminationGracePeriodSeconds: 30
        dnsPolicy: ClusterFirst
        serviceAccountName: openshift-perceiver
        serviceAccount: openshift-perceiver
EOF

oc create -f perceivers.yml

echo "Your configuration is at config.yml, click enter to proceed installing, or edit it bbefore continuing"
cat << EOF > perceptor-scanner.yml
apiVersion: v1
kind: List
metadata:
  name: "bds-perceptor components"
  resourceVersion: "0.0.1"
items:
- apiVersion: v1
  kind: ReplicationController
  metadata:
    name: perceptor-scanner
    labels:
      app: perceptor-app
  spec:
    replicas: 2
    selector:
      name: bds-perceptor
    template:
      metadata:
        labels:
          name: bds-perceptor
        name: perceptor-scanner
      spec:
        volumes:
          - emptyDir: {}
            name: "var-images"
          - name: dir-docker-socket
            hostPath:
              path: /var/run/docker.sock
          - name: perceptor-scanner-config
            configMap:
              name: perceptor-scanner-config
          - name: perceptor-imagefacade-config
            configMap:
              name: perceptor-imagefacade-config
        containers:
          - name: perceptor-scanner
            image: gcr.io/gke-verification/blackducksoftware/perceptor-scanner:latest
            imagePullPolicy: Always
            ports:
              - containerPort: 3003
            resources:
              requests:
                memory: 2Gi
                cpu: 50m
              limits:
                cpu: 500m
            volumeMounts:
              - mountPath: /var/images
                name: var-images
              - name: perceptor-scanner-config
                mountPath: /etc/perceptor_scanner
            terminationMessagePath: /dev/termination-log
          - name: perceptor-imagefacade
            image: gcr.io/gke-verification/blackducksoftware/perceptor-imagefacade:latest
            imagePullPolicy: Always
            ports:
              - containerPort: 3004
            resources:
              requests:
                memory: 2Gi
                cpu: 50m
              limits:
                cpu: 500m
            volumeMounts:
              - mountPath: /var/images
                name: var-images
              - mountPath: /var/run/docker.sock
                name: dir-docker-socket
              - name: perceptor-imagefacade-config
                mountPath: /etc/perceptor_imagefacade
            terminationMessagePath: /dev/termination-log
            securityContext:
              privileged: true
        restartPolicy: Always
        terminationGracePeriodSeconds: 30
        dnsPolicy: ClusterFirst
        serviceAccountName: perceptor-scanner-sa
        serviceAccount: perceptor-scanner-sa
EOF

oc create -f perceptor-scanner.yml

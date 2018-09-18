# Perceptor-protoform: a cloud native administration utility for Black Duck ecosystem components:

- perceptor
- perceivers (openshift, kube, ...)
- perceptor-image-facade
- perceptor-scanner
- blackduck hub

# Quick start

## Preconditions

 - have `kubectl` set up and configured

## Steps

 - Clone this repo

```
git clone git@github.com:blackducksoftware/perceptor-protoform.git
cd perceptor-protoform
```

 - Find the deploy script

```
cd install
```

 - run the script

```
./install.sh my-favorite-namespace <your_favorite_hub_registration_key> master
```

 - create a Hub

```
kubectl create -f ../examples/hub.yaml
```

 - create an OpsSight instance

```
kubectl create -f ../examples/opssight.yaml
```

 - create an alert instance

```
kubectl create -f ../examples/alert.yaml
```

# Tested Cluster Versions

Protoform has been run against the following clusters:

- Kubernetes 1.9.1 and 1.10
- Openshift Origin 3.6, 3.7, 3.9, 3.10

# Prerequisites

The user running the installation should be able to create service accounts with in-cluster API RBAC capabilities, and launch pods within them.  Specifically, protoform assumes access to oadm (for openshift users) or the ability to define RBAC objects (for kubernetes users).  

Protoform will attempt to detect your cluster type, and bootstrap all necessary components as needed.  This is done via environment variables, but the implementation is highly fluid right now, and we are leaning towards command line options once basic hardening of the core functionality is done.

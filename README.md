# WIP : Not usable yet !!! We are moving components into this repository currently.

## Perceptor-protoform: a cloud native administration utility for your scanning platform.

Protoform is a cloud native installation utility for blackduck's distributed system framework for scanning platforms over time
for perceptor (core), perceptor-convex, perceivers.

## Running

Make sure you have a kubectl client installed locally, which is configured and logged in.  

Then...

```
git clone https://github.com/blackducksoftware/perceptor-protoform.git
cd perceptor-protoform

# Sets up ACLS, you may need to modify this script.
./pre-install.sh

# Wrapper to the containerized installer.  No need to modify it.
./install.sh

# TODO , jay will delete this after he finishes viperizing things tonite.
./post-hack.sh
```

... Thats it! Now check that your pods are running:

```
kubectl get pods -n bds-perceptor
```

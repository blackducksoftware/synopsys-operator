# NOTE we are moving components into this repository currently and it is under alot of flux !

## Perceptor-protoform: a cloud native administration utility for the perceptor ecosystem components:

- perceptor
- perceivers (openshift, kube, ...)
- perceptor-image-facade
- perceptor-scan

To run it, clone this repo and simply run: `install.sh`

## Prerequisites

The user running the installation should be able to create service accounts with in-cluster API RBAC capabilities, and launch pods within them.  Specifically, protoform assumes access to oadm (for openshift users) or the ability to define RBAC objects (for kubernetes users).  

Protoform will attempt to detect your cluster type, and bootstrap all necessary components as needed.  This is done via environment variables, but the implementation is highly fluid right now, and we are leaning towards command line options once basica hardening of the core functionality is done.

## Run without cloning the source

Note that you can easily run without cloning, just pull down `install.sh` and `pre-install.sh` into the same directory.

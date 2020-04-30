# Moved !!!
Since the Synopsys Operator is decommissioned and it is no longer used to manage and deploy Alert applications from version 5.3.1 and later and Black Duck 2020.4.0 and later, we have renamed the source code repository as Synopsysctl which can be found [here] https://github.com/blackducksoftware/synopsysctl

Please refer to [Synopsysctl] (https://synopsys.atlassian.net/wiki/spaces/BDLM/pages/34373652/Synopsysctl) for an instructions to manage and deploy Synopsys application

# Overview

**Synopsys Operator** is a cloud-native administration utility for Synopsys software.  Synopsys Operator assists in the deployment and management of Synopsys software in cloud-native environments (i.e., **Kubernetes** and **OpenShift**). 

Once the Operator is installed, you can leverage it to easily deploy and manage Synopsys software like **Black Duck**, **OpsSight Connector**, and **Black Duck Alert**.

To learn more about Synopsys Operator, go to the [Synopsys Operator wiki](https://github.com/blackducksoftware/synopsys-operator/wiki).

# Quick start

## Prerequisites:

1. kube cluster, user account which has owner role.
   This can be granted using rbac, you'll specifically need the cluster-admin role.

2. synopsys-ctl binary: [download here](https://github.com/blackducksoftware/synopsys-operator/releases)

## Blackduck Installation

1. deploy CRDs: `path/to/synopsysctl deploy --enable-blackduck --cluster-scoped --enable-alert`

2. create a BlackDuck instance: `./synopsysctl create blackduck myhub --admin-password a --postgres-password p --user-password u --expose-ui LOADBALANCER --persistent-storage=true --enable-binary-analysis=true --enable-source-code-upload=true --size small --version 2019.10.1`

3. find the webapp ip address: `kubectl get services -n myhub | grep -i loadbalancer`

4. visit the webapp ip address at `https://<ip address from step 3>` in web browser, ignore security warning, log in using username and password

5. register hub using registration key

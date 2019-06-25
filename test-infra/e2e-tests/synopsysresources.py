from kubernetes import client, config
import sys
import subprocess
import yaml
import json
import time
import requests
from ctl import *
import logging


class Helper():
    def __init__(self, synopsysctl_version="latest", synopsysctl_path=None, kubectl_path=None, oc_path=None):
        self.synopsysctl_version = synopsysctl_version
        self.synopsysctl_path = synopsysctl_path
        self.synopsysctl = None
        if self.synopsysctl_path != None:
            self.synopsysctl = Synopsysctl(
                version=self.synopsysctl_version, executable=self.synopsysctl_path)

        self.kubectl_path = kubectl_path
        self.kubectl = None
        if self.kubectl_path != None:
            self.kubectl = Kubectl(self.kubectl_path)

        self.oc_path = oc_path
        self.oc = None
        if self.oc_path != None:
            self.oc = Oc(self.oc_path)

        self.terminal = Terminal()
        self.v1 = client.CoreV1Api()

    # def clusterLogIn(self, cluster_ip, username, password):
    #     command = "login {} --username={} --password={} --insecure-skip-tls-verify=true".format(
    #         cluster_ip, username, password)
    #     out, err = self.kubectl.exec(command)
    #     if err != None:
    #         logging.error(str(err))
    #         sys.exit(1)

    def waitForPodsRunning(self, namespace, label="", retry=10):
        """INPUTS
        - namespace: namespace of pods
        - label: label to identify pods
        - retry: number of times to check if pods are running
        OUTPUTS
        - error message
        """
        logging.debug("waiting for pods in namespace '{}' with label '{}'".format(
            namespace, label))
        for i in range(retry):
            retryDelay(i, retry)
            podList = self.v1.list_namespaced_pod(
                namespace, label_selector=label)
            for pod in podList.items:
                if pod.status.phase != "Running":
                    break
                return None
        return "pods failed to start"

    def waitForPodsDeleted(self, namespace, label="", retry=10):
        """INPUTS
        - namespace: namespace of pods
        - label: label to identify pods
        - retry: number of times to check if pods are deleted
        OUTPUTS
        - error message
        """
        logging.debug("waiting for pods to terminate in namespace '{}' with label '{}'".format(
            namespace, label))
        for i in range(retry):
            retryDelay(i, retry)
            podList = self.v1.list_namespaced_pod(
                namespace, label_selector=label)
            if len(podList.items) == 0:
                return None
        return "pods failed to stop running"

    def waitForNamespaceDelete(self, namespace, retry=10):
        logging.debug(
            "waiting for namespace '{}' to terminate".format(namespace))
        for i in range(retry):
            retryDelay(i, retry)
            nsList = self.v1.list_namespace()
            if namespace not in nsList.items:
                return None
        return "namespace failed to delete"

    def checkIfCrdExists(self, CRD):
        api_instance = client.CustomObjectsApi()
        group = "alerts.synopsys.com"
        version = ""
        plural = ""
        name = ""
        try:
            api_response = api_instance.get_cluster_custom_object(
                group, version, plural, name)
            return True
        except:
            print(
                "Exception when calling CustomObjectsApi->get_cluster_custom_object: %s\n" % e)

    def checkResourceExists(self, resource, resource_name):
        command = "get {} --no-headers".format(resource)
        out, err = self.kubectl.exec(command)
        if err != None:
            return err
        resources = out.split('\n')
        resource_names = [resource.split()[0]
                          for resource in resources if resource != '']
        logging.debug("Resource Names: "+str(resource_names))
        logging.debug("Found Resource: "+str(resource_name in resource_names))
        return resource_name in resource_names


class SynopsysResource:
    def __init__(self, helper):
        self.testHelper = helper
        self.label = ""

    def waitUntilRunning(self, namespace):
        return self.testHelper.waitForPodsRunning(namespace, self.label)

    def waitUntilDeleted(self, namespace):
        return self.testHelper.waitForPodsDeleted(namespace, self.label)

    def CRDExists(self, CRDName):
        command = "get {}s".format(CRDName)
        return self.testHelper.kubectl.exec(command)


class SynopsysOperator(SynopsysResource):
    def __init__(self, helper):
        self.testHelper = helper
        self.label = "app=synopsys-operator"

    def deploy(self, namespace, version="2019.4.1"):
        if self.testHelper.synopsysctl != None:
            self.testHelper.synopsysctl.deploySynopsysOperatorDefault()
        else:
            self.deploy_old(namespace, "",
                            synopsys_operator_url=self.testHelper.synopsysctl_path)

    def deploy_old(self, namespace, reg_key, synopsys_operator_url):
        # Download Synopsys Operator - https://github.com/blackducksoftware/synopsys-operator/archive/2018.12.0.tar.gz
        command = "wget {}".format(synopsys_operator_url)
        logging.debug("Command: {}".format(command))
        r = subprocess.call(command, shell=True, stdout=subprocess.PIPE)
        # Uncompress and un-tar the operator file
        command = "gunzip 2018.12.0.tar.gz"
        logging.debug("Command: {}".format(command))
        r = subprocess.call(command, shell=True, stdout=subprocess.PIPE)
        command = "tar -xvf 2018.12.0.tar"
        logging.debug("Command: {}".format(command))
        r = subprocess.call(command, shell=True, stdout=subprocess.PIPE)
        # Clean up an old operator
        command = "./cleanup.sh {}".format(namespace)
        logging.debug("Command: {}".format(command))
        subprocess.call(command, cwd="synopsys-operator-2018.12.0/install/openshift",
                        shell=True, stdout=subprocess.PIPE)
        self.testHelper.waitForNamespaceDelete(namespace)
        # Install the operator
        command = "./install.sh --blackduck-registration-key tmpkey"
        logging.debug("Command: {}".format(command))
        p = subprocess.Popen(command, cwd="synopsys-operator-2018.12.0/install/openshift",
                             shell=True, stdout=subprocess.PIPE, stdin=subprocess.PIPE)
        p.communicate(input=b'\n\n')
        self.testHelper.waitForPodsRunning(namespace)
        # Clean up Operator Tar File
        command = "rm 2018.12.0.tar"
        logging.debug("Command: {}".format(command))
        r = subprocess.call(command, shell=True, stdout=subprocess.PIPE)
        # Clean up Operator Folder
        command = "rm -rf synopsys-operator-2018.12.0"
        logging.debug("Command: {}".format(command))
        r = subprocess.call(command, shell=True, stdout=subprocess.PIPE)

    def destroy(self, namespace):
        if self.testHelper.synopsysctl != None:
            self.testHelper.synopsysctl.destroyDefault()
        elif self.testHelper.kubectl != None:
            self.testHelper.kubectl.exec("delete ns {}".format(namespace))
        elif self.testHelper.oc != None:
            self.testHelper.oc.exec("delete ns {}".format(namespace))
        else:
            return "no clients to remove Synopsys Operator"

    def waitForCRDs(self):
        a = Alert(self.testHelper)
        err = a.waitForCRD()
        if err != None:
            return err
        b = BlackDuck(self.testHelper)
        b.waitForCRD()
        if err != None:
            return err
        o = OpsSight(self.testHelper)
        o.waitForCRD()
        if err != None:
            return err


class Alert(SynopsysResource):
    def __init__(self, helper):
        self.testHelper = helper
        self.label = "app=alert"

    def deploy(self, namespace, version="4.0.0"):
        if self.testHelper.synopsysctl != None:
            if version in ["4.0.0", "3.1.0"]:
                self.testHelper.synopsysctl.exec(
                    "create alert {}".format(namespace))

    def waitForCRD(self, retry=10):
        logging.debug("verifying Alert CRD exists...")
        for i in range(retry):
            retryDelay(i, retry)
            if self.CRDExists() == True:
                return None
        return "Alert CRD failed to start"

    def CRDExists(self):
        out, err = super(Alert, self).CRDExists("alert")
        return err == None


class BlackDuck(SynopsysResource):
    def __init__(self, helper):
        self.testHelper = helper
        self.label = "app=blackduck"

    def deploy(self, namespace, version="2019.6.0"):
        if self.testHelper.synopsysctl != None:
            if version in ["2019.6.0"]:
                return self.testHelper.synopsysctl.exec(
                    "create blackduck bd --admin-password a --postgres-password p --user-password u")
            else:
                return self.testHelper.synopsysctl.exec("create blackduck bd")
            return None
        return "failed to create Black Duck"

    def waitForCRD(self, retry=10):
        logging.debug("verifying Black Duck CRD exists...")
        for i in range(retry):
            retryDelay(i, retry)
            if self.CRDExists():
                return None
        return "Black Duck CRD failed to start"

    def CRDExists(self):
        out, err = super(BlackDuck, self).CRDExists("blackduck")
        return err == None


class OpsSight(SynopsysResource):
    def __init__(self, helper):
        self.testHelper = helper
        self.label = "app=opssight"

    def deploy(self, namespace, version="2.2.3"):
        if self.testHelper.synopsysctl != None:
            self.testHelper.synopsysctl.exec("create opssight opssight-test")
        else:
            self.testHelper.deploy_old(namespace)

    def deploy_old(self, namespace):
        # Delete opssight instance if already exists
        if self.testHelper.checkResourceExists("opssights", namespace):
            command = "kubectl delete opssights opssight-test"
            logging.debug("Command: {}".format(command))
            r = subprocess.call(command, shell=True, stdout=subprocess.PIPE)
        # Delete opssight namespace if already exists
        if self.testHelper.checkResourceExists("ns", namespace):
            command = "kubectl delete ns opssight-test"
            logging.debug("Command: {}".format(command))
            r = subprocess.call(command, shell=True, stdout=subprocess.PIPE)
            self.testHelper.waitForNamespaceDelete("opssight-test")
        # Get Opssight yaml
        command = "wget https://raw.githubusercontent.com/blackducksoftware/opssight-connector/release-2.2.x/examples/opssight.json"
        logging.debug("Command: {}".format(command))
        r = subprocess.call(command, shell=True, stdout=subprocess.PIPE)
        time.sleep(2)
        # Create Opssight from yaml
        command = "kubectl create -f opssight.json"
        logging.debug("Command: {}".format(command))
        r = subprocess.call(command, shell=True, stdout=subprocess.PIPE)
        self.testHelper.waitForPodsRunning("opssight-test")
        # Clean up Opssight yaml
        command = "rm opssight.json"
        logging.debug("Command: {}".format(command))
        r = subprocess.call(command, shell=True, stdout=subprocess.PIPE)

    def waitForCRD(self, retry=10):
        logging.debug("verifying OpsSight CRD exists...")
        for i in range(retry):
            retryDelay(i, retry)
            if self.CRDExists():
                return None
        return "OpsSight CRD failed to start"

    def CRDExists(self):
        out, err = super(OpsSight, self).CRDExists("opssight")
        return err == None

    def addHubToConfig(self, v1, namespace, hub_host):
        try:
            # Read the current Config Map Body Object
            opssight_cm = v1.read_namespaced_config_map('opssight', namespace)
            opssight_data = opssight_cm.data
            opssight_data_json = json.loads(opssight_data['opssight.json'])
            if hub_host not in opssight_data_json['Hub']['Hosts']:
                opssight_data_json['Hub']['Hosts'].append(hub_host)
            opssight_data['opssight.json'] = json.dumps(opssight_data_json)
            # Update the Config Map with new Cofig Map Body Object
            opssight_cm.data = opssight_data
            v1.patch_namespaced_config_map('opssight', namespace, opssight_cm)
        except Exception as e:
            logging.debug("Exception when editing OpsSight Config: %s\n" % e)
            sys.exit(1)

    def setSkyfireReplica(self, v1, namespace, count):
        try:
            # Read the current Config Map Body Object
            skyfire_rc = v1.read_namespaced_replication_controller(
                'skyfire', namespace)
            skyfire_rc_spec = skyfire_rc.spec
            skyfire_rc_spec.replicas = count
            skyfire_rc.spec = skyfire_rc_spec
            # Update the Replication Controller
            v1.patch_namespaced_replication_controller(
                'skyfire', namespace, skyfire_rc)
        except Exception as e:
            logging.debug(
                "Exception when editing Skyfire Replication Controller: %s\n" % e)
            sys.exit(1)
        return self.testHelper.waitForPodsRunning(namespace)

    def getSkyfireRoute(self, namespace):
        skyfire_route = ""
        try:
            # Expose the service if route doesn't exist
            command = "kubectl get routes -n {} --no-headers".format(namespace)
            r = subprocess.run(command, shell=True, stdout=subprocess.PIPE)
            routes = r.stdout.split(b'\n')
            route_names = [route.split()[0]
                           for route in routes if route != b'']
            if b'skyfire' not in route_names:
                r = subprocess.run(
                    "kubectl expose service skyfire -n {}".format(namespace), shell=True, stdout=subprocess.PIPE)
                # Parse Routes for Skyfire URL
                command = "kubectl get routes -n {} --no-headers".format(
                    namespace)
                r = subprocess.run(command, shell=True, stdout=subprocess.PIPE)
                routes = r.stdout.split(b'\n')
            routes = [route.split() for route in routes if route != b'']
            skyfire_route = [route[1]
                             for route in routes if route[0] == b'skyfire'][0]
            logging.debug("Skyfire Route: %s", skyfire_route)
        except Exception as e:
            logging.debug("Exception when exposing Skyfire Route: %s\n" % e)
            sys.exit(1)
        return skyfire_route


def retryDelay(count, retries, delay=4):
    if count != 0:
        time.sleep(delay)
        logging.debug(" > retrying ({}/{})...".format(count, retries-1))

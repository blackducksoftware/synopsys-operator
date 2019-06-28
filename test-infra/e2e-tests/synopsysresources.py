import json
import logging
import subprocess
import time

import yaml
from kubernetes import client, config, watch

from ctl import Synopsysctl, Terminal


class Helper():
    def __init__(self, synopsysctl_version="latest", synopsysctl_path=None):
        self.synopsysctl_version = synopsysctl_version
        self.synopsysctl_path = synopsysctl_path
        self.synopsysctl = None
        if self.synopsysctl_path is not None:
            self.synopsysctl = Synopsysctl(
                version=self.synopsysctl_version,
                executable=self.synopsysctl_path)

        # kubernetes-python apiClient
        self.apiClient = client.api_client.ApiClient()

        # k8s apis
        self.corev1Api = client.CoreV1Api()
        self.apiextensionsV1beta1Api = client.ApiextensionsV1beta1Api()
        self.customObjectsApi = client.CustomObjectsApi()
        self.watcher = watch.Watch()

    # def clusterLogIn(self, cluster_ip, username, password):
    #     command = "login {} --username={} --password={} --insecure-skip-tls-verify=true".format(
    #         cluster_ip, username, password)
    #     out, err = self.kubectl.exec(command)
    #     if err != None:
    #         logging.error(str(err))
    #         sys.exit(1)

    def arePodsRunning(self, namespace, label="", timeout_seconds=60, watch=True):
        """INPUTS
        - namespace: namespace of pods
        - label: label to identify pods
        - retry: number of times to check if pods are running
        OUTPUTS
        - error message
        """
        logging.debug("waiting for pods in namespace '{}' with label '{}'".format(
            namespace, label))
        stream = self.corev1Api.list_namespaced_pod(namespace=namespace,
                                                    label_selector=label,
                                                    pretty=True,
                                                    timeout_seconds=timeout_seconds,
                                                    watch=watch
                                                    )
        for event in stream:
            logging.info(
                f"Event: {event['type']} {event['object'].kind} {event['object'].metadata.name}")

        for pod in stream.items:
            if pod.status.phase != "Running":
                break
            return True
        return False

    def arePodsDeleted(self, namespace, label="", retry=100):
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
            podList = self.corev1Api.list_namespaced_pod(
                namespace, label_selector=label)
            if len(podList.items) == 0:
                return True
        return False

    def isNamespaceDeleted(self, namespace, retry=10):
        logging.debug(
            "waiting for namespace '{}' to terminate".format(namespace))
        for i in range(retry):
            retryDelay(i, retry)
            nsList = self.corev1Api.list_namespace()
            if namespace not in nsList.items:
                return True
        return False

    def doesCrdExist(self, name):
        api_response = self.apiextensionsV1beta1Api.read_custom_resource_definition(
            name)
        return True if api_response is not None else False

    def doesCrExist(self, group, version, plural, name):
        api_response = self.customObjectsApi.get_cluster_custom_object(
            group, version, plural, name)
        return True if api_response is not None else False


class SynopsysResource:
    def __init__(self, helper=Helper()):
        self.testHelper = helper
        self.label = ""

    def arePodsRunning(self, namespace):
        return self.testHelper.arePodsRunning(namespace, self.label)

    def arePodsDeleted(self, namespace):
        return self.testHelper.arePodsDeleted(namespace, self.label)

    def doesCrdExist(self, CRDName):
        return self.testHelper.doesCrdExist(CRDName)


class SynopsysOperator(SynopsysResource):
    def __init__(self, helper=Helper()):
        self.testHelper = helper
        self.label = "app=synopsys-operator"

    def deploy(self, namespace, version="2019.4.1"):
        if self.testHelper.synopsysctl != None:
            return self.testHelper.synopsysctl.deploySynopsysOperatorDefault()
        else:
            return "cannot deploy cause testHelper.synopsysctl is None"

    # def deploy_old(self, namespace, reg_key, synopsys_operator_url):
    #     # Download Synopsys Operator - https://github.com/blackducksoftware/synopsys-operator/archive/2018.12.0.tar.gz
    #     command = "wget {}".format(synopsys_operator_url)
    #     logging.debug("Command: {}".format(command))
    #     r = subprocess.call(command, shell=True, stdout=subprocess.PIPE)
    #     # Uncompress and un-tar the operator file
    #     command = "gunzip 2018.12.0.tar.gz"
    #     logging.debug("Command: {}".format(command))
    #     r = subprocess.call(command, shell=True, stdout=subprocess.PIPE)
    #     command = "tar -xvf 2018.12.0.tar"
    #     logging.debug("Command: {}".format(command))
    #     r = subprocess.call(command, shell=True, stdout=subprocess.PIPE)
    #     # Clean up an old operator
    #     command = "./cleanup.sh {}".format(namespace)
    #     logging.debug("Command: {}".format(command))
    #     subprocess.call(command, cwd="synopsys-operator-2018.12.0/install/openshift",
    #                     shell=True, stdout=subprocess.PIPE)
    #     self.testHelper.isNamespaceDeleted(namespace)
    #     # Install the operator
    #     command = "./install.sh --blackduck-registration-key tmpkey"
    #     logging.debug("Command: {}".format(command))
    #     p = subprocess.Popen(command, cwd="synopsys-operator-2018.12.0/install/openshift",
    #                          shell=True, stdout=subprocess.PIPE, stdin=subprocess.PIPE)
    #     p.communicate(input=b'\n\n')
    #     self.testHelper.arePodsRunning(namespace)
    #     # Clean up Operator Tar File
    #     command = "rm 2018.12.0.tar"
    #     logging.debug("Command: {}".format(command))
    #     r = subprocess.call(command, shell=True, stdout=subprocess.PIPE)
    #     # Clean up Operator Folder
    #     command = "rm -rf synopsys-operator-2018.12.0"
    #     logging.debug("Command: {}".format(command))
    #     r = subprocess.call(command, shell=True, stdout=subprocess.PIPE)

    def destroy(self, namespace):
        if self.testHelper.synopsysctl != None:
            return self.testHelper.synopsysctl.destroyDefault()
        else:
            return "no clients to remove Synopsys Operator"

    def didCrdsComeUp(self):
        a = Alert(self.testHelper)
        b = BlackDuck(self.testHelper)
        o = OpsSight(self.testHelper)
        return a.didCrdComeUp() and b.didCrdComeUp() and o.didCrdComeUp()


class Alert(SynopsysResource):
    def __init__(self, helper):
        self.testHelper = helper
        self.label = "app=alert"

    def deploy(self, namespace, version="4.0.0"):
        if self.testHelper.synopsysctl != None:
            if version in ["4.0.0", "3.1.0"]:
                return self.testHelper.synopsysctl.exec(f"create alert {namespace}")
            else:
                return "Not a valid alert version"
        else:
            return "Yo dog you missing that synopsysctl, call er up"

    def didCrdComeUp(self, retry=10):
        logging.debug("verifying Alert CRD exists...")
        for i in range(retry):
            retryDelay(i, retry)
            if self.doesCrdExist():
                return True
        return False

    def doesCrdExist(self):
        return super(Alert, self).doesCrdExist("alerts.synopsys.com")


class BlackDuck(SynopsysResource):
    def __init__(self, helper):
        self.testHelper = helper
        self.label = "app=blackduck"

    def deploy(self, namespace, version="2019.6.0"):
        if self.testHelper.synopsysctl != None:
            if version in ["2019.6.0"]:
                return self.testHelper.synopsysctl.exec(
                    f"create blackduck {namespace} --admin-password a --postgres-password p --user-password u")
            else:
                return self.testHelper.synopsysctl.exec(f"create blackduck {namespace}")
        return "failed to create Black Duck"

    def didCrdComeUp(self, retry=10):
        logging.debug("verifying Black Duck CRD exists...")
        for i in range(retry):
            retryDelay(i, retry)
            if self.doesCrdExist():
                return True
        return False

    def doesCrdExist(self):
        return super(BlackDuck, self).doesCrdExist("blackducks.synopsys.com")


class OpsSight(SynopsysResource):
    def __init__(self, helper):
        self.testHelper = helper
        self.label = "app=opssight"

    def deploy(self, namespace, version="2.2.3"):
        if self.testHelper.synopsysctl != None:
            return self.testHelper.synopsysctl.exec("create opssight opssight-test")
        else:
            return "no synopsysctl"
        # else:
        #     self.testHelper.deploy_old(namespace)

    # def deploy_old(self, namespace):
    #     # Delete opssight instance if already exists
    #     if self.testHelper.doesCrExist("opssights", namespace):
    #         command = "kubectl delete opssights opssight-test"
    #         logging.debug("Command: {}".format(command))
    #         r = subprocess.call(command, shell=True, stdout=subprocess.PIPE)
    #     # Delete opssight namespace if already exists
    #     if self.testHelper.doesCrExist("ns", namespace):
    #         command = "kubectl delete ns opssight-test"
    #         logging.debug("Command: {}".format(command))
    #         r = subprocess.call(command, shell=True, stdout=subprocess.PIPE)
    #         self.testHelper.isNamespaceDeleted("opssight-test")
    #     # Get Opssight yaml
    #     command = "wget https://raw.githubusercontent.com/blackducksoftware/opssight-connector/release-2.2.x/examples/opssight.json"
    #     logging.debug("Command: {}".format(command))
    #     r = subprocess.call(command, shell=True, stdout=subprocess.PIPE)
    #     time.sleep(2)
    #     # Create Opssight from yaml
    #     command = "kubectl create -f opssight.json"
    #     logging.debug("Command: {}".format(command))
    #     r = subprocess.call(command, shell=True, stdout=subprocess.PIPE)
    #     self.testHelper.arePodsRunning("opssight-test")
    #     # Clean up Opssight yaml
    #     command = "rm opssight.json"
    #     logging.debug("Command: {}".format(command))
    #     r = subprocess.call(command, shell=True, stdout=subprocess.PIPE)

    def didCrdComeUp(self, retry=10):
        logging.debug("verifying OpsSight CRD exists...")
        for i in range(retry):
            retryDelay(i, retry)
            if self.doesCrdExist():
                return True
        return False

    def doesCrdExist(self):
        return super(OpsSight, self).doesCrdExist("opssights.synopsys.com")

    # def addHubToConfig(self, v1, namespace, hub_host):
    #     try:
    #         # Read the current Config Map Body Object
    #         opssight_cm = v1.read_namespaced_config_map('opssight', namespace)
    #         opssight_data = opssight_cm.data
    #         opssight_data_json = json.loads(opssight_data['opssight.json'])
    #         if hub_host not in opssight_data_json['Hub']['Hosts']:
    #             opssight_data_json['Hub']['Hosts'].append(hub_host)
    #         opssight_data['opssight.json'] = json.dumps(opssight_data_json)
    #         # Update the Config Map with new Cofig Map Body Object
    #         opssight_cm.data = opssight_data
    #         v1.patch_namespaced_config_map('opssight', namespace, opssight_cm)
    #     except Exception as e:
    #         logging.debug("Exception when editing OpsSight Config: %s\n" % e)
    #         sys.exit(1)

    # def setSkyfireReplica(self, v1, namespace, count):
    #     try:
    #         # Read the current Config Map Body Object
    #         skyfire_rc = v1.read_namespaced_replication_controller(
    #             'skyfire', namespace)
    #         skyfire_rc_spec = skyfire_rc.spec
    #         skyfire_rc_spec.replicas = count
    #         skyfire_rc.spec = skyfire_rc_spec
    #         # Update the Replication Controller
    #         v1.patch_namespaced_replication_controller(
    #             'skyfire', namespace, skyfire_rc)
    #     except Exception as e:
    #         logging.debug(
    #             "Exception when editing Skyfire Replication Controller: %s\n" % e)
    #         sys.exit(1)
    #     return self.testHelper.arePodsRunning(namespace)

    # def getSkyfireRoute(self, namespace):
    #     skyfire_route = ""
    #     try:
    #         # Expose the service if route doesn't exist
    #         command = "kubectl get routes -n {} --no-headers".format(namespace)
    #         r = subprocess.run(command, shell=True, stdout=subprocess.PIPE)
    #         routes = r.stdout.split(b'\n')
    #         route_names = [route.split()[0]
    #                        for route in routes if route != b'']
    #         if b'skyfire' not in route_names:
    #             r = subprocess.run(
    #                 "kubectl expose service skyfire -n {}".format(namespace), shell=True, stdout=subprocess.PIPE)
    #             # Parse Routes for Skyfire URL
    #             command = "kubectl get routes -n {} --no-headers".format(
    #                 namespace)
    #             r = subprocess.run(command, shell=True, stdout=subprocess.PIPE)
    #             routes = r.stdout.split(b'\n')
    #         routes = [route.split() for route in routes if route != b'']
    #         skyfire_route = [route[1]
    #                          for route in routes if route[0] == b'skyfire'][0]
    #         logging.debug("Skyfire Route: %s", skyfire_route)
    #     except Exception as e:
    #         logging.debug("Exception when exposing Skyfire Route: %s\n" % e)
    #         sys.exit(1)
    #     return skyfire_route


def retryDelay(count, retries, delay=4):
    if count != 0:
        time.sleep(delay)
        logging.debug(" > retrying ({}/{})...".format(count, retries-1))

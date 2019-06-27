from kubernetes import client, config, utils
from synopsysresources import *
from ctl import *
import sys
import subprocess
import yaml
import json
import time
import requests
import pytest
import logging

try:
    config.load_kube_config()
except:
    logging.error("FAILED to load kube config")

client.rest.logger.setLevel("INFO")
logging.basicConfig(level=logging.DEBUG)

synopsysctl_path = "synopsysctl"

image = "-i docker.io/blackducksoftware/synopsys-operator:2019.6.0-RC"


@pytest.mark.smoke
def test_synopsysctlSmoke():
    logging.info("TEST: test_synopsysctlSmoke")
    return
    # Create clients
    h = Helper(synopsysctl_path=synopsysctl_path)
    synopsysctl = Synopsysctl(executable=synopsysctl_path, version="latest")
    so = SynopsysOperator(h)

    alert = Alert(h)
    blackDuck = BlackDuck(h)
    opsSight = OpsSight(h)

    # Deploy Synopsys Operator
    out, err = synopsysctl.exec(
        f"deploy {image} --cluster-scoped --enable-alert --enable-blackduck --enable-opssight")
    if err != None:
        raise Exception(err)
    so.waitUntilRunning("synopsys-operator")
    if not alert.didCrdComeUp():
        raise Exception("Alert CRD failed to come up")
    if not blackDuck.didCrdComeUp():
        raise Exception("Black Duck CRD failed to come up")
    if not opsSight.didCrdComeUp():
        raise Exception("OpsSight CRD failed to come up")

    # Create each Synopsys resource
    out, err = synopsysctl.exec("create alert alt --persistent-storage=false")
    if err != None:
        raise Exception(err)
    if not alert.waitUntilRunning("alt"):
        raise Exception("Alert failed to start running")

    out, err = synopsysctl.exec(
        "create blackduck bd --admin-password a --postgres-password p --user-password u")
    if err != None:
        raise Exception(err)
    if not blackDuck.waitUntilRunning("bd"):
        raise Exception("Black Duck failed to start running")

    out, err = synopsysctl.exec("create opssight ops")
    if err != None:
        raise Exception(err)
    if not opsSight.waitUntilRunning("ops"):
        raise Exception("OpsSight failed to start running")

    # Get each Synopsys resource
    out, err = synopsysctl.exec("get alerts")
    if err != None:
        raise Exception(err)
    if "alt" not in out:
        raise Exception("Failed to get Alert alt")
    out, err = synopsysctl.exec("get blackducks")
    if err != None:
        raise Exception(err)
    if "bd" not in out:
        raise Exception("Failed to get Black Duck bd")
    out, err = synopsysctl.exec("get opssights")
    if err != None:
        raise Exception(err)
    if "ops" not in out:
        raise Exception("Failed to get OpsSight ops")

    # Describe each Synopsys resource
    out, err = synopsysctl.exec("describe alert alt")
    if err != None:
        raise Exception(err)
    out, err = synopsysctl.exec("describe blackduck bd")
    if err != None:
        raise Exception(err)
    out, err = synopsysctl.exec("describe opssight ops")
    if err != None:
        raise Exception(err)

    # Delete each Synopsys resource
    out, err = synopsysctl.exec("delete alert alt")
    if err != None:
        raise Exception(err)
    if not alert.waitUntilDeleted("alt"):
        raise Exception("Failed to delete Alert alt")

    out, err = synopsysctl.exec("delete blackduck bd")
    if err != None:
        raise Exception(err)
    if not blackDuck.waitUntilDeleted("bd"):
        raise Exception("Failed to delete Black Duck bd")

    out, err = synopsysctl.exec("delete opssight ops")
    if err != None:
        raise Exception(err)
    if not opsSight.waitUntilDeleted("ops"):
        raise Exception("Failed to delete OpsSight ops")

    # Destroy Synopsys Operator
    out, err = synopsysctl.exec("destroy")
    if err != None:
        assert err
    if not so.waitUntilDeleted("synopsys-operator"):
        raise Exception("Failed to delete Synopsys Operator")

    if alert.CRDExists():
        raise Exception("Alert CRD didn't delete")
    if blackDuck.CRDExists():
        raise Exception("Black Duck CRD didn't delete")
    if opsSight.CRDExists():
        raise Exception("OpsSight CRD didn't delete")


def test_mockAlert():
    logging.info("TEST: test_mockAlert")
    # Create clients
    h = Helper(synopsysctl_path=synopsysctl_path)
    synopsysctl = Synopsysctl(executable=synopsysctl_path, version="latest")
    terminal = Terminal()

    # Deploy the Synopsys Operator
    synopsysctl.exec(
        "deploy {image} --cluster-scoped --enable-alert --enable-blackduck --enable-opssight")
    so = SynopsysOperator(h)
    so.waitUntilRunning("synopsys-operator")
    if not so.didCrdsComeUp():
        raise Exception("CRDs failed to come up")

    # Generate files
    altJSON, err = synopsysctl.exec(
        "create alert alt-json --persistent-storage false --mock json")
    f = open("alt.json", "w")
    f.write(altJSON)
    f.close()

    altYAML, err = synopsysctl.exec(
        "create alert alt-yaml --persistent-storage false --mock yaml")
    f = open("alt.yaml", "w")
    f.write(altYAML)
    f.close()

    altKubeJSON, err = synopsysctl.exec(
        "create alert alt-kube-json --persistent-storage false --mock-kube json")
    f = open("alt-kube.json", "w")
    f.write(altKubeJSON)
    f.close()

    altKubeYAML, err = synopsysctl.exec(
        "create alert alt-kube-yaml --persistent-storage false --mock-kube yaml")
    f = open("alt-kube.yaml", "w")
    f.write(altKubeYAML)
    f.close()

    # Create namespaces
    h.corev1Api.create_namespace("alt-json")
    h.corev1Api.create_namespace("alt-yaml")
    h.corev1Api.create_namespace("alt-kube-json")
    h.corev1Api.create_namespace("alt-kube-yaml")

    # Create Alert instances in namespaces
    utils.create_from_yaml(h.apiClient, "alt.json")
    utils.create_from_yaml(h.apiClient, "alt.yaml")
    utils.create_from_yaml(h.apiClient, "alt-kube.json")
    utils.create_from_yaml(h.apiClient, "alt-kube.yaml")

    # Wait for Alert instances to be running
    alert = Alert(h)
    alert.waitUntilRunning("alt-json")
    alert.waitUntilRunning("alt-yaml")
    alert.waitUntilRunning("alt-kube-json")
    alert.waitUntilRunning("alt-kube-yaml")

    # Delete the CRDs
    h.customObjectsApi.delete_cluster_custom_object(
        "synopsys.com", "apiextensions.k8s.io/v1beta1", "alerts", "alt-json", client.V1DeleteOptions())
    h.customObjectsApi.delete_cluster_custom_object(
        "synopsys.com", "apiextensions.k8s.io/v1beta1", "alerts", "alt-yaml", client.V1DeleteOptions())

    h.corev1Api.delete_namespace("alt-json", client.V1DeleteOptions)
    h.corev1Api.delete_namespace("alt-yaml", client.V1DeleteOptions())
    h.corev1Api.delete_namespace("alt-kube-json", client.V1DeleteOptions())
    h.corev1Api.delete_namespace("alt-kube-yaml", client.V1DeleteOptions())

    # Verify Alert instances deleted
    alert.waitUntilDeleted("alt-json")
    alert.waitUntilDeleted("alt-yaml")

    # Delete the files
    terminal.exec("rm alt.json")
    terminal.exec("rm alt.yaml")
    terminal.exec("rm alt-kube.json")
    terminal.exec("rm alt-kube.yaml")

    # Remove the Synopsys Operator
    so.destroy("synopsys-operator")
    so.waitUntilDeleted("synopsys-operator")


def test_namespacedOperations():
    logging.info("TEST: test_namespacedOperations")
    # Create clients
    h = Helper(synopsysctl_path="synopsysctl_path")
    synopsysctl = Synopsysctl(executable=synopsysctl_path, version="latest")
    so = SynopsysOperator(h)
    alt = Alert(h)
    bd = BlackDuck(h)

    # Deploy Synopsys Operator into namespace 1
    synopsysctl.exec("deploy {} --enable-alert -n test-space1".format(image))
    so.waitUntilRunning(namespace="test-space1")
    alt.didCrdComeUp()

    # Deploy Synopsys Operator into namespace 2
    synopsysctl.exec("deploy --enable-blackduck -n test-space2")
    so.waitUntilRunning(namespace="test-space2")
    bd.waitForCRD()

    # Create an Alert instance in namespace 1
    synopsysctl.exec("create alert alt -n test-space1")
    alt.waitUntilRunning(namespace="test-space1")

    # Create a Black Duck instance in namespace 2
    synopsysctl.exec(
        "create blackduck bd --admin-password a --postgres-password p --user-password u -n test-space2")
    bd.waitUntilRunning(namespace="test-space2")

    # Delete the Alert instance in namespace 1
    synopsysctl.exec("delete alert alt -n test-space1")
    alt.waitUntilDeleted(namespace="test-space1")

    # Delete the Black Duck instance in namespace 2
    synopsysctl.exec("delete blackduck bd -n test-space2")
    alt.waitUntilDeleted(namespace="test-space2")

    # Remove Synopsys Operator from namespace 1
    synopsysctl.exec("destroy test-space1")
    so.waitUntilDeleted(namespace="test-space1")

    # Remove Synopsys Operator from namespace 2
    synopsysctl.exec("destroy test-space2")
    so.waitUntilDeleted(namespace="test-space2")


def test_nodeAffinity():
    logging.info("TEST: test_nodeAffinity")
    # Create Helpers
    h = Helper(synopsysctl_path=synopsysctl_path)
    synopsysctl = Synopsysctl(executable=synopsysctl_path, version="latest")
    bd = BlackDuck(h)

    # Deploy Synopsys Operator
    synopsysctl.exec(
        "deploy {} --cluster-scoped --enable-blackduck".format(image))
    so = SynopsysOperator(h)
    so.waitUntilRunning("synopsys-operator")

    # Create Black Duck with Node Affinity
    bd.waitForCRD()
    synopsysctl.exec(
        "create blackduck bd --node-affinity-file-path ~/gocode/src/github.com/blackducksoftware/synopsys-operator/test-infra/e2e-tests/node.json --admin-password a --postgres-password p --user-password u --persistent-storage false")
    bd.waitUntilRunning("bd")

    # Check if Documentation pod was configured correctly
    documentationPod = h.corev1Api.list_namespaced_pod(
        namespace="bd", label_selector='component=documentation').items[0]
    docPodNodeAffinity = documentationPod.spec.affinity.node_affinity
    docPodNodeTerms = docPodNodeAffinity.required_during_scheduling_ignored_during_execution.node_selector_terms
    found = False
    for term in docPodNodeTerms:
        for match in term.match_expressions:
            foundKey = match.key == "beta.kubernetes.io/arch"
            foundOperator = match.operator == "In"
            foundVal = match.values == ["amd64"]
            if foundKey and foundOperator and foundVal:
                found = True
                break
        if found:
            break
    if not found:
        raise Exception("couldn't determine Node Affinity")

    # cleanup
    synopsysctl.exec("delete blackduck bd")
    bd.waitUntilDeleted(namespace="bd")
    synopsysctl.exec("destroy")


@pytest.mark.v2019_4_0
def test_2019_4_0():
    # test_exampleFail()
    # test_examplePass()

    # test_skyfire()
    # test_smoke()
    # test_mockAlert()
    # test_mockBlackDuck()
    # test_mockOpsSight()
    # test_namespacedOperations()
    # test_nodeAffinity()
    pass


def main():
    h = Helper(synopsysctl_path=synopsysctl_path)
    try:
        test_synopsysctlSmoke()
    except Exception as e:
        print(f"Exception in main: {e}")
    finally:
        logging.debug("sending delete namespace events to k8s api")
        h.corev1Api.delete_namespace("alt", client.V1DeleteOptions())
        h.corev1Api.delete_namespace("bd", client.V1DeleteOptions())
        h.corev1Api.delete_namespace("ops", client.V1DeleteOptions())
        h.corev1Api.delete_namespace(
            "synopsys-operator", client.V1DeleteOptions())

        logging.debug("sending delete crd events to k8s api")
        h.apiextensionsV1beta1Api.delete_custom_resource_definition(
            "alerts.synopsys.com", client.V1DeleteOptions())
        h.apiextensionsV1beta1Api.delete_custom_resource_definition(
            "blackducks.synopsys.com", client.V1DeleteOptions())
        h.apiextensionsV1beta1Api.delete_custom_resource_definition(
            "opssights.synopsys.com", client.V1DeleteOptions())


if __name__ == '__main__':
    logging.debug("running main")
    main()

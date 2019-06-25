from kubernetes import client, config
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

synopsysctl_path = "../../../../../../../Desktop/ctls/synopsysctl-latest/synopsysctl"
kubectl_path = "../../../../../../../Desktop/ctls/kubectl/kubectl"

image = "-i docker.io/blackducksoftware/synopsys-operator:2019.6.0-RC"

def test_synopsysctlSmoke():
    logging.info("TEST: test_synopsysctlSmoke")
    # Create clients
    h = Helper(synopsysctl_path=synopsysctl_path, kubectl_path=kubectl_path)
    synopsysctl = Synopsysctl(path=synopsysctl_path, version="latest")
    so = SynopsysOperator(h)

    alert = Alert(h)
    blackDuck = BlackDuck(h)
    opsSight = OpsSight(h)

    # Deploy Synopsys Operator
    out, err = synopsysctl.exec(
        "deploy --cluster-scoped --enable-alert --enable-blackduck --enable-opssight")
    if err != None:
        assert err
    so.waitUntilRunning("synopsys-operator")
    err = alert.waitForCRD()
    if err != None:
        assert err
    err = blackDuck.waitForCRD()
    if err != None:
        assert err
    err = opsSight.waitForCRD()
    if err != None:
        assert err

    # Create each Synopsys resource
    out, err = synopsysctl.exec("create alert alt")
    if err != None:
        assert err
    out, err = synopsysctl.exec(
        "create blackduck bd --admin-password a --postgres-password p --user-password u")
    if err != None:
        assert err
    out, err = synopsysctl.exec("create opssight ops")
    if err != None:
        assert err
    if alert.waitUntilRunning("alt") != None:
        assert 0
    if blackDuck.waitUntilRunning("bd") != None:
        assert 0
    if opsSight.waitUntilRunning("ops") != None:
        assert 0

    # Stop each Synopsys resource
    out, err = synopsysctl.exec("stop alert alt")
    if err != None:
        assert err
    out, err = synopsysctl.exec("stop blackduck bd")
    if err != None:
        assert err
    out, err = synopsysctl.exec("stop opssight ops")

    # Start each Synopsys resource
    out, err = synopsysctl.exec("start alert alt")
    if err != None:
        assert err
    out, err = synopsysctl.exec("start blackduck bd")
    if err != None:
        assert err
    out, err = synopsysctl.exec("start opssight ops")

    # Get each Synopsys resource
    out, err = synopsysctl.exec("get alerts")
    if err != None:
        assert 0
    if "alt" not in out:
        assert 0
    out, err = synopsysctl.exec("get blackducks")
    if err != None:
        assert 0
    if "bd" not in out:
        assert 0
    out, err = synopsysctl.exec("get opssights")
    if err != None:
        assert 0
    if "ops" not in out:
        assert 0

    # Describe each Synopsys resource
    out, err = synopsysctl.exec("describe alert alt")
    if err != None:
        assert 0
    out, err = synopsysctl.exec("describe blackduck bd")
    if err != None:
        assert 0
    out, err = synopsysctl.exec("describe opssight ops")
    if err != None:
        assert 0

    # Update each Synopsys resource
    out, err = synopsysctl.exec("update alert alt")
    if err != None:
        assert 0
    out, err = synopsysctl.exec("update blackduck bd")
    if err != None:
        assert 0
    out, err = synopsysctl.exec("update opssight ops")
    if err != None:
        assert 0

    # Delete each Synopsys resource
    out, err = synopsysctl.exec("delete alert alt")
    if err != None:
        assert 0
    out, err = synopsysctl.exec("delete blackduck bd")
    if err != None:
        assert 0
    out, err = synopsysctl.exec("delete opssight ops")
    if err != None:
        assert 0
    if alert.waitUntilDeleted("alt") != None:
        assert 0
    if blackDuck.waitUntilDeleted("bd") != None:
        assert 0
    if opsSight.waitUntilDeleted("ops") != None:
        assert 0

    # Destroy Synopsys Operator
    out, err = synopsysctl.exec("destroy")
    if err != None:
        assert err
    if so.waitUntilDeleted("synopsys-operator") != None:
        assert 0

    if alert.CRDExists():
        raise Exception("Alert CRD didn't delete")
    if blackDuck.CRDExists():
        raise Exception("Black Duck CRD didn't delete")
    if opsSight.CRDExists():
        raise Exception("OpsSight CRD didn't delete")


def test_mockAlert():
    logging.info("TEST: test_mockAlert")
    # Create clients
    h = Helper(synopsysctl_path=synopsysctl_path, kubectl_path=kubectl_path)
    synopsysctl = Synopsysctl(path=synopsysctl_path, version="latest")
    kubectl = Kubectl(path=kubectl_path)
    terminal = Terminal()

    # Deploy the Synopsys Operator
    synopsysctl.exec(
        "deploy {} --cluster-scoped --enable-alert --enable-blackduck --enable-opssight".format(image))
    so = SynopsysOperator(h)
    so.waitUntilRunning("synopsys-operator")
    so.waitForCRDs()

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
    kubectl.exec("create ns alt-json")
    kubectl.exec("create ns alt-yaml")
    kubectl.exec("create ns alt-kube-json")
    kubectl.exec("create ns alt-kube-yaml")

    # Create Alert instances in namespaces
    kubectl.exec("create -f alt.json")
    kubectl.exec("create -f alt.yaml")
    kubectl.exec("create -f alt-kube.json")
    kubectl.exec("create -f alt-kube.yaml")

    # Wait for Alert instances to be running
    alert = Alert(h)
    alert.waitUntilRunning("alt-json")
    alert.waitUntilRunning("alt-yaml")
    alert.waitUntilRunning("alt-kube-json")
    alert.waitUntilRunning("alt-kube-yaml")

    # Delete the CRDs
    kubectl.exec("delete alert alt-json")
    kubectl.exec("delete alert alt-yaml")

    # Delete the namespaces
    kubectl.exec("delete ns alt-json")
    kubectl.exec("delete ns alt-yaml")
    kubectl.exec("delete ns alt-kube-json")
    kubectl.exec("delete ns alt-kube-yaml")

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
    h = Helper(synopsysctl_path="synopsysctl_path", kubectl_path=kubectl_path)
    synopsysctl = Synopsysctl(path=synopsysctl_path, version="latest")
    so = SynopsysOperator(h)
    alt = Alert(h)
    bd = BlackDuck(h)

    # Deploy Synopsys Operator into namespace 1
    synopsysctl.exec("deploy {} --enable-alert -n test-space1".format(image))
    so.waitUntilRunning(namespace="test-space1")
    alt.waitForCRD()

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
    h = Helper(synopsysctl_path="synopsysctl_path", kubectl_path=kubectl_path)
    synopsysctl = Synopsysctl(path=synopsysctl_path, version="latest")
    bd = BlackDuck(h)

    # Deploy Synopsys Operator
    synopsysctl.exec(
        "deploy {} --cluster-scoped --enable-blackduck".format(image))
    so = SynopsysOperator(h)
    so.waitUntilRunning("synopsys-operator")

    # Create Black Duck with Node Affinity
    bd.waitForCRD()
    synopsysctl.exec(
        "create blackduck bd --node-affinity-file-path node.json --admin-password a --postgres-password p --user-password u --persistent-storage false")
    bd.waitUntilRunning("bd")

    # Check if Documentation pod was configured correctly
    pods = h.v1.list_namespaced_pod("bd").items
    documentationPod = None
    for pod in pods:
        if "documentation" in pod.metadata.name:
            documentationPod = pod
            break
    docPodNodeAffinity = documentationPod.spec.affinity.node_affinity
    docPodNodeTerms = docPodNodeAffinity.required_during_scheduling_ignored_during_execution.node_selector_terms
    found = False
    for term in docPodNodeTerms:
        for match in term.match_experssions:
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


# test_mockAlert()
test_nodeAffinity()

import json
import logging

import pytest
import yaml
from kubernetes import client, config, utils

from ctl import Synopsysctl, Terminal
from synopsysresources import (Alert, BlackDuck, Helper, OpsSight,
                               SynopsysOperator)

SYNOPSYSCTL_PATH = "synopsysctl"
IMAGE = "gcr.io/saas-hub-stg/blackducksoftware/synopsys-operator:release-2019.6.x"


@pytest.mark.smoke
def test_smoke():
    logging.info("TEST: test_smoke")
    # SETUP BEGIN
    # Create clients
    my_helper = Helper(synopsysctl_path=SYNOPSYSCTL_PATH)
    my_synopsysctl = Synopsysctl(executable=SYNOPSYSCTL_PATH, version="latest")
    my_synopsys_operator = SynopsysOperator(helper=my_helper)
    my_alert = Alert(helper=my_helper)
    my_black_duck = BlackDuck(helper=my_helper)
    my_opssight = OpsSight(helper=my_helper)
    # Deploy Synopsys Operator
    command = f"deploy --synopsys-operator-image {IMAGE} --cluster-scoped --enable-alert --enable-blackduck --enable-opssight"
    _, err = my_synopsysctl.exec(command=command)
    if err:
        raise Exception(
            f"Synopsys Operator did not deploy with command {command}, error: {err}")
    if my_synopsys_operator.arePodsRunning("synopsys-operator") is False:
        raise Exception("Synopsys Operator pods are not running")
    # Verify Crds appear
    if my_alert.didCrdComeUp() is False:
        raise Exception("Alert CRD failed to come up")
    if my_black_duck.didCrdComeUp() is False:
        raise Exception("Black Duck CRD failed to come up")
    if my_opssight.didCrdComeUp() is False:
        raise Exception("OpsSight CRD failed to come up")
    # SETUP END

    # Create Alert
    _, err = my_synopsysctl.exec("create alert alt --persistent-storage=false")
    if err:
        raise Exception(f"Exception creating alert: {err}")
    if my_alert.arePodsRunning("alt") is False:
        raise Exception("Alert pods are not running")
    # Create Black Duck
    _, err = my_synopsysctl.exec(
        "create blackduck bd --admin-password a --postgres-password p --user-password u")
    if err:
        raise Exception(f"Exception creating Black Duck: {err}")
    if not my_black_duck.arePodsRunning("bd"):
        raise Exception("Black Duck pods are not running")
    # Create OpsSight
    _, err = my_synopsysctl.exec("create opssight ops")
    if err:
        raise Exception(f"Exception creating OpsSight: {err}")
    if not my_opssight.arePodsRunning("ops"):
        raise Exception("OpsSight pods are not running")

    # Check if CRs are created
    out, err = my_synopsysctl.exec("get alerts")
    if err:
        raise Exception(err)
    if "alt" not in out:
        raise Exception("Failed to get Alert alt")
    out, err = my_synopsysctl.exec("get blackducks")
    if err:
        raise Exception(err)
    if "bd" not in out:
        raise Exception("Failed to get Black Duck bd")
    out, err = my_synopsysctl.exec("get opssights")
    if err:
        raise Exception(err)
    if "ops" not in out:
        raise Exception("Failed to get OpsSight ops")

    # Describe each Synopsys resource
    _, err = my_synopsysctl.exec("describe alert alt")
    if err:
        raise Exception(err)
    _, err = my_synopsysctl.exec("describe blackduck bd")
    if err:
        raise Exception(err)
    _, err = my_synopsysctl.exec("describe opssight ops")
    if err:
        raise Exception(err)

    # Delete each Synopsys resource
    _, err = my_synopsysctl.exec("delete alert alt")
    if err:
        raise Exception(err)
    if my_alert.arePodsDeleted("alt") is False:
        raise Exception("Failed to delete Alert alt")

    _, err = my_synopsysctl.exec("delete blackduck bd")
    if err:
        raise Exception(err)
    if my_black_duck.arePodsDeleted("bd") is False:
        raise Exception("Failed to delete Black Duck bd")

    _, err = my_synopsysctl.exec("delete opssight ops")
    if err:
        raise Exception(err)
    if my_opssight.arePodsDeleted("ops") is False:
        raise Exception("Failed to delete OpsSight ops")

    # CLEANUP
    # Destroy Synopsys Operator
    err = my_synopsysctl.destroyDefault()
    if err:
        assert err
    if my_synopsys_operator.arePodsDeleted("synopsys-operator") is False:
        raise Exception("Failed to delete Synopsys Operator")

    if my_alert.doesCrdExist() is False:
        raise Exception("Alert CRD didn't delete")
    if my_black_duck.doesCrdExist() is False:
        raise Exception("Black Duck CRD didn't delete")
    if my_opssight.doesCrdExist() is False:
        raise Exception("OpsSight CRD didn't delete")


def main():
    try:
        config.load_kube_config()
    except:
        logging.error("FAILED to load kube config")

    client.rest.logger.setLevel("INFO")
    logging.basicConfig(level=logging.DEBUG)

    h = Helper(synopsysctl_path=SYNOPSYSCTL_PATH)
    my_synopsys_operator = SynopsysOperator(helper=h)

    try:
        test_smoke()
    except Exception as e:
        print(f"Exception in main: {e}")
    finally:
        logging.debug("sending delete namespace events to k8s api")
        h.corev1Api.delete_namespace(
            name="alt",
            body=client.V1DeleteOptions())
        h.corev1Api.delete_namespace(
            name="bd",
            body=client.V1DeleteOptions())
        h.corev1Api.delete_namespace(
            name="ops",
            body=client.V1DeleteOptions())
        h.corev1Api.delete_namespace(
            name="synopsys-operator",
            body=client.V1DeleteOptions())

        logging.debug("sending delete crd events to k8s api")
        h.apiextensionsV1beta1Api.delete_custom_resource_definition(
            name="alerts.synopsys.com",
            body=client.V1DeleteOptions())
        h.apiextensionsV1beta1Api.delete_custom_resource_definition(
            name="blackducks.synopsys.com",
            body=client.V1DeleteOptions())
        h.apiextensionsV1beta1Api.delete_custom_resource_definition(
            name="opssights.synopsys.com",
            body=client.V1DeleteOptions())

        # CLEANUP
        # Destroy Synopsys Operator
        err = h.synopsysctl.destroyDefault()
        if err:
            raise Exception(err)
        if my_synopsys_operator.arePodsDeleted("synopsys-operator") is False:
            raise Exception("Failed to delete Synopsys Operator")
        if my_synopsys_operator.didCrdsComeUp is True:
            raise Exception("Crds didn't delete")


if __name__ == '__main__':
    logging.debug("running main")
    main()

# def test_mockAlert():
#     logging.info("TEST: test_mockAlert")
#     # Create clients
#     h = Helper(synopsysctl_path=SYNOPSYSCTL_PATH)
#     synopsysctl = Synopsysctl(
#         executable=SYNOPSYSCTL_PATH, version="latest")
#     terminal = Terminal()

#     # Deploy the Synopsys Operator
#     synopsysctl.exec(
#         f"deploy --synopsys-operator-image {IMAGE} --cluster-scoped --enable-alert")
#     so = SynopsysOperator(h)
#     so.arePodsRunning("synopsys-operator")
#     if not so.didCrdsComeUp():
#         raise Exception("CRDs failed to come up")

#     # Generate files
#     altJSON, err = synopsysctl.exec(
#         "create alert alt-json --persistent-storage false --mock json")
#     if err:
#         raise Exception(err)
#     f = open("alt.json", "w")
#     f.write(altJSON)
#     f.close()

#     altYAML, err = synopsysctl.exec(
#         "create alert alt-yaml --persistent-storage false --mock yaml")
#     f = open("alt.yaml", "w")
#     f.write(altYAML)
#     f.close()

#     altKubeJSON, err = synopsysctl.exec(
#         "create alert alt-kube-json --persistent-storage false --mock-kube json")
#     f = open("alt-kube.json", "w")
#     f.write(altKubeJSON)
#     f.close()

#     altKubeYAML, err = synopsysctl.exec(
#         "create alert alt-kube-yaml --persistent-storage false --mock-kube yaml")
#     f = open("alt-kube.yaml", "w")
#     f.write(altKubeYAML)
#     f.close()

#     # Create namespaces
#     h.corev1Api.create_namespace("alt-json")
#     h.corev1Api.create_namespace("alt-yaml")
#     h.corev1Api.create_namespace("alt-kube-json")
#     h.corev1Api.create_namespace("alt-kube-yaml")

#     # Create Alert instances in namespaces
#     utils.create_from_yaml(h.apiClient, "alt.json")
#     utils.create_from_yaml(h.apiClient, "alt.yaml")
#     utils.create_from_yaml(h.apiClient, "alt-kube.json")
#     utils.create_from_yaml(h.apiClient, "alt-kube.yaml")

#     # Wait for Alert instances to be running
#     alert = Alert(h)
#     alert.arePodsRunning("alt-json")
#     alert.arePodsRunning("alt-yaml")
#     alert.arePodsRunning("alt-kube-json")
#     alert.arePodsRunning("alt-kube-yaml")

#     # Delete the CRDs
#     h.customObjectsApi.delete_cluster_custom_object(
#         "synopsys.com", "apiextensions.k8s.io/v1beta1", "alerts", "alt-json", client.V1DeleteOptions())
#     h.customObjectsApi.delete_cluster_custom_object(
#         "synopsys.com", "apiextensions.k8s.io/v1beta1", "alerts", "alt-yaml", client.V1DeleteOptions())

#     h.corev1Api.delete_namespace(
#         name="alt-json",
#         body=client.V1DeleteOptions())
#     h.corev1Api.delete_namespace(
#         name="alt-yaml",
#         body=client.V1DeleteOptions())
#     h.corev1Api.delete_namespace(
#         name="alt-kube-json",
#         body=client.V1DeleteOptions())
#     h.corev1Api.delete_namespace(
#         name="alt-kube-yaml",
#         body=client.V1DeleteOptions())

#     # Verify Alert instances deleted
#     alert.arePodsDeleted("alt-json")
#     alert.arePodsDeleted("alt-yaml")

#     # Delete the files
#     terminal.exec("rm alt.json")
#     terminal.exec("rm alt.yaml")
#     terminal.exec("rm alt-kube.json")
#     terminal.exec("rm alt-kube.yaml")

#     # Remove the Synopsys Operator
#     err = synopsysctl.destroyDefault()
#     if err:
#         raise Exception(err)
#     so.arePodsDeleted("synopsys-operator")


# def test_namespacedOperations():
#     logging.info("TEST: test_namespacedOperations")
#     # Create clients
#     h = Helper(synopsysctl_path="synopsysctl_path")
#     synopsysctl = Synopsysctl(
#         executable=SYNOPSYSCTL_PATH, version="latest")
#     so = SynopsysOperator(h)
#     alt = Alert(h)
#     bd = BlackDuck(h)

#     # Deploy Synopsys Operator into namespace 1
#     synopsysctl.exec(
#         f"deploy --synopsys-operator-image {IMAGE} --enable-alert -n test-space1")
#     so.arePodsRunning(namespace="test-space1")
#     alt.didCrdComeUp()

#     # Deploy Synopsys Operator into namespace 2
#     synopsysctl.exec("deploy --enable-blackduck -n test-space2")
#     so.arePodsRunning(namespace="test-space2")
#     bd.didCrdComeUp()

#     # Create an Alert instance in namespace 1
#     synopsysctl.exec("create alert alt -n test-space1")
#     alt.arePodsRunning(namespace="test-space1")

#     # Create a Black Duck instance in namespace 2
#     synopsysctl.exec(
#         "create blackduck bd --admin-password a --postgres-password p --user-password u -n test-space2")
#     bd.arePodsRunning(namespace="test-space2")

#     # Delete the Alert instance in namespace 1
#     synopsysctl.exec("delete alert alt -n test-space1")
#     alt.arePodsDeleted(namespace="test-space1")

#     # Delete the Black Duck instance in namespace 2
#     synopsysctl.exec("delete blackduck bd -n test-space2")
#     alt.arePodsDeleted(namespace="test-space2")

#     # Remove Synopsys Operator from namespace 1
#     synopsysctl.exec("destroy test-space1")
#     so.arePodsDeleted(namespace="test-space1")

#     # Remove Synopsys Operator from namespace 2
#     synopsysctl.exec("destroy test-space2")
#     so.arePodsDeleted(namespace="test-space2")


# def test_nodeAffinity():
#     logging.info("TEST: test_nodeAffinity")
#     # Create Helpers
#     h = Helper(synopsysctl_path=SYNOPSYSCTL_PATH)
#     synopsysctl = Synopsysctl(
#         executable=SYNOPSYSCTL_PATH, version="latest")
#     bd = BlackDuck(h)

#     # Deploy Synopsys Operator
#     synopsysctl.exec(
#         f"deploy --synopsys-operator-image {IMAGE} --cluster-scoped --enable-blackduck")
#     so = SynopsysOperator(h)
#     so.arePodsRunning("synopsys-operator")

#     # Create Black Duck with Node Affinity
#     bd.didCrdComeUp()
#     synopsysctl.exec(
#         "create blackduck bd --node-affinity-file-path ~/gocode/src/github.com/blackducksoftware/synopsys-operator/test-infra/e2e-tests/node.json --admin-password a --postgres-password p --user-password u --persistent-storage false")
#     bd.arePodsRunning("bd")

#     # Check if Documentation pod was configured correctly
#     documentationPod = h.corev1Api.list_namespaced_pod(
#         namespace="bd", label_selector='component=documentation').items[0]
#     docPodNodeAffinity = documentationPod.spec.affinity.node_affinity
#     docPodNodeTerms = docPodNodeAffinity.required_during_scheduling_ignored_during_execution.node_selector_terms
#     found = False
#     for term in docPodNodeTerms:
#         for match in term.match_expressions:
#             foundKey = match.key == "beta.kubernetes.io/arch"
#             foundOperator = match.operator == "In"
#             foundVal = match.values == ["amd64"]
#             if foundKey and foundOperator and foundVal:
#                 found = True
#                 break
#         if found:
#             break
#     if not found:
#         raise Exception("couldn't determine Node Affinity")

#     # cleanup
#     synopsysctl.exec("delete blackduck bd")
#     bd.arePodsDeleted(namespace="bd")
#     synopsysctl.exec("destroy")


# @pytest.mark.v2019_4_0
# def test_2019_4_0():
#     # test_exampleFail()
#     # test_examplePass()
#     # test_skyfire()
#     # test_smoke()
#     # test_mockAlert()
#     # test_mockBlackDuck()
#     # test_mockOpsSight()
#     # test_namespacedOperations()
#     # test_nodeAffinity()
#     pass

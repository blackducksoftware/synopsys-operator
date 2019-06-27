import json
import subprocess
import sys
import time

import requests
import yaml
from kubernetes import client, config

from synopsysresources import Helper, OpsSight

client.rest.logger.setLevel("INFO")


def skyfireTest():
    if len(sys.argv) != 5:
        print("Usage:\npython3 jenkins-script <hub-host> <cluster-ip:port> <username> <password>")
        sys.exit(1)

    # Parameters to be passed in
    namespace = "opssight-test"
    hub_host = sys.argv[1]
    cluster_ip = "https://" + sys.argv[2]
    username = sys.argv[3]
    password = sys.argv[4]
    print("namespace: %s", namespace)
    print("hub host: %s", hub_host)
    print("cluster ip: %s", cluster_ip)
    print("username: %s", username)
    print("password: %s", password)

    t = Helper()

    # Login to the Cluster
    #print("Logging In...")
    #t.clusterLogIn(cluster_ip, username, password)

    # Create Kubernetes Client
    print("Creating Kube Client...")
    config.load_kube_config()
    v1 = client.CoreV1Api()

    # Deploy the Synopsys Operator
    #operator_namespace = "synopsys-operator"
    # operator_reg_key = "abcd" # cannot be numbers
    #operator_version = "master"
    #deployOperator(operator_namespace, operator_reg_key, operator_version)

    # Create OpsSight from Yaml File
    print("Creating OpsSight...")
    opssight_namespace = "opssight-test"
    ops = OpsSight()
    ops.deploy(namespace=opssight_namespace)

    # Edit OpsSight Config to have hub url
    print("Adding Hub to OpsSight Config...")
    ops.addHubToConfig(v1, opssight_namespace, hub_host)

    # Create one instance of skyfire
    print("Creating instance of skyfire")
    ops.setSkyfireReplica(v1, opssight_namespace, 1)

    # Get the route for Skyfire
    print("Getting Skyfire Route...")
    skyfire_route = ops.getSkyfireRoute(namespace)

    # curl to start skyfire tests
    print("Starting Skyfire Tests...")
    print("Route: %s", skyfire_route)
    for _ in range(10):
        try:
            url = "http://{}/starttest".format(skyfire_route.decode("utf-8"))
            r = requests.post(url, data={'nothing': 'nothing'}, verify=False)
            print(url)
            if 200 <= r.status_code <= 299:
                break
            time.sleep(2)
        except Exception as e:
            print("Exception when starting skyfire tests: %s\n" % e)

    # curl to get skyfire results
    print("Getting Skyfire Results...")
    results = None
    try:
        for _ in range(100):
            url = "http://{}/testsuite".format(skyfire_route.decode("utf-8"))
            r = requests.get(url, verify=False)
            results = r.json()
            print(results)
            if 200 <= r.status_code <= 299:
                if results['state'] == 'FINISHED':
                    break
                else:
                    time.sleep(2)
                    continue
    except Exception as e:
        print("Exception when getting skyfire results: %s\n" % e)
        sys.exit(1)

    # Remove Skyfire Instance
    print("Removing Skyfire Instance...")
    ops.setSkyfireReplica(v1, opssight_namespace, 0)

    # Remove OpsSight Instance by deleting the namespace
    command = "kubectl delete ns opssight-test"
    print("Command: {}".format(command))
    r = subprocess.call(command, shell=True, stdout=subprocess.PIPE)
    t.isNamespaceDeleted("opssight-test")

    # print out the results
    return results['summary']


def main():
    skyfireTest()


print(main())

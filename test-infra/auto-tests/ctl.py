from kubernetes import client, config
import os
import sys
import subprocess
import yaml
import json
import time
import requests
import logging


class Terminal:
    def exec(self, command, retry=3):
        logging.debug(" > Command: {}".format(command))
        for i in range(retry):
            if i != 0:
                logging.debug(
                    "failed exec, retrying ({}/{})...".format(i, retry-1))
                time.sleep(3)
            r = subprocess.run(command, shell=True, stdout=subprocess.PIPE)
            if r.returncode == 0:
                return r.stdout.decode("utf-8"), None
        return None, r.stderr


class Ctl:
    def __init__(self, path=""):
        self.path = path
        self.terminal = Terminal()

    def exec(self, command="", retry=3):
        self.verify()
        command = "./{} {}".format(self.path, command)
        return self.terminal.exec(command=command, retry=retry)

    def verify(self):
        f = open(self.path, 'r')  # exception if binary doesn't exist
        f.close()


class Synopsysctl(Ctl):
    def __init__(self, version="latest", path="synopsysctl"):
        self.path = path
        self.version = version
        self.terminal = Terminal()

    def deployDefault(self):
        self.verify()
        command = ""
        if self.version in ["latest", "2019.4.2]"]:
            command = "deploy --cluster-scoped --enable-alert --enable-blackduck --enable-opssight"
        elif self.version in ["2019.4.1", "2019.4.0"]:
            command = "deploy"

        self.exec(command)

    def destroyDefault(self):
        self.verify()
        command = "destroy"
        self.exec(command)


class Kubectl(Ctl):
    def __init__(self, path="kubectl"):
        self.path = path
        self.terminal = Terminal()


class Oc(Ctl):
    def __init__(self, path="oc"):
        self.path = path
        self.terminal = Terminal()

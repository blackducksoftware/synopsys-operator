# Prow

Prow is a CI system that offers various features such as rich Github automation,
and running tests in Jenkins or on a Kubernetes cluster. You can read more about
Prow in [upstream docs][0].

# Creating the cluster

Create the GKE cluster, the role bindings and the GitHub secrets. For details, see <https://github.com/kubernetes/test-infra/blob/master/prow/getting_started.md>.

One secret that needs to be setup is a Github token from the bot account that is
going to manage PRs and issues. The token needs the `repo` and `read:org` scopes
enabled. The bot account also needs to be added as a collaborator in the repository
it is going to manage.

To automate the installation of Prow, store the webhook secret as `hmac` and the bot
token as `oauth` inside the `test-infra/prow` directory. Then, installing Prow involves
running the following command:
```
make prow
```

# What is installed

`hook` is installed that manages receiving webhooks from Github and reacting
appropriately on Github. `deck` is installed as the Prow frontend. Last, `tide`
is also installed that takes care of merging pull requests that pass all tests
and satisfy a set of label requirements.


# Useful commands:
```
bazel run //prow/cmd/checkconfig -- --plugin-config=/Users/bhutwala/gocode/src/github.com/blackducksoftware/synopsys-operator/test-infra/prow/plugins.yaml --config-path=/Users/bhutwala/gocode/src/github.com/blackducksoftware/synopsys-operator/test-infra/prow/config.yaml
```


# Other useful sources:
https://github.com/kubernetes/test-infra/blob/master/prow/getting_started_deploy.md
https://github.com/knative/test-infra/blob/master/ci/prow_setup.md
https://github.com/Azure/aks-engine/tree/master/.prowci

[0]: https://github.com/kubernetes/test-infra/tree/master/prow#prow

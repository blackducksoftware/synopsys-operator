Welcome to the Synopsys Operator Helm Chart!
This is an UNSUPPORTED method to deploy the Black Duck Operator.
That said, it has been tested and works. Feedback and PR are welcome. :)

To get started, configure the values file with your Black Duck configuration.

Please provide:

  • Black Duck Admin Username (default sysadmin)
  
  • Black Duck Admin Password (default YmxhY2tkdWNr)
  
  • Black Duck Registration Key (no default provided)

When ready, navigate to the directory containing the synopsys-operator folder then deploy the chart:
To avoid deploying to default namespace with helm default release name, specify these values. Example:

"helm install ./synopsys-operator --name=synopsys-operator --namespace synopsys-operator"


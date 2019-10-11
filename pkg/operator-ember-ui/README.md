# Onprem User Interface

Welcome to the Onprem User Interface repository.  

This repository contains the front-end files for the UI of Synopsysctl (a tool for managing Synopsys resources in your cluster).  

You can find the Synopsys Operator Repository [here](https://github.com/blackducksoftware/synopsys-operator).  

# Prerequisites

You will need the following things properly installed on your computer.

* [Git](https://git-scm.com/)
* [Node.js](https://nodejs.org/) (with npm)
* [Ember CLI](https://ember-cli.com/)
* [Google Chrome](https://google.com/chrome/)
* [Gobuffalo Packr](https://github.com/gobuffalo/packr)

# Quick Start

(Note: This method does not use the synopsysctl server so requests from the User Interface cannot be completed)  

## Installation and Set Up

* `git clone <repository-url>` this repository
* `cd onprem-ui`
* `npm install`

## Running / Development

* `ember build --environment production`
* `ember serve`
* Visit your app at [http://localhost:4200](http://localhost:4200).

# Serve the UI with Synopsysctl

### Download the repository into the synospys-operator repo
* `git clone <repository-url>` 
* `cd onprem-ui`
* `cp -r * ~/<path-to-synopsys-operator-repo>/pkg/operator-ember-ui/.`

### Build the User Interface into the /dist directory
* `cd ~/<path-to-synopsysctl>/pkg/operator-ember-ui/`
* `ember build --environment production`

### Store the contents of /dist into a package for the synopsysctl binary
* `cd ../..`
* `packr` (you may need to install this binary)  

### Build the synopsysctl binary with the packaged UI files
* `cd cmd/synopsysctl`
* `go build`
* `./synopsysctl serve-ui --port 8081 -v debug`
* Visit your app at [http://localhost:8081](http://localhost:8081)

# User Interace Features
### Deploy Polaris
From the `Deploy Polaris` tab the User can enter information to configure an instance of Polaris. When the `Submit` button is selected the information is sent to the back end API at /ensure_polaris.

### Update Polaris
From the `Update Polaris` tab the User can first select a Polaris instance by entering the namespace of the instance. When the `Submit` button is selected a request will be sent to the back end at /get_polaris to get information about the Polaris instance currently running in the cluster. The information will be populated into input fields that the user can update. They can then hit `Submit` which will send the updated information to the back end at /ensure_polaris. (Note: The User must always specify the license key path)  

# The Synopsysctl Server API

### The Interface for Polaris

HTTP requests between this UI and the synopsysctl server use this interface to transfer information about Polaris.  

```
polarisConfig: {
    version: "",
    environmentDNS: "",
    storageClass: "",
    namespace: "",
    imagePullSecrets: "",
    //postgresHost: "",
    //postgresPort: "",
    postgresUsername: "",
    postgresPassword: "",
    postgresSize: "",
    smtpHost: "",
    smtpPort: "",
    smtpUsername: "",
    smtpPassword: "",
    smtpSenderEmail: "",
    uploadServerSize: "",
    eventstoreSize: "",
    mongoDBSize: "",
    downloadServerSize: "",
    enableReporting: false,
    reportStorageSize: "",
    organizationDescription: "",
    organizationName: "",
    organizationAdminName: "",
    organizationAdminUsername: "",
    organizationAdminEmail: "",
    coverityLicensePath: "",
}
```

### Endpoint: /get_polaris
Functionality: Returns information about an instance of Polaris running in the cluser.  

Request: This User Interface sends a string of Polaris' namespace to the back end server.  
Response: The back end server returns the polarisConfig from above.  

### Endpoint: /ensure_polaris
Funcitonality: Ensure the the information about the Polaris instance is what is running in the cluster. It will create the instance or update the instance as necessary.  

Request: This User Interface sends the polarisConfig from above to the back end server.  
Response: The back end server does not return a response.  
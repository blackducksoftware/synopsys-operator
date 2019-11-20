/*
Copyright (C) 2019 Synopsys, Inc.

Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements. See the NOTICE file
distributed with this work for additional information
regarding copyright ownership. The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied. See the License for the
specific language governing permissions and limitations
under the License.
*/

package synopsysctl

import (
	"encoding/json"
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	// "os"
	// "time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	// blackduckv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/polaris"
	// soperator "github.com/blackducksoftware/synopsys-operator/pkg/soperator"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/gobuffalo/packr"
)

// Port to serve on
var serverPort = "8081"

// serveUICmd serves a User Interface on localhost that allows the user to interact with
// Synopsys Operator
var serveUICmd = &cobra.Command{
	Use:   "serve-ui",
	Short: "Provides a User Interface service on localhost for Synopsys Operator",
	RunE:  ServeUICmd,
}

func init() {
	// Add the serve-ui command to synopsysctl
	rootCmd.AddCommand(serveUICmd)

	// Flags for serveUICmd
	serveUICmd.Flags().StringVarP(&serverPort, "port", "p", serverPort, "Port to access the User Interface")
}

// ServeUICmd is the RunE cobra command
// This function starts a server on localhost that serves a User Interface
// for Synopsys Operator
func ServeUICmd(cmd *cobra.Command, args []string) error {
	log.Debug("serving User Interface...")

	// Pack the front-end html/css/etc. files into a "box" that can be
	// provided with the syopsysctl binary using the packr cli
	box := packr.NewBox("../operator-ui-ember/dist")

	// Create a Router to listen and serve User Interface requests
	router := mux.NewRouter()

	// // TODO: add operator for Black Duck for 12.0 release
	// // api route - deploy_operator deploys Synopsys Operator into the cluster
	// router.HandleFunc("/api/deploy_operator", func(w http.ResponseWriter, r *http.Request) {
	// 	log.Infof("handling request: %q\n", html.EscapeString(r.URL.Path))
	// 	reqBody, err := ioutil.ReadAll(r.Body)
	// 	if err != nil {
	// log.Errorf("failed to read request body: %+v", err)
	// return
	// 	}
	// 	log.Debugf("request data: %s", reqBody)
	// 	err = deployOperatorRequest(reqBody)
	// 	if err != nil {
	// 		log.Errorf("error deploying Synopsys Operator: %s\n", err)
	// 		return
	// 	}
	// 	log.Infof("successfully deployed Synopsys Operator")
	// })

	// api route - deploy_polaris deploys Polaris into the cluster
	router.HandleFunc("/api/ensure_polaris", func(w http.ResponseWriter, r *http.Request) {
		log.Infof("handling request: %q\n", html.EscapeString(r.URL.Path))
		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
			errMsg := fmt.Sprintf("%s", err)
			http.Error(w, errMsg, 404)
		}

		pConfig := PolarisUIRequestResponse{}
		err = json.Unmarshal(reqBody, &pConfig)
		if err != nil {
			errMsg := fmt.Sprintf("failed to unmarshal Polaris request: %s", err)
			log.Errorf(errMsg)
			http.Error(w, errMsg, 400)
			return
		}

		// Check Secret to see if polaris already exists so a new instance needs to be made
		if pConfig.Namespace == "" {
			pConfig.Namespace = "default"
		}
		namespace = pConfig.Namespace // set global namespace value for getPolarisFromSecret and ensurePolaris
		oldPolaris, err := getPolarisFromSecret()
		if err != nil {
			errMsg := fmt.Sprintf("failed to get Secret for Polaris: %s", err)
			log.Errorf(errMsg)
			http.Error(w, errMsg, 404)
			return
		}
		requestIsUpdatingPolaris := oldPolaris != nil

		defaultPolarisObj := polaris.GetPolarisDefault()
		polarisObj, err := convertPolarisUIResponseToPolarisObject(defaultPolarisObj, pConfig, requestIsUpdatingPolaris)
		if err != nil {
			errMsg := fmt.Sprintf("failed to convert Request data to Polaris Object: %s", err)
			log.Errorf(errMsg)
			http.Error(w, errMsg, 422)
			return
		}

		err = checkPolarisRequestFields(pConfig, oldPolaris, requestIsUpdatingPolaris)
		if err != nil {
			errMsg := fmt.Sprintf("invalid request fields: %s", err)
			log.Errorf(errMsg)
			http.Error(w, errMsg, 422)
			return
		}

		err = ensurePolaris(polarisObj, requestIsUpdatingPolaris)
		if err != nil {
			errMsg := fmt.Sprintf("error ensuring Polaris: %s", err)
			log.Errorf(errMsg)
			http.Error(w, errMsg, 500)
			return
		}

		log.Infof("successfully ensured Polaris")
	})

	// api route - get_polaris returns Polaris specification to the client
	router.HandleFunc("/api/get_polaris", func(w http.ResponseWriter, r *http.Request) {
		log.Infof("handling request: %q\n", html.EscapeString(r.URL.Path))
		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			errMsg := fmt.Sprintf("failed to read request body: %+v", err)
			log.Errorf(errMsg)
			http.Error(w, errMsg, 400)
			return
		}

		// Get Polaris from secret
		namespace = string(reqBody)
		polarisObj, err := getPolarisFromSecret()
		if err != nil {
			errMsg := fmt.Sprintf("error getting Polaris: %s", err)
			log.Errorf(errMsg)
			http.Error(w, errMsg, 500)
			return
		}
		if polarisObj == nil {
			errMsg := fmt.Sprintf("there is no Secret data for a Polaris instance in namespace '%s'", namespace)
			log.Errorf(errMsg)
			http.Error(w, errMsg, 404)
			return
		}
		response, err := convertPolarisObjToUIResponse(*polarisObj)
		if err != nil {
			errMsg := fmt.Sprintf("error converting PolarisObj to UI Response: %s", err)
			log.Errorf(errMsg)
			http.Error(w, errMsg, 500)
			return
		}
		responseByte, err := json.Marshal(response)
		if err != nil {
			errMsg := fmt.Sprintf("error converting Polaris to bytes: %s", err)
			log.Errorf(errMsg)
			http.Error(w, errMsg, 500)
			return
		}
		w.Write(responseByte)
	})

	// api route - get_polaris_defaults returns Polaris default specifications to the client
	router.HandleFunc("/api/get_polaris_defaults", func(w http.ResponseWriter, r *http.Request) {
		log.Infof("handling request: %q\n", html.EscapeString(r.URL.Path))
		polarisObj := polaris.GetPolarisDefault()
		response, err := convertPolarisObjToUIResponse(*polarisObj)
		if err != nil {
			errMsg := fmt.Sprintf("error converting PolarisObj to UI Response: %s", err)
			log.Errorf(errMsg)
			http.Error(w, errMsg, 500)
			return
		}
		responseByte, err := json.Marshal(response)
		if err != nil {
			errMsg := fmt.Sprintf("error converting Polaris to bytes: %s", err)
			log.Errorf(errMsg)
			http.Error(w, errMsg, 500)
			return
		}
		w.Write(responseByte)
	})

	// TODO: add Black Duck to UI for 12.0 release
	// api route - deploy_black_duck deploys Black Duck into the cluster
	// router.HandleFunc("/api/deploy_black_duck", func(w http.ResponseWriter, r *http.Request) {
	// 	log.Infof("handling request: %q\n", html.EscapeString(r.URL.Path))
	// 	reqBody, err := ioutil.ReadAll(r.Body)
	// 	if err != nil {
	// log.Errorf("failed to read request body: %+v", err)
	// return
	// 	}
	// 	log.Debugf("request data: %s", reqBody)
	// 	err = createBlackDuckCRRequest(reqBody)
	// 	if err != nil {
	// 		log.Errorf("error creating Black Duck: %s", err)
	// 		return
	// 	}
	// 	log.Infof("successfully created Black Duck")
	// })

	// Create a Handler to serve the front-end files that are accessed
	// through the index route /
	spa := spaHandler{
		box: &box,
	}
	router.PathPrefix("/").Handler(spa)

	// Display helpful information for the user to interact with
	// the User Interface
	fmt.Printf("==================================\n")
	fmt.Printf("Serving at: http://localhost:%s\n", serverPort)
	fmt.Printf("api:\n")
	// fmt.Printf("  - /api/deploy_operator\n") // TODO: add operator for Black Duck for 12.0 release
	fmt.Printf("  - /api/ensure_polaris\n")
	fmt.Printf("  - /api/get_polaris\n")
	fmt.Printf("  - /api/get_polaris_defaults\n")
	// fmt.Printf("  - /api/deploy_black_duck\n") // TODO: add Black Duck to UI for 12.0 release
	fmt.Printf("==================================\n")
	fmt.Printf("\n")

	// Serving the server
	// TODO - the UI team needs to verify that these permissions are okay
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", serverPort), handlers.CORS(handlers.AllowedHeaders([]string{"X-Requested-With", "Content-type", "Authorization"}), handlers.AllowedMethods([]string{"GET", "POST", "PUT", "HEAD", "OPTIONS"}), handlers.AllowedOrigins([]string{"*"}))(router)))

	return nil
}

// Modified From: https://github.com/gorilla/mux#serving-single-page-applications
// spaHandler implements the http.Handler interface, so we can use it
// to respond to HTTP requests. The path to the static directory and
// path to the index file within that static directory are used to
// serve the SPA in the given static directory.
type spaHandler struct {
	box *packr.Box
}

// Modified From: https://github.com/gorilla/mux#serving-single-page-applications
// ServeHTTP inspects the URL path to locate a file within the static dir
// on the SPA handler. If a file is found, it will be served. If not, the
// file located at the index path on the SPA handler will be served. This
// is suitable behavior for serving an SPA (single page application).
func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// get the absolute path to prevent directory traversal
	path, err := filepath.Abs(r.URL.Path)
	if err != nil {
		// if we failed to get the absolute path respond with a 400 bad request
		// and stop
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// check whether a file exists at the given path
	if !h.box.Has(path) {
		// file does not exist, serve index.html
		file, err := h.box.Open("index.html")
		if err != nil {
			log.Errorf("failed to read index file from box")
		}
		http.ServeContent(w, r, path, time.Now(), file)
		return
	}

	// otherwise, use http.FileServer to serve the static dir
	http.FileServer(h.box).ServeHTTP(w, r)
}

/* USER INTERFACE REQUEST DATA STRUCTS */

// // TODO: add operator for Black Duck for 12.0 release
// // DeployOperatorUIRequestConfig represents the format that
// // the front-end should send its data for deploying
// // Synopsys Operator
// type DeployOperatorUIRequestConfig struct {
// 	Namespace       string `json:"namespace"`
// 	ClusterScoped   bool   `json:"clusterScoped"`
// 	EnableAlert     bool   `json:"enableAlert"`
// 	EnableBlackDuck bool   `json:"enableBlackDuck"`
// 	EnableOpsSight  bool   `json:"enableOpsSight"`
// 	EnablePolaris   bool   `json:"enablePolaris"`
// 	ExposeMetrics   string `json:"exposeMetrics"`
// 	ExposeUI        string `json:"exposeUI"`
// 	MetricsImage    string `json:"metricsImage"`
// 	OperatorImage   string `json:"operatorImage"`
// }

// PolarisUIRequestResponse represents the format that
// the front-end and back-end can communicate Polaris data
type PolarisUIRequestResponse struct {
	Version   string `json:"version"`
	Namespace string `json:"namespace"`

	FullyQualifiedDomainName string `json:"fullyQualifiedDomainName"`
	IngressClass             string `json:"ingressClass"`

	GCPServiceAccountPath string `json:"GCPServiceAccountPath"`
	PolarisLicensePath    string `json:"polarisLicensePath"`
	CoverityLicensePath   string `json:"coverityLicensePath"`

	SMTPHost                 string `json:"smtpHost"`
	SMTPPort                 string `json:"smtpPort"`
	SMTPUsername             string `json:"smtpUsername"`
	SMTPPassword             string `json:"smtpPassword"`
	SMTPSenderEmail          string `json:"smtpSenderEmail"`
	SMTPTLSTrustedHosts      string `json:"smtpTLSTrustedHosts"`
	SMTPTlsIgnoreInvalidCert bool   `json:"smtpTlsIgnoreInvalidCert"`

	PostgresHost            string `json:"postgresHost"`
	PostgresPort            string `json:"postgresPort"`
	PostgresSSLMode         string `json:"postgresSSLMode"`
	PostgresUsername        string `json:"postgresUsername"`
	PostgresPassword        string `json:"postgresPassword"`
	EnablePostgresContainer bool   `json:"enablePostgresContainer"`
	PostgresSize            string `json:"postgresSize"`

	StorageClass string `json:"storageClass"`

	UploadServerSize   string `json:"uploadServerSize"`
	EventstoreSize     string `json:"eventstoreSize"`
	MongoDBSize        string `json:"mongoDBSize"`
	DownloadServerSize string `json:"downloadServerSize"`

	EnableReporting   bool   `json:"enableReporting"`
	ReportStorageSize string `json:"reportStorageSize"`

	OrganizationDescription   string `json:"organizationDescription"`
	OrganizationAdminName     string `json:"organizationAdminName"`
	OrganizationAdminUsername string `json:"organizationAdminUsername"`
	OrganizationAdminEmail    string `json:"organizationAdminEmail"`

	ImagePullSecrets string `json:"imagePullSecrets"`
	Registry         string `json:"registry"`
}

// // TODO: add Black Duck to UI for 12.0 release
// // BlackDuckUIRequest represents the format that
// // the front-end should send its data for deploying
// // Black Duck
// type BlackDuckUIRequest struct {
// 	Name                              string `json:"name"`
// 	Namespace                         string `json:"namespace"`
// 	Version                           string `json:"version"`
// 	LicenseKey                        string `json:"licenseKey"`
// 	DbMigrate                         bool   `json:"dbMigrate"`
// 	Size                              string `json:"size"`
// 	ExposeService                     string `json:"exposeService"`
// 	BlackDuckType                     string `json:"blackDuckType"`
// 	UseBinaryUploads                  bool   `json:"useBinaryUploads"`
// 	EnableSourceUploads               bool   `json:"enableSourceUploads"`
// 	LivenessProbes                    bool   `json:"livenessProbes"`
// 	PersistentStorage                 bool   `json:"persistentStorage"`
// 	CloneDB                           string `json:"cloneDB"`
// 	PVCStorageClass                   string `json:"PVCStorageClass"`
// 	ScanType                          string `json:"scanType"`
// 	ExternalDatabase                  bool   `json:"externalDatabase"`
// 	ExternalPostgresSQLHost           string `json:"externalPostgresSQLHost"`
// 	ExternalPostgresSQLPort           string `json:"externalPostgresSQLPort"`
// 	ExternalPostgresSQLAdminUser      string `json:"externalPostgresSQLAdminUser"`
// 	ExternalPostgresSQLAdminPassword  string `json:"externalPostgresSQLAdminPassword"`
// 	ExternalPostgresSQLUser           string `json:"externalPostgresSQLUser"`
// 	ExternalPostgresSQLUserPassword   string `json:"externalPostgresSQLUserPassword"`
// 	EnableSSL                         bool   `json:"enableSSL"`
// 	PostgresSQLUserPassword           string `json:"postgresSQLUserPassword"`
// 	PostgresSQLAdminPassword          string `json:"postgresSQLAdminPassword"`
// 	PostgresSQLPostgresPassword       string `json:"postgresSQLPostgresPassword"`
// 	CertificateName                   string `json:"certificateName"`
// 	CustomCACertificateAuthentication bool   `json:"customCACertificateAuthentication"`
// 	ProxyRootCertificate              string `json:"proxyRootCertificate"`
// 	ContainerImageTags                string `json:"containerImageTags"`
// 	EnvironmentVariables              string `json:"environmentVariables"`
// 	NodeAffinityJSON                  string `json:"nodeAffinityJSON"`
// }

/* COMMANDS TO DEPLOY/CREATE RESOURCES FROM UI REQUEST STRUCTS */

// // TODO: add operator for Black Duck for 12.0 release
// // createPolarisCRsRequest deploys Synopsys Operator into your cluster by
// // populating the global variables used by cmd_deploy and then running
// // the cmd_deploy logic
// // data -  a []byte representation of DeployOperatorUIRequestConfig
// func deployOperatorRequest(data []byte) error {
// 	oConfig := DeployOperatorUIRequestConfig{}
// 	err := json.Unmarshal(data, &oConfig)
// 	if err != nil {
// 		return fmt.Errorf("failed to unmarshal Synopsys Operator request: %s", err)
// 	}
// 	// Set Global Flags for cmd_deploy - use defaults for other flags
// 	if oConfig.Namespace != "" {
// 		operatorNamespace = oConfig.Namespace
// 	}
// 	if oConfig.ExposeUI != "" {
// 		exposeUI = oConfig.ExposeUI
// 	}
// 	if oConfig.OperatorImage != "" {
// 		synopsysOperatorImage = oConfig.OperatorImage
// 	}
// 	if oConfig.MetricsImage != "" {
// 		metricsImage = oConfig.MetricsImage
// 	}
// 	if oConfig.ExposeMetrics != "" {
// 		exposeMetrics = oConfig.ExposeMetrics
// 	}
// 	isEnabledAlert = oConfig.EnableAlert
// 	isEnabledBlackDuck = oConfig.EnableBlackDuck
// 	isEnabledOpsSight = oConfig.EnableOpsSight
// 	isEnabledPrm = oConfig.EnablePolaris
// 	isClusterScoped = oConfig.ClusterScoped

// 	/* TODO: Change from being a duplicate of cmd_deploy */

// 	// validate each CRD enable parameter is enabled/disabled and cluster scope are from supported values
// 	crds, err := getEnabledCrds()
// 	if err != nil {
// 		return err
// 	}

// 	//// Add Size CRD
// 	// TODO
// 	//crds = append(crds, util.SizeCRDName)

// 	// Get CRD configs
// 	crdConfigs, err := getCrdConfigs(operatorNamespace, isClusterScoped, crds)
// 	if err != nil {
// 		return err
// 	}
// 	if len(crdConfigs) <= 0 {
// 		return fmt.Errorf("no resources are enabled (include flag(s): --enable-alert --enable-blackduck --enable-opssight --enable-polaris )")
// 	}
// 	// Create Synopsys Operator Spec
// 	soperatorSpec, err := getSpecToDeploySOperator(crds)
// 	if err != nil {
// 		return err
// 	}

// 	// check if namespace exist in namespace scope, if not throw an error
// 	if !isClusterScoped {
// 		_, err = kubeClient.CoreV1().Namespaces().Get(operatorNamespace, metav1.GetOptions{})
// 		if err != nil {
// 			return fmt.Errorf("please create the namespace '%s' to deploy the Synopsys Operator in namespace scoped", operatorNamespace)
// 		}
// 	}

// 	// check if operator is already installed
// 	_, err = util.GetOperatorNamespace(kubeClient, operatorNamespace)
// 	if err == nil {
// 		return fmt.Errorf("the Synopsys Operator instance is already deployed in namespace '%s'", namespace)
// 	}

// 	log.Infof("deploying Synopsys Operator in namespace '%s'...", operatorNamespace)

// 	log.Debugf("creating custom resource definitions")
// 	err = deployCrds(operatorNamespace, isClusterScoped, crdConfigs)
// 	if err != nil {
// 		return err
// 	}

// 	log.Debugf("creating Synopsys Operator components")
// 	sOperatorCreater := soperator.NewCreater(false, restconfig, kubeClient)
// 	err = sOperatorCreater.UpdateSOperatorComponents(soperatorSpec)
// 	if err != nil {
// 		return fmt.Errorf("error deploying Synopsys Operator due to %+v", err)
// 	}

// 	log.Infof("successfully submitted Synopsys Operator into namespace '%s'", operatorNamespace)
// 	return nil
// }

// checkPolarisRequestFields returns an error if a required field is missing from
// a request from the front end UI
func checkPolarisRequestFields(polarisUIRequestConfig PolarisUIRequestResponse, oldPolaris *polaris.Polaris, updating bool) error {
	// Check required UI request fields
	if polarisUIRequestConfig.Version == "" {
		return fmt.Errorf("field required: Version")
	}
	if polarisUIRequestConfig.FullyQualifiedDomainName == "" {
		return fmt.Errorf("field required: FullyQualifiedDomainName")
	}

	if polarisUIRequestConfig.SMTPHost == "" {
		return fmt.Errorf("field required: SMTPHost")
	}
	if polarisUIRequestConfig.SMTPPort == "" {
		return fmt.Errorf("field required: SMTPPort")
	}
	if polarisUIRequestConfig.SMTPUsername == "" {
		return fmt.Errorf("field required: SMTPUsername")
	}
	if polarisUIRequestConfig.SMTPPassword == "" {
		return fmt.Errorf("field required: SMTPPassword")
	}
	if polarisUIRequestConfig.SMTPSenderEmail == "" {
		return fmt.Errorf("field required: SMTPSenderEmail")
	}

	if polarisUIRequestConfig.OrganizationDescription == "" {
		return fmt.Errorf("field required: OrganizationDescription")
	}
	if polarisUIRequestConfig.OrganizationAdminName == "" {
		return fmt.Errorf("field required: OrganizationAdminName")
	}
	if polarisUIRequestConfig.OrganizationAdminUsername == "" {
		return fmt.Errorf("field required: OrganizationAdminUsername")
	}
	if polarisUIRequestConfig.OrganizationAdminEmail == "" {
		return fmt.Errorf("field required: OrganizationAdminEmail")
	}
	if polarisUIRequestConfig.Namespace == "" {
		return fmt.Errorf("field required: Namespace")
	}
	if !updating { // not required if updating a polaris instance
		if polarisUIRequestConfig.PolarisLicensePath == "" {
			return fmt.Errorf("field required: PolarisLicensePath")
		}
		if polarisUIRequestConfig.CoverityLicensePath == "" {
			return fmt.Errorf("field required: CoverityLicensePath")
		}
		if polarisUIRequestConfig.GCPServiceAccountPath == "" {
			return fmt.Errorf("field required: GCPServiceAccountPath")
		}
	}

	if !polarisUIRequestConfig.EnablePostgresContainer {
		if polarisUIRequestConfig.PostgresHost == "" {
			return fmt.Errorf("field required when using an external Postgres database: PostgresHost")
		}
		if polarisUIRequestConfig.PostgresPort == "" {
			return fmt.Errorf("field required when using an external Postgres database: PostgresPort")
		}
		if polarisUIRequestConfig.PostgresUsername == "" {
			return fmt.Errorf("field required when using an external Postgres database: PostgresUsername")
		}
	}
	if polarisUIRequestConfig.PostgresPassword == "" {
		return fmt.Errorf("field required: PostgresPassword")
	}

	// if updating and reporting size was changed
	if updating && polarisUIRequestConfig.ReportStorageSize != oldPolaris.ReportingSpec.ReportStorageDetails.Storage.StorageSize {
		// if enableReporting was not "turned on"
		if !(polarisUIRequestConfig.EnableReporting == true && oldPolaris.EnableReporting == false) {
			return fmt.Errorf("reporting size cannot be updated, it can only be set when enabling reporting for the first time")
		}
	}

	return nil
}

// convertPolarisUIResponseToPolarisObject maps the values in a PolarisUIRequestResponse to values
// in a Polaris Spec for deployment. It handles the values for updating and creating a Polaris instance.
// - polarisObj: the default Polaris Spec or the Polaris Spec being updated
// - polarisUIRequestConfig: the values received from the UI
// - updating: if true then updating; if false then creating
//   - updating: some values should be overwritten if the value provided is empty (ex: turning off a value)
//   - creating: default values should not be overwritten if non-required values are empty
func convertPolarisUIResponseToPolarisObject(polarisObj *polaris.Polaris, polarisUIRequestConfig PolarisUIRequestResponse, updating bool) (*polaris.Polaris, error) {
	// Set the Namespace to default if one is not provided
	if polarisUIRequestConfig.Namespace == "" {
		polarisObj.Namespace = "default"
	} else {
		polarisObj.Namespace = polarisUIRequestConfig.Namespace
	}

	polarisObj.Version = polarisUIRequestConfig.Version
	polarisObj.EnvironmentDNS = polarisUIRequestConfig.FullyQualifiedDomainName

	// If the user is not updating (aka creating) then the path is required so it is set here
	// If the user is updating then only overwrite the old value if the user provides a path
	// - Note: synopsysctl does not save the path information so it isn't given to the User in the UI, thus
	//   if this field is empty during an update it means they want the pervious value
	if !updating || polarisUIRequestConfig.GCPServiceAccountPath != "" {
		GCPServiceAccount, err := util.ReadFileData(polarisUIRequestConfig.GCPServiceAccountPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read GCP Service Account from file: %s", err)
		}
		polarisObj.GCPServiceAccount = GCPServiceAccount
	}
	if !updating || polarisUIRequestConfig.CoverityLicensePath != "" {
		CoverityLicense, err := util.ReadFileData(polarisUIRequestConfig.CoverityLicensePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read coverity license from file: %+v", err)
		}
		polarisObj.Licenses.Coverity = CoverityLicense
	}
	if !updating || polarisUIRequestConfig.PolarisLicensePath != "" {
		PolarisLicense, err := util.ReadFileData(polarisUIRequestConfig.PolarisLicensePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read polaris license from file: %+v", err)
		}
		polarisObj.Licenses.Polaris = PolarisLicense
	}

	// If the user is updating then overwrite the old values even if these fields are empty
	// If the user is creating then only overwrite the default value if the new value is not empty
	if updating || polarisUIRequestConfig.ImagePullSecrets != "" {
		polarisObj.ImagePullSecrets = polarisUIRequestConfig.ImagePullSecrets
	}
	if updating || polarisUIRequestConfig.Registry != "" {
		polarisObj.Registry = polarisUIRequestConfig.Registry
	}
	if updating || polarisUIRequestConfig.StorageClass != "" {
		polarisObj.StorageClass = polarisUIRequestConfig.StorageClass
	}
	if updating || polarisUIRequestConfig.IngressClass != "" {
		polarisObj.IngressClass = polarisUIRequestConfig.IngressClass
	}

	// CONFIGURE POSTGRES
	polarisObj.PolarisDBSpec.PostgresDetails.IsInternal = polarisUIRequestConfig.EnablePostgresContainer
	if polarisUIRequestConfig.EnablePostgresContainer { // Using the Postgres provided by synopsysctl (internal)
		polarisObj.PolarisDBSpec.PostgresDetails.SSLMode = polaris.PostgresSSLModeDisable
		if polarisUIRequestConfig.PostgresSize != "" {
			polarisObj.PolarisDBSpec.PostgresDetails.Storage.StorageSize = polarisUIRequestConfig.PostgresSize
		}
	} else { // Using the Postgres provided by the user (external)
		polarisObj.PolarisDBSpec.PostgresDetails.Host = polarisUIRequestConfig.PostgresHost
		var postPort int64
		var err error
		if polarisUIRequestConfig.PostgresPort != "" {
			postPort, err = strconv.ParseInt(polarisUIRequestConfig.PostgresPort, 0, 64)
			if err != nil {
				return nil, fmt.Errorf("falied to convert postgres port to an int: %s", polarisUIRequestConfig.PostgresPort)
			}
		}
		polarisObj.PolarisDBSpec.PostgresDetails.Port = int(postPort)
		polarisObj.PolarisDBSpec.PostgresDetails.Username = polarisUIRequestConfig.PostgresUsername
		switch polaris.PostgresSSLMode(polarisUIRequestConfig.PostgresSSLMode) {
		case polaris.PostgresSSLModeDisable:
			polarisObj.PolarisDBSpec.PostgresDetails.SSLMode = polaris.PostgresSSLModeDisable
		//case polaris.PostgresSSLModeAllow:
		//	polarisObj.PolarisDBSpec.PostgresDetails.SSLMode = polaris.PostgresSSLModeAllow
		//case polaris.PostgresSSLModePrefer:
		//	polarisObj.PolarisDBSpec.PostgresDetails.SSLMode = polaris.PostgresSSLModePrefer
		case polaris.PostgresSSLModeRequire:
			polarisObj.PolarisDBSpec.PostgresDetails.SSLMode = polaris.PostgresSSLModeRequire
		default:
			return nil, fmt.Errorf("%s is an invalid value", polarisUIRequestConfig.PostgresSSLMode)
		}
	}
	polarisObj.PolarisDBSpec.PostgresDetails.Password = polarisUIRequestConfig.PostgresPassword

	// CONFIGURE SMTP - SMTPHost, SMTPPort, SMTPUsername, SMTPPassword, SMTPSenderEmail are always required
	polarisObj.PolarisDBSpec.SMTPDetails.Host = polarisUIRequestConfig.SMTPHost
	var sPort int64
	var err error
	if polarisUIRequestConfig.SMTPPort != "" {
		sPort, err = strconv.ParseInt(polarisUIRequestConfig.SMTPPort, 0, 64)
		if err != nil {
			return nil, fmt.Errorf("falied to convert smtp port to an int: %s", polarisUIRequestConfig.SMTPPort)
		}
	}
	polarisObj.PolarisDBSpec.SMTPDetails.Port = int(sPort)
	polarisObj.PolarisDBSpec.SMTPDetails.Username = polarisUIRequestConfig.SMTPUsername
	polarisObj.PolarisDBSpec.SMTPDetails.Password = polarisUIRequestConfig.SMTPPassword
	polarisObj.PolarisDBSpec.SMTPDetails.SenderEmail = polarisUIRequestConfig.SMTPSenderEmail

	polarisObj.PolarisDBSpec.SMTPDetails.TLSTrustedHosts = polarisUIRequestConfig.SMTPTLSTrustedHosts
	polarisObj.PolarisDBSpec.SMTPDetails.TLSCheckServerIdentity = !polarisUIRequestConfig.SMTPTlsIgnoreInvalidCert

	// CONFIGURE STORAGE
	// use defaults/previous values if they aren't provided
	if polarisUIRequestConfig.UploadServerSize != "" {
		polarisObj.PolarisDBSpec.UploadServerDetails.Storage.StorageSize = polarisUIRequestConfig.UploadServerSize
	}
	if polarisUIRequestConfig.EventstoreSize != "" {
		polarisObj.PolarisDBSpec.EventstoreDetails.Storage.StorageSize = polarisUIRequestConfig.EventstoreSize
	}
	if polarisUIRequestConfig.MongoDBSize != "" {
		polarisObj.PolarisDBSpec.MongoDBDetails.Storage.StorageSize = polarisUIRequestConfig.MongoDBSize
	}
	if polarisUIRequestConfig.DownloadServerSize != "" {
		polarisObj.PolarisSpec.DownloadServerDetails.Storage.StorageSize = polarisUIRequestConfig.DownloadServerSize
	}

	// CONFIGURE REPORTING
	polarisObj.EnableReporting = polarisUIRequestConfig.EnableReporting
	if polarisUIRequestConfig.EnableReporting {
		// use default/precious value if it isn't provided
		if polarisUIRequestConfig.ReportStorageSize != "" {
			polarisObj.ReportingSpec.ReportStorageDetails.Storage.StorageSize = polarisUIRequestConfig.ReportStorageSize
		}
	}

	// CONFIGURE ORGANIZATION - these fields are always required
	polarisObj.OrganizationDetails.OrganizationProvisionOrganizationDescription = polarisUIRequestConfig.OrganizationDescription
	polarisObj.OrganizationDetails.OrganizationProvisionAdminName = polarisUIRequestConfig.OrganizationAdminName
	polarisObj.OrganizationDetails.OrganizationProvisionAdminUsername = polarisUIRequestConfig.OrganizationAdminUsername
	polarisObj.OrganizationDetails.OrganizationProvisionAdminEmail = polarisUIRequestConfig.OrganizationAdminEmail

	return polarisObj, nil
}

// convertPolarisObjToUIResponse converts the Polaris Object Specification into the Response format
// that can be sent to the User Interface
func convertPolarisObjToUIResponse(polarisObj polaris.Polaris) (*PolarisUIRequestResponse, error) {
	polarisUIRequestConfig := &PolarisUIRequestResponse{}

	// Populate Polaris Config Fields
	polarisUIRequestConfig.Namespace = polarisObj.Namespace
	polarisUIRequestConfig.Version = polarisObj.Version
	polarisUIRequestConfig.FullyQualifiedDomainName = polarisObj.EnvironmentDNS

	// synopsysctl does not save the file paths - the data is not sent to the UI
	polarisUIRequestConfig.GCPServiceAccountPath = ""
	polarisUIRequestConfig.CoverityLicensePath = ""
	polarisUIRequestConfig.PolarisLicensePath = ""

	polarisUIRequestConfig.ImagePullSecrets = polarisObj.ImagePullSecrets
	polarisUIRequestConfig.Registry = polarisObj.Registry
	polarisUIRequestConfig.StorageClass = polarisObj.StorageClass
	polarisUIRequestConfig.IngressClass = polarisObj.IngressClass

	// POSTGRES
	polarisUIRequestConfig.PostgresHost = polarisObj.PolarisDBSpec.PostgresDetails.Host
	polarisUIRequestConfig.PostgresPort = strconv.FormatInt(int64(polarisObj.PolarisDBSpec.PostgresDetails.Port), 10)
	polarisUIRequestConfig.PostgresUsername = polarisObj.PolarisDBSpec.PostgresDetails.Username
	polarisUIRequestConfig.PostgresPassword = polarisObj.PolarisDBSpec.PostgresDetails.Password
	polarisUIRequestConfig.PostgresSize = polarisObj.PolarisDBSpec.PostgresDetails.Storage.StorageSize
	polarisUIRequestConfig.EnablePostgresContainer = polarisObj.PolarisDBSpec.PostgresDetails.IsInternal
	polarisUIRequestConfig.PostgresSSLMode = string(polarisObj.PolarisDBSpec.PostgresDetails.SSLMode)

	// SMTP
	polarisUIRequestConfig.SMTPHost = polarisObj.PolarisDBSpec.SMTPDetails.Host
	polarisUIRequestConfig.SMTPPort = strconv.FormatInt(int64(polarisObj.PolarisDBSpec.SMTPDetails.Port), 10)
	polarisUIRequestConfig.SMTPUsername = polarisObj.PolarisDBSpec.SMTPDetails.Username
	polarisUIRequestConfig.SMTPPassword = polarisObj.PolarisDBSpec.SMTPDetails.Password
	polarisUIRequestConfig.SMTPSenderEmail = polarisObj.PolarisDBSpec.SMTPDetails.SenderEmail
	polarisUIRequestConfig.SMTPTLSTrustedHosts = polarisObj.PolarisDBSpec.SMTPDetails.TLSTrustedHosts
	polarisUIRequestConfig.SMTPTlsIgnoreInvalidCert = !polarisObj.PolarisDBSpec.SMTPDetails.TLSCheckServerIdentity

	// STORAGE
	polarisUIRequestConfig.UploadServerSize = polarisObj.PolarisDBSpec.UploadServerDetails.Storage.StorageSize
	polarisUIRequestConfig.EventstoreSize = polarisObj.PolarisDBSpec.EventstoreDetails.Storage.StorageSize
	polarisUIRequestConfig.MongoDBSize = polarisObj.PolarisDBSpec.MongoDBDetails.Storage.StorageSize
	polarisUIRequestConfig.DownloadServerSize = polarisObj.PolarisSpec.DownloadServerDetails.Storage.StorageSize

	// REPORTING
	polarisUIRequestConfig.EnableReporting = polarisObj.EnableReporting
	polarisUIRequestConfig.ReportStorageSize = polarisObj.ReportingSpec.ReportStorageDetails.Storage.StorageSize

	// ORGANIZATIONS
	polarisUIRequestConfig.OrganizationDescription = polarisObj.OrganizationDetails.OrganizationProvisionOrganizationDescription
	polarisUIRequestConfig.OrganizationAdminName = polarisObj.OrganizationDetails.OrganizationProvisionAdminName
	polarisUIRequestConfig.OrganizationAdminUsername = polarisObj.OrganizationDetails.OrganizationProvisionAdminUsername
	polarisUIRequestConfig.OrganizationAdminEmail = polarisObj.OrganizationDetails.OrganizationProvisionAdminEmail

	return polarisUIRequestConfig, nil
}

// // TODO: add Black Duck to UI for 12.0 release
// // createBlackDuckCRRequest generates the CR for Black Duck from the request data
// // provided by the User Interface
// // data -  a []byte representation of BlackDuckUIRequest
// func createBlackDuckCRRequest(data []byte) error {
// 	bdConfig := BlackDuckUIRequest{}
// 	err := json.Unmarshal(data, &bdConfig)
// 	if err != nil {
// 		return fmt.Errorf("failed to unmarshal Black Duck request: %s", err)
// 	}
// 	blackDuckCR := &blackduckv1.Blackduck{
// 		ObjectMeta: v1.ObjectMeta{
// 			Name:      bdConfig.Name,
// 			Namespace: bdConfig.Namespace,
// 		},
// 		Spec: blackduckv1.BlackduckSpec{
// 			Namespace:         bdConfig.Namespace,
// 			Size:              bdConfig.Size,
// 			Version:           bdConfig.Version,
// 			ExposeService:     bdConfig.ExposeService,
// 			PVCStorageClass:   bdConfig.PVCStorageClass,
// 			LivenessProbes:    bdConfig.LivenessProbes,
// 			ScanType:          bdConfig.ScanType,
// 			PersistentStorage: bdConfig.PersistentStorage,
// 			LicenseKey:        bdConfig.LicenseKey,
// 			AdminPassword:     bdConfig.PostgresSQLAdminPassword,
// 			UserPassword:      bdConfig.PostgresSQLUserPassword,
// 			PostgresPassword:  bdConfig.PostgresSQLPostgresPassword,
// 		},
// 	}
// 	blackDuckCR.Kind = "Blackduck"
// 	blackDuckCR.APIVersion = "synopsys.com/v1"

// 	_, err = util.CreateBlackduck(blackDuckClient, blackDuckCR.Namespace, blackDuckCR)
// 	if err != nil {
// 		return fmt.Errorf("error creating Black Duck '%s' in namespace '%s' due to %+v", blackDuckCR.Name, blackDuckCR.Namespace, err)
// 	}
// 	return nil
// }

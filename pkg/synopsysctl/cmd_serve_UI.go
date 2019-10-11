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
		}

		pConfig := PolarisUIRequestResponse{}
		err = json.Unmarshal(reqBody, &pConfig)
		if err != nil {
			log.Errorf("failed to unmarshal Polaris request: %s", err)
			return
		}
		err = checkRequiredPolarisRequestFields(pConfig)
		if err != nil {
			log.Errorf("missing data to ensure Polaris: %s", err)
			return
		}

		polarisObj, err := convertPolarisUIResponseToPolarisObject(pConfig)
		if err != nil {
			log.Errorf("failed to convert Request data to Polaris Object: %s", err)
			return
		}

		// Check Secret to see if polaris already exists
		namespace = polarisObj.Namespace // set global namespace value for getPolarisFromSecret and ensurePolaris
		oldPolaris, err := getPolarisFromSecret()
		if err != nil {
			log.Errorf("failed to get Secret for Polaris: %s", err)
			return
		}
		if oldPolaris == nil { // create new Polaris
			err = ensurePolaris(polarisObj, false, true)
			if err != nil {
				log.Errorf("error ensuring Polaris: %s", err)
				return
			}
		} else { // update Polaris
			err = ensurePolaris(polarisObj, true, false)
			if err != nil {
				log.Errorf("error ensuring Polaris: %s", err)
				return
			}
		}

		log.Infof("successfully ensured Polaris")
	})

	// api route - get_polaris returns Polaris specification to the client
	router.HandleFunc("/api/get_polaris", func(w http.ResponseWriter, r *http.Request) {
		log.Infof("handling request: %q\n", html.EscapeString(r.URL.Path))
		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Errorf("failed to read request body: %+v", err)
			return
		}

		// Get Polaris from secret
		namespace = string(reqBody)
		polarisObj, err := getPolarisFromSecret()
		if err != nil {
			log.Errorf("error getting Polaris: %s", err)
			emptyResonse, _ := json.Marshal(PolarisUIRequestResponse{})
			w.Write(emptyResonse)
			return // TODO - return a 404 and handle on the front end side
		}
		response, err := convertPolarisObjToUIResponse(*polarisObj)
		if err != nil {
			log.Errorf("error converting PolarisObj to Response: %s", err)
			emptyResonse, _ := json.Marshal(PolarisUIRequestResponse{})
			w.Write(emptyResonse)
			return // TODO - return a 404 and handle on the front end side
		}
		responseByte, err := json.Marshal(response)
		if err != nil {
			log.Errorf("error converting Polaris to bytes: %s", err)
			emptyResonse, _ := json.Marshal(PolarisUIRequestResponse{})
			w.Write(emptyResonse)
			return // TODO - return a 404 and handle on the front end side
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
	Version          string `json:"version"`
	EnvironmentDNS   string `json:"environmentDNS"`
	ImagePullSecrets string `json:"imagePullSecrets"`
	StorageClass     string `json:"storageClass"`
	Namespace        string `json:"namespace"`

	PostgresHost     string `json:"postgresHost"`
	PostgresPort     string `json:"postgresPort"`
	PostgresUsername string `json:"postgresUsername"`
	PostgresPassword string `json:"postgresPassword"`
	PostgresSize     string `json:"postgresSize"`

	SMTPHost        string `json:"smtpHost"`
	SMTPPort        string `json:"smtpPort"`
	SMTPUsername    string `json:"smtpUsername"`
	SMTPPassword    string `json:"smtpPassword"`
	SMTPSenderEmail string `json:"smtpSenderEmail"`

	UploadServerSize   string `json:"uploadServerSize"`
	EventstoreSize     string `json:"eventstoreSize"`
	MongoDBSize        string `json:"mongoDBSize"`
	DownloadServerSize string `json:"downloadServerSize"`

	EnableReporting   bool   `json:"enableReporting"`
	ReportStorageSize string `json:"reportStorageSize"`

	OrganizationDescription   string `json:"organizationDescription"`
	OrganizationName          string `json:"organizationName"`
	OrganizationAdminName     string `json:"organizationAdminName"`
	OrganizationAdminUsername string `json:"organizationAdminUsername"`
	OrganizationAdminEmail    string `json:"organizationAdminEmail"`
	CoverityLicensePath       string `json:"coverityLicensePath"`
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

// checkRequiredPolarisRequestFields returns an error if a required field is missing from
// a request from the front end UI
func checkRequiredPolarisRequestFields(polarisUIRequestConfig PolarisUIRequestResponse) error {
	// Check required UI request fields
	if polarisUIRequestConfig.Version == "" {
		return fmt.Errorf("field required: Version")
	}
	if polarisUIRequestConfig.EnvironmentDNS == "" {
		return fmt.Errorf("field required: EnvironmentDNS")
	}
	if polarisUIRequestConfig.PostgresUsername == "" {
		return fmt.Errorf("field required: PostgresUsername")
	}
	if polarisUIRequestConfig.PostgresPassword == "" {
		return fmt.Errorf("field required: PostgresPassword")
	}

	if polarisUIRequestConfig.SMTPHost == "" {
		return fmt.Errorf("field required: SMTPHost")
	}
	if polarisUIRequestConfig.SMTPPort == "" {
		return fmt.Errorf("field required: SMTPPort")
	}
	if polarisUIRequestConfig.SMTPSenderEmail == "" {
		return fmt.Errorf("field required: SMTPSenderEmail")
	}

	if polarisUIRequestConfig.OrganizationDescription == "" {
		return fmt.Errorf("field required: OrganizationDescription")
	}
	if polarisUIRequestConfig.OrganizationName == "" {
		return fmt.Errorf("field required: OrganizationName")
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
	if polarisUIRequestConfig.CoverityLicensePath == "" {
		return fmt.Errorf("field required: CoverityLicensePath")
	}
	return nil
}

// convertPolarisUIResponseToPolarisObject takes the fields in polarisUIRequestConfig and maps
// them into the Polaris Object Specification needed to create Polaris
func convertPolarisUIResponseToPolarisObject(polarisUIRequestConfig PolarisUIRequestResponse) (*polaris.Polaris, error) {
	// Populate Polaris Config Fields
	polarisObj := *polaris.GetPolarisDefault()
	if polarisUIRequestConfig.Namespace == "" {
		polarisObj.Namespace = "default"
	} else {
		polarisObj.Namespace = polarisUIRequestConfig.Namespace
	}
	polarisObj.Version = polarisUIRequestConfig.Version
	polarisObj.EnvironmentDNS = polarisUIRequestConfig.EnvironmentDNS
	data, err := util.ReadFileData(polarisUIRequestConfig.ImagePullSecrets)
	if err != nil {
		panic(err)
	}
	polarisObj.GCPServiceAccount = data
	// TODO - Postgres host and port are not supported
	// polarisObj.PolarisDBSpec.PostgresDetails.Host = polarisUIRequestConfig.PostgresHost
	// var postPort int64
	// if polarisUIRequestConfig.PostgresPort != "" {
	// 	postPort, err = strconv.ParseInt(polarisUIRequestConfig.PostgresPort, 0, 64)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("falied to convert postgres port to an int: %s", polarisUIRequestConfig.PostgresPort)
	// 	}
	// }
	// polarisObj.PolarisDBSpec.PostgresDetails.Port = int(postPort)

	polarisObj.PolarisDBSpec.PostgresDetails.Username = polarisUIRequestConfig.PostgresUsername
	polarisObj.PolarisDBSpec.PostgresDetails.Password = polarisUIRequestConfig.PostgresPassword
	polarisObj.PolarisDBSpec.PostgresDetails.Storage.StorageSize = polarisUIRequestConfig.PostgresSize

	polarisObj.PolarisDBSpec.SMTPDetails.Host = polarisUIRequestConfig.SMTPHost
	var sPort int64
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

	polarisObj.PolarisDBSpec.UploadServerDetails.Storage.StorageSize = polarisUIRequestConfig.UploadServerSize
	polarisObj.PolarisDBSpec.EventstoreDetails.Storage.StorageSize = polarisUIRequestConfig.EventstoreSize
	polarisObj.PolarisDBSpec.MongoDBDetails.Storage.StorageSize = polarisUIRequestConfig.MongoDBSize
	polarisObj.PolarisSpec.DownloadServerDetails.Storage.StorageSize = polarisUIRequestConfig.DownloadServerSize

	polarisObj.EnableReporting = polarisUIRequestConfig.EnableReporting
	if polarisUIRequestConfig.EnableReporting {
		polarisObj.ReportingSpec.ReportStorageDetails.Storage.StorageSize = polarisUIRequestConfig.ReportStorageSize
	}

	polarisObj.OrganizationDetails.OrganizationProvisionOrganizationDescription = polarisUIRequestConfig.OrganizationDescription
	polarisObj.OrganizationDetails.OrganizationProvisionOrganizationName = polarisUIRequestConfig.OrganizationName
	polarisObj.OrganizationDetails.OrganizationProvisionAdminName = polarisUIRequestConfig.OrganizationAdminName
	polarisObj.OrganizationDetails.OrganizationProvisionAdminUsername = polarisUIRequestConfig.OrganizationAdminUsername
	polarisObj.OrganizationDetails.OrganizationProvisionAdminEmail = polarisUIRequestConfig.OrganizationAdminEmail
	data, err = util.ReadFileData(polarisUIRequestConfig.CoverityLicensePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read coverity license from file: %+v", err)
	}
	// TODO - add validating file format
	polarisObj.Licenses.Coverity = data

	return &polarisObj, nil
}

// convertPolarisObjToUIResponse converts the Polaris Object Specification into the Response format
// that can be sent to the User Interface
func convertPolarisObjToUIResponse(polarisObj polaris.Polaris) (*PolarisUIRequestResponse, error) {
	// Populate Polaris Config Fields
	polarisUIRequestConfig := &PolarisUIRequestResponse{}
	polarisUIRequestConfig.Namespace = polarisObj.Namespace

	polarisUIRequestConfig.Version = polarisObj.Version
	polarisUIRequestConfig.EnvironmentDNS = polarisObj.EnvironmentDNS
	polarisUIRequestConfig.ImagePullSecrets = polarisObj.ImagePullSecrets
	// TODO - Postgres host and port are not supported
	// polarisUIRequestConfig.PostgresHost = polarisObj.PolarisDBSpec.PostgresDetails.Host
	// polarisUIRequestConfig.PostgresPort  = strconv.FormatInt(polarisObj.PolarisDBSpec.PostgresDetails.Port)
	polarisUIRequestConfig.PostgresUsername = polarisObj.PolarisDBSpec.PostgresDetails.Username
	polarisUIRequestConfig.PostgresPassword = polarisObj.PolarisDBSpec.PostgresDetails.Password
	polarisUIRequestConfig.PostgresSize = polarisObj.PolarisDBSpec.PostgresDetails.Storage.StorageSize

	polarisUIRequestConfig.SMTPHost = polarisObj.PolarisDBSpec.SMTPDetails.Host
	polarisUIRequestConfig.SMTPPort = strconv.FormatInt(int64(polarisObj.PolarisDBSpec.SMTPDetails.Port), 10)
	polarisUIRequestConfig.SMTPUsername = polarisObj.PolarisDBSpec.SMTPDetails.Username
	polarisUIRequestConfig.SMTPPassword = polarisObj.PolarisDBSpec.SMTPDetails.Password
	polarisUIRequestConfig.SMTPSenderEmail = polarisObj.PolarisDBSpec.SMTPDetails.SenderEmail

	polarisUIRequestConfig.UploadServerSize = polarisObj.PolarisDBSpec.UploadServerDetails.Storage.StorageSize
	polarisUIRequestConfig.EventstoreSize = polarisObj.PolarisDBSpec.EventstoreDetails.Storage.StorageSize
	polarisUIRequestConfig.MongoDBSize = polarisObj.PolarisDBSpec.MongoDBDetails.Storage.StorageSize
	polarisUIRequestConfig.DownloadServerSize = polarisObj.PolarisSpec.DownloadServerDetails.Storage.StorageSize

	polarisUIRequestConfig.EnableReporting = polarisObj.EnableReporting
	polarisUIRequestConfig.ReportStorageSize = polarisObj.ReportingSpec.ReportStorageDetails.Storage.StorageSize

	polarisUIRequestConfig.OrganizationDescription = polarisObj.OrganizationDetails.OrganizationProvisionOrganizationDescription
	polarisUIRequestConfig.OrganizationName = polarisObj.OrganizationDetails.OrganizationProvisionOrganizationName
	polarisUIRequestConfig.OrganizationAdminName = polarisObj.OrganizationDetails.OrganizationProvisionAdminName
	polarisUIRequestConfig.OrganizationAdminUsername = polarisObj.OrganizationDetails.OrganizationProvisionAdminUsername
	polarisUIRequestConfig.OrganizationAdminEmail = polarisObj.OrganizationDetails.OrganizationProvisionAdminEmail
	polarisUIRequestConfig.CoverityLicensePath = "" // TODO - User needs to provide License Path every time, change to not require on Updates (may involve front end)

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

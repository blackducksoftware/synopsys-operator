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

	synopsysV1 "github.com/blackducksoftware/synopsys-operator/api/v1"
	"github.com/blackducksoftware/synopsys-operator/soperator"
	"github.com/blackducksoftware/synopsys-operator/utils"
	"github.com/gobuffalo/packr"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	box := packr.NewBox("../../operator-ui-ember/dist")

	// Create a Router to listen and serve User Interface requests
	router := mux.NewRouter()

	// api route - deploy_operator deploys Synopsys Operator into the cluster
	router.HandleFunc("/api/deploy_operator", func(w http.ResponseWriter, r *http.Request) {
		log.Infof("handling request: %q\n", html.EscapeString(r.URL.Path))
		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}
		log.Debugf("request data: %s\n\n", reqBody)
		err = deployOperatorRequest(reqBody)
		if err != nil {
			log.Errorf("error deploying Synopsys Operator: %s\n", err)
			return
		}
		log.Infof("successfully deployed Synopsys Operator")
	})

	// api route - deploy_polaris deploys Polaris into the cluster
	router.HandleFunc("/api/deploy_polaris", func(w http.ResponseWriter, r *http.Request) {
		log.Infof("handling request: %q\n", html.EscapeString(r.URL.Path))
		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}
		log.Debugf("request data: %s\n\n", reqBody)

		err = createPolarisCRsRequest(reqBody)
		if err != nil {
			log.Errorf("error deploying Polaris: %s\n", err)
			return
		}
		log.Infof("successfully deployed Polaris")
	})

	// api route - deploy_black_duck deploys Black Duck into the cluster
	router.HandleFunc("/api/deploy_black_duck", func(w http.ResponseWriter, r *http.Request) {
		log.Infof("handling request: %q\n", html.EscapeString(r.URL.Path))
		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}
		log.Debugf("request data: %s\n\n", reqBody)
		err = createBlackDuckCRRequest(reqBody)
		if err != nil {
			log.Errorf("error creating Black Duck: %s", err)
			return
		}
		log.Infof("successfully created Black Duck")
	})

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
	fmt.Printf("  - /api/deploy_operator\n")
	fmt.Printf("  - /api/deploy_polaris\n")
	fmt.Printf("  - /api/deploy_black_duck\n")
	fmt.Printf("==================================\n")
	fmt.Printf("\n")

	// Serving the server
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
			fmt.Printf("[ERROR] reading index file from box\n")
		}
		http.ServeContent(w, r, path, time.Now(), file)
		return
	}

	// otherwise, use http.FileServer to serve the static dir
	http.FileServer(h.box).ServeHTTP(w, r)
}

/* USER INTERFACE REQUEST DATA STRUCTS */

// DeployOperatorUIRequestConfig represents the format that
// the front-end should send its data for deploying
// Synopsys Operator
type DeployOperatorUIRequestConfig struct {
	Namespace       string `json:"namespace"`
	ClusterScoped   bool   `json:"clusterScoped"`
	EnableAlert     bool   `json:"enableAlert"`
	EnableBlackDuck bool   `json:"enableBlackDuck"`
	EnableOpsSight  bool   `json:"enableOpsSight"`
	EnablePolaris   bool   `json:"enablePolaris"`
	ExposeMetrics   string `json:"exposeMetrics"`
	ExposeUI        string `json:"exposeUI"`
	MetricsImage    string `json:"metricsImage"`
	OperatorImage   string `json:"operatorImage"`
}

// PolarisUIRequest represents the format that
// the front-end should send its data for deploying
// Polaris
type PolarisUIRequest struct {
	Version          string `json:"version"`
	EnvironmentName  string `json:"environmentName"`
	EnvironmentDNS   string `json:"environmentDNS"`
	ImagePullSecrets string `json:"imagePullSecrets"`
	StorageClass     string `json:"storageClass"`
	Namespace        string `json:"namespace"`

	PostgresHost     string `json:"postgresHost"`
	PostgresPort     string `json:"postgresPort"`
	PostgresUsername string `json:"postgresUsername"`
	PostgresPassword string `json:"postgresPassword"`
	PostgresSize     string `json:"postgresSize"`

	SMTPHost     string `json:"smtpHost"`
	SMTPPort     string `json:"smtpPort"`
	SMTPUsername string `json:"smtpUsername"`
	SMTPPassword string `json:"smtpPassword"`

	UploadServerSize string `json:"uploadServerSize"`
	EventstoreSize   string `json:"eventstoreSize"`
}

// BlackDuckUIRequest represents the format that
// the front-end should send its data for deploying
// Black Duck
type BlackDuckUIRequest struct {
	Name                              string `json:"name"`
	Namespace                         string `json:"namespace"`
	Version                           string `json:"version"`
	LicenseKey                        string `json:"licenseKey"`
	DbMigrate                         bool   `json:"dbMigrate"`
	Size                              string `json:"size"`
	ExposeService                     string `json:"exposeService"`
	BlackDuckType                     string `json:"blackDuckType"`
	UseBinaryUploads                  bool   `json:"useBinaryUploads"`
	EnableSourceUploads               bool   `json:"enableSourceUploads"`
	LivenessProbes                    bool   `json:"livenessProbes"`
	PersistentStorage                 bool   `json:"persistentStorage"`
	CloneDB                           string `json:"cloneDB"`
	PVCStorageClass                   string `json:"PVCStorageClass"`
	ScanType                          string `json:"scanType"`
	ExternalDatabase                  bool   `json:"externalDatabase"`
	ExternalPostgresSQLHost           string `json:"externalPostgresSQLHost"`
	ExternalPostgresSQLPort           string `json:"externalPostgresSQLPort"`
	ExternalPostgresSQLAdminUser      string `json:"externalPostgresSQLAdminUser"`
	ExternalPostgresSQLAdminPassword  string `json:"externalPostgresSQLAdminPassword"`
	ExternalPostgresSQLUser           string `json:"externalPostgresSQLUser"`
	ExternalPostgresSQLUserPassword   string `json:"externalPostgresSQLUserPassword"`
	EnableSSL                         bool   `json:"enableSSL"`
	PostgresSQLUserPassword           string `json:"postgresSQLUserPassword"`
	PostgresSQLAdminPassword          string `json:"postgresSQLAdminPassword"`
	PostgresSQLPostgresPassword       string `json:"postgresSQLPostgresPassword"`
	CertificateName                   string `json:"certificateName"`
	CustomCACertificateAuthentication bool   `json:"customCACertificateAuthentication"`
	ProxyRootCertificate              string `json:"proxyRootCertificate"`
	ContainerImageTags                string `json:"containerImageTags"`
	EnvironmentVariables              string `json:"environmentVariables"`
	NodeAffinityJSON                  string `json:"nodeAffinityJSON"`
}

/* COMMANDS TO DEPLOY/CREATE RESOURCES FROM UI REQUEST STRUCTS */

// createPolarisCRsRequest deploys Synopsys Operator into your cluster by
// populating the global variables used by cmd_deploy and then running
// the cmd_deploy logic
// data -  a []byte representation of DeployOperatorUIRequestConfig
func deployOperatorRequest(data []byte) error {
	oConfig := DeployOperatorUIRequestConfig{}
	err := json.Unmarshal(data, &oConfig)
	if err != nil {
		return fmt.Errorf("failed to unmarshal Synopsys Operator request: %s", err)
	}
	// Set Global Flags for cmd_deploy - use defaults for other flags
	if oConfig.Namespace != "" {
		operatorNamespace = oConfig.Namespace
	}
	if oConfig.ExposeUI != "" {
		exposeUI = oConfig.ExposeUI
	}
	if oConfig.OperatorImage != "" {
		synopsysOperatorImage = oConfig.OperatorImage
	}
	if oConfig.MetricsImage != "" {
		metricsImage = oConfig.MetricsImage
	}
	if oConfig.ExposeMetrics != "" {
		exposeMetrics = oConfig.ExposeMetrics
	}
	isEnabledAlert = oConfig.EnableAlert
	isEnabledBlackDuck = oConfig.EnableBlackDuck
	isEnabledOpsSight = oConfig.EnableOpsSight
	isEnabledPrm = oConfig.EnablePolaris
	isClusterScoped = oConfig.ClusterScoped

	/* TODO: Change from being a duplicate of cmd_deploy */

	// validate each CRD enable parameter is enabled/disabled and cluster scope are from supported values
	crds, err := getEnabledCrds()
	if err != nil {
		return err
	}

	//// Add Size CRD
	// TODO
	//crds = append(crds, util.SizeCRDName)

	// Get CRD configs
	crdConfigs, err := getCrdConfigs(operatorNamespace, isClusterScoped, crds)
	if err != nil {
		return err
	}
	if len(crdConfigs) <= 0 {
		return fmt.Errorf("no resources are enabled (include flag(s): --enable-alert --enable-blackduck --enable-opssight --enable-polaris )")
	}
	// Create Synopsys Operator Spec
	soperatorSpec, err := getSpecToDeploySOperator(crds)
	if err != nil {
		return err
	}

	// check if namespace exist in namespace scope, if not throw an error
	if !isClusterScoped {
		_, err = kubeClient.CoreV1().Namespaces().Get(operatorNamespace, v1.GetOptions{})
		if err != nil {
			return fmt.Errorf("please create the namespace '%s' to deploy the Synopsys Operator in namespace scoped", operatorNamespace)
		}
	}

	// check if operator is already installed
	_, err = utils.GetOperatorNamespace(kubeClient, operatorNamespace)
	if err == nil {
		return fmt.Errorf("the Synopsys Operator instance is already deployed in namespace '%s'", namespace)
	}

	log.Infof("deploying Synopsys Operator in namespace '%s'...", operatorNamespace)

	log.Debugf("creating custom resource definitions")
	err = deployCrds(operatorNamespace, isClusterScoped, crdConfigs)
	if err != nil {
		return err
	}

	log.Debugf("creating Synopsys Operator components")
	sOperatorCreater := soperator.NewCreater(false, restconfig, kubeClient)
	err = sOperatorCreater.UpdateSOperatorComponents(soperatorSpec)
	if err != nil {
		return fmt.Errorf("error deploying Synopsys Operator due to %+v", err)
	}

	log.Infof("successfully submitted Synopsys Operator into namespace '%s'", operatorNamespace)
	return nil
}

// createPolarisCRsRequest generates the three CRs for Polaris from the request data
// provided by the User Interface
// data -  a []byte representation of PolarisUIRequest
func createPolarisCRsRequest(data []byte) error {
	pConfig := PolarisUIRequest{}
	err := json.Unmarshal(data, &pConfig)
	if err != nil {
		return fmt.Errorf("failed to unmarshal Polaris request: %s", err)
	}

	polarisAuthCR, polarisDBCR, polarisCR, err := convertPolarisUIResponseToCRs(pConfig)
	if err != nil {
		return err
	}

	if _, err := utils.CreateAuthServer(restClient, polarisAuthCR); err != nil {
		return fmt.Errorf("error creating Polaris Auth: %s", err)
	}
	log.Debugf("successfully created Polaris Auth '%s' in namespace '%s'", polarisAuthCR.Name, polarisAuthCR.Namespace)

	if _, err := utils.CreatePolarisDB(restClient, polarisDBCR); err != nil {
		return fmt.Errorf("error creating Polaris DB: %s", err)
	}
	log.Debugf("successfully created Polaris DB '%s' in namespace '%s'", polarisDBCR.Name, polarisDBCR.Namespace)

	if _, err := utils.CreatePolaris(restClient, polarisCR); err != nil {
		return fmt.Errorf("error creating Polaris: %s", err)
	}
	log.Debugf("successfully created Polaris '%s' in namespace '%s'", polarisCR.Name, polarisCR.Namespace)

	return nil
}

// convertPolarisUIResponseToCRs takes the fields in polarisUIRequestConfig and maps
// them into the three CRs needed to create Polaris: AuthServer, PolarisDB, and Polaris
func convertPolarisUIResponseToCRs(polarisUIRequestConfig PolarisUIRequest) (*synopsysV1.AuthServer, *synopsysV1.PolarisDB, *synopsysV1.Polaris, error) {
	// Populate Auth Service
	auth := &synopsysV1.AuthServer{}
	authSpec := &synopsysV1.AuthServerSpec{}
	authSpec.Namespace = polarisUIRequestConfig.Namespace
	authSpec.Version = polarisUIRequestConfig.Version
	authSpec.EnvironmentDNS = polarisUIRequestConfig.EnvironmentDNS
	authSpec.EnvironmentName = polarisUIRequestConfig.EnvironmentName
	authSpec.ImagePullSecrets = polarisUIRequestConfig.ImagePullSecrets
	auth.Spec = *authSpec

	// Populate Polaris Database
	polarisDB := &synopsysV1.PolarisDB{}
	polarisDBSpec := &synopsysV1.PolarisDBSpec{}
	polarisDBSpec.Namespace = polarisUIRequestConfig.Namespace
	polarisDBSpec.Version = polarisUIRequestConfig.Version
	polarisDBSpec.EnvironmentDNS = polarisUIRequestConfig.EnvironmentDNS
	polarisDBSpec.EnvironmentName = polarisUIRequestConfig.EnvironmentName
	polarisDBSpec.ImagePullSecrets = polarisUIRequestConfig.ImagePullSecrets
	//polarisDBSpec.PostgresStorageDetails.StorageClass = &polarisUIRequestConfig.StorageClass
	//polarisDBSpec.UploadServerDetails.Storage.StorageClass = &polarisUIRequestConfig.StorageClass
	polarisDBSpec.PostgresDetails.Host = polarisUIRequestConfig.PostgresHost
	postPort, err := strconv.ParseInt(polarisUIRequestConfig.PostgresPort, 0, 32)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("falied to convert postgres port to an int: %s", polarisUIRequestConfig.PostgresPort)
	}
	polarisDBSpec.PostgresDetails.Port = int32(postPort)
	polarisDBSpec.PostgresDetails.Username = polarisUIRequestConfig.PostgresUsername
	polarisDBSpec.PostgresDetails.Password = polarisUIRequestConfig.PostgresPassword
	polarisDBSpec.PostgresStorageDetails.StorageSize = polarisUIRequestConfig.PostgresSize
	polarisDBSpec.UploadServerDetails.Storage.StorageSize = polarisUIRequestConfig.UploadServerSize
	polarisDBSpec.EventstoreDetails.Storage.StorageSize = polarisUIRequestConfig.EventstoreSize

	polarisDBSpec.SMTPDetails.Host = polarisUIRequestConfig.SMTPHost
	sPort, err := strconv.ParseInt(polarisUIRequestConfig.SMTPPort, 0, 32)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("falied to convert smtp port to an int: %s", polarisUIRequestConfig.PostgresPort)
	}
	polarisDBSpec.SMTPDetails.Port = int32(sPort)
	polarisDBSpec.SMTPDetails.Username = polarisUIRequestConfig.SMTPUsername
	polarisDBSpec.SMTPDetails.Password = polarisUIRequestConfig.SMTPPassword
	polarisDB.Spec = *polarisDBSpec

	// Populate Polaris
	polaris := &synopsysV1.Polaris{}
	polarisSpec := &synopsysV1.PolarisSpec{}
	polarisSpec.Namespace = polarisUIRequestConfig.Namespace
	polarisSpec.Version = polarisUIRequestConfig.Version
	polarisSpec.EnvironmentDNS = polarisUIRequestConfig.EnvironmentDNS
	polarisSpec.EnvironmentName = polarisUIRequestConfig.EnvironmentName
	polarisSpec.ImagePullSecrets = polarisUIRequestConfig.ImagePullSecrets
	polaris.Spec = *polarisSpec

	return auth, polarisDB, polaris, nil
}

// createBlackDuckCRRequest generates the CR for Black Duck from the request data
// provided by the User Interface
// data -  a []byte representation of BlackDuckUIRequest
func createBlackDuckCRRequest(data []byte) error {
	bdConfig := BlackDuckUIRequest{}
	err := json.Unmarshal(data, &bdConfig)
	if err != nil {
		return fmt.Errorf("failed to unmarshal Black Duck request: %s", err)
	}
	blackDuckCR := &synopsysV1.Blackduck{
		ObjectMeta: v1.ObjectMeta{
			Name:      bdConfig.Name,
			Namespace: bdConfig.Namespace,
		},
		Spec: synopsysV1.BlackduckSpec{
			Namespace:         bdConfig.Namespace,
			Size:              bdConfig.Size,
			Version:           bdConfig.Version,
			ExposeService:     bdConfig.ExposeService,
			PVCStorageClass:   bdConfig.PVCStorageClass,
			LivenessProbes:    bdConfig.LivenessProbes,
			ScanType:          bdConfig.ScanType,
			PersistentStorage: bdConfig.PersistentStorage,
			LicenseKey:        bdConfig.LicenseKey,
			AdminPassword:     bdConfig.PostgresSQLAdminPassword,
			UserPassword:      bdConfig.PostgresSQLUserPassword,
			PostgresPassword:  bdConfig.PostgresSQLPostgresPassword,
		},
	}
	blackDuckCR.Kind = "Blackduck"
	blackDuckCR.APIVersion = "synopsys.com/v1"
	_, err = utils.CreateBlackduck(restClient, blackDuckCR)
	if err != nil {
		return fmt.Errorf("error creating Black Duck CR: %s", err)
	}
	return nil
}

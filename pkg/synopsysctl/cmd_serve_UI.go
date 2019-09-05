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
	"os"
	"path/filepath"
	"strconv"
	"time"

	// "os"
	// "time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	synopsysV1 "github.com/blackducksoftware/synopsys-operator/meta-builder/api/v1"
	"github.com/blackducksoftware/synopsys-operator/meta-builder/utils"
	blackduckv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/size"
	"github.com/blackducksoftware/synopsys-operator/pkg/soperator"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Port to serve on
var serverPort = "8081"

// serveUICmd edits Synopsys resources
var serveUICmd = &cobra.Command{
	Use:   "serve-ui",
	Short: "Starts a service running the User Interface and listens for events",
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debug("Starting User Interface Server...")

		// Start Running a backend server that listens for input from the User Interface
		router := mux.NewRouter()

		// api route - deploy Operator
		router.HandleFunc("/api/deploy_operator", func(w http.ResponseWriter, r *http.Request) {
			fmt.Printf("Handler Deploy Operator: %q\n", html.EscapeString(r.URL.Path))
			reqBody, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Data from Operator Body: %s\n\n", reqBody)
			err = deployOperator(reqBody)
			if err != nil {
				fmt.Printf("[ERROR] Failed to deploy Operator: %s\n", err)
			}
		})

		// api route - deploy Polaris
		router.HandleFunc("/api/deploy_polaris", func(w http.ResponseWriter, r *http.Request) {
			fmt.Printf("Handler Deploy Polaris: %q\n", html.EscapeString(r.URL.Path))
			reqBody, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Data from Polaris Body: %s\n\n", reqBody)
			auth, db, polaris := createPolarisSpec(reqBody)
			fmt.Printf("authSpec: %+v\n", p1)
			fmt.Printf("polarisDBSpec: %+v\n", p2)
			fmt.Printf("polarisSpec: %+v\n", p3)
			if _, err := utils.CreateAuthServer(restconfig, auth); err != nil {
				fmt.Printf("[ERROR] failed to deploy Polaris: %s", err)
				return
			}
			if _, err := utils.CreatePolarisDB(restconfig, db); err != nil {
				fmt.Printf("[ERROR] failed to deploy Polaris: %s", err)
				return
			}
			if _, err := utils.CreatePolaris(restconfig, polaris); err != nil {
				fmt.Printf("[ERROR] failed to deploy Polaris: %s", err)
				return
			}
		})

		// api route - deploy Black Duck
		router.HandleFunc("/api/deploy_black_duck", func(w http.ResponseWriter, r *http.Request) {
			fmt.Printf("Handler Deploy Black Duck: %q\n", html.EscapeString(r.URL.Path))
			reqBody, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Data from Black Duck Body: %s\n\n", reqBody)
			bd := createBlackDuckSpec(reqBody)
			fmt.Printf("Black Duck CR: %+v\n\n", bd)
			_, err = utils.CreateBlackduck(restconfig, bd)
			if err != nil {
				fmt.Errorf("[ERROR] failed to create Black Duck '%s' in namespace '%s' due to %+v", bd.Name, bd.Namespace, err)
				return
			}
			fmt.Printf("Successfully deployed Black Duck: %+v\n\n", bd.Name)
		})

		// Serve files for UI
		spa := spaHandler{
			staticPath: "../../operator-ui-ember/dist",
			indexPath:  "../../operator-ui-ember/dist/index.html",
		}
		router.PathPrefix("/").Handler(spa)

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
	},
}

// Copied From: https://github.com/gorilla/mux#serving-single-page-applications
// spaHandler implements the http.Handler interface, so we can use it
// to respond to HTTP requests. The path to the static directory and
// path to the index file within that static directory are used to
// serve the SPA in the given static directory.
type spaHandler struct {
	staticPath string
	indexPath  string
}

// Copied From: https://github.com/gorilla/mux#serving-single-page-applications
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

	// prepend the path with the path to the static directory
	path = filepath.Join(h.staticPath, path)

	// check whether a file exists at the given path
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		// file does not exist, serve index.html
		http.ServeFile(w, r, filepath.Join(h.staticPath, h.indexPath))
		return
	} else if err != nil {
		// if we got an error (that wasn't that the file doesn't exist) stating the
		// file, return a 500 internal server error and stop
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// otherwise, use http.FileServer to serve the static dir
	http.FileServer(http.Dir(h.staticPath)).ServeHTTP(w, r)
}

func init() {
	rootCmd.AddCommand(serveUICmd)
	serveUICmd.Flags().StringVarP(&serverPort, "port", "p", serverPort, "Port to listen for UI requests")
}

type operatorUIRequestConfig struct {
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

func deployOperator(data []byte) error {
	oConfig := operatorUIRequestConfig{}
	err := json.Unmarshal(data, &oConfig)
	if err != nil {
		fmt.Printf("Failed to Unmarshal: %s\n\n", err)
		return nil
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
	isEnabledBlackDuck = oConfig.EnableAlert
	isEnabledOpsSight = oConfig.EnableAlert
	isEnabledPrm = oConfig.EnablePolaris
	isClusterScoped = oConfig.ClusterScoped

	/* Call the same functions as cmd_deploy */

	// validate each CRD enable parameter is enabled/disabled and cluster scope are from supported values
	crds, err := getEnabledCrds()
	if err != nil {
		return err
	}

	// Add Size CRD
	crds = append(crds, util.SizeCRDName)

	log.Debugf("Got CRDs to Enable: %+v", crds)

	// Get CRD configs
	crdConfigs, err := getCrdConfigs(operatorNamespace, isClusterScoped, crds)
	if err != nil {
		return err
	}
	if len(crdConfigs) <= 1 {
		return fmt.Errorf("no resources are enabled (include flag(s): --enable-alert --enable-blackduck --enable-opssight )")
	}
	// Create Synopsys Operator Spec
	soperatorSpec, err := getSpecToDeploySOperator(crds)
	if err != nil {
		return err
	}

	// check if namespace exist in namespace scope, if not throw an error
	if !isClusterScoped {
		_, err = util.GetNamespace(kubeClient, operatorNamespace)
		if err != nil {
			return fmt.Errorf("please create the namespace '%s' to deploy the Synopsys Operator in namespace scoped", operatorNamespace)
		}
	}

	// check if operator is already installed
	_, err = util.GetOperatorNamespace(kubeClient, operatorNamespace)
	if err == nil {
		return fmt.Errorf("the Synopsys Operator instance is already deployed in namespace '%s'", namespace)
	}

	log.Infof("deploying Synopsys Operator in namespace '%s'...", operatorNamespace)

	log.Debugf("creating custom resource definitions")
	err = deployCrds(operatorNamespace, isClusterScoped, crdConfigs)
	if err != nil {
		return err
	}

	// In some rare cases, the CRD may not be available when we create the default sizes.
	log.Debugf("Waiting for CRDs...")
	if err := util.WaitForCRD(util.SizeCRDName, time.Second, time.Minute*3, apiExtensionClient); err != nil {
		return err
	}

	// Create default sizes
	log.Debugf("Getting Default Sizes...")
	for _, v := range size.GetAllDefaultSizes() {
		_, err = sizeClient.SynopsysV1().Sizes(operatorNamespace).Create(v)
		if err != nil {
			return err
		}
	}

	log.Debugf("creating Synopsys Operator components")
	sOperatorCreater := soperator.NewCreater(false, restconfig, kubeClient)
	err = sOperatorCreater.UpdateSOperatorComponents(soperatorSpec)
	if err != nil {
		return fmt.Errorf("error deploying Synopsys Operator due to %+v", err)
	}

	log.Debugf("creating Metrics components")
	promtheusSpec := soperator.NewPrometheus(operatorNamespace, metricsImage, exposeMetrics, restconfig, kubeClient)
	err = sOperatorCreater.UpdatePrometheus(promtheusSpec)
	if err != nil {
		return fmt.Errorf("error deploying metrics due to %s", err)
	}

	log.Infof("successfully submitted Synopsys Operator into namespace '%s'", operatorNamespace)
	return nil
}

type blackDuckUIRequestConfig struct {
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
	PostgresSQLUserPassword           string `json:"postgresSQLUserPassword"`
	PostgresSQLAdminPassword          string `json:"postgresSQLAdminPassword"`
	PostgresSQLPostgresPassword       string `json:"postgresSQLPostgresPassword"`
	CertificateName                   string `json:"certificateName"`
	CustomCACertificateAuthentication bool   `json:"customCACertificateAuthentication"`
	ProxyRootCertificate              string `json:"proxyRootCertificate"`
}

func createBlackDuckSpec(data []byte) *blackduckv1.Blackduck {
	bdConfig := blackDuckUIRequestConfig{}
	err := json.Unmarshal(data, &bdConfig)
	if err != nil {
		fmt.Printf("Failed to Unmarshal: %s\n\n", err)
		return nil
	}
	blackDuck := &blackduckv1.Blackduck{
		ObjectMeta: metav1.ObjectMeta{
			Name:      bdConfig.Name,
			Namespace: bdConfig.Namespace,
		},
		Spec: blackduckv1.BlackduckSpec{
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
	blackDuck.Kind = "Blackduck"
	blackDuck.APIVersion = "synopsys.com/v1"
	return blackDuck
}

type PolarisUIRequestConfig struct {
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

func createPolarisSpec(data []byte) (*synopsysV1.AuthServerSpec, *synopsysV1.PolarisDBSpec, *synopsysV1.PolarisSpec) {
	pConfig := PolarisUIRequestConfig{}
	err := json.Unmarshal(data, &pConfig)
	if err != nil {
		fmt.Printf("Failed to Unmarshal: %s\n\n", err)
		return nil, nil, nil
	}
	return convertPolarisUIResponseToSpecs(pConfig)
}

func convertPolarisUIResponseToSpecs(polarisUIRequestConfig PolarisUIRequestConfig) (*synopsysV1.AuthServerSpec, *synopsysV1.PolarisDBSpec, *synopsysV1.PolarisSpec) {
	// Populate Auth Service Spec
	authSpec := &synopsysV1.AuthServerSpec{}
	authSpec.Namespace = polarisUIRequestConfig.Namespace
	authSpec.Version = polarisUIRequestConfig.Version
	authSpec.EnvironmentDNS = polarisUIRequestConfig.EnvironmentDNS
	authSpec.EnvironmentName = polarisUIRequestConfig.EnvironmentName
	authSpec.ImagePullSecrets = polarisUIRequestConfig.ImagePullSecrets

	// Populate Polaris Database Spec
	polarisDBSpec := &synopsysV1.PolarisDBSpec{}
	polarisDBSpec.Namespace = polarisUIRequestConfig.Namespace
	polarisDBSpec.Version = polarisUIRequestConfig.Version
	polarisDBSpec.EnvironmentDNS = polarisUIRequestConfig.EnvironmentDNS
	polarisDBSpec.EnvironmentName = polarisUIRequestConfig.EnvironmentName
	polarisDBSpec.ImagePullSecrets = polarisUIRequestConfig.ImagePullSecrets
	polarisDBSpec.PostgresStorageDetails.StorageClass = &polarisUIRequestConfig.StorageClass
	polarisDBSpec.UploadServerDetails.Storage.StorageClass = &polarisUIRequestConfig.StorageClass
	polarisDBSpec.PostgresDetails.Host = polarisUIRequestConfig.PostgresHost
	postPort, err := strconv.ParseInt(polarisUIRequestConfig.PostgresPort, 0, 32)
	if err != nil {
		fmt.Printf("[ERROR]: Falied to convert port to an int %s\n", polarisUIRequestConfig.PostgresPort)
	}
	castedPostPort := int32(postPort)
	polarisDBSpec.PostgresDetails.Port = &castedPostPort
	polarisDBSpec.PostgresDetails.Username = polarisUIRequestConfig.PostgresUsername
	polarisDBSpec.PostgresDetails.Password = polarisUIRequestConfig.PostgresPassword
	polarisDBSpec.PostgresStorageDetails.StorageSize = polarisUIRequestConfig.PostgresSize
	polarisDBSpec.UploadServerDetails.Storage.StorageSize = polarisUIRequestConfig.UploadServerSize
	polarisDBSpec.EventstoreDetails.StorageSize = polarisUIRequestConfig.EventstoreSize

	polarisDBSpec.SMTPDetails.Host = polarisUIRequestConfig.SMTPHost
	sPort, err := strconv.ParseInt(polarisUIRequestConfig.SMTPPort, 0, 32)
	if err != nil {
		fmt.Printf("[ERROR]: Falied to convert port to an int %s\n", polarisUIRequestConfig.SMTPPort)
	}
	castedSPort := int32(sPort)
	polarisDBSpec.SMTPDetails.Port = &castedSPort
	polarisDBSpec.SMTPDetails.Username = polarisUIRequestConfig.SMTPUsername
	polarisDBSpec.SMTPDetails.Password = polarisUIRequestConfig.SMTPPassword

	// Populate Polaris Spec
	polarisSpec := &synopsysV1.PolarisSpec{}
	polarisSpec.Namespace = polarisUIRequestConfig.Namespace
	polarisSpec.Version = polarisUIRequestConfig.Version
	polarisSpec.EnvironmentDNS = polarisUIRequestConfig.EnvironmentDNS
	polarisSpec.EnvironmentName = polarisUIRequestConfig.EnvironmentName
	polarisSpec.ImagePullSecrets = polarisUIRequestConfig.ImagePullSecrets

	return authSpec, polarisDBSpec, polarisSpec
}

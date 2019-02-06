// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"flag"
	"fmt"
	"path/filepath"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	horizoncomponents "github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/horizon/pkg/deployer"
	blackduckv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	opssightv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	blackduckclientset "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/client/clientset/versioned"
	opssightclientset "github.com/blackducksoftware/synopsys-operator/pkg/opssight/client/clientset/versioned"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a synopsys resource (ex: blackduck, opssight)",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("create called")
	},
}

var blackduckCmd = &cobra.Command{
	Use:   "blackduck",
	Short: "Create an instance of a Blackduck",
	Run: func(cmd *cobra.Command, args []string) {
		// Create kubernetes Clientset
		var kubeconfig *string
		if home := homeDir(); home != "" {
			kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		} else {
			kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
		}
		flag.Parse()
		restconfig, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			panic(err.Error())
		}

		// Create namespace for the Blackduck
		namespaceDeployer, err := deployer.NewDeployer(restconfig)
		ns := horizoncomponents.NewNamespace(horizonapi.NamespaceConfig{
			// APIVersion:  "string",
			// ClusterName: "string",
			Name:      namespace,
			Namespace: namespace,
		})
		namespaceDeployer.AddNamespace(ns)
		err = namespaceDeployer.Run()
		if err != nil {
			fmt.Printf("Error deploying namespace for the Blackduck with Horizon : %s\n", err)
			return
		}

		// Create Spec for a Blackduck CRD
		blackduck := &blackduckv1.Blackduck{}
		populateBlackduckConfig(blackduck)
		fmt.Printf("%+v\n", blackduck)

		blackduckClient, err := blackduckclientset.NewForConfig(restconfig)

		_, err = blackduckClient.SynopsysV1().Blackducks(namespace).Create(blackduck)
		if err != nil {
			fmt.Printf("Error creating the Blackduck : %s\n", err)
			return
		}
	},
}

var opssightCmd = &cobra.Command{
	Use:   "opssight",
	Short: "create an instance of OpsSight",
	Run: func(cmd *cobra.Command, args []string) {
		// Create kubernetes Clientset
		var kubeconfig *string
		if home := homeDir(); home != "" {
			kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		} else {
			kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
		}
		flag.Parse()
		restconfig, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			panic(err.Error())
		}

		// Create namespace for the OpsSight
		namespaceDeployer, err := deployer.NewDeployer(restconfig)
		ns := horizoncomponents.NewNamespace(horizonapi.NamespaceConfig{
			// APIVersion:  "string",
			// ClusterName: "string",
			Name:      namespace,
			Namespace: namespace,
		})
		namespaceDeployer.AddNamespace(ns)
		err = namespaceDeployer.Run()
		if err != nil {
			fmt.Printf("Error deploying namespace for the Blackduck with Horizon : %s\n", err)
			return
		}

		// Create OpsSight Spec
		opssight := &opssightv1.OpsSight{}
		populateOpssightConfig(opssight)
		opssightClient, err := opssightclientset.NewForConfig(restconfig)
		_, err = opssightClient.SynopsysV1().OpsSights(namespace).Create(opssight)
		if err != nil {
			fmt.Printf("Error creating the OpsSight : %s\n", err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(createCmd)

	blackduckCmd.Flags().StringVar(&create_blackduck_size, "size", create_blackduck_size, "blackduck size - small, medium, large")
	blackduckCmd.Flags().BoolVar(&create_blackduck_persistentStorage, "persistent-storage", create_blackduck_persistentStorage, "enable persistent storage")
	blackduckCmd.Flags().BoolVar(&create_blackduck_LivenessProbes, "liveness-probes", create_blackduck_LivenessProbes, "enable liveness probes")
	createCmd.AddCommand(blackduckCmd)

	createCmd.AddCommand(opssightCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func populateBlackduckConfig(bd *blackduckv1.Blackduck) {
	bdSpec := blackduckv1.BlackduckSpec{
		Namespace: namespace,
		Size:      create_blackduck_size,
		//DbPrototype:       "string",
		//ExternalPostgres:  "*PostgresExternalDBConfig",
		//PVCStorageClass:   "string",
		LivenessProbes: create_blackduck_LivenessProbes,
		//ScanType:          "string",
		PersistentStorage: create_blackduck_persistentStorage,
		//PVC:               "[]PVC",
		CertificateName: "string",
		//Certificate:       "string",
		//CertificateKey:    "string",
		//ProxyCertificate:  "string",
		//Type:              "string",
		DesiredState: "string",
		//Environs:          "[]string",
		//ImageRegistries:   "[]string",
		//ImageUIDMap:       "map[string]int64",
		//LicenseKey:        "string",
	}
	bd.ObjectMeta = metav1.ObjectMeta{
		Name: namespace,
	}
	bd.Spec = bdSpec
}

func populateOpssightConfig(opssight *opssightv1.OpsSight) {
	perceptor := opssightv1.Perceptor{
		Name:                           "opssight-core",
		Image:                          "docker.io/blackducksoftware/opssight-core:master",
		Port:                           3001,
		CheckForStalledScansPauseHours: 999999,
		StalledScanClientTimeoutHours:  999999,
		ModelMetricsPauseSeconds:       15,
		UnknownImagePauseMilliseconds:  15000,
		ClientTimeoutMilliseconds:      100000,
	}

	scanner := opssightv1.Scanner{
		Name:                 "opssight-scanner",
		Image:                "docker.io/blackducksoftware/opssight-scanner:master",
		Port:                 3003,
		ClientTimeoutSeconds: 600,
	}

	imageFacade := opssightv1.ImageFacade{
		Name:               "opssight-image-getter",
		Image:              "docker.io/blackducksoftware/opssight-image-getter:master",
		Port:               3004,
		InternalRegistries: []opssightv1.RegistryAuth{},
		ImagePullerType:    "skopeo",
		ServiceAccount:     "opssight-scanner",
	}

	scannerPod := opssightv1.ScannerPod{
		Name:         "opssight-scanner",
		Scanner:      &scanner,
		ImageFacade:  &imageFacade,
		ReplicaCount: 1,
		//ImageDirectory string       `json:"imageDirectory"`
	}

	imagePerceiver := opssightv1.ImagePerceiver{
		Name:  "opssight-image-processor",
		Image: "docker.io/blackducksoftware/opssight-image-processor:${TAG}",
	}

	podPerceiver := opssightv1.PodPerceiver{
		Name:  "opssight-pod-processor",
		Image: "docker.io/blackducksoftware/opssight-pod-processor:${TAG}",
		//NamespaceFilter string `json:"namespaceFilter,omitempty"`
	}

	perceiver := opssightv1.Perceiver{
		EnableImagePerceiver:      false,
		EnablePodPerceiver:        true,
		ImagePerceiver:            &imagePerceiver,
		PodPerceiver:              &podPerceiver,
		AnnotationIntervalSeconds: 30,
		DumpIntervalMinutes:       30,
		ServiceAccount:            "opssight-processor",
		Port:                      3002,
	}

	prometheus := opssightv1.Prometheus{
		Name:  "prometheus",
		Image: "docker.io/prom/prometheus:v2.1.0",
		Port:  9090,
	}

	skyfire := opssightv1.Skyfire{
		Name:                         "skyfire",
		Image:                        "gcr.io/saas-hub-stg/blackducksoftware/pyfire:master",
		Port:                         3005,
		PrometheusPort:               3006,
		ServiceAccount:               "skyfire",
		HubClientTimeoutSeconds:      120,
		HubDumpPauseSeconds:          240,
		KubeDumpIntervalSeconds:      60,
		PerceptorDumpIntervalSeconds: 60,
	}

	blackduckSpec := blackduckv1.BlackduckSpec{
		//Namespace       :  "string",
		Size:        "small",
		DbPrototype: "",
		//ExternalPostgres  *PostgresExternalDBConfig `json:"externalPostgres,omitempty"`
		//PVCStorageClass   string                    `json:"pvcStorageClass,omitempty"`
		//LivenessProbes    bool                      `json:"livenessProbes"`
		//ScanType          string                    `json:"scanType,omitempty"`
		PersistentStorage: false,
		//PVC               []PVC                     `json:"pvc,omitempty"`
		CertificateName: "default",
		//Certificate       string                    `json:"certificate,omitempty"`
		//CertificateKey    string                    `json:"certificateKey,omitempty"`
		//ProxyCertificate  string                    `json:"proxyCertificate,omitempty"`
		Type: "worker",
		//DesiredState      string                    `json:"desiredState"`
		Environs: []string{
			"HUB_VERSION:2018.12.2",
		},
		ImageRegistries: []string{
			"docker.io/blackducksoftware/blackduck-authentication:2018.12.2",
			"docker.io/blackducksoftware/blackduck-documentation:2018.12.2",
			"docker.io/blackducksoftware/blackduck-jobrunner:2018.12.2",
			"docker.io/blackducksoftware/blackduck-registration:2018.12.2",
			"docker.io/blackducksoftware/blackduck-scan:2018.12.2",
			"docker.io/blackducksoftware/blackduck-webapp:2018.12.2",
			"docker.io/blackducksoftware/blackduck-cfssl:1.0.0",
			"docker.io/blackducksoftware/blackduck-logstash:1.0.2",
			"docker.io/blackducksoftware/blackduck-nginx:1.0.0",
			"docker.io/blackducksoftware/blackduck-solr:1.0.0",
			"docker.io/blackducksoftware/blackduck-zookeeper:1.0.0",
		},
		//ImageUIDMap       map[string]int64          `json:"imageUidMap,omitempty"`
		LicenseKey: "LICENSE_KEY",
	}

	blackduck := opssightv1.Blackduck{
		//Hosts               []string `json:"hosts"`
		User: "sysadmin",
		//Port                int      `json:"port"`
		ConcurrentScanLimit: 2,
		TotalScanLimit:      1000,
		//PasswordEnvVar      string   `json:"passwordEnvVar"`
		InitialCount: 0,
		MaxCount:     0,
		//DeleteHubThresholdPercentage int               `json:"deleteHubThresholdPercentage"`
		BlackduckSpec: &blackduckSpec,
	}

	opsSightSpec := opssightv1.OpsSightSpec{
		Namespace: namespace,
		//State:         "string",
		Perceptor:     &perceptor,
		ScannerPod:    &scannerPod,
		Perceiver:     &perceiver,
		Prometheus:    &prometheus,
		EnableSkyfire: false,
		Skyfire:       &skyfire,
		Blackduck:     &blackduck,
		EnableMetrics: true,
		DefaultCPU:    "300m",
		DefaultMem:    "1300Mi",
		LogLevel:      "debug",
		//ConfigMapName: "string",
		SecretName: "blackduck",
	}
	opssight.ObjectMeta = metav1.ObjectMeta{
		Name: namespace,
	}
	opssight.Spec = opsSightSpec
}
